package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"interview-sim/model"
	"interview-sim/repository"
)

// ---------- key helpers ----------

func questionKey(questionID string) string {
	return "question:" + questionID
}

func questionIndexKey(jobTitle, difficulty, qType string) string {
	return fmt.Sprintf("questions:index:%s:%s:%s", jobTitle, difficulty, qType)
}

func userCollectKey(userID string) string {
	return "user:collect:question:" + userID
}

// fillDefault 将空字符串替换为 "通用"，用于 prompt / key 构建
func fillDefault(s string) string {
	if s == "" {
		return "通用"
	}
	return s
}

// ---------- 题目读取 ----------

// getQuestionByID 从 Redis Hash 读取单道题目
func getQuestionByID(questionID string) (*model.QuestionItem, error) {
	m, err := repository.HGetAll(questionKey(questionID))
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}
	q := &model.QuestionItem{}
	if err := q.FromRedisHash(m); err != nil {
		return nil, err
	}
	return q, nil
}

// ---------- 公开接口 ----------

// ListQuestions 列出题目（支持缓存、AI 生成、keyword 过滤和分页）
func ListQuestions(jobTitle, difficulty, questionType, keyword string, page, pageSize int) ([]model.QuestionItem, int, error) {
	jt := fillDefault(jobTitle)
	diff := fillDefault(difficulty)
	qt := fillDefault(questionType)

	indexKey := questionIndexKey(jt, diff, qt)

	// 1. 从 ZSet 取所有题目 ID
	ids, err := repository.ZRevRange(indexKey, 0, -1)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, 0, err
	}

	// 2. 缓存未命中 → AI 生成
	if len(ids) == 0 {
		if genErr := generateAndCacheQuestions(jt, diff, qt, indexKey); genErr != nil {
			return nil, 0, genErr
		}
		ids, err = repository.ZRevRange(indexKey, 0, -1)
		if err != nil {
			return nil, 0, err
		}
	}

	// 3. 批量读取题目 Hash
	var all []model.QuestionItem
	for _, id := range ids {
		q, err := getQuestionByID(id)
		if err != nil || q == nil {
			continue
		}
		all = append(all, *q)
	}

	// 4. keyword 过滤（本地，content 字符串包含）
	if keyword != "" {
		var filtered []model.QuestionItem
		for _, q := range all {
			if strings.Contains(q.Content, keyword) {
				filtered = append(filtered, q)
			}
		}
		all = filtered
	}

	total := len(all)

	// 5. 分页
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []model.QuestionItem{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

// generateAndCacheQuestions 调用 DeepSeek 生成 20 道题并写入 Redis
func generateAndCacheQuestions(jobTitle, difficulty, questionType, indexKey string) error {
	prompt := fmt.Sprintf(
		`你是一个技术面试题目生成器。请为【%s】岗位生成20道%s难度的面试题，题目类型为%s。
要求返回JSON数组，格式如下：
[{"questionId":"","content":"题目内容","jobTitle":"%s","difficulty":"%s","tags":["标签1","标签2"],"type":"%s","answerCount":0,"avgScore":0,"createdBy":"system","createdAt":%d}]
只返回JSON数组，不要其他内容。`,
		jobTitle, difficulty, questionType,
		jobTitle, difficulty, questionType,
		time.Now().Unix(),
	)

	raw, err := Chat(prompt)
	if err != nil {
		return fmt.Errorf("DeepSeek 调用失败: %w", err)
	}
	raw = cleanJSON(raw)

	var items []model.QuestionItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return fmt.Errorf("解析题目 JSON 失败: %w", err)
	}

	now := time.Now().Unix()
	for i := range items {
		// 生成 questionId
		if items[i].QuestionID == "" {
			items[i].QuestionID = uuid.New().String()
		}
		if items[i].CreatedAt == 0 {
			items[i].CreatedAt = now
		}

		// 写 Hash（永久）
		hash := items[i].ToRedisHash()
		if err := repository.HSetMap(questionKey(items[i].QuestionID), hash); err != nil {
			continue
		}

		// 写 ZSet 索引（score = createdAt）
		_ = repository.ZAdd(indexKey, float64(items[i].CreatedAt), items[i].QuestionID)
	}

	// 设置索引 24h TTL
	_ = repository.Expire(indexKey, 24*time.Hour)
	return nil
}

