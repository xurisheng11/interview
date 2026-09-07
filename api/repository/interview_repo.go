package repository

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"interview-sim/model"

	"github.com/go-redis/redis/v8"
)

const (
	sessionKeyPrefix     = "interview:session:"
	questionsCachePrefix = "interview:questions:"
	userInterviewsPrefix = "user:interviews:"
)

func sessionKey(interviewID string) string     { return sessionKeyPrefix + interviewID }
func questionsCacheKey(cacheKey string) string { return questionsCachePrefix + cacheKey }
func userInterviewsKey(userID string) string   { return userInterviewsPrefix + userID }

// SaveSession 保存面试会话（7天TTL）
func SaveSession(session *model.InterviewSession) error {
	jsonStr, err := session.ToJSON()
	if err != nil {
		return err
	}
	return Set(sessionKey(session.InterviewID), jsonStr, 7*24*time.Hour)
}

// GetSession 获取面试会话（Redis 未命中时回源 MySQL 并回填）
func GetSession(interviewID string) (*model.InterviewSession, error) {
	raw, err := Get(sessionKey(interviewID))
	if err == redis.Nil {
		// Redis 未命中，尝试 MySQL 回源；失败保持原有 not found 语义
		if session := loadSessionFromMySQL(interviewID); session != nil {
			return session, nil
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return model.SessionFromJSON(raw)
}

// loadSessionFromMySQL 从 MySQL 加载会话并回填 Redis
// 已完成会话永久保存，其余按 7 天 TTL（与 SaveSession/CompleteSession 语义一致）
func loadSessionFromMySQL(interviewID string) *model.InterviewSession {
	if !MySQLAvailable() {
		return nil
	}
	data, err := QuerySessionData(interviewID)
	if err != nil {
		log.Printf("loadSessionFromMySQL 查询失败: %v", err)
		return nil
	}
	if data == "" {
		return nil
	}
	session, err := model.SessionFromJSON(data)
	if err != nil || session.InterviewID == "" {
		log.Printf("loadSessionFromMySQL 解析失败: %v", err)
		return nil
	}
	if session.Status == "completed" {
		if err := SetPermanent(sessionKey(interviewID), data); err != nil {
			log.Printf("loadSessionFromMySQL 回填会话失败: %v", err)
		}
	} else {
		if err := Set(sessionKey(interviewID), data, 7*24*time.Hour); err != nil {
			log.Printf("loadSessionFromMySQL 回填会话失败: %v", err)
		}
	}
	// 回填用户面试历史索引 zset
	if session.UserID != "" {
		if err := ZAdd(userInterviewsKey(session.UserID), float64(session.StartTime.Unix()), session.InterviewID); err != nil {
			log.Printf("loadSessionFromMySQL 回填 user:interviews 失败: %v", err)
		}
	}
	log.Printf("MySQL 回源面试会话成功: %s", interviewID)
	return session
}

// UpdateSessionStatus 更新会话状态和当前索引
func UpdateSessionStatus(interviewID, status string, currentIndex int) error {
	session, err := GetSession(interviewID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found: %s", interviewID)
	}
	session.Status = status
	session.CurrentIndex = currentIndex
	return SaveSession(session)
}

// UpdateSessionAnswer 更新答案记录
func UpdateSessionAnswer(interviewID string, index int, answer *model.AnswerRecord) error {
	session, err := GetSession(interviewID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found: %s", interviewID)
	}
	if session.Answers == nil {
		session.Answers = make(map[int]*model.AnswerRecord)
	}
	session.Answers[index] = answer
	session.CurrentIndex = index + 1
	return SaveSession(session)
}

// PauseSession 暂停面试（重置TTL为7天）
func PauseSession(interviewID string) error {
	session, err := GetSession(interviewID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found: %s", interviewID)
	}
	now := time.Now()
	session.Status = "paused"
	session.PauseTime = &now
	return SaveSession(session)
}

// CompleteSession 完成面试
func CompleteSession(interviewID string) error {
	session, err := GetSession(interviewID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found: %s", interviewID)
	}
	session.Status = "completed"
	// 完成后永久保存（移除过期）
	jsonStr, err := session.ToJSON()
	if err != nil {
		return err
	}
	if err := SetPermanent(sessionKey(interviewID), jsonStr); err != nil {
		return err
	}
	return nil
}

// AddUserInterview 记录用户面试索引（ZSet，score=时间戳）
func AddUserInterview(userID, interviewID string) error {
	score := float64(time.Now().Unix())
	return ZAdd(userInterviewsKey(userID), score, interviewID)
}

// GetUserInterviewIDs 获取用户所有面试ID（最近50条；zset 为空时从 MySQL 重建）
func GetUserInterviewIDs(userID string) ([]string, error) {
	ids, err := ZRevRange(userInterviewsKey(userID), 0, 49)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 && MySQLAvailable() {
		// 从 MySQL 按 user_id 查最近 50 条重建 zset
		mids, times, qerr := QueryUserSessionIDs(userID, 50)
		if qerr != nil {
			log.Printf("GetUserInterviewIDs MySQL 回源失败: %v", qerr)
			return ids, nil
		}
		for i, id := range mids {
			if err := ZAdd(userInterviewsKey(userID), float64(times[i].Unix()), id); err != nil {
				log.Printf("GetUserInterviewIDs 回填 zset 失败: %v", err)
			}
		}
		if len(mids) > 0 {
			log.Printf("MySQL 回源用户面试索引成功: %s，共 %d 条", userID, len(mids))
		}
		return mids, nil
	}
	return ids, nil
}

// GetQuestionsCache 获取题目缓存
func GetQuestionsCache(cfg *model.InterviewConfig) ([]model.Question, error) {
	key := buildQuestionsCacheKey(cfg)
	raw, err := Get(questionsCacheKey(key))
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var questions []model.Question
	if err := json.Unmarshal([]byte(raw), &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

// SetQuestionsCache 缓存题目（24h）
func SetQuestionsCache(cfg *model.InterviewConfig, questions []model.Question) error {
	key := buildQuestionsCacheKey(cfg)
	b, err := json.Marshal(questions)
	if err != nil {
		return err
	}
	return Set(questionsCacheKey(key), string(b), 24*time.Hour)
}

func buildQuestionsCacheKey(cfg *model.InterviewConfig) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%v", cfg.JobTitle, cfg.Difficulty, cfg.Experience, cfg.Round, cfg.FocusAreas)
	return fmt.Sprintf("%x", md5.Sum([]byte(raw)))
}
