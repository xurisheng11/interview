package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"interview-sim/config"
	"interview-sim/model"
)

// ---- DeepSeek API 结构 ----

type dsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type dsRequest struct {
	Model       string      `json:"model"`
	Messages    []dsMessage `json:"messages"`
	Temperature float64     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens"`
}

type dsChoice struct {
	Message dsMessage `json:"message"`
}

type dsResponse struct {
	Choices []dsChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat 调用 DeepSeek，失败自动重试 2 次
func Chat(prompt string) (string, error) {
	reqBody := dsRequest{
		Model:       "deepseek-chat",
		Messages:    []dsMessage{{Role: "user", Content: prompt}},
		Temperature: 0.7,
		MaxTokens:   8192,
	}
	body, _ := json.Marshal(reqBody)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		result, err := doChat(body)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("DeepSeek 调用失败（已重试）: %w", lastErr)
}

func doChat(body []byte) (string, error) {
	url := config.Cfg.DeepSeekBaseURL + "/v1/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Cfg.DeepSeekAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var dsResp dsResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %s", string(respBody))
	}
	if dsResp.Error != nil {
		return "", errors.New(dsResp.Error.Message)
	}
	if len(dsResp.Choices) == 0 {
		return "", errors.New("DeepSeek 返回空结果")
	}
	return strings.TrimSpace(dsResp.Choices[0].Message.Content), nil
}

// ---- 题目生成 ----
// resumeContext 为可选参数，传入时会附加到 prompt 中，使题目与候选人简历强关联
func GenerateQuestions(cfg *model.InterviewConfig, resumeContext ...string) ([]model.Question, error) {
	roundReq := getRoundRequirement(cfg.Round)
	count := config.Cfg.InterviewQuestionCount

	// 获取面试形式描述
	interviewTypeDesc := getInterviewTypeDesc(cfg.InterviewTypes)

	resumeBlock := ""
	if len(resumeContext) > 0 && resumeContext[0] != "" {
		resumeBlock = resumeContext[0]
	}

	prompt := fmt.Sprintf(`你是一位资深技术面试官，请为以下面试场景生成 %d 道面试题目。

面试信息：
- 目标岗位：%s
- 面试难度：%s
- 工作经验：%s
- 面试轮次：%s
- 重点方向：%s
- 补充说明：%s
- 面试形式：%s
%s

题目构成要求：
%s

请严格按照以下 JSON 数组格式返回，不要包含任何其他文字、代码块标记：
[{"index":0,"content":"题目内容","tags":["知识点"],"difficulty":"easy","estimatedMinutes":3,"type":"basic"}]

difficulty 取值：easy/medium/hard
type 取值：basic/algorithm/design/hr`,
		count,
		cfg.JobTitle,
		cfg.Difficulty,
		cfg.Experience,
		getRoundName(cfg.Round),
		strings.Join(cfg.FocusAreas, "、"),
		cfg.Remark,
		interviewTypeDesc,
		resumeBlock,
		roundReq,
	)

	raw, err := Chat(prompt)
	if err != nil {
		return nil, err
	}

	// 清理可能的代码块标记
	raw = cleanJSON(raw)

	var questions []model.Question
	if err := json.Unmarshal([]byte(raw), &questions); err != nil {
		return nil, fmt.Errorf("解析题目失败: %w, raw: %s", err, raw[:minInt(200, len(raw))])
	}
	return questions, nil
}

// ---- 答案点评 ----

type ReviewResult struct {
	Score              int      `json:"score"`
	Pros               []string `json:"pros"`
	Cons               []string `json:"cons"`
	ReferenceAnswer    string   `json:"referenceAnswer"`
	ExpressionScore    int      `json:"expressionScore,omitempty"`
	ExpressionFeedback string   `json:"expressionFeedback,omitempty"`
	DetectedVerbalTics []string `json:"detectedVerbalTics,omitempty"` // 识别到的口头禅
}

