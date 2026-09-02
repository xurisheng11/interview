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

// ---- 简历解析 ----

// ParseResumeText 调用 AI 从原始文本中提取结构化简历内容
func ParseResumeText(rawText string) (*model.ResumeContent, error) {
	prompt := fmt.Sprintf(`你是一个简历解析专家，请从以下简历文本中提取结构化信息。

简历文本：
%s

请严格按照以下 JSON 格式返回，不要包含任何其他文字、代码块标记：
{
  "jobTitle": "目标岗位（从简历中推断，如无则返回空字符串）",
  "workExperience": [
    {"company": "公司名", "position": "职位", "duration": "时间段", "desc": "工作描述"}
  ],
  "projects": [
    {"name": "项目名", "role": "担任角色", "stack": "技术栈", "desc": "项目描述"}
  ],
  "skills": ["技能1", "技能2", "技能3"]
}

如果某项为空则返回空数组或空字符串，不要省略字段。`, rawText)

	raw, err := Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 解析简历失败: %w", err)
	}
	raw = cleanJSON(raw)

	var content model.ResumeContent
	if err := json.Unmarshal([]byte(raw), &content); err != nil {
		return nil, fmt.Errorf("解析结构化内容失败: %w", err)
	}
	content.RawText = rawText
	return &content, nil
}

// ---- 简历分析 ----

// AnalyzeResume 对简历进行 AI 评分分析，异步触发
func AnalyzeResume(resumeID string) {
	go func() {
		// 标记为 analyzing
		_ = repository.UpdateResumeAnalysis(resumeID, "analyzing", nil)

		r, err := repository.GetResume(resumeID)
		if err != nil || r == nil {
			_ = repository.UpdateResumeAnalysis(resumeID, "failed", nil)
			return
		}

		analysis, err := doAnalyze(&r.ParsedContent)
		if err != nil {
			_ = repository.UpdateResumeAnalysis(resumeID, "failed", nil)
			return
		}

		_ = repository.UpdateResumeAnalysis(resumeID, "done", analysis)
	}()
}

func doAnalyze(content *model.ResumeContent) (*model.ResumeAnalysis, error) {
	prompt := fmt.Sprintf(`你是一位专业HR和简历评审专家，请对以下简历内容进行综合评分。

目标岗位：%s
工作经历：%s
项目经历：%s
技能关键词：%s

请从以下4个维度分别评分（0-100分），并给出至少3条改进建议：
1. 项目质量（project_quality）：项目描述的深度、技术含量、量化成果
2. 技能完整性（skill_completeness）：技能关键词覆盖度、表述规范性
3. 格式规范（format_compliance）：信息完整性、结构清晰度、描述专业性
4. 亮点提炼（highlights）：差异化竞争力、个人特色、成就展示

请严格按照以下 JSON 格式返回，不要包含任何其他文字、代码块标记：
{
  "totalScore": 85,
  "dimensions": [
    {"name": "项目质量", "score": 80},
    {"name": "技能完整性", "score": 90},
    {"name": "格式规范", "score": 85},
    {"name": "亮点提炼", "score": 75}
  ],
  "suggestions": ["改进建议1", "改进建议2", "改进建议3"]
}`,
		content.JobTitle,
		formatWorkExp(content.WorkExperience),
		formatProjects(content.Projects),
		strings.Join(content.Skills, "、"),
	)

	raw, err := Chat(prompt)
	if err != nil {
		return nil, err
	}
	raw = cleanJSON(raw)

	var analysis model.ResumeAnalysis
	if err := json.Unmarshal([]byte(raw), &analysis); err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %w", err)
	}

	// 校验分数范围
	if analysis.TotalScore < 0 {
		analysis.TotalScore = 0
	}
	if analysis.TotalScore > 100 {
		analysis.TotalScore = 100
	}
	for i := range analysis.Dimensions {
		if analysis.Dimensions[i].Score < 0 {
			analysis.Dimensions[i].Score = 0
		}
		if analysis.Dimensions[i].Score > 100 {
			analysis.Dimensions[i].Score = 100
		}
	}
	analysis.AnalyzedAt = time.Now()
	return &analysis, nil
}

func formatWorkExp(entries []model.WorkEntry) string {
	if len(entries) == 0 {
		return "无"
	}
	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("[%s @ %s (%s)]: %s\n", e.Position, e.Company, e.Duration, e.Desc))
	}
	return sb.String()
}

