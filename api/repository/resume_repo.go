package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"interview-sim/model"

	"github.com/go-redis/redis/v8"
)

const (
	resumeKeyPrefix      = "resume:"
	userResumesKeyPrefix = "user:resumes:"
)

func resumeKey(id string) string          { return resumeKeyPrefix + id }
func userResumesKey(userID string) string { return userResumesKeyPrefix + userID }

// SaveResume 保存简历记录（永久）
func SaveResume(r *model.ResumeRecord) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if err := SetPermanent(resumeKey(r.ID), string(b)); err != nil {
		return err
	}
	// 写入用户简历有序集合，score = 上传时间戳
	return ZAdd(userResumesKey(r.UserID), float64(r.UploadedAt.Unix()), r.ID)
}

// GetResume 获取单条简历记录（Redis 未命中时回源 MySQL 并回填）
func GetResume(id string) (*model.ResumeRecord, error) {
	raw, err := Get(resumeKey(id))
	if err == redis.Nil {
		// Redis 未命中，尝试 MySQL 回源；失败保持原有 not found 语义
		if r := loadResumeFromMySQL(id); r != nil {
			return r, nil
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var r model.ResumeRecord
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// loadResumeFromMySQL 从 MySQL 加载简历并回填 Redis（resume:{id} + user:resumes:{userId} zset）
func loadResumeFromMySQL(id string) *model.ResumeRecord {
	if !MySQLAvailable() {
		return nil
	}
	data, err := QueryResumeData(id)
	if err != nil {
		log.Printf("loadResumeFromMySQL 查询失败: %v", err)
		return nil
	}
	if data == "" {
		return nil
	}
	var r model.ResumeRecord
	if err := json.Unmarshal([]byte(data), &r); err != nil || r.ID == "" {
		log.Printf("loadResumeFromMySQL 解析失败: %v", err)
		return nil
	}
	if err := SetPermanent(resumeKey(r.ID), data); err != nil {
		log.Printf("loadResumeFromMySQL 回填简历失败: %v", err)
	}
	if r.UserID != "" {
		if err := ZAdd(userResumesKey(r.UserID), float64(r.UploadedAt.Unix()), r.ID); err != nil {
			log.Printf("loadResumeFromMySQL 回填 user:resumes 失败: %v", err)
		}
	}
	log.Printf("MySQL 回源简历成功: %s", id)
	return &r
}

// UpdateResumeAnalysis 更新简历分析状态和结果
func UpdateResumeAnalysis(id, status string, analysis *model.ResumeAnalysis) error {
	r, err := GetResume(id)
	if err != nil || r == nil {
		return fmt.Errorf("简历不存在: %s", id)
	}
	r.AnalysisStatus = status
	if analysis != nil {
		r.Analysis = analysis
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return SetPermanent(resumeKey(id), string(b))
}

// DeleteResume 删除简历记录
func DeleteResume(userID, id string) error {
	if err := Del(resumeKey(id)); err != nil {
		return err
	}
	return ZRem(userResumesKey(userID), id)
}

// GetUserResumeIDs 获取用户简历 ID 列表（最新50条，时间倒序；zset 为空时从 MySQL 重建）
func GetUserResumeIDs(userID string) ([]string, error) {
	ids, err := ZRevRange(userResumesKey(userID), 0, 49)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 && MySQLAvailable() {
		// 从 MySQL 按 user_id 查最近 50 条重建 zset
		mids, times, qerr := QueryUserResumeIDs(userID, 50)
		if qerr != nil {
			log.Printf("GetUserResumeIDs MySQL 回源失败: %v", qerr)
			return ids, nil
		}
		for i, id := range mids {
			if err := ZAdd(userResumesKey(userID), float64(times[i].Unix()), id); err != nil {
				log.Printf("GetUserResumeIDs 回填 zset 失败: %v", err)
			}
		}
		if len(mids) > 0 {
			log.Printf("MySQL 回源用户简历索引成功: %s，共 %d 条", userID, len(mids))
		}
		return mids, nil
	}
	return ids, nil
}

// AddResumeInterview 记录简历面试索引（同时加入用户面试列表）
func AddResumeInterview(userID, interviewID string) error {
	score := float64(time.Now().Unix())
	return ZAdd(userInterviewsKey(userID), score, interviewID)
}