// ReviewAnswer 对答案进行 AI 点评，语音/视频模式时附加非语言指标与表情分析
func ReviewAnswer(question *model.Question, answer string, cfg *model.InterviewConfig, metrics *model.NonVerbalMetrics, face *model.FaceEmotionInfo) (*ReviewResult, error) {
	prompt := fmt.Sprintf(`【面试评分任务】严格按以下步骤评分：

【第一步：判断答案是否切题】（必须先做）
对照题目要求，判断候选人答案是否直接回答了问题：
- 题目要求：%s
- 候选人答案：%s

【评分标准】（★关键规则★）
★ 如果答案与题目要求不相关、不切题、答非所问，必须给 0-19 分！
★ "不切题"包括：仅回复数字/字母/符号、完全无关的内容、无法理解的乱码、明显敷衍的套话
★ "基本正确但有缺陷"才给 20-59 分（有一定相关性，但有错误或遗漏）
★ "基本正确"才给 60-89 分（切题且大部分正确）
★ "优秀"才给 90-100 分（切题、准确、完整、有深度）

分数量级参考（严禁跨档给分）：
- 0-19分：完全不切题 / 答非所问 / 乱码 / 无意义内容
- 20-39分：略微相关但严重偏离 / 几乎没有正确内容
- 40-59分：部分相关 / 有较多错误或遗漏
- 60-79分：切题且基本正确 / 覆盖主要要点
- 80-89分：切题且正确 / 要点较完整 / 表达较清晰
- 90-100分：切题且完全正确 / 要点全覆盖 / 有深度见解

【第二步：给出点评】（选 1-2 条）
- 优点（仅当分数>=60时填写）
- 缺点/不足

【第三步：写出参考答案】（必须包含完整要点）

请严格按照以下 JSON 格式返回，不要包含任何其他文字：
{"score":85,"pros":["优点1","优点2"],"cons":["不足1","不足2"],"referenceAnswer":"参考答案"}`,
		question.Content,
		answer,
	)

	// 视频模式：附加非语言指标上下文
	if metrics != nil {
		verbalTicInfo := ""
		if len(metrics.VerbalTics) > 0 {
			verbalTicInfo = fmt.Sprintf("\n识别到的口头禅：%s", strings.Join(metrics.VerbalTics, "、"))
		}
		if metrics.StutterCount > 0 {
			verbalTicInfo += fmt.Sprintf("\n口吃/不流畅次数：%d（转写文本中检测到重复字/词，如“我我”“然后然后”）", metrics.StutterCount)
		}
		// 视频面试：附加表情抓帧观察
		faceInfo := ""
		if face != nil {
			smileText := "未微笑"
			if face.Smile == 1 {
				smileText = "微笑"
			}
			faceInfo = fmt.Sprintf("\n[表情观察] 视频面试抓帧识别到候选人表情：情绪“%s”（置信度%.0f%%），%s", face.Name, face.Probability*100, smileText)
		}

		prompt += fmt.Sprintf(`

[语音表达数据]（候选人采用语音答题，以下为语音转写后统计的客观指标）%s
语速：%.0f字/分钟（面试推荐区间120-160）；作答时长：%d秒；思考时长：%d秒；停顿次数：%d%s

请基于上述指标 + 转写文本，额外输出口头表达专项评价，在 JSON 中新增以下字段：
- expressionScore（0-100，综合口头表达流畅度、逻辑性、自信度的表达得分）
- expressionFeedback（分点给出具体的口头表达改进建议，必须依次覆盖以下维度，每个维度 1 句：
  1) 语速：偏慢/适中/偏快，给出调整到120-160字每分钟的建议；
  2) 自信度：从转写文本的犹豫词、填充词、断句、模棱两可表达（如“可能/好像/大概/应该”）推断是否自信，给出增强自信的表达建议；
  3) 口头禅：若检测到高频口头禅（如“然后/这个/那个/嗯”），指出并建议用停顿或逻辑连接词替代；
  4) 普通话标准度：无法直接听音，请从转写质量推断（若出现大量同音错别字、乱码、词不成句，提示发音可能不够清晰标准；若转写流畅则肯定其清晰度）；
  5) 条理性：回答是否总分总/要点分明，建议用“第一/第二/第三”等结构化表达；
  6) 流畅度：若口吃/不流畅次数大于0，指出字词重复现象并建议放慢开头语速、用短暂停顿替代重复；若无则肯定流畅度%s
- detectedVerbalTics（从转写文本中检测到的口头禅列表，如["然后","这个","嗯"]，无则返回空数组）`,
			faceInfo,
			metrics.SpeechRate, metrics.Duration, metrics.ThinkDuration, metrics.PauseCount, verbalTicInfo,
			map[bool]string{true: "；\n  7) 表情管理与紧张度：结合[表情观察]判断是否紧张（害怕/悲伤/生气/惊讶等负面情绪提示紧张，自然/高兴提示放松），若紧张给出改善建议（答题前深呼吸、开场有意识微笑、目光看摄像头而非屏幕）；若放松则肯定其镜头前表现力", false: ""}[face != nil],
		)
	}

	raw, err := Chat(prompt)
	if err != nil {
		return nil, err
	}
	raw = cleanJSON(raw)

	var result ReviewResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("解析点评失败: %w", err)
	}
	// 确保分数在 0-100
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 100 {
		result.Score = 100
	}
	if result.ExpressionScore < 0 {
		result.ExpressionScore = 0
	}
	if result.ExpressionScore > 100 {
		result.ExpressionScore = 100
	}
	return &result, nil
}

