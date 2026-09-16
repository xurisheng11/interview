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

const wordDocHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`

const wordDocFooter = `<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr></w:body></w:document>`

const wordContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`

const wordRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`

// wordStyle 段落样式预设
type wordStyle struct {
	size   int // 字号（磅）
	bold   bool
	color  string // RGB，空为默认黑
	indent int    // 左缩进（半角字符数）
}

var (
	wsTitle    = wordStyle{size: 22, bold: true, color: "1F4E79"}
	wsH1       = wordStyle{size: 16, bold: true, color: "1F4E79"}
	wsH2       = wordStyle{size: 13, bold: true, color: "2E74B5"}
	wsBody     = wordStyle{size: 11}
	wsLabel    = wordStyle{size: 11, bold: true, color: "595959"}
	wsListItem = wordStyle{size: 11, indent: 2}
	wsGood     = wordStyle{size: 11, color: "1E7D32", indent: 2}
	wsBad      = wordStyle{size: 11, color: "C0392B", indent: 2}
)

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// wordPara 一段文字（含换行则拆成多段）
func wordPara(text string, st wordStyle) string {
	var sb strings.Builder
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		sb.WriteString("<w:p><w:pPr>")
		if st.indent > 0 {
			sb.WriteString(fmt.Sprintf(`<w:ind w:left="%d"/>`, st.indent*240))
		}
		sb.WriteString("</w:pPr><w:r><w:rPr>")
		sb.WriteString(fmt.Sprintf(`<w:rFonts w:eastAsia="微软雅黑"/>`))
		if st.bold {
			sb.WriteString("<w:b/>")
		}
		if st.color != "" {
			sb.WriteString(fmt.Sprintf(`<w:color w:val="%s"/>`, st.color))
		}
		sb.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, st.size*2, st.size*2))
		sb.WriteString("</w:rPr>")
		sb.WriteString(fmt.Sprintf("<w:t xml:space=\"preserve\">%s</w:t>", xmlEscape(line)))
		sb.WriteString("</w:r></w:p>")
	}
	return sb.String()
}

// wordKV 一行"标签：值"
func wordKV(label, value string) string {
	return `<w:p><w:pPr></w:pPr>` +
		`<w:r><w:rPr><w:rFonts w:eastAsia="微软雅黑"/><w:b/><w:color w:val="595959"/><w:sz w:val="22"/></w:rPr><w:t xml:space="preserve">` + xmlEscape(label+"：") + `</w:t></w:r>` +
		`<w:r><w:rPr><w:rFonts w:eastAsia="微软雅黑"/><w:sz w:val="22"/></w:rPr><w:t xml:space="preserve">` + xmlEscape(value) + `</w:t></w:r></w:p>`
}

func wordEmpty() string { return `<w:p/>` }

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

// BuildWordReport 将面试报告渲染为 .docx 文件字节
func BuildWordReport(report *model.InterviewReport) ([]byte, error) {
	var body strings.Builder

	// ===== 标题与基本信息 =====
	title := "面试评测报告"
	if report.JobTitle != "" {
		title += " · " + report.JobTitle
	}
	body.WriteString(wordPara(title, wsTitle))
	if report.CompanyName != "" {
		body.WriteString(wordPara("目标公司："+report.CompanyName, wsBody))
	}
	meta := []string{}
	if v, ok := wordRoundNames[report.Round]; ok {
		meta = append(meta, v)
	} else if report.Round != "" {
		meta = append(meta, report.Round)
	}
	if v, ok := wordDifficultyNames[report.Difficulty]; ok {
		meta = append(meta, "难度 "+v)
	}
	if report.Mode == "video_call" {
		meta = append(meta, "视频面试")
	} else if report.Mode == "video" {
		meta = append(meta, "语音面试")
	} else {
		meta = append(meta, "文字面试")
	}
	body.WriteString(wordPara(strings.Join(meta, "｜"), wsBody))
	body.WriteString(wordKV("开始时间", fmtWordTime(report.StartTime)))
	body.WriteString(wordKV("结束时间", fmtWordTime(report.EndTime)))
	body.WriteString(wordKV("总用时", fmtDuration(report.TotalSeconds)))
	body.WriteString(wordEmpty())

	// ===== 总体成绩 =====
	body.WriteString(wordPara("一、总体成绩", wsH1))
	body.WriteString(wordKV("综合得分", fmt.Sprintf("%d 分（%s）", report.TotalScore, report.Grade)))
	passText := wordPassNames[report.PassStatus]
	if passText == "" {
		passText = report.PassStatus
	}
	if report.PassReason != "" {
		passText += "（" + report.PassReason + "）"
	}
	body.WriteString(wordKV("通过判定", passText))
	body.WriteString(wordKV("答题情况", fmt.Sprintf("共 %d 题，作答 %d 题，跳过 %d 题",
		report.TotalCount, report.AnsweredCount, report.SkippedCount)))
	if isVoiceMode(report.Mode) {
		if report.AvgExpressionScore > 0 {
			body.WriteString(wordKV("平均表达得分", fmt.Sprintf("%d 分", report.AvgExpressionScore)))
		}
		if report.AvgSpeechRate > 0 {
			body.WriteString(wordKV("平均语速", fmt.Sprintf("%.0f 字/分钟", report.AvgSpeechRate)))
		}
		if report.AvgThinkDuration > 0 {
			body.WriteString(wordKV("平均思考时长", fmtDuration(report.AvgThinkDuration)))
		}
		if report.FaceSummary != nil {
			fs := report.FaceSummary
			body.WriteString(wordKV("面部表情分析", fmt.Sprintf("有效抓帧 %d 帧；主要情绪：%s；紧张信号帧占比 %.0f%%（%s）；微笑帧占比 %.0f%%",
				fs.Samples, fs.DominantName, fs.TenseRatio*100, fs.NervousLevel, fs.SmileRatio*100)))
		}
	}
	if len(report.ModuleScores) > 0 {
		body.WriteString(wordPara("知识点模块得分：", wsLabel))
		for _, m := range report.ModuleScores {
			body.WriteString(wordPara(fmt.Sprintf("· %s：%.0f 分（%s，共 %d 题）", m.Module, m.AvgScore, m.Level, m.Count), wsListItem))
		}
	}
	body.WriteString(wordEmpty())

	// ===== 逐题点评 =====
	body.WriteString(wordPara("二、逐题点评", wsH1))
	for i, q := range report.Questions {
		head := fmt.Sprintf("第 %d 题", i+1)
		if q.Skipped || strings.TrimSpace(q.UserAnswer) == "" {
			head += "（跳过 / 未作答）"
		} else {
			head += fmt.Sprintf("（得分 %d）", q.Score)
		}
		body.WriteString(wordPara(head, wsH2))
		body.WriteString(wordKV("题目", q.Content))
		if strings.TrimSpace(q.UserAnswer) != "" {
			body.WriteString(wordKV("我的回答", q.UserAnswer))
		}
		for _, p := range q.Pros {
			body.WriteString(wordPara("✔ "+p, wsGood))
		}
		for _, c := range q.Cons {
			body.WriteString(wordPara("✘ "+c, wsBad))
		}
		if strings.TrimSpace(q.ReferenceAnswer) != "" {
			body.WriteString(wordKV("参考答案", q.ReferenceAnswer))
		}
		if q.ExpressionFeedback != "" {
			body.WriteString(wordKV("表达点评", q.ExpressionFeedback))
		}
		body.WriteString(wordEmpty())
	}

	// ===== AI 综合评价 =====
	if report.AISummary != nil {
		body.WriteString(wordPara("三、AI 综合评价", wsH1))
		if len(report.AISummary.Strengths) > 0 {
			body.WriteString(wordPara("优势亮点：", wsLabel))
			for _, s := range report.AISummary.Strengths {
				body.WriteString(wordPara("· "+s, wsGood))
			}
		}
		if len(report.AISummary.Weaknesses) > 0 {
			body.WriteString(wordPara("待改进项：", wsLabel))
			for _, w := range report.AISummary.Weaknesses {
				line := "· " + w.Point
				if w.Suggestion != "" {
					line += " —— 建议：" + w.Suggestion
				}
				if w.Resource != "" {
					line += "（学习资料：" + w.Resource + "）"
				}
				body.WriteString(wordPara(line, wsBad))
			}
		}
		if report.AISummary.Roadmap != "" {
			body.WriteString(wordKV("提升路线", report.AISummary.Roadmap))
		}
	} else if report.ExpressionSummary != "" {
		body.WriteString(wordPara("三、表达能力综合评价", wsH1))
	}
	if report.ExpressionSummary != "" {
		if report.AISummary == nil {
			body.WriteString(wordPara(report.ExpressionSummary, wsBody))
		} else {
			body.WriteString(wordEmpty())
			body.WriteString(wordKV("表达能力评价", report.ExpressionSummary))
		}
	}

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