// GetQuestion 获取单道题目
func GetQuestion(questionID string) (*model.QuestionItem, error) {
	q, err := getQuestionByID(questionID)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, errors.New("题目不存在")
	}
	return q, nil
}

// PracticeQuestion 练习题目（AI 点评 + 更新统计）
func PracticeQuestion(userID, questionID, answer string) (map[string]interface{}, error) {
	q, err := GetQuestion(questionID)
	if err != nil {
		return nil, err
	}

	// 构造用于 ReviewAnswer 所需的 model.Question 及 InterviewConfig
	mq := &model.Question{
		Content: q.Content,
		Tags:    q.Tags,
	}
	cfg := &model.InterviewConfig{
		JobTitle:   q.JobTitle,
		Difficulty: q.Difficulty,
	}

	result, err := ReviewAnswer(mq, answer, cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("AI 点评失败: %w", err)
	}

	// 更新 answerCount（HIncrBy）
	newCount, _ := repository.HIncrBy(questionKey(questionID), "answerCount", 1)

	// 重新计算加权平均分：newAvg = (oldAvg*(newCount-1) + score) / newCount
	oldAvgStr, _ := repository.HGet(questionKey(questionID), "avgScore")
	var oldAvg float64
	if oldAvgStr != "" {
		fmt.Sscanf(oldAvgStr, "%f", &oldAvg)
	}
	var newAvg float64
	if newCount <= 1 {
		newAvg = float64(result.Score)
	} else {
		newAvg = (oldAvg*float64(newCount-1) + float64(result.Score)) / float64(newCount)
	}
	_ = repository.HSet(questionKey(questionID), "avgScore", newAvg)

	return map[string]interface{}{
		"score":           result.Score,
		"pros":            result.Pros,
		"cons":            result.Cons,
		"referenceAnswer": result.ReferenceAnswer,
		"answerCount":     newCount,
		"avgScore":        newAvg,
	}, nil
}

// CollectQuestion 收藏题目
func CollectQuestion(userID, questionID string) error {
	return repository.SAdd(userCollectKey(userID), questionID)
}

// UncollectQuestion 取消收藏
func UncollectQuestion(userID, questionID string) error {
	return repository.SRem(userCollectKey(userID), questionID)
}

// GetCollectedQuestions 获取用户收藏的所有题目
func GetCollectedQuestions(userID string) ([]model.QuestionItem, error) {
	ids, err := repository.SMembers(userCollectKey(userID))
	if err != nil {
		return nil, err
	}
	var result []model.QuestionItem
	for _, id := range ids {
		q, err := getQuestionByID(id)
		if err != nil || q == nil {
			continue
		}
		result = append(result, *q)
	}
	if result == nil {
		result = []model.QuestionItem{}
	}
	return result, nil
}

// CreateQuestion 管理员创建题目
func CreateQuestion(q *model.QuestionItem) error {
	if q.QuestionID == "" {
		q.QuestionID = uuid.New().String()
	}
	if q.CreatedAt == 0 {
		q.CreatedAt = time.Now().Unix()
	}

	// 写 Hash（永久）
	hash := q.ToRedisHash()
	if err := repository.HSetMap(questionKey(q.QuestionID), hash); err != nil {
		return err
	}

	// 写 ZSet 索引（score = createdAt）
	indexKey := questionIndexKey(
		fillDefault(q.JobTitle),
		fillDefault(q.Difficulty),
		fillDefault(q.Type),
	)
	return repository.ZAdd(indexKey, float64(q.CreatedAt), q.QuestionID)
}

// UpdateQuestion 管理员更新题目字段
func UpdateQuestion(questionID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	return repository.HSetMap(questionKey(questionID), updates)
}

// DeleteQuestion 管理员删除题目
func DeleteQuestion(questionID string) error {
	// 先读出题目信息以便从 ZSet 中移除
	q, err := getQuestionByID(questionID)
	if err != nil {
		return err
	}

	// 删除 Hash
	if err := repository.Del(questionKey(questionID)); err != nil {
		return err
	}

	if q == nil {
		return nil
	}

	// 从对应 ZSet 索引中移除
	indexKey := questionIndexKey(
		fillDefault(q.JobTitle),
		fillDefault(q.Difficulty),
		fillDefault(q.Type),
	)
	return repository.ZRem(indexKey, questionID)
}
