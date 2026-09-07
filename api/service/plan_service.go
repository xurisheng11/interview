package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"interview-sim/model"
	"interview-sim/repository"

	"github.com/go-redis/redis/v8"
)

var ErrPlanNotFound = errors.New("冲刺计划不存在")

func sprintPlanKey(userId string) string {
	return "sprint_plan:" + userId
}

// 题型类别 id -> 中文名（与前端学习中心五大类保持一致）
var weakAreaNames = map[string]string{
	"basic":         "基础知识",
	"project":       "项目经验",
	"algorithm":     "算法与数据结构",
	"system_design": "系统设计",
	"behavior":      "行为面试",
}

var experienceNames = map[string]string{
	"fresh": "应届生/在校生",
	"1-3":   "1-3年经验",
	"3-5":   "3-5年经验",
	"5+":    "5年以上经验",
}

// GenerateSprintPlan 根据用户配置调用 AI 生成个性化冲刺计划并持久化
func GenerateSprintPlan(userId string, cfg *model.SprintPlanConfig) (*model.SprintPlan, error) {
	if cfg.JobTitle == "" {
		return nil, errors.New("目标岗位不能为空")
	}
	if cfg.TotalDays < 7 || cfg.TotalDays > 180 {
		return nil, errors.New("冲刺天数需在 7-180 天之间")
	}
	if cfg.DailyMinutes <= 0 {
		cfg.DailyMinutes = 60
	}

	prompt := buildSprintPlanPrompt(cfg)
	raw, err := Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 生成计划失败: %w", err)
	}

	var plan model.SprintPlan
	if err := json.Unmarshal([]byte(cleanJSON(raw)), &plan); err != nil {
		return nil, fmt.Errorf("计划解析失败，请重试: %w", err)
	}
	if len(plan.Phases) == 0 {
		return nil, errors.New("AI 生成的计划为空，请重试")
	}

	// 服务端补充元数据，避免依赖 AI 输出
	plan.Config = *cfg
	plan.StartDate = time.Now().Format("2006-01-02")
	plan.CreatedAt = time.Now()
	for pi := range plan.Phases {
		for ti := range plan.Phases[pi].Tasks {
			plan.Phases[pi].Tasks[ti].Done = false
		}
	}

	if err := saveSprintPlan(userId, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetSprintPlan 获取用户当前冲刺计划（Redis 未命中时回源 MySQL 并回填）
func GetSprintPlan(userId string) (*model.SprintPlan, error) {
	raw, err := repository.Get(sprintPlanKey(userId))
	if err != nil && err != redis.Nil {
		// 真实 Redis 错误原样上抛，避免故障时误用 MySQL 旧数据覆盖 Redis 新数据
		return nil, err
	}
	if err == redis.Nil || raw == "" {
		// Redis 未命中，尝试 MySQL 回源；回源失败保持原有 not found 语义
		if repository.MySQLAvailable() {
			if data, derr := repository.QuerySprintPlanData(userId); derr == nil && data != "" {
				// 回填 Redis（计划永久保存）
				_ = repository.SetPermanent(sprintPlanKey(userId), data)
				raw = data
			}
		}
		if raw == "" {
			return nil, ErrPlanNotFound
		}
	}
	var plan model.SprintPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("计划数据损坏: %w", err)
	}
	return &plan, nil
}

// UpdateSprintTask 勾选/取消某个阶段任务的完成状态
func UpdateSprintTask(userId string, phaseIndex, taskIndex int, done bool) (*model.SprintPlan, error) {
	plan, err := GetSprintPlan(userId)
	if err != nil {
		return nil, err
	}
	if phaseIndex < 0 || phaseIndex >= len(plan.Phases) {
		return nil, errors.New("阶段索引无效")
	}
	if taskIndex < 0 || taskIndex >= len(plan.Phases[phaseIndex].Tasks) {
		return nil, errors.New("任务索引无效")
	}
	plan.Phases[phaseIndex].Tasks[taskIndex].Done = done
	if err := saveSprintPlan(userId, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// DeleteSprintPlan 删除当前冲刺计划
func DeleteSprintPlan(userId string) error {
	return repository.Del(sprintPlanKey(userId))
}

func saveSprintPlan(userId string, plan *model.SprintPlan) error {
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	return repository.SetPermanent(sprintPlanKey(userId), string(data))
}

func buildSprintPlanPrompt(cfg *model.SprintPlanConfig) string {
	company := cfg.TargetCompany
	if company == "" {
		company = "不限"
	}
	expName := experienceNames[cfg.Experience]
	if expName == "" {
		expName = "未填写"
	}
	var weakList []string
	for _, w := range cfg.WeakAreas {
		if name, ok := weakAreaNames[w]; ok {
			weakList = append(weakList, name)
		}
	}
	weakDesc := "无特别薄弱项"
	if len(weakList) > 0 {
		weakDesc = strings.Join(weakList, "、")
	}

	return fmt.Sprintf(`你是一位资深面试辅导教练。请为求职者定制一份 %d 天的面试冲刺计划。

求职者情况：
- 目标岗位：%s
- 目标公司：%s
- 工作经验：%s
- 每天可投入学习时间：约 %d 分钟
- 自认为薄弱的环节：%s

计划要求：
1. 将 %d 天划分为 3-5 个循序渐进的阶段（如基础巩固→专项突破→模拟实战→冲刺复盘），阶段天数区间连续且覆盖全部天数
2. 每个阶段给出 4-6 条具体任务，任务量要匹配每日可投入时间，薄弱环节安排更多任务
3. 任务应结合以下平台功能，并在 action 字段标注（仅可用这些值，无对应功能则留空字符串）：
   - practice：题库分类练习
   - mock：发起 AI 模拟面试
   - resume：完善/分析简历
   - company：查看公司面试知识库
   - community：阅读社区面经文章
4. 给出 3-5 条每日例行任务（dailyRoutine）
5. 在关键节点设置 2-4 个里程碑检验点（milestones），如"第X天完成一次完整模拟面试，目标70分"
6. tips 给出一段 50 字以内的冲刺心态建议

请严格只返回 JSON（不要 markdown 代码块），格式如下：
{
  "title": "计划标题（15字以内，包含天数和岗位）",
  "summary": "计划总体思路（60字以内）",
  "phases": [
    {
      "name": "阶段名称",
      "dayStart": 1,
      "dayEnd": 14,
      "goal": "阶段目标（30字以内）",
      "tasks": [
        { "content": "具体任务描述", "action": "practice" }
      ]
    }
  ],
  "dailyRoutine": ["每日例行任务1"],
  "milestones": [
    { "day": 14, "content": "检验点描述" }
  ],
  "tips": "冲刺建议"
}`, cfg.TotalDays, cfg.JobTitle, company, expName, cfg.DailyMinutes, weakDesc, cfg.TotalDays)
}
