package model

import (
	"encoding/json"
	"fmt"
	"strconv"
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
