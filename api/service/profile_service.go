package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"interview-sim/model"
	"interview-sim/repository"
)

var ErrOldPasswordWrong = errors.New("原密码错误")

// GetProfile 获取用户个人信息
func GetProfile(userId string) (map[string]interface{}, error) {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	return map[string]interface{}{
		"userId":    user.UserID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"avatar":    user.Avatar,
		"bio":       user.Bio,
		"email":     user.Email,
		"phone":     user.Phone,
		"createdAt": user.CreatedAt.Format(time.RFC3339),
		"role":      user.Role,
	}, nil
}

// UpdateProfile 更新用户个人信息
func UpdateProfile(userId, nickname, avatar, bio string) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	// avatar 允许为空（清除头像）
	updates["avatar"] = avatar
	updates["bio"] = bio
	return repository.HSetMap("user:"+userId, updates)
}

// ChangePassword 修改密码
func ChangePassword(userId, oldPwd, newPwd string) error {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return ErrOldPasswordWrong
	}

	// 加密新密码
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新 Redis Hash 中的 passwordHash 字段
	return repository.HSet("user:"+userId, "passwordHash", string(hash))
}

// GetStats 获取用户面试统计数据
func GetStats(userId string) (map[string]interface{}, error) {
	ids, err := repository.ZRevRange("user:interviews:"+userId, 0, -1)
	if err != nil {
		return nil, err
	}

	totalCount := 0
	var scores []int
	maxScore := 0
	jobTitleFreq := map[string]int{}

	for _, id := range ids {
		// 读取面试会话获取岗位信息
		session, err := repository.GetSession(id)
		if err != nil || session == nil {
			continue
		}

		// 读取报告以获取综合得分
		reportRaw, err := repository.Get(fmt.Sprintf("report:%s:%s", userId, id))
		if err != nil || reportRaw == "" {
			// 无报告则跳过得分统计，但仍计入总次数
			totalCount++
			if session.Config.JobTitle != "" {
				jobTitleFreq[session.Config.JobTitle]++
			}
			continue
		}

		var report model.InterviewReport
		if err := json.Unmarshal([]byte(reportRaw), &report); err != nil {
			continue
		}

		totalCount++
		score := report.TotalScore
		scores = append(scores, score)
		if score > maxScore {
			maxScore = score
		}
		if session.Config.JobTitle != "" {
			jobTitleFreq[session.Config.JobTitle]++
		}
	}

	// 计算平均分
	avgScore := 0.0
	if len(scores) > 0 {
		sum := 0
		for _, s := range scores {
			sum += s
		}
		avgScore = float64(sum) / float64(len(scores))
	}

	// 找出最常见岗位
	topJobTitle := ""
	topCount := 0
	for jt, cnt := range jobTitleFreq {
		if cnt > topCount {
			topCount = cnt
			topJobTitle = jt
		}
	}

	return map[string]interface{}{
		"totalCount":  totalCount,
		"avgScore":    avgScore,
		"maxScore":    maxScore,
		"topJobTitle": topJobTitle,
	}, nil
}

// GetScoreTrend 获取最近30次面试得分趋势
func GetScoreTrend(userId string) ([]map[string]interface{}, error) {
	ids, err := repository.ZRevRange("user:interviews:"+userId, 0, 29)
	if err != nil {
		return nil, err
	}

	var trend []map[string]interface{}

	for _, id := range ids {
		session, err := repository.GetSession(id)
		if err != nil || session == nil {
			continue
		}
		if session.Status != "completed" {
			continue
		}

		reportRaw, err := repository.Get(fmt.Sprintf("report:%s:%s", userId, id))
		if err != nil || reportRaw == "" {
			continue
		}

		var report model.InterviewReport
		if err := json.Unmarshal([]byte(reportRaw), &report); err != nil {
			continue
		}

		dateStr := time.Unix(session.StartTime.Unix(), 0).Format("2006-01-02")

		trend = append(trend, map[string]interface{}{
			"date":     dateStr,
			"score":    report.TotalScore,
			"jobTitle": session.Config.JobTitle,
		})
	}

	if trend == nil {
		trend = []map[string]interface{}{}
	}
	return trend, nil
}

// GetCollections 获取用户收藏（题目 + 文章）
func GetCollections(userId string) (map[string]interface{}, error) {
	questions, err := GetCollectedQuestions(userId)
	if err != nil {
		return nil, fmt.Errorf("获取收藏题目失败: %w", err)
	}

	articles, err := GetCollectedArticles(userId)
	if err != nil {
		return nil, fmt.Errorf("获取收藏文章失败: %w", err)
	}

	return map[string]interface{}{
		"questions": questions,
		"articles":  articles,
	}, nil
}
