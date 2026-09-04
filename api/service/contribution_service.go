package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"interview-sim/model"
	"interview-sim/repository"
)

// ---------- key helpers ----------

func contributedQKey(id string) string           { return "contributed:Q:" + id }
func companyContributionsKey(company string) string {
	// company name → ZSet (score=createdAt, member=questionID)
	return "contributed:byCompany:" + strings.ToLower(company)
}
func userContributionsKey(userID string) string  { return "contributed:byUser:" + userID }
func userVotesKey(questionID string) string      { return "contributed:votes:" + questionID }
func userReportsKey(questionID string) string    { return "contributed:reports:" + questionID }
func contributionRecordKey(id string) string    { return "contribution:record:" + id }
func userCreditsKey(userID string) string        { return "user:credits:" + userID }

// ---------- 贡献题目 ----------

// SubmitContributedQuestion 用户提交面试题
func SubmitContributedQuestion(userID string, req *ContributeQuestionReq) (*model.ContributedQuestion, error) {
	if len(req.Questions) == 0 {
		return nil, fmt.Errorf("至少需要提交1道题目")
	}

	var saved []*model.ContributedQuestion
	for _, q := range req.Questions {
		if strings.TrimSpace(q.Content) == "" {
			continue
		}

		cq := &model.ContributedQuestion{
			ID:            uuid.New().String(),
			Company:       req.Company,
			JobTitle:      req.JobTitle,
			Content:       strings.TrimSpace(q.Content),
			QuestionType:  q.QuestionType,
			Difficulty:    q.Difficulty,
			Tags:          q.Tags,
			Year:          req.Year,
			Round:         req.Round,
			ContributorID: userID,
			Status:        "pending",
			CreatedAt:     time.Now().Unix(),
		}
		if cq.Difficulty == "" {
			cq.Difficulty = "medium"
		}
		if cq.QuestionType == "" {
			cq.QuestionType = "technical"
		}

		hash := cq.ToRedisHash()
		if err := repository.HSetMap(contributedQKey(cq.ID), hash); err != nil {
			continue
		}

		// 加入全局集合（用于管理员审核列表）
		_ = repository.SAdd("contributed:all:ids", cq.ID)
		// 写入公司索引
		_ = repository.ZAdd(companyContributionsKey(req.Company), float64(cq.CreatedAt), cq.ID)
		// 写入用户贡献记录
		_ = repository.ZAdd(userContributionsKey(userID), float64(cq.CreatedAt), cq.ID)

		saved = append(saved, cq)
	}

	if len(saved) == 0 {
		return nil, fmt.Errorf("题目内容不能为空")
	}

	// 记录贡献流水（用于后续奖励判定）
	record := &model.QuestionContribution{
		ID:            uuid.New().String(),
		UserID:        userID,
		Company:       req.Company,
		JobTitle:      req.JobTitle,
		Questions:     fmt.Sprintf("%d道题", len(saved)),
		RewardStatus:  "pending",
		CreditReward:  0,
		CreatedAt:     time.Now().Unix(),
	}
	b, _ := json.Marshal(saved)
	record.Questions = string(b)

	b2, _ := json.Marshal(record)
	_ = repository.Set(contributionRecordKey(record.ID), string(b2), 30*24*time.Hour)

	return saved[0], nil
}

// ContributedQuestionItem 请求体
type ContributeQuestionReq struct {
	Company   string                     `json:"company" binding:"required"`
	JobTitle  string                     `json:"jobTitle" binding:"required"`
	Year      int                        `json:"year"`
	Round     string                     `json:"round"`
	Questions []ContributeQuestionItem   `json:"questions" binding:"required,min=1"`
}

type ContributeQuestionItem struct {
	Content      string   `json:"content" binding:"required"`
	QuestionType string   `json:"questionType"` // "technical" | "behavioral"
	Difficulty   string   `json:"difficulty"`    // "easy" | "medium" | "hard"
	Tags         []string `json:"tags"`
}

// ---------- 获取公司贡献的题目（供面试使用）----------

