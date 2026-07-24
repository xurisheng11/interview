package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"interview-sim/model"
	"interview-sim/repository"
)

type CreateInterviewReq struct {
	JobTitle   string   `json:"jobTitle" binding:"required"`
	Difficulty string   `json:"difficulty" binding:"required"`
	Experience string   `json:"experience" binding:"required"`
	Round      string   `json:"round" binding:"required"`
	FocusAreas []string `json:"focusAreas"`
	Remark     string   `json:"remark"`
	Mode       string   `json:"mode"` // "text" | "video"，默认 "text"
}

type SubmitAnswerReq struct {
	QuestionIndex    int                    `json:"questionIndex" binding:"min=0"`
	Answer           string                 `json:"answer" binding:"required"`
	NonVerbalMetrics *model.NonVerbalMetrics `json:"nonVerbalMetrics,omitempty"` // 视频模式附加
}

// CreateInterview 创建面试（含题目生成）
func CreateInterview(userID string, req *CreateInterviewReq) (*model.InterviewSession, error) {
	cfg := &model.InterviewConfig{
		JobTitle:   req.JobTitle,
		Difficulty: req.Difficulty,
		Experience: req.Experience,
		Round:      req.Round,
		FocusAreas: req.FocusAreas,
		Remark:     req.Remark,
		Mode:       req.Mode,
	}
	if cfg.Mode == "" {
		cfg.Mode = "text"
	}

	// 1. 查题目缓存
	questions, err := repository.GetQuestionsCache(cfg)
	if err != nil || questions == nil {
		// 2. 调 DeepSeek 生成
		questions, err = GenerateQuestions(cfg)
		if err != nil {
			return nil, fmt.Errorf("题目生成失败: %w", err)
		}
		// 3. 写缓存
		_ = repository.SetQuestionsCache(cfg, questions)
	}

	// 创建会话
	session := &model.InterviewSession{
		InterviewID:  uuid.New().String(),
		UserID:       userID,
		Config:       *cfg,
		Questions:    questions,
		CurrentIndex: 0,
		Answers:      make(map[int]*model.AnswerRecord),
		Status:       "ongoing",
		Mode:         cfg.Mode,
		StartTime:    time.Now(),
	}

	// 保存到 Redis
	if err := repository.SaveSession(session); err != nil {
		return nil, err
	}
	// 记录用户面试索引
	_ = repository.AddUserInterview(userID, session.InterviewID)

	return session, nil
}

// SubmitAnswer 提交单题答案（AI 点评）
func SubmitAnswer(userID, interviewID string, req *SubmitAnswerReq) (*ReviewResult, error) {
	session, err := repository.GetSession(interviewID)
	if err != nil || session == nil {
		return nil, fmt.Errorf("面试不存在")
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("无权操作")
	}
	if req.QuestionIndex < 0 || req.QuestionIndex >= len(session.Questions) {
		return nil, fmt.Errorf("题目索引超出范围")
	}

	question := &session.Questions[req.QuestionIndex]
	result, err := ReviewAnswer(question, req.Answer, &session.Config, req.NonVerbalMetrics)
	if err != nil {
		return nil, fmt.Errorf("AI 点评失败: %w", err)
	}

	// 保存答案记录
	record := &model.AnswerRecord{
		UserAnswer:         req.Answer,
		Score:              result.Score,
		Pros:               result.Pros,
		Cons:               result.Cons,
		ReferenceAnswer:    result.ReferenceAnswer,
		Skipped:            false,
		SubmittedAt:        time.Now().Format(time.RFC3339),
		ExpressionScore:    result.ExpressionScore,
		ExpressionFeedback: result.ExpressionFeedback,
		NonVerbalMetrics:   req.NonVerbalMetrics,
	}
	_ = repository.UpdateSessionAnswer(interviewID, req.QuestionIndex, record)
	return result, nil
}

// SkipQuestion 跳过题目（本地标记，前端处理，此接口可选）
func SkipQuestion(userID, interviewID string, questionIndex int) error {
	session, err := repository.GetSession(interviewID)
	if err != nil || session == nil {
		return fmt.Errorf("面试不存在")
	}
	if session.UserID != userID {
		return fmt.Errorf("无权操作")
	}
	record := &model.AnswerRecord{
		Skipped:     true,
		Score:       0,
		SubmittedAt: time.Now().Format(time.RFC3339),
	}
	return repository.UpdateSessionAnswer(interviewID, questionIndex, record)
}

// PauseInterview 暂停面试
func PauseInterview(userID, interviewID string) error {
	session, err := repository.GetSession(interviewID)
	if err != nil || session == nil {
		return fmt.Errorf("面试不存在")
	}
	if session.UserID != userID {
		return fmt.Errorf("无权操作")
	}
	return repository.PauseSession(interviewID)
}

// GetInterviewSession 获取面试会话
func GetInterviewSession(userID, interviewID string) (*model.InterviewSession, error) {
	session, err := repository.GetSession(interviewID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("面试不存在")
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("无权操作")
	}
	return session, nil
}

// GetInterviewList 获取用户面试列表
func GetInterviewList(userID string) ([]*model.InterviewListItem, error) {
	ids, err := repository.GetUserInterviewIDs(userID)
	if err != nil {
		return nil, err
	}
	var list []*model.InterviewListItem
	for _, id := range ids {
		session, err := repository.GetSession(id)
		if err != nil || session == nil {
			continue
		}
		item := &model.InterviewListItem{
			InterviewID: session.InterviewID,
			JobTitle:    session.Config.JobTitle,
			Round:       session.Config.Round,
			Difficulty:  session.Config.Difficulty,
			Status:      session.Status,
			StartTime:   session.StartTime.Format(time.RFC3339),
		}
		// 如果已完成，尝试从报告读取得分
		if session.Status == "completed" {
			reportRaw, err := repository.Get(fmt.Sprintf("report:%s:%s", userID, id))
			if err == nil && reportRaw != "" {
				var report model.InterviewReport
				if jsonErr := json.Unmarshal([]byte(reportRaw), &report); jsonErr == nil {
					item.TotalScore = report.TotalScore
				}
			}
		}
		list = append(list, item)
	}
	return list, nil
}
