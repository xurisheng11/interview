package model

import "time"

// SprintPlanConfig 用户定制冲刺计划时的配置
type SprintPlanConfig struct {
	JobTitle      string   `json:"jobTitle"`      // 目标岗位
	TargetCompany string   `json:"targetCompany"` // 目标公司（可选）
	TotalDays     int      `json:"totalDays"`     // 冲刺天数：30/60/90
	DailyMinutes  int      `json:"dailyMinutes"`  // 每日可投入时间（分钟）
	Experience    string   `json:"experience"`    // 工作经验
	WeakAreas     []string `json:"weakAreas"`     // 薄弱环节（题型类别）
}

// SprintTask 计划中的单个任务
type SprintTask struct {
	Content string `json:"content"` // 任务内容
	Action  string `json:"action"`  // 站内功能跳转：practice/mock/resume/company/community，可为空
	Done    bool   `json:"done"`    // 是否已完成
}

// SprintPhase 冲刺阶段
type SprintPhase struct {
	Name     string       `json:"name"`     // 阶段名称
	DayStart int          `json:"dayStart"` // 起始天（含）
	DayEnd   int          `json:"dayEnd"`   // 结束天（含）
	Goal     string       `json:"goal"`     // 阶段目标
	Tasks    []SprintTask `json:"tasks"`    // 阶段任务清单
}

// SprintMilestone 里程碑检验点
type SprintMilestone struct {
	Day     int    `json:"day"`     // 第几天
	Content string `json:"content"` // 检验内容
}

// SprintPlan 完整冲刺计划
type SprintPlan struct {
	Title        string            `json:"title"`        // 计划标题
	Summary      string            `json:"summary"`      // 计划总述
	Config       SprintPlanConfig  `json:"config"`       // 生成时的配置
	StartDate    string            `json:"startDate"`    // 开始日期 YYYY-MM-DD
	Phases       []SprintPhase     `json:"phases"`       // 阶段列表
	DailyRoutine []string          `json:"dailyRoutine"` // 每日例行任务
	Milestones   []SprintMilestone `json:"milestones"`   // 里程碑
	Tips         string            `json:"tips"`         // 冲刺建议
	CreatedAt    time.Time         `json:"createdAt"`    // 创建时间
}
