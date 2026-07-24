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

func GenerateQuestions(cfg *model.InterviewConfig) ([]model.Question, error) {
	roundReq := getRoundRequirement(cfg.Round)
	count := config.Cfg.InterviewQuestionCount

	prompt := fmt.Sprintf(`你是一位资深技术面试官，请为以下面试场景生成 %d 道面试题目。

面试信息：
- 目标岗位：%s
- 面试难度：%s
- 工作经验：%s
- 面试轮次：%s
- 重点方向：%s
- 补充说明：%s

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
}

// ReviewAnswer 对答案进行 AI 点评，视频模式时附加非语言指标分析
func ReviewAnswer(question *model.Question, answer string, cfg *model.InterviewConfig, metrics *model.NonVerbalMetrics) (*ReviewResult, error) {
	prompt := fmt.Sprintf(`你是一位专业技术面试官，请对以下面试答案进行评分和点评。

题目：%s
知识点：%s
岗位：%s，难度：%s
候选人答案：%s

请严格按照以下 JSON 格式返回，不要包含任何其他文字、代码块标记：
{"score":85,"pros":["优点1","优点2"],"cons":["不足1","不足2"],"referenceAnswer":"参考答案（支持Markdown）"}

评分标准：完整准确90-100分，基本正确60-80分，有误20-50分，完全错误0-10分。`,
		question.Content,
		strings.Join(question.Tags, "、"),
		cfg.JobTitle,
		cfg.Difficulty,
		answer,
	)

	// 视频模式：附加非语言指标上下文
	if metrics != nil {
		prompt += fmt.Sprintf(`

[语音表达数据]
语速：%.0f字/分钟（推荐范围120-150），停顿次数：%d次，作答时长：%d秒。
请额外在JSON中新增字段：expressionScore（0-100，评估口头表达流畅度、逻辑性、语言规范性）和 expressionFeedback（针对口头表达的具体改进建议，含语速和停顿反馈）。`,
			metrics.SpeechRate, metrics.PauseCount, metrics.Duration,
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

	prompt := fmt.Sprintf(`你是一位资深面试顾问，请根据以下视频面试数据给出整体口头表达能力评价。

岗位：%s
完成题数：%d
平均语速：%.0f字/分钟（推荐120-150）
平均表达得分：%d分

请严格按照以下 JSON 格式返回，不要包含任何其他文字：
{"summary":"整体口头表达能力评价（100字以内）","suggestions":["改进建议1","改进建议2","改进建议3"]}`,
		jobTitle, len(answers), avgRate, avgScore,
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

// ---- AI 知识文章生成 ----

// GenerateArticle 调用 DeepSeek 生成知识文章，返回填充好的 model.Article（articleId 留空，由调用方生成）
func GenerateArticle(topic, jobCategory string) (*model.Article, error) {
	prompt := fmt.Sprintf(`你是一个技术知识博主。请围绕【%s】这个知识点，为【%s】岗位求职者撰写一篇高质量的备考知识文章。
要求：
1. 内容深度适合面试备考，覆盖核心概念、常见考点和实战建议
2. 正文使用 Markdown 格式，包含标题、代码块、列表等
3. 文章长度 800-1200 字
请严格按以下 JSON 格式返回，只返回 JSON，不要其他内容：
{"title":"文章标题","content":"Markdown格式正文","tags":["标签1","标签2","标签3"]}`,
		topic, jobCategory,
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
	default:
		return "均衡分配各类题目"
	}
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