// ---- 报告综合评价 ----

type AISummary struct {
	Strengths  []string       `json:"strengths"`
	Weaknesses []WeaknessItem `json:"weaknesses"`
	Roadmap    string         `json:"roadmap"`
}

type WeaknessItem struct {
	Point      string `json:"point"`
	Suggestion string `json:"suggestion"`
	Resource   string `json:"resource"`
}

func GenerateReportSummary(jobTitle, round string, totalScore int, scoresSummary string) (*AISummary, error) {
	prompt := fmt.Sprintf(`你是一位资深面试顾问，请根据以下面试数据生成综合评价。

岗位：%s，轮次：%s，综合得分：%d
各题得分摘要：%s

请严格按照以下 JSON 格式返回，不要包含任何其他文字、代码块标记：
{"strengths":["优势1","优势2","优势3"],"weaknesses":[{"point":"不足1","suggestion":"改进建议","resource":"https://..."},{"point":"不足2","suggestion":"改进建议","resource":"https://..."},{"point":"不足3","suggestion":"改进建议","resource":"https://..."}],"roadmap":"备考路线图文字描述"}`,
		jobTitle, getRoundName(round), totalScore, scoresSummary,
	)

	raw, err := Chat(prompt)
	if err != nil {
		return nil, err
	}
	raw = cleanJSON(raw)

	var summary AISummary
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return nil, fmt.Errorf("解析综合评价失败: %w", err)
	}
	return &summary, nil
}

