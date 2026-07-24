package model

import (
	"time"
)

// ReportQuestion 报告中的题目详情
type ReportQuestion struct {
	Index            int      `json:"index"`
	Content          string   `json:"content"`
	Tags             []string `json:"tags"`
	Difficulty       string   `json:"difficulty"`
	EstimatedMinutes int      `json:"estimatedMinutes"`
	Type             string   `json:"type"`
	UserAnswer       string   `json:"userAnswer"`
	Score            int      `json:"score"`
	Pros             []string `json:"pros"`
	Cons             []string `json:"cons"`
	ReferenceAnswer  string   `json:"referenceAnswer"`
	Skipped          bool     `json:"skipped"`

	// 视频面试专有字段
	ExpressionScore    int               `json:"expressionScore,omitempty"`
	ExpressionFeedback string            `json:"expressionFeedback,omitempty"`
	NonVerbalMetrics   *NonVerbalMetrics `json:"nonVerbalMetrics,omitempty"`
}

// ModuleScore 知识点模块得分
type ModuleScore struct {
	Module   string  `json:"module"`
	Count    int     `json:"count"`
	AvgScore float64 `json:"avgScore"`
	Level    string  `json:"level"` // 优秀/良好/及格/待提升
}

// WeaknessDetail 不足详情（含建议）
type WeaknessDetail struct {
	Point      string `json:"point"`
	Suggestion string `json:"suggestion"`
	Resource   string `json:"resource"`
}

// AISummaryReport AI综合评价
type AISummaryReport struct {
	Strengths  []string         `json:"strengths"`
	Weaknesses []WeaknessDetail `json:"weaknesses"`
	Roadmap    string           `json:"roadmap"`
}

// InterviewReport 完整面试报告
type InterviewReport struct {
	InterviewID   string           `json:"interviewId"`
	UserID        string           `json:"userId"`
	JobTitle      string           `json:"jobTitle"`
	Difficulty    string           `json:"difficulty"`
	Round         string           `json:"round"`
	Experience    string           `json:"experience"`
	TotalScore    int              `json:"totalScore"`
	Grade         string           `json:"grade"`      // 优秀/良好/及格/待提升
	PassStatus    string           `json:"passStatus"` // pass/pending/fail
	PassReason    string           `json:"passReason"`
	StartTime     time.Time        `json:"startTime"`
	EndTime       time.Time        `json:"endTime"`
	TotalSeconds  int              `json:"totalSeconds"`
	AnsweredCount int              `json:"answeredCount"`
	SkippedCount  int              `json:"skippedCount"`
	TotalCount    int              `json:"totalCount"`
	Questions     []ReportQuestion `json:"questions"`
	ModuleScores  []ModuleScore    `json:"moduleScores"`
	AISummary     *AISummaryReport `json:"aiSummary"`
	CreatedAt     time.Time        `json:"createdAt"`

	// 视频面试专有字段
	Mode               string  `json:"mode"`
	ExpressionSummary  string  `json:"expressionSummary,omitempty"`
	AvgExpressionScore int     `json:"avgExpressionScore,omitempty"`
	AvgSpeechRate      float64 `json:"avgSpeechRate,omitempty"`
}

// CalcAvgExpressionScore 计算平均表达得分（仅视频模式）
func CalcAvgExpressionScore(questions []ReportQuestion) int {
	total, count := 0, 0
	for _, q := range questions {
		if !q.Skipped && q.ExpressionScore > 0 {
			total += q.ExpressionScore
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

// CalcAvgSpeechRate 计算平均语速（仅视频模式）
func CalcAvgSpeechRate(questions []ReportQuestion) float64 {
	total := 0.0
	count := 0
	for _, q := range questions {
		if !q.Skipped && q.NonVerbalMetrics != nil && q.NonVerbalMetrics.SpeechRate > 0 {
			total += q.NonVerbalMetrics.SpeechRate
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// CalcGrade 根据分数返回等级
func CalcGrade(score int) string {
	switch {
	case score >= 90:
		return "优秀"
	case score >= 75:
		return "良好"
	case score >= 60:
		return "及格"
	default:
		return "待提升"
	}
}

// CalcPassStatus 判断通过状态
func CalcPassStatus(score, totalCount, skippedCount int) (status, reason string) {
	if totalCount > 0 && float64(skippedCount)/float64(totalCount) > 0.5 {
		return "fail", "跳过题目过多"
	}
	if score >= 75 {
		return "pass", ""
	}
	if score >= 60 {
		return "pending", ""
	}
	return "fail", ""
}
