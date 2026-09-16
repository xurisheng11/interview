package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"time"

	"interview-sim/model"
)

// isVoiceMode 语音类模式（语音面试 video / 视频面试 video_call）均需做表达分析
func isVoiceMode(mode string) bool {
	return mode == "video" || mode == "video_call"
}

// ============ Word 报告生成（.docx，OOXML 最小子集） ============
// 不引入第三方依赖：docx 本质是 zip 包内几个 XML，直接手工拼装。
// 视觉设计：主题色横幅封面 → 基本信息表 → 得分大卡 → 模块得分表 → 逐题点评（色条题头+引用回答）→ AI 评价。

const wordDocHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`

const wordDocFooter = `<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134"/></w:sectPr></w:body></w:document>`

const wordContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`

const wordRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`

// 主题色板（与小程序品牌蓝一致）
const (
	wcPrimary = "1F4E79" // 深蓝（横幅/标题）
	wcAccent  = "2E74B5" // 亮蓝（H1/边条）
	wcSoft    = "EEF4FA" // 浅蓝底（题头/隔行）
	wcCard    = "F7FAFD" // 卡片底（引用）
	wcLine    = "D5DFEA" // 表格线
	wcGood    = "1E7D32" // 绿（优势/通过）
	wcBad     = "C0392B" // 红（不足/未通过）
	wcWarn    = "C77700" // 橙（及格）
	wcGray    = "6A6A6A" // 次要文字
)

// wordStyle 文字样式预设
type wordStyle struct {
	size   int // 字号（磅）
	bold   bool
	color  string // RGB，空为默认黑
	indent int    // 左缩进（半角字符数）
	align  string // center/right，空为左对齐
	before int    // 段前距（1/20 磅）
	after  int    // 段后距（1/20 磅）
}

// wordParaOpt 段落级装饰
type wordParaOpt struct {
	fill         string // 段落底纹色
	borderLeft   string // 左侧色条
	borderBottom string // 底部横线
}

var (
	wsCoverTitle = wordStyle{size: 26, bold: true, color: "FFFFFF", align: "center"}
	wsCoverSub   = wordStyle{size: 12, color: "D9E7F5", align: "center"}
	wsH1         = wordStyle{size: 15, bold: true, color: wcPrimary, before: 240, after: 120}
	wsH2         = wordStyle{size: 12, bold: true, color: wcAccent, before: 160, after: 60}
	wsBody       = wordStyle{size: 11, after: 40}
	wsLabel      = wordStyle{size: 11, bold: true, color: wcGray}
	wsGraySmall  = wordStyle{size: 9, color: wcGray}
	wsQuote      = wordStyle{size: 11, color: "404040", after: 40}
	wsGood       = wordStyle{size: 11, color: wcGood, indent: 1, after: 20}
	wsBad        = wordStyle{size: 11, color: wcBad, indent: 1, after: 20}
	wsItem       = wordStyle{size: 11, indent: 1, after: 20}
)

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// wordRunXml 一个文本 run
func wordRunXml(text string, st wordStyle) string {
	var sb strings.Builder
	sb.WriteString(`<w:r><w:rPr><w:rFonts w:eastAsia="微软雅黑"/>`)
	if st.bold {
		sb.WriteString("<w:b/>")
	}
	if st.color != "" {
		sb.WriteString(fmt.Sprintf(`<w:color w:val="%s"/>`, st.color))
	}
	sb.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, st.size*2, st.size*2))
	sb.WriteString("</w:rPr>")
	sb.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, xmlEscape(text)))
	sb.WriteString("</w:r>")
	return sb.String()
}

