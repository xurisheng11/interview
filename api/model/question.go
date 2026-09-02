package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// QuestionItem 题库题目
type QuestionItem struct {
	QuestionID  string   `json:"questionId"`
	Content     string   `json:"content"`
	JobTitle    string   `json:"jobTitle"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	Type        string   `json:"type"`
	AnswerCount int      `json:"answerCount"`
	AvgScore    float64  `json:"avgScore"`
	CreatedBy   string   `json:"createdBy"`
	CreatedAt   int64    `json:"createdAt"`
}

// ToRedisHash 序列化为 Redis Hash（tags 字段用 JSON 字符串存储）
func (q *QuestionItem) ToRedisHash() map[string]interface{} {
	tagsJSON, _ := json.Marshal(q.Tags)
	return map[string]interface{}{
		"questionId":  q.QuestionID,
		"content":     q.Content,
		"jobTitle":    q.JobTitle,
		"difficulty":  q.Difficulty,
		"tags":        string(tagsJSON),
		"type":        q.Type,
		"answerCount": q.AnswerCount,
		"avgScore":    q.AvgScore,
		"createdBy":   q.CreatedBy,
		"createdAt":   q.CreatedAt,
	}
}

// FromRedisHash 从 Redis Hash 反序列化
func (q *QuestionItem) FromRedisHash(m map[string]string) error {
	q.QuestionID = m["questionId"]
	q.Content = m["content"]
	q.JobTitle = m["jobTitle"]
	q.Difficulty = m["difficulty"]
	q.Type = m["type"]
	q.CreatedBy = m["createdBy"]

	if v, ok := m["tags"]; ok && v != "" {
		if err := json.Unmarshal([]byte(v), &q.Tags); err != nil {
			return fmt.Errorf("parse tags: %w", err)
		}
	}

	if v, ok := m["answerCount"]; ok && v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("parse answerCount: %w", err)
		}
		q.AnswerCount = n
	}

	if v, ok := m["avgScore"]; ok && v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("parse avgScore: %w", err)
		}
		q.AvgScore = f
	}

	if v, ok := m["createdAt"]; ok && v != "" {
		t, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("parse createdAt: %w", err)
		}
		q.CreatedAt = t
	}

	return nil
}

// ContributedQuestion 用户贡献的面试题
type ContributedQuestion struct {
	ID            string   `json:"id"`
	Company       string   `json:"company"`
	JobTitle      string   `json:"jobTitle"`
	Content       string   `json:"content"`
	QuestionType  string   `json:"questionType"` // "technical" | "behavioral"
	Difficulty    string   `json:"difficulty"`   // "easy" | "medium" | "hard"
	Tags          []string `json:"tags"`
	Year          int      `json:"year"`         // 题目年份，如 2024
	Round         string   `json:"round"`        // 面试轮次，如 "一面","二面","HR面"
	ContributorID string   `json:"contributorId"`
	Status        string   `json:"status"` // "pending" | "approved" | "rejected" | "verified"
	HelpfulCount  int      `json:"helpfulCount"`  // 觉得有用
	UselessCount  int      `json:"uselessCount"`  // 觉得没用
	UseCount      int      `json:"useCount"`      // 被其他用户使用（面试时调用）次数
	ReportCount   int      `json:"reportCount"`   // 被举报次数
	ApprovedAt    int64    `json:"approvedAt,omitempty"`
	CreatedAt     int64    `json:"createdAt"`
}

// ToRedisHash 序列化为 Redis Hash
func (c *ContributedQuestion) ToRedisHash() map[string]interface{} {
	tagsJSON, _ := json.Marshal(c.Tags)
	return map[string]interface{}{
		"id":             c.ID,
		"company":        c.Company,
		"jobTitle":       c.JobTitle,
		"content":        c.Content,
		"questionType":   c.QuestionType,
		"difficulty":     c.Difficulty,
		"tags":           string(tagsJSON),
		"year":           c.Year,
		"round":          c.Round,
		"contributorId":  c.ContributorID,
		"status":         c.Status,
		"helpfulCount":   c.HelpfulCount,
		"uselessCount":   c.UselessCount,
		"useCount":       c.UseCount,
		"reportCount":    c.ReportCount,
		"approvedAt":     c.ApprovedAt,
		"createdAt":      c.CreatedAt,
	}
}

// FromRedisHash 从 Redis Hash 反序列化
func (c *ContributedQuestion) FromRedisHash(m map[string]string) error {
	c.ID = m["id"]
	c.Company = m["company"]
	c.JobTitle = m["jobTitle"]
	c.Content = m["content"]
	c.QuestionType = m["questionType"]
	c.Difficulty = m["difficulty"]
	c.QuestionType = m["questionType"]
	c.ContributorID = m["contributorId"]
	c.Status = m["status"]
	c.Round = m["round"]

	if v, ok := m["tags"]; ok && v != "" {
		json.Unmarshal([]byte(v), &c.Tags)
	}
	if v, ok := m["year"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &c.Year)
	}
	if v, ok := m["helpfulCount"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &c.HelpfulCount)
	}
	if v, ok := m["uselessCount"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &c.UselessCount)
	}
	if v, ok := m["useCount"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &c.UseCount)
	}
	if v, ok := m["reportCount"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &c.ReportCount)
	}
	if v, ok := m["approvedAt"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &c.ApprovedAt)
	}
	if v, ok := m["createdAt"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &c.CreatedAt)
	}
	_ = time.Now()
	return nil
}

// QuestionContribution 用户贡献记录（提交题目时生成）
type QuestionContribution struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Company   string `json:"company"`
	JobTitle  string `json:"jobTitle"`
	// 题目列表（JSON 字符串）
	Questions string `json:"questions"`
	// 奖励状态
	RewardStatus string `json:"rewardStatus"` // "unrewarded" | "rewarded"
	CreditReward int    `json:"creditReward"`  // 本次奖励的积分
	CreatedAt   int64  `json:"createdAt"`
}

// ToRedisHash 序列化为 Redis Hash
func (qc *QuestionContribution) ToRedisHash() map[string]interface{} {
	return map[string]interface{}{
		"id":            qc.ID,
		"userId":        qc.UserID,
		"company":       qc.Company,
		"jobTitle":      qc.JobTitle,
		"questions":     qc.Questions,
		"rewardStatus":  qc.RewardStatus,
		"creditReward":  qc.CreditReward,
		"createdAt":     qc.CreatedAt,
	}
}

// FromRedisHash 反序列化
func (qc *QuestionContribution) FromRedisHash(m map[string]string) error {
	qc.ID = m["id"]
	qc.UserID = m["userId"]
	qc.Company = m["company"]
	qc.JobTitle = m["jobTitle"]
	qc.Questions = m["questions"]
	qc.RewardStatus = m["rewardStatus"]

	if v, ok := m["creditReward"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &qc.CreditReward)
	}
	if v, ok := m["createdAt"]; ok && v != "" {
		fmt.Sscanf(v, "%d", &qc.CreatedAt)
	}
	return nil
}
