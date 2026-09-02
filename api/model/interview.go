package model

import (
	"encoding/json"
	"time"
)

// InterviewConfig 面试配置
type InterviewConfig struct {
	JobTitle       string   `json:"jobTitle"`
	Difficulty     string   `json:"difficulty"`   // junior/middle/senior
	Experience     string   `json:"experience"`   // fresh/1-3/3-5/5+
	Round          string   `json:"round"`        // round1/round2/round3
	FocusAreas     []string `json:"focusAreas"`
	Remark         string   `json:"remark"`
	Mode           string   `json:"mode"`         // text | video，默认 text
	InterviewTypes []string `json:"interviewTypes"` // 面试形式: structured, semi-structured, random
}

// NonVerbalMetrics 非语言行为指标（视频面试模式专有）
type NonVerbalMetrics struct {
	SpeechRate      float64  `json:"speechRate"`      // 每分钟字数（WPM）
	PauseCount      int      `json:"pauseCount"`      // 停顿次数（>2秒算一次）
	Duration        int      `json:"duration"`        // 作答时长（秒）
	ThinkDuration   int      `json:"thinkDuration"`   // 思考时长（秒）
	VerbalTics      []string `json:"verbalTics"`       // 识别到的口头禅列表
}

// Question 面试题目
type Question struct {
	Index            int      `json:"index"`
	Content          string   `json:"content"`
	Tags             []string `json:"tags"`
	Difficulty       string   `json:"difficulty"`
	EstimatedMinutes int      `json:"estimatedMinutes"`
	Type             string   `json:"type"`
}

// AnswerRecord 单题作答记录
type AnswerRecord struct {
	UserAnswer      string   `json:"userAnswer"`
	Score           int      `json:"score"`
	Pros            []string `json:"pros"`
	Cons            []string `json:"cons"`
	ReferenceAnswer string   `json:"referenceAnswer"`
	Skipped         bool     `json:"skipped"`
	SubmittedAt     string   `json:"submittedAt"`

	// 视频面试专有字段（omitempty，文字模式不写入）
	ExpressionScore    int               `json:"expressionScore,omitempty"`
	ExpressionFeedback string            `json:"expressionFeedback,omitempty"`
	NonVerbalMetrics  *NonVerbalMetrics `json:"nonVerbalMetrics,omitempty"`
}

// InterviewSession 面试会话（存 Redis）
type InterviewSession struct {
	InterviewID  string                `json:"interviewId"`
	UserID       string                `json:"userId"`
	Config       InterviewConfig       `json:"config"`
	Questions    []Question            `json:"questions"`
	CurrentIndex int                   `json:"currentIndex"`
	Answers      map[int]*AnswerRecord `json:"answers"`
	Status       string                `json:"status"` // ongoing/paused/completed
	Mode         string                `json:"mode"`   // text | video，默认 text
	StartTime    time.Time             `json:"startTime"`
	PauseTime    *time.Time            `json:"pauseTime,omitempty"`

	// 新增字段
	CompanyID           string   `json:"companyId"`            // 目标公司ID
	CompanyName         string   `json:"companyName"`          // 目标公司名称
	InterviewTypes      []string `json:"interviewTypes"`       // 面试形式: structured, semi-structured, random
	ThinkTime           int      `json:"thinkTime"`           // 思考时间(秒)
	VirtualBackground   bool     `json:"virtualBackground"`    // 虚拟背景
	BgStyle             string   `json:"bgStyle"`             // 背景样式
}

func (s *InterviewSession) ToJSON() (string, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

func SessionFromJSON(raw string) (*InterviewSession, error) {
	var s InterviewSession
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, err
	}
	if s.Answers == nil {
		s.Answers = make(map[int]*AnswerRecord)
	}
	return &s, nil
}

// InterviewListItem 面试历史列表项
type InterviewListItem struct {
	InterviewID string `json:"interviewId"`
	JobTitle    string `json:"jobTitle"`
	Round       string `json:"round"`
	Difficulty  string `json:"difficulty"`
	Status      string `json:"status"`
	TotalScore  int    `json:"totalScore"`
	StartTime   string `json:"startTime"`
	CompanyName string `json:"companyName,omitempty"` // 新增:目标公司
	Mode        string `json:"mode,omitempty"`        // 新增:面试模式
}