// wordParaRuns 单段（多个 run 混排），带底纹/边条等装饰
func wordParaRuns(runs []wordRunSpec, st wordStyle, opt wordParaOpt) string {
	var sb strings.Builder
	sb.WriteString("<w:p><w:pPr>")
	// pPr 子元素需按 schema 顺序：pBdr → shd → spacing → ind → jc
	if opt.borderLeft != "" || opt.borderBottom != "" {
		sb.WriteString("<w:pBdr>")
		if opt.borderLeft != "" {
			sb.WriteString(fmt.Sprintf(`<w:left w:val="single" w:sz="28" w:space="6" w:color="%s"/>`, opt.borderLeft))
		}
		if opt.borderBottom != "" {
			sb.WriteString(fmt.Sprintf(`<w:bottom w:val="single" w:sz="6" w:space="6" w:color="%s"/>`, opt.borderBottom))
		}
		sb.WriteString("</w:pBdr>")
	}
	if opt.fill != "" {
		sb.WriteString(fmt.Sprintf(`<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, opt.fill))
	}
	sb.WriteString(fmt.Sprintf(`<w:spacing w:before="%d" w:after="%d" w:line="312" w:lineRule="auto"/>`, st.before, st.after))
	left := st.indent * 240
	if opt.fill != "" && left < 113 {
		left = 113 // 底纹段落加一点内边距更像"卡片"
	}
	if left > 0 {
		sb.WriteString(fmt.Sprintf(`<w:ind w:left="%d"`, left))
		if opt.fill != "" {
			sb.WriteString(` w:right="113"`)
		}
		sb.WriteString("/>")
	}
	if st.align != "" {
		sb.WriteString(fmt.Sprintf(`<w:jc w:val="%s"/>`, st.align))
	}
	sb.WriteString("</w:pPr>")
	for _, r := range runs {
		sb.WriteString(wordRunXml(r.text, r.st))
	}
	sb.WriteString("</w:p>")
	return sb.String()
}

type wordRunSpec struct {
	text string
	st   wordStyle
}

// wordPara 单样式段落（含换行则拆成多段）
func wordPara(text string, st wordStyle, opt wordParaOpt) string {
	var sb strings.Builder
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, line := range lines {
		s := st
		if i == len(lines)-1 && s.after == 0 {
			s.after = 40
		}
		sb.WriteString(wordParaRuns([]wordRunSpec{{text: line, st: s}}, s, opt))
	}
	return sb.String()
}

// wordKV 一行"标签：值"（标签灰色加粗）
func wordKV(label, value string) string {
	return wordParaRuns([]wordRunSpec{
		{text: label + "：", st: wsLabel},
		{text: value, st: wsBody},
	}, wsBody, wordParaOpt{})
}

func wordEmpty(after int) string {
	return fmt.Sprintf(`<w:p><w:pPr><w:spacing w:after="%d"/></w:pPr></w:p>`, after)
}

// wordCellSpec 表格单元格
type wordCellSpec struct {
	text  string
	st    wordStyle
	width int // dxa
	fill  string
}

func wordCellXml(c wordCellSpec) string {
	var sb strings.Builder
	sb.WriteString("<w:tc><w:tcPr>")
	if c.width > 0 {
		sb.WriteString(fmt.Sprintf(`<w:tcW w:w="%d" w:type="dxa"/>`, c.width))
	} else {
		sb.WriteString(`<w:tcW w:w="0" w:type="auto"/>`)
	}
	if c.fill != "" {
		sb.WriteString(fmt.Sprintf(`<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, c.fill))
	}
	sb.WriteString(`<w:vAlign w:val="center"/></w:tcPr>`)
	sb.WriteString(wordPara(c.text, c.st, wordParaOpt{}))
	sb.WriteString("</w:tc>")
	return sb.String()
}