// GenerateVideoExpressionSummary 生成视频面试表达能力综合评价
func GenerateVideoExpressionSummary(jobTitle string, answers []*model.AnswerRecord) (string, error) {
	avgRate := calcAvgSpeechRateFromAnswers(answers)
	avgScore := calcAvgExpressionScoreFromAnswers(answers)
	avgThinkDuration := calcAvgThinkDurationFromAnswers(answers)
	ticReport := model.CalcVerbalTicReportFromAnswers(answers)

	ticInfo := ""
	if ticReport != nil && len(ticReport.DetectedTics) > 0 {
		ticInfo = fmt.Sprintf("\n口头禅检测：检测到 %d 次口头禅，包括：%s", ticReport.TicFrequency, strings.Join(ticReport.DetectedTics, "、"))
	}

	// 拼接平均思考时间信息
	thinkInfo := ""
	if avgThinkDuration > 0 {
		thinkInfo = fmt.Sprintf("\n平均思考时间：%d秒（从看到题目到开始作答的平均间隔）", avgThinkDuration)
	}

	// 口吃/不流畅统计
	stutterTotal := 0
	for _, a := range answers {
		if a != nil && a.NonVerbalMetrics != nil {
			stutterTotal += a.NonVerbalMetrics.StutterCount
		}
	}
	stutterInfo := ""
	if stutterTotal > 0 {
		stutterInfo = fmt.Sprintf("\n口吃/不流畅：全场累计检测到 %d 次字词重复", stutterTotal)
	}

	// 视频面试：面部情绪抓帧统计
	faceInfo := ""
	faceSamples, faceTense, faceSmile := 0, 0, 0
	for _, a := range answers {
		if a == nil || a.FaceEmotion == nil {
			continue
		}
		faceSamples++
		if a.FaceEmotion.Type >= 2 { // 非自然/高兴 → 紧张信号
			faceTense++
		}
		if a.FaceEmotion.Smile == 1 {
			faceSmile++
		}
	}
	if faceSamples > 0 {
		faceInfo = fmt.Sprintf("\n面部表情分析：%d 帧抓帧中，%d 帧呈紧张信号情绪（惊讶/生气/悲伤/厌恶/害怕），%d 帧微笑", faceSamples, faceTense, faceSmile)
	}

	prompt := fmt.Sprintf(`你是一位资深面试顾问，请根据以下视频面试数据给出整体口头表达能力评价。

岗位：%s
完成题数：%d
平均语速：%.0f字/分钟（推荐120-150）
平均表达得分：%d分%s%s%s%s

请严格按照以下 JSON 格式返回，不要包含任何其他文字：
{"summary":"整体口头表达能力评价（100字以内）","suggestions":["改进建议1","改进建议2","改进建议3"]}
要求：若有口吃或面部紧张数据，suggestions 必须包含对应的改善建议（如放慢起语、停顿替代重复、深呼吸、开场微笑练习等）。`,
		jobTitle, len(answers), avgRate, avgScore, ticInfo, thinkInfo, stutterInfo, faceInfo,
	)

	raw, err := Chat(prompt)
	if err != nil {
		return "", err
	}
	raw = cleanJSON(raw)

	var tmp struct {
		Summary     string   `json:"summary"`
		Suggestions []string `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return "", fmt.Errorf("解析表达评价失败: %w", err)
	}
	// 将建议拼接到 summary 中返回
	if len(tmp.Suggestions) > 0 {
		return tmp.Summary + "\n改进建议：" + strings.Join(tmp.Suggestions, "；"), nil
	}
	return tmp.Summary, nil
}

func calcAvgSpeechRateFromAnswers(answers []*model.AnswerRecord) float64 {
	total := 0.0
	count := 0
	for _, a := range answers {
		if a != nil && !a.Skipped && a.NonVerbalMetrics != nil && a.NonVerbalMetrics.SpeechRate > 0 {
			total += a.NonVerbalMetrics.SpeechRate
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func calcAvgExpressionScoreFromAnswers(answers []*model.AnswerRecord) int {
	total, count := 0, 0
	for _, a := range answers {
		if a != nil && !a.Skipped && a.ExpressionScore > 0 {
			total += a.ExpressionScore
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

// calcAvgThinkDurationFromAnswers 从 AnswerRecord 列表计算平均思考时长（秒）
func calcAvgThinkDurationFromAnswers(answers []*model.AnswerRecord) int {
	total, count := 0, 0
	for _, a := range answers {
		if a != nil && !a.Skipped && a.NonVerbalMetrics != nil && a.NonVerbalMetrics.ThinkDuration > 0 {
			total += a.NonVerbalMetrics.ThinkDuration
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

// ---- AI 知识文章生成 ----

// GenerateArticle 调用 DeepSeek 生成知识文章，返回填充好的 model.Article（articleId 留空，由调用方生成）
// jobCategoryCNMap 将英文分类映射为中文岗位描述
var jobCategoryCNMap = map[string]string{
	"backend":    "后端开发",
	"frontend":   "前端开发",
	"bigdata":    "大数据开发",
	"ai":         "AI/算法工程师",
	"accounting": "会计/财务",
	"general":    "通用",
	"all":        "通用技术",
}

func getJobCategoryCN(jobCategory string) string {
	if cn, ok := jobCategoryCNMap[jobCategory]; ok {
		return cn
	}
	return jobCategory
}

func GenerateArticle(topic, jobCategory string) (*model.Article, error) {
	jobCN := getJobCategoryCN(jobCategory)
	prompt := fmt.Sprintf(`你是一位资深技术面试辅导专家。请围绕【%s】这个知识点，为【%s】岗位求职者撰写一篇高质量的面试备考文章。

要求：
1. 内容必须紧密围绕用户输入的【%s】这个知识点展开，不要偏离
2. 文章结构必须包含以下部分：
   - 核心概念定义（用简洁的话解释清楚）
   - 常见面试考点（高频问题，每个问题给出参考答案要点）
   - 深度追问方向（面试官可能追问的延伸问题）
   - 易错点/误区提醒
   - 实战代码示例或图解说明（如适用）
3. 难度：覆盖从基础到进阶的完整学习路径
4. 正文使用 Markdown 格式，包含标题、代码块、表格、列表等丰富排版
5. 文章长度 1000-1500 字，确保内容充实有深度
6. 标签请从知识点中提取3-5个关键词
7. 标题要直接体现知识点，让人一眼看出文章主题

请严格按以下 JSON 格式返回，只返回 JSON，不要其他内容：
{"title":"文章标题（直接体现知识点）","content":"Markdown格式正文","tags":["标签1","标签2","标签3"]}`,
		topic, jobCN, topic,
	)

	raw, err := Chat(prompt)
	if err != nil {
		return nil, err
	}
	raw = cleanJSON(raw)

	var tmp struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return nil, fmt.Errorf("解析文章失败: %w", err)
	}

	article := &model.Article{
		Title:       tmp.Title,
		Content:     tmp.Content,
		Tags:        tmp.Tags,
		JobCategory: jobCategory,
		AuthorID:    "ai",
		CreatedAt:   time.Now().Unix(),
	}
	return article, nil
}

// ---- 辅助函数 ----

func getRoundName(round string) string {
	switch round {
	case "round1":
		return "一面（基础）"
	case "round2":
		return "二面（技术深度）"
	case "round3":
		return "三面（综合/HR）"
	case "comprehensive":
		return "综合面试"
	default:
		return round
	}
}

func getRoundRequirement(round string) string {
	switch round {
	case "round1":
		return "一面：必含自我介绍(1题) + 基础知识题(60%) + 简单算法/逻辑题(20%) + 项目经历简述(20%)"
	case "round2":
		return "二面：技术深度题(50%) + 系统设计题(30%) + 项目经验追问(20%)"
	case "round3":
		return "三面：综合能力题(40%) + 职业规划(30%) + HR类问题(30%)"
	case "comprehensive":
		return "综合面试：必含自我介绍(1题) + 基础知识题(25%) + 技术深度题(25%) + 项目经验题(25%) + 综合/HR题(25%)"
	default:
		return "均衡分配各类题目"
	}
}

func getInterviewTypeDesc(types []string) string {
	if len(types) == 0 {
		return "半结构化面试（最常见）"
	}
	var descs []string
	for _, t := range types {
		switch t {
		case "structured":
			descs = append(descs, "结构化面试")
		case "semi-structured":
			descs = append(descs, "半结构化面试")
		case "random":
			descs = append(descs, "随机问答")
		}
	}
	return strings.Join(descs, "、")
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	// 去除 ```json ... ``` 包裹
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) > 2 {
			s = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	return strings.TrimSpace(s)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