// GetCompanyContributedQuestions 获取公司贡献的题目（审核通过的）
// 支持模糊匹配：如果精确匹配不到，会搜索包含该关键词的公司
func GetCompanyContributedQuestions(company, jobTitle string) ([]model.ContributedQuestion, error) {
	// 1. 先尝试精确匹配
	ids, err := repository.ZRevRange(companyContributionsKey(company), 0, -1)
	if err == nil && len(ids) > 0 {
		result := filterContributedQuestions(ids, jobTitle)
		if len(result) > 0 {
			return result, nil
		}
	}

	// 2. 精确匹配为空，尝试模糊匹配：搜索所有公司，查找名称包含搜索词的公司
	return fuzzySearchContributedQuestions(company, jobTitle)
}

// fuzzySearchContributedQuestions 模糊搜索公司题目
func fuzzySearchContributedQuestions(keyword, jobTitle string) ([]model.ContributedQuestion, error) {
	// 扫描 contributed:byCompany:* 所有 key
	var cursor uint64
	var matchedIds []string
	keywordLower := strings.ToLower(keyword)

	for {
		keys, nextCursor, err := repository.Scan("contributed:byCompany:*", 100, &cursor)
		if err != nil {
			break
		}
		for _, key := range keys {
			// 提取公司名称（去掉前缀）
			companyName := strings.TrimPrefix(key, "contributed:byCompany:")
			// 解码 URL 编码的公司名
			companyName, _ = url.QueryUnescape(companyName)
			companyNameLower := strings.ToLower(companyName)

			// 模糊匹配：搜索词是公司名的子串，或者公司名是搜索词的子串
			if strings.Contains(companyNameLower, keywordLower) ||
				strings.Contains(keywordLower, companyNameLower) {
				// 获取该公司下的所有题目 ID
				ids, _ := repository.ZRevRange(key, 0, -1)
				matchedIds = append(matchedIds, ids...)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return filterContributedQuestions(matchedIds, jobTitle), nil
}

// filterContributedQuestions 过滤题目列表，支持岗位过滤
func filterContributedQuestions(ids []string, jobTitle string) []model.ContributedQuestion {
	seen := make(map[string]bool)
	var result []model.ContributedQuestion

	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true

		m, err := repository.HGetAll(contributedQKey(id))
		if err != nil || len(m) == 0 {
			continue
		}
		cq := &model.ContributedQuestion{}
		if err := cq.FromRedisHash(m); err != nil {
			continue
		}

		// 只返回已审核通过的题目
		if cq.Status != "approved" && cq.Status != "verified" {
			continue
		}
		// 岗位过滤（如果指定了）
		if jobTitle != "" && cq.JobTitle != jobTitle {
			continue
		}
		result = append(result, *cq)
	}

	if result == nil {
		result = []model.ContributedQuestion{}
	}
	return result
}

// ---------- 投票验证 ----------

// VoteContributedQuestion 对贡献题目投票
func VoteContributedQuestion(userID, questionID, vote string) error {
	if vote != "helpful" && vote != "useless" {
		return fmt.Errorf("vote 必须是 helpful 或 useless")
	}

	// 防止重复投票
	hasVoted, _ := repository.SIsMember(userVotesKey(questionID), userID)
	if hasVoted {
		return fmt.Errorf("您已经投过票了")
	}

	// 记录投票
	_ = repository.SAdd(userVotesKey(questionID), userID)

	// 更新计数
	key := contributedQKey(questionID)
	if vote == "helpful" {
		_, _ = repository.HIncrBy(key, "helpfulCount", 1)
		// 检查是否达到 verified 阈值（5个有用票）
		m, _ := repository.HGetAll(key)
		if m != nil {
			var cq model.ContributedQuestion
			cq.FromRedisHash(m)
			if cq.HelpfulCount >= 5 && cq.Status == "approved" {
				_ = repository.HSet(key, "status", "verified")
			}
		}
	} else {
		_, _ = repository.HIncrBy(key, "uselessCount", 1)
	}

	return nil
}

// ReportContributedQuestion 举报题目
func ReportContributedQuestion(userID, questionID, reason string) error {
	// 防止重复举报
	hasReported, _ := repository.SIsMember(userReportsKey(questionID), userID)
	if hasReported {
		return fmt.Errorf("您已经举报过此题")
	}

	_ = repository.SAdd(userReportsKey(questionID), userID)
	_, _ = repository.HIncrBy(contributedQKey(questionID), "reportCount", 1)

	// 举报超过3次，标记待人工审核
	m, _ := repository.HGetAll(contributedQKey(questionID))
	if m != nil {
		var cq model.ContributedQuestion
		cq.FromRedisHash(m)
		if cq.ReportCount >= 3 {
			_ = repository.HSet(contributedQKey(questionID), "status", "pending")
		}
	}

	return nil
}

// ---------- 用户贡献列表 ----------

// GetUserContributions 获取用户的贡献记录
func GetUserContributions(userID string) ([]model.ContributedQuestion, error) {
	ids, err := repository.ZRevRange(userContributionsKey(userID), 0, -1)
	if err != nil {
		return nil, err
	}

	var result []model.ContributedQuestion
	for _, id := range ids {
		m, err := repository.HGetAll(contributedQKey(id))
		if err != nil || len(m) == 0 {
			continue
		}
		cq := &model.ContributedQuestion{}
		if err := cq.FromRedisHash(m); err != nil {
			continue
		}
		result = append(result, *cq)
	}

	if result == nil {
		result = []model.ContributedQuestion{}
	}
	return result, nil
}

// GetContributedQuestion 获取单道题目
func GetContributedQuestion(questionID string) (*model.ContributedQuestion, error) {
	m, err := repository.HGetAll(contributedQKey(questionID))
	if err != nil || len(m) == 0 {
		return nil, fmt.Errorf("题目不存在")
	}
	cq := &model.ContributedQuestion{}
	if err := cq.FromRedisHash(m); err != nil {
		return nil, err
	}
	return cq, nil
}

// ListContributedQuestionsForReview 列出待审核题目（管理员用）
func ListContributedQuestionsForReview(status string, page, pageSize int) ([]model.ContributedQuestion, int, error) {
	// 从 Redis 中遍历所有贡献题目（这里简化为全量扫描，在生产环境应加索引）
	// 实际项目中应该维护一个 pending 队列或 ZSet
	keys, _ := repository.SMembers("contributed:all:ids")
	if keys == nil {
		keys = []string{}
	}

	var all []model.ContributedQuestion
	for _, id := range keys {
		m, err := repository.HGetAll(contributedQKey(id))
		if err != nil || len(m) == 0 {
			continue
		}
		cq := &model.ContributedQuestion{}
		if err := cq.FromRedisHash(m); err != nil {
			continue
		}
		if status != "" && cq.Status != status {
			continue
		}
		all = append(all, *cq)
	}

	total := len(all)
	start := (page - 1) * pageSize
	if start >= total {
		return []model.ContributedQuestion{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

// ApproveContributedQuestion 管理员审核通过题目
func ApproveContributedQuestion(questionID string) error {
	key := contributedQKey(questionID)
	_ = repository.HSet(key, "status", "approved")
	_ = repository.HSet(key, "approvedAt", time.Now().Unix())
	// 将 ID 加入全局集合（用于管理员列表）
	_ = repository.SAdd("contributed:all:ids", questionID)
	return nil
}

// RejectContributedQuestion 管理员驳回题目
func RejectContributedQuestion(questionID string) error {
	return repository.HSet(contributedQKey(questionID), "status", "rejected")
}

// ---------- 积分互惠机制 ----------

const (
	// CreditRuleSubmit 提交题目奖励积分
	CreditRuleSubmit = 2
	// CreditRuleVerified 题目被审核通过奖励
	CreditRuleVerified = 3
	// CreditRuleUsed 题目被他人使用（面试时调用）奖励
	CreditRuleUsed = 1
	// CreditCostPerQuestion 每使用一次优质题库消耗积分
	CreditCostPerQuestion = 1
	// CreditInitial 免费初始积分（新用户）
	CreditInitial = 10
)

// AwardCredit 奖励用户积分
func AwardCredit(userID string, amount int) error {
	key := userCreditsKey(userID)

	// 检查是否已有积分记录
	exists, err := repository.Exists(key)
	if err != nil {
		return fmt.Errorf("检查积分记录失败: %w", err)
	}

	if !exists {
		// 新用户首次：初始化为初始积分（不含本次奖励）
		// 奖励会在后面通过 HIncrBy 追加，确保记录存在
		_ = repository.HSet(key, "balance", CreditInitial)
		_ = repository.Persist(key)
	}

	// 增加积分（包含本次奖励）
	_, err = repository.HIncrBy(key, "balance", int64(amount))
	return err
}

// GetUserCredits 获取用户积分
func GetUserCredits(userID string) (int, error) {
	key := userCreditsKey(userID)

	// 检查 key 是否存在
	exists, err := repository.Exists(key)
	if err != nil {
		return 0, err
	}
	if !exists {
		// 用户从未有过积分，返回初始积分
		return CreditInitial, nil
	}

	// 获取积分余额
	v, err := repository.HGet(key, "balance")
	if err != nil {
		return 0, fmt.Errorf("获取积分失败: %w", err)
	}
	if v == "" {
		// 异常情况：key 存在但 balance 字段为空，返回 0
		return 0, nil
	}

	var balance int
	fmt.Sscanf(v, "%d", &balance)
	return balance, nil
}

// DeductCredit 扣除积分（查看优质题目时）
func DeductCredit(userID string, amount int) error {
	balance, err := GetUserCredits(userID)
	if err != nil {
		return err
	}
	if balance < amount {
		return fmt.Errorf("积分不足，当前剩余 %d，需要 %d", balance, amount)
	}
	key := userCreditsKey(userID)

	// 确保 key 存在
	exists, _ := repository.Exists(key)
	if !exists {
		// 如果 key 不存在，先初始化
		_ = repository.HSet(key, "balance", CreditInitial-amount)
		_ = repository.Persist(key)
		return nil
	}

	// 用 HIncrBy 扣减
	_, err = repository.HIncrBy(key, "balance", int64(-amount))
	return err
}

// GetVerifiedQuestions 获取优质已验证题目（需要消耗积分或贡献过题目）
func GetVerifiedQuestions(company, jobTitle string, page, pageSize int) ([]model.ContributedQuestion, int, error) {
	all, err := GetCompanyContributedQuestions(company, jobTitle)
	if err != nil {
		return nil, 0, err
	}

	// 只返回 verified 状态的（高质量题目）
	var verified []model.ContributedQuestion
	for _, q := range all {
		if q.Status == "verified" {
			verified = append(verified, q)
		}
	}

	total := len(verified)
	start := (page - 1) * pageSize
	if start >= total {
		return []model.ContributedQuestion{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return verified[start:end], total, nil
}

// OnQuestionUsed 题目被调用时触发（增加使用计数，给贡献者奖励积分）
func OnQuestionUsed(questionID string) error {
	_, _ = repository.HIncrBy(contributedQKey(questionID), "useCount", 1)

	// 获取题目信息，给贡献者加积分
	cq, err := GetContributedQuestion(questionID)
	if err != nil || cq.ContributorID == "" {
		return nil
	}
	if cq.Status == "verified" || cq.Status == "approved" {
		_ = AwardCredit(cq.ContributorID, CreditRuleUsed)
	}
	return nil
}

// ---------- 搜索公司题目的降级兜底 ----------

// SearchQuestionsFallback 搜不到公司时的降级方案
func SearchQuestionsFallback(company, jobTitle string) (map[string]interface{}, error) {
	// 1. 尝试从全量贡献题库中搜索相关公司名（模糊匹配）
	// 2. 搜索类似岗位的题目
	// 3. 最终兜底：AI 基于岗位生成题目

	// 先从贡献题库中找近似公司
	allKeys, _ := repository.SMembers("contributed:all:ids")
	var similar []model.ContributedQuestion
	for _, id := range allKeys {
		m, _ := repository.HGetAll(contributedQKey(id))
		if m == nil || len(m) == 0 {
			continue
		}
		cq := &model.ContributedQuestion{}
		cq.FromRedisHash(m)
		if cq.Status != "approved" && cq.Status != "verified" {
			continue
		}
		// 简单模糊：包含公司名关键词
		if strings.Contains(cq.Company, company) || strings.Contains(company, cq.Company) {
			similar = append(similar, *cq)
		}
	}

	result := map[string]interface{}{
		"found":       false,
		"company":     company,
		"jobTitle":    jobTitle,
		"similar":     similar,
		"aiGenerated": false,
		"message":     fmt.Sprintf("未找到【%s】的专属题库，以下是相似公司或岗位的题目", company),
	}

	if len(similar) == 0 {
		// 完全没有，用 AI 生成岗位通用题目
		result["message"] = fmt.Sprintf("未找到【%s】的专属题库，已为您生成该岗位的通用面试题", company)
		result["aiGenerated"] = true
	}

	return result, nil
}