// wordTableXml 渲染带细线边框的表格；调用方负责给表头行上底色
func wordTableXml(rows [][]wordCellSpec) string {
	var sb strings.Builder
	sb.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/>`)
	sb.WriteString(fmt.Sprintf(`<w:tblBorders><w:top w:val="single" w:sz="4" w:color="%s"/><w:left w:val="single" w:sz="4" w:color="%s"/><w:bottom w:val="single" w:sz="4" w:color="%s"/><w:right w:val="single" w:sz="4" w:color="%s"/><w:insideH w:val="single" w:sz="4" w:color="%s"/><w:insideV w:val="single" w:sz="4" w:color="%s"/></w:tblBorders>`, wcLine, wcLine, wcLine, wcLine, wcLine, wcLine))
	sb.WriteString(`<w:tblCellMar><w:left w:w="120" w:type="dxa"/><w:right w:w="120" w:type="dxa"/></w:tblCellMar>`)
	sb.WriteString("</w:tblPr>")
	for _, row := range rows {
		sb.WriteString("<w:tr>")
		for _, cell := range row {
			sb.WriteString(wordCellXml(cell))
		}
		sb.WriteString("</w:tr>")
	}
	sb.WriteString("</w:tbl>")
	return sb.String()
}

func fmtWordTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

func fmtDuration(seconds int) string {
	m, s := seconds/60, seconds%60
	if m > 0 {
		return fmt.Sprintf("%d分%d秒", m, s)
	}
	return fmt.Sprintf("%d秒", s)
}

var wordRoundNames = map[string]string{
	"round1": "一面（基础轮）", "round2": "二面（进阶轮）", "round3": "三面（高管轮）",
	"comprehensive": "综合面试",
}

var wordDifficultyNames = map[string]string{"easy": "简单", "medium": "中等", "hard": "困难"}

var wordPassNames = map[string]string{"pass": "通过（建议进入下一轮）", "pending": "待定", "fail": "未通过"}

// wordScoreColor 分数→主题色
func wordScoreColor(score int) string {
	switch {
	case score >= 90:
		return wcGood
	case score >= 75:
		return wcAccent
	case score >= 60:
		return wcWarn
	default:
		return wcBad
	}
}

func modeName(mode string) string {
	switch mode {
	case "video_call":
		return "视频面试"
	case "video":
		return "语音面试"
	default:
		return "文字面试"
	}
}

// wordH1 带底部分隔线的章节标题
func wordH1(text string) string {
	return wordPara(text, wsH1, wordParaOpt{borderBottom: wcLine})
}

// BuildWordReport 将面试报告渲染为 .docx 文件字节
func BuildWordReport(report *model.InterviewReport) ([]byte, error) {
	var body strings.Builder

	// ===== 封面横幅 =====
	title := "面试评测报告"
	if report.JobTitle != "" {
		title = report.JobTitle + " · 面试评测报告"
	}
	coverSub := report.CompanyName
	if coverSub == "" {
		coverSub = "AI 模拟面试"
	}
	metaLine := []string{}
	if v, ok := wordRoundNames[report.Round]; ok {
		metaLine = append(metaLine, v)
	} else if report.Round != "" {
		metaLine = append(metaLine, report.Round)
	}
	if v, ok := wordDifficultyNames[report.Difficulty]; ok {
		metaLine = append(metaLine, "难度 "+v)
	}
	metaLine = append(metaLine, modeName(report.Mode))

	body.WriteString(wordPara(title, wsCoverTitle, wordParaOpt{fill: wcPrimary}))
	body.WriteString(wordPara(coverSub, wsCoverSub, wordParaOpt{fill: wcPrimary}))
	body.WriteString(wordPara(strings.Join(metaLine, "｜")+"　"+fmtWordTime(report.EndTime), wsCoverSub, wordParaOpt{fill: wcPrimary}))
	// 横幅底部亮蓝细线收尾
	body.WriteString(wordParaRuns(nil, wordStyle{after: 200}, wordParaOpt{borderBottom: wcAccent}))
	body.WriteString(wordEmpty(120))

	// ===== 总体成绩大卡 =====
	body.WriteString(wordH1("一、总体成绩"))
	passText := wordPassNames[report.PassStatus]
	if passText == "" {
		passText = report.PassStatus
	}
	scoreColor := wordScoreColor(report.TotalScore)
	cardRuns := []wordRunSpec{
		{text: fmt.Sprintf("%d", report.TotalScore), st: wordStyle{size: 40, bold: true, color: scoreColor}},
		{text: " 分", st: wordStyle{size: 14, bold: true, color: scoreColor}},
		{text: "　" + report.Grade, st: wordStyle{size: 16, bold: true, color: scoreColor}},
	}
	body.WriteString(wordParaRuns(cardRuns, wordStyle{align: "center", before: 120, after: 120}, wordParaOpt{fill: wcSoft}))
	passStyle := wordStyle{size: 12, bold: true, color: wcGood, align: "center", after: 120}
	switch report.PassStatus {
	case "fail":
		passStyle.color = wcBad
	case "pending":
		passStyle.color = wcWarn
	}
	if report.PassReason != "" {
		passText += "（" + report.PassReason + "）"
	}
	body.WriteString(wordParaRuns([]wordRunSpec{{text: "判定结论：" + passText, st: passStyle}}, passStyle, wordParaOpt{}))

	// 基本信息表（两列 KV）
	kvSt := wordStyle{size: 10, color: "404040"}
	infoRows := [][]wordCellSpec{}
	addInfo := func(k, v string) {
		infoRows = append(infoRows, []wordCellSpec{
			{text: k, st: wordStyle{size: 10, bold: true, color: wcGray}, width: 2200, fill: wcSoft},
			{text: v, st: kvSt, width: 6600},
		})
	}
	if report.CompanyName != "" {
		addInfo("目标公司", report.CompanyName)
	}
	if report.JobTitle != "" {
		addInfo("应聘职位", report.JobTitle)
	}
	addInfo("面试模式", modeName(report.Mode))
	addInfo("开始时间", fmtWordTime(report.StartTime))
	addInfo("结束时间", fmtWordTime(report.EndTime))
	addInfo("总用时", fmtDuration(report.TotalSeconds))
	addInfo("答题情况", fmt.Sprintf("共 %d 题，作答 %d 题，跳过 %d 题", report.TotalCount, report.AnsweredCount, report.SkippedCount))
	if isVoiceMode(report.Mode) {
		if report.AvgExpressionScore > 0 {
			addInfo("平均表达得分", fmt.Sprintf("%d 分", report.AvgExpressionScore))
		}
		if report.AvgSpeechRate > 0 {
			addInfo("平均语速", fmt.Sprintf("%.0f 字/分钟", report.AvgSpeechRate))
		}
		if report.AvgThinkDuration > 0 {
			addInfo("平均思考时长", fmtDuration(report.AvgThinkDuration))
		}
		if report.FaceSummary != nil {
			fs := report.FaceSummary
			addInfo("面部表情分析", fmt.Sprintf("有效抓帧 %d 帧；主要情绪：%s；紧张信号帧占比 %.0f%%（%s）；微笑帧占比 %.0f%%",
				fs.Samples, fs.DominantName, fs.TenseRatio*100, fs.NervousLevel, fs.SmileRatio*100))
		}
	}
	body.WriteString(wordTableXml(infoRows))
	body.WriteString(wordEmpty(160))

	// ===== 知识点模块得分表 =====
	if len(report.ModuleScores) > 0 {
		body.WriteString(wordPara("知识点模块得分", wsH2, wordParaOpt{}))
		headerSt := wordStyle{size: 10, bold: true, color: "FFFFFF"}
		modRows := [][]wordCellSpec{{
			{text: "知识模块", st: headerSt, width: 3400, fill: wcAccent},
			{text: "题数", st: headerSt, width: 1200, fill: wcAccent},
			{text: "平均分", st: headerSt, width: 1600, fill: wcAccent},
			{text: "掌握程度", st: headerSt, width: 2600, fill: wcAccent},
		}}
		for i, m := range report.ModuleScores {
			zebra := ""
			if i%2 == 1 {
				zebra = wcCard
			}
			modRows = append(modRows, []wordCellSpec{
				{text: m.Module, st: kvSt, width: 3400, fill: zebra},
				{text: fmt.Sprintf("%d", m.Count), st: wordStyle{size: 10, align: "center", color: "404040"}, width: 1200, fill: zebra},
				{text: fmt.Sprintf("%.0f", m.AvgScore), st: wordStyle{size: 10, bold: true, align: "center", color: wordScoreColor(int(m.AvgScore + 0.5))}, width: 1600, fill: zebra},
				{text: m.Level, st: wordStyle{size: 10, align: "center", color: "404040"}, width: 2600, fill: zebra},
			})
		}
		body.WriteString(wordTableXml(modRows))
		body.WriteString(wordEmpty(160))
	}

	// ===== 逐题点评 =====
	body.WriteString(wordH1("二、逐题点评"))
	for i, q := range report.Questions {
		// 题头：亮蓝左条 + 浅蓝底
		headRuns := []wordRunSpec{{text: fmt.Sprintf(" Q%d  ", i+1), st: wordStyle{size: 12, bold: true, color: wcPrimary}}}
		if q.Skipped || strings.TrimSpace(q.UserAnswer) == "" {
			headRuns = append(headRuns, wordRunSpec{text: "跳过 / 未作答", st: wordStyle{size: 11, bold: true, color: wcGray}})
		} else {
			headRuns = append(headRuns, wordRunSpec{text: fmt.Sprintf("%d 分", q.Score), st: wordStyle{size: 12, bold: true, color: wordScoreColor(q.Score)}})
		}
		if q.Difficulty != "" {
			diff := wordDifficultyNames[q.Difficulty]
			if diff == "" {
				diff = q.Difficulty
			}
			headRuns = append(headRuns, wordRunSpec{text: "　难度 " + diff, st: wsGraySmall})
		}
		body.WriteString(wordParaRuns(headRuns, wsH2, wordParaOpt{fill: wcSoft, borderLeft: wcAccent}))
		body.WriteString(wordPara("题目："+q.Content, wordStyle{size: 11, bold: true, before: 60, after: 60}, wordParaOpt{}))
		if strings.TrimSpace(q.UserAnswer) != "" {
			body.WriteString(wordPara("我的回答　"+q.UserAnswer, wsQuote, wordParaOpt{fill: wcCard, borderLeft: wcLine}))
		}
		for _, p := range q.Pros {
			body.WriteString(wordPara("✔ "+p, wsGood, wordParaOpt{}))
		}
		for _, c := range q.Cons {
			body.WriteString(wordPara("✘ "+c, wsBad, wordParaOpt{}))
		}
		if q.ExpressionFeedback != "" {
			body.WriteString(wordPara("🎙 表达点评　"+q.ExpressionFeedback, wsQuote, wordParaOpt{borderLeft: wcLine}))
		}
		if strings.TrimSpace(q.ReferenceAnswer) != "" {
			body.WriteString(wordPara("💡 参考答案　"+q.ReferenceAnswer, wordStyle{size: 9, color: wcGray, indent: 1, after: 60}, wordParaOpt{}))
		}
		body.WriteString(wordEmpty(120))
	}

	// ===== AI 综合评价 =====
	if report.AISummary != nil {
		body.WriteString(wordH1("三、AI 综合评价"))
		if len(report.AISummary.Strengths) > 0 {
			body.WriteString(wordPara("✦ 优势亮点", wordStyle{size: 12, bold: true, color: wcGood, before: 120, after: 60}, wordParaOpt{}))
			for _, s := range report.AISummary.Strengths {
				body.WriteString(wordPara("· "+s, wsGood, wordParaOpt{}))
			}
		}
		if len(report.AISummary.Weaknesses) > 0 {
			body.WriteString(wordPara("✦ 待改进项", wordStyle{size: 12, bold: true, color: wcBad, before: 120, after: 60}, wordParaOpt{}))
			for _, w := range report.AISummary.Weaknesses {
				line := "· " + w.Point
				if w.Suggestion != "" {
					line += "　→ 建议：" + w.Suggestion
				}
				if w.Resource != "" {
					line += "（学习资料：" + w.Resource + "）"
				}
				body.WriteString(wordPara(line, wsBad, wordParaOpt{}))
			}
		}
		if report.AISummary.Roadmap != "" {
			body.WriteString(wordPara("✦ 提升路线", wordStyle{size: 12, bold: true, color: wcAccent, before: 120, after: 60}, wordParaOpt{}))
			body.WriteString(wordPara(report.AISummary.Roadmap, wsItem, wordParaOpt{borderLeft: wcAccent, fill: wcCard}))
		}
	}
	if report.ExpressionSummary != "" {
		if report.AISummary == nil {
			body.WriteString(wordH1("三、表达能力综合评价"))
		} else {
			body.WriteString(wordPara("✦ 表达能力评价", wordStyle{size: 12, bold: true, color: wcAccent, before: 160, after: 60}, wordParaOpt{}))
		}
		body.WriteString(wordPara(report.ExpressionSummary, wsQuote, wordParaOpt{fill: wcCard, borderLeft: wcLine}))
	}

	// ===== 结束落款 =====
	body.WriteString(wordEmpty(200))
	body.WriteString(wordPara("—— 本报告由「AI 模拟面试」小程序生成 ——", wordStyle{size: 9, color: wcGray, align: "center"}, wordParaOpt{}))

	document := wordDocHeader + body.String() + wordDocFooter

	// ===== 打包 zip 成 docx =====
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	parts := map[string]string{
		"[Content_Types].xml": wordContentTypes,
		"_rels/.rels":         wordRels,
		"word/document.xml":   document,
	}
	for name, content := range parts {
		f, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