func formatProjects(entries []model.ProjectEntry) string {
	if len(entries) == 0 {
		return "无"
	}
	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("[%s | %s | 栈: %s]: %s\n", e.Name, e.Role, e.Stack, e.Desc))
	}
	return sb.String()
}

// ---- 简历驱动面试题目生成 ----

// GenerateResumeQuestions 根据简历内容生成面试题目
func GenerateResumeQuestions(content *model.ResumeContent) ([]model.Question, error) {
	prompt := fmt.Sprintf(`你是一位资深技术面试官，请根据以下候选人简历，生成8-12道有针对性的面试题目。

目标岗位：%s
工作经历：%s
项目经历：%s
技能关键词：%s

题目要求：
- 每个项目经历至少1道深挖题（项目背景、技术选型、难点攻克）
- 每个核心技能至少1道考察题（原理、实践、边界情况）
- 包含1-2道综合能力题（系统设计或经验总结）
- 题目总数控制在8-15道

请严格按照以下 JSON 数组格式返回，不要包含任何其他文字、代码块标记：
[{"index":0,"content":"题目内容","tags":["知识点"],"difficulty":"middle","estimatedMinutes":5,"type":"basic"}]

difficulty 取值：junior/middle/senior
type 取值：basic/algorithm/design/hr`,
		content.JobTitle,
		formatWorkExp(content.WorkExperience),
		formatProjects(content.Projects),
		strings.Join(content.Skills, "、"),
	)

	raw, err := Chat(prompt)
	if err != nil {
		return nil, err
	}
	raw = cleanJSON(raw)

	var questions []model.Question
	if err := json.Unmarshal([]byte(raw), &questions); err != nil {
		return nil, fmt.Errorf("解析题目失败: %w", err)
	}

	// 校验题目数量和内容
	var valid []model.Question
	for _, q := range questions {
		if strings.TrimSpace(q.Content) != "" {
			valid = append(valid, q)
		}
	}
	if len(valid) < 5 {
		return nil, fmt.Errorf("生成题目数量不足（最少5道），实际: %d", len(valid))
	}
	if len(valid) > 15 {
		valid = valid[:15]
	}
	// 重新编排 index
	for i := range valid {
		valid[i].Index = i
	}
	return valid, nil
}

// ---- 简历面试会话创建 ----

type ResumeSessionExtended struct {
	model.InterviewSession
	Source   string `json:"source"`
	ResumeID string `json:"resumeId"`
}

func (s *ResumeSessionExtended) ToJSON() (string, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

// CreateResumeInterview 创建基于简历的面试会话（含重试）
func CreateResumeInterview(userID, resumeID string) (*ResumeSessionExtended, error) {
	r, err := repository.GetResume(resumeID)
	if err != nil || r == nil {
		return nil, fmt.Errorf("简历不存在")
	}
	if r.UserID != userID {
		return nil, fmt.Errorf("无权操作")
	}
	if r.AnalysisStatus == "failed" || r.AnalysisStatus == "pending" || r.AnalysisStatus == "analyzing" {
		return nil, fmt.Errorf("简历分析未完成，无法发起面试")
	}

	// 生成题目（失败重试一次）
	questions, err := GenerateResumeQuestions(&r.ParsedContent)
	if err != nil {
		questions, err = GenerateResumeQuestions(&r.ParsedContent)
		if err != nil {
			return nil, fmt.Errorf("题目生成失败: %w", err)
		}
	}

	jobTitle := r.ParsedContent.JobTitle
	if jobTitle == "" {
		jobTitle = "通用岗位"
	}

	session := &ResumeSessionExtended{
		InterviewSession: model.InterviewSession{
			InterviewID: uuid.New().String(),
			UserID:      userID,
			Config: model.InterviewConfig{
				JobTitle:   jobTitle,
				Difficulty: "middle",
				Round:      "round1",
				Mode:       "text",
			},
			Questions:    questions,
			CurrentIndex: 0,
			Answers:      make(map[int]*model.AnswerRecord),
			Status:       "ongoing",
			Mode:         "text",
			StartTime:    time.Now(),
		},
		Source:   "resume",
		ResumeID: resumeID,
	}

	// 序列化并保存到 Redis（7天 TTL）
	b, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	if err := repository.Set(
		fmt.Sprintf("interview:session:%s", session.InterviewID),
		string(b),
		7*24*time.Hour,
	); err != nil {
		return nil, err
	}

	// 记录用户面试索引
	_ = repository.AddResumeInterview(userID, session.InterviewID)

	return session, nil
}
