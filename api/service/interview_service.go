package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"interview-sim/model"
	"interview-sim/repository"
)

type CreateInterviewReq struct {
	JobTitle        string   `json:"jobTitle" binding:"required"`
	Difficulty      string   `json:"difficulty" binding:"required"` // easy/medium/hard
	Experience      string   `json:"experience"`                    // fresh/1-3/3-5/5+ （可选，未传则从用户profile读取）
	Round           string   `json:"round" binding:"required"`      // round1/round2/round3
	FocusAreas     []string `json:"focusAreas"`
	Remark         string   `json:"remark"`
	Mode           string   `json:"mode"` // "text" | "video"，默认 "text"

	// 新增字段
	CompanyID       string   `json:"companyId"`
	CompanyName     string   `json:"companyName"`
	InterviewTypes  []string `json:"interviewTypes"`  // ["structured", "semi-structured", "random"]
	ThinkTime       int      `json:"thinkTime"`       // 思考时间(秒)
	VirtualBackground bool  `json:"virtualBackground"` // 虚拟背景
	BgStyle         string   `json:"bgStyle"`        // 背景样式: office, blue, gray, blur

	// 简历关联（可选，传了则生成题目时注入简历上下文）
	ResumeID        string   `json:"resumeId"`
}

type SubmitAnswerReq struct {
	QuestionIndex    int                    `json:"questionIndex" binding:"min=0"`
	Answer           string                 `json:"answer" binding:"required"`
	NonVerbalMetrics *model.NonVerbalMetrics `json:"nonVerbalMetrics,omitempty"` // 视频模式附加
	ThinkDuration    int                    `json:"thinkDuration,omitempty"` // 思考时长(秒)
	VerbalTics      []string               `json:"verbalTics,omitempty"`    // 识别到的口头禅
}

// formatResumeContext 将简历结构化内容格式化为 AI prompt 上下文
func formatResumeContext(content *model.ResumeContent) string {
	var sb strings.Builder
	sb.WriteString("\n\n【候选人简历背景】（请在出题时围绕以下经历深挖）\n")

	if content.JobTitle != "" {
		sb.WriteString(fmt.Sprintf("目标岗位：%s\n", content.JobTitle))
	}

	if len(content.WorkExperience) > 0 {
		sb.WriteString("工作经历：\n")
		for _, w := range content.WorkExperience {
			sb.WriteString(fmt.Sprintf("  - [%s @ %s (%s)]: %s\n", w.Position, w.Company, w.Duration, w.Desc))
		}
	}

	if len(content.Projects) > 0 {
		sb.WriteString("项目经历：\n")
		for _, p := range content.Projects {
			sb.WriteString(fmt.Sprintf("  - [%s | 角色: %s | 技术栈: %s]: %s\n", p.Name, p.Role, p.Stack, p.Desc))
		}
	}

	if len(content.Skills) > 0 {
		sb.WriteString(fmt.Sprintf("技能关键词：%s\n", strings.Join(content.Skills, "、")))
	}

	return sb.String()
}

// CreateInterview 创建面试（含题目生成，可选注入简历上下文）
func CreateInterview(userID string, req *CreateInterviewReq) (*model.InterviewSession, error) {
	// 如果没有传工作经验，从用户 profile 中读取
	experience := req.Experience
	if experience == "" {
		if profile, err := GetProfile(userID); err == nil {
			if exp, ok := profile["experience"].(string); ok && exp != "" {
				experience = exp
			}
		}
	}
	// 仍然没有，使用默认值
	if experience == "" {
		experience = "fresh"
	}

	cfg := &model.InterviewConfig{
		JobTitle:   req.JobTitle,
		Difficulty: req.Difficulty,
		Experience: experience,
		Round:      req.Round,
		FocusAreas: req.FocusAreas,
		Remark:     req.Remark,
		Mode:       req.Mode,
	}

	// 设置默认值
	if cfg.Mode == "" {
		cfg.Mode = "text"
	}
	if req.ThinkTime <= 0 {
		req.ThinkTime = 15 // 默认15秒思考时间
	}
	if req.BgStyle == "" {
		req.BgStyle = "blur"
	}

	// 创建会话
	session := &model.InterviewSession{
		InterviewID:  uuid.New().String(),
		UserID:       userID,
		Config:       *cfg,
		CurrentIndex: 0,
		Answers:      make(map[int]*model.AnswerRecord),
		Status:       "ongoing",
		Mode:         cfg.Mode,
		StartTime:    time.Now(),

		// 新增字段
		CompanyID:      req.CompanyID,
		CompanyName:    req.CompanyName,
		InterviewTypes: req.InterviewTypes,
		ThinkTime:      req.ThinkTime,
		VirtualBackground: req.VirtualBackground,
		BgStyle:        req.BgStyle,
	}

	// 1. 如果传了简历 ID，读取简历内容注入到题目生成
	var resumeContext string
	if req.ResumeID != "" {
		if r, err := repository.GetResume(req.ResumeID); err == nil && r != nil && r.UserID == userID {
			if len(r.ParsedContent.Projects) > 0 || len(r.ParsedContent.WorkExperience) > 0 || len(r.ParsedContent.Skills) > 0 {
				resumeContext = formatResumeContext(&r.ParsedContent)
				session.ResumeID = req.ResumeID
			}
		}
	}

	// 2. 查题目缓存
	questions, err := repository.GetQuestionsCache(cfg)
	if err != nil || questions == nil {
		// 3. 调 DeepSeek 生成（注入简历上下文）
		questions, err = GenerateQuestions(cfg, resumeContext)
		if err != nil {
			return nil, fmt.Errorf("题目生成失败: %w", err)
		}
		// 4. 写缓存
		_ = repository.SetQuestionsCache(cfg, questions)
	}

	session.Questions = questions

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

	// 构建非语言指标扩展
	var extendedMetrics *model.NonVerbalMetrics
	if req.NonVerbalMetrics != nil {
		extendedMetrics = req.NonVerbalMetrics
		// 添加口头禅信息
		if len(req.VerbalTics) > 0 {
			extendedMetrics.VerbalTics = req.VerbalTics
		}
		// 添加思考时长
		if req.ThinkDuration > 0 {
			extendedMetrics.ThinkDuration = req.ThinkDuration
		}
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
		NonVerbalMetrics:   extendedMetrics,
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
			CompanyName: session.CompanyName,
			Mode:        session.Mode,
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
