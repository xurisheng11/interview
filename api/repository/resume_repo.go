package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"interview-sim/model"
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

// GetResume 获取单条简历记录
func GetResume(id string) (*model.ResumeRecord, error) {
	raw, err := Get(resumeKey(id))
	if err == redis.Nil {
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

// GetUserResumeIDs 获取用户简历 ID 列表（最新50条，时间倒序）
func GetUserResumeIDs(userID string) ([]string, error) {
	return ZRevRange(userResumesKey(userID), 0, 49)
}

// AddResumeInterview 记录简历面试索引（同时加入用户面试列表）
func AddResumeInterview(userID, interviewID string) error {
	score := float64(time.Now().Unix())
	return ZAdd(userInterviewsKey(userID), score, interviewID)
}
