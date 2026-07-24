package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Article 知识文章
type Article struct {
	ArticleID    string   `json:"articleId"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Tags         []string `json:"tags"`
	JobCategory  string   `json:"jobCategory"`
	AuthorID     string   `json:"authorId"`
	LikeCount    int      `json:"likeCount"`
	CollectCount int      `json:"collectCount"`
	CommentCount int      `json:"commentCount"`
	CreatedAt    int64    `json:"createdAt"`
	HotScoreVal  float64  `json:"hotScore"`
}

// HotScore 计算热度分数（likeCount + collectCount + commentCount）
func (a *Article) HotScore() float64 {
	return float64(a.LikeCount + a.CollectCount + a.CommentCount)
}

// ToRedisHash 序列化为 Redis Hash（tags 字段用 JSON 字符串存储）
func (a *Article) ToRedisHash() map[string]interface{} {
	tagsJSON, _ := json.Marshal(a.Tags)
	return map[string]interface{}{
		"articleId":    a.ArticleID,
		"title":        a.Title,
		"content":      a.Content,
		"tags":         string(tagsJSON),
		"jobCategory":  a.JobCategory,
		"authorId":     a.AuthorID,
		"likeCount":    a.LikeCount,
		"collectCount": a.CollectCount,
		"commentCount": a.CommentCount,
		"createdAt":    a.CreatedAt,
		"hotScore":     a.HotScoreVal,
	}
}

// FromRedisHash 从 Redis Hash 反序列化
func (a *Article) FromRedisHash(m map[string]string) error {
	a.ArticleID = m["articleId"]
	a.Title = m["title"]
	a.Content = m["content"]
	a.JobCategory = m["jobCategory"]
	a.AuthorID = m["authorId"]

	if v, ok := m["tags"]; ok && v != "" {
		if err := json.Unmarshal([]byte(v), &a.Tags); err != nil {
			return fmt.Errorf("parse tags: %w", err)
		}
	}

	if v, ok := m["likeCount"]; ok && v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("parse likeCount: %w", err)
		}
		a.LikeCount = n
	}

	if v, ok := m["collectCount"]; ok && v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("parse collectCount: %w", err)
		}
		a.CollectCount = n
	}

	if v, ok := m["commentCount"]; ok && v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("parse commentCount: %w", err)
		}
		a.CommentCount = n
	}

	if v, ok := m["createdAt"]; ok && v != "" {
		t, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("parse createdAt: %w", err)
		}
		a.CreatedAt = t
	}

	if v, ok := m["hotScore"]; ok && v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("parse hotScore: %w", err)
		}
		a.HotScoreVal = f
	}

	return nil
}

// Comment 评论
type Comment struct {
	CommentID string `json:"commentId"`
	ArticleID string `json:"articleId"`
	UserID    string `json:"userId"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
}

// ToRedisHash 序列化为 Redis Hash
func (c *Comment) ToRedisHash() map[string]interface{} {
	return map[string]interface{}{
		"commentId": c.CommentID,
		"articleId": c.ArticleID,
		"userId":    c.UserID,
		"content":   c.Content,
		"createdAt": c.CreatedAt,
	}
}

// FromRedisHash 从 Redis Hash 反序列化
func (c *Comment) FromRedisHash(m map[string]string) error {
	c.CommentID = m["commentId"]
	c.ArticleID = m["articleId"]
	c.UserID = m["userId"]
	c.Content = m["content"]

	if v, ok := m["createdAt"]; ok && v != "" {
		t, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("parse createdAt: %w", err)
		}
		c.CreatedAt = t
	}

	return nil
}
