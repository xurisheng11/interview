package model

import (
	"encoding/json"
	"time"
)

// InterviewConfig 面试配置
type InterviewConfig struct {
	JobTitle string `json:"jobTitle"`
	// Difficulty: easy/medium/hard（与题目难度对齐）
	Difficulty     string   `json:"difficulty"` // easy/medium/hard
	Experience     string   `json:"experience"` // fresh/1-3/3-5/5+
	Round          string   `json:"round"`      // round1/round2/round3/comprehensive（综合面试）
	FocusAreas     []string `json:"focusAreas"`
	Remark         string   `json:"remark"`
	Mode           string   `json:"mode"`           // text | video(语音) | video_call(视频)，默认 text
	InterviewTypes []string `json:"interviewTypes"` // 面试形式: structured, semi-structured, random
}

// NonVerbalMetrics 非语言行为指标（语音/视频面试模式专有）
type NonVerbalMetrics struct {
	SpeechRate    float64  `json:"speechRate"`    // 每分钟字数（WPM）
	PauseCount    int      `json:"pauseCount"`    // 停顿次数（>2秒算一次）
	Duration      int      `json:"duration"`      // 作答时长（秒）
	ThinkDuration int      `json:"thinkDuration"` // 思考时长（秒）
	VerbalTics    []string `json:"verbalTics"`    // 识别到的口头禅列表
	StutterCount  int      `json:"stutterCount,omitempty"` // 口吃/不流畅次数（转写文本中重复字/词）
}

// FaceEmotionInfo 视频面试抓帧情绪识别结果（腾讯云人脸属性 DetectFaceAttributes）
type FaceEmotionInfo struct {
	Type        int     `json:"type"`        // 情绪类型：0自然 1高兴 2惊讶 3生气 4悲伤 5厌恶 6害怕
	Name        string  `json:"name"`        // 情绪中文名
	Probability float64 `json:"probability"` // 识别置信度 [0,1]
	Smile       int     `json:"smile"`       // 是否微笑：0否 1是
	SmileProb   float64 `json:"smileProb"`   // 微笑置信度
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

	// 语音/视频面试专有字段（omitempty，文字模式不写入）
	ExpressionScore    int               `json:"expressionScore,omitempty"`
	ExpressionFeedback string            `json:"expressionFeedback,omitempty"`
	NonVerbalMetrics   *NonVerbalMetrics `json:"nonVerbalMetrics,omitempty"`
	FaceEmotion        *FaceEmotionInfo  `json:"faceEmotion,omitempty"` // 本题视频抓帧情绪（提交时从 session.FaceEmotions 合并）
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
	Mode         string                `json:"mode"`   // text | video(语音) | video_call(视频)，默认 text
	StartTime    time.Time             `json:"startTime"`
	PauseTime    *time.Time            `json:"pauseTime,omitempty"`

	// 新增字段
	CompanyID         string   `json:"companyId"`          // 目标公司ID
	CompanyName       string   `json:"companyName"`        // 目标公司名称
	InterviewTypes    []string `json:"interviewTypes"`     // 面试形式: structured, semi-structured, random
	ThinkTime         int      `json:"thinkTime"`          // 思考时间(秒)
	VirtualBackground bool     `json:"virtualBackground"`  // 虚拟背景
	BgStyle           string   `json:"bgStyle"`            // 背景样式
	ResumeID          string   `json:"resumeId,omitempty"` // 关联简历ID（用于简历关联出题）

	// 视频面试（video_call）逐题情绪抓帧结果，提交答案时合并进 AnswerRecord
	FaceEmotions map[int]*FaceEmotionInfo `json:"faceEmotions,omitempty"`
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
