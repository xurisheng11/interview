package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"interview-sim/model"
	"interview-sim/repository"
)

const reportKeyPrefix = "report:"
const shareKeyPrefix = "report:share:"

func reportKey(userID, interviewID string) string {
	return reportKeyPrefix + userID + ":" + interviewID
}
func shareKey(token string) string { return shareKeyPrefix + token }

// GenerateReport 生成并保存完整面试报告
func GenerateReport(userID, interviewID string) (*model.InterviewReport, error) {
	session, err := repository.GetSession(interviewID)
	if err != nil || session == nil {
		return nil, fmt.Errorf("面试会话不存在")
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("无权操作")
	}

	// 标记会话完成
	_ = repository.CompleteSession(interviewID)

	now := time.Now()
	totalSeconds := int(now.Sub(session.StartTime).Seconds())

	// 构建题目详情
	questions := make([]model.ReportQuestion, len(session.Questions))
	var answeredCount, skippedCount int
	totalScore := 0
	moduleMap := map[string][]int{}

	for i, q := range session.Questions {
		rq := model.ReportQuestion{
			Index:            q.Index,
			Content:          q.Content,
			Tags:             q.Tags,
			Difficulty:       q.Difficulty,
			EstimatedMinutes: q.EstimatedMinutes,
			Type:             q.Type,
		}
		if ans, ok := session.Answers[i]; ok {
			rq.UserAnswer = ans.UserAnswer
			rq.Score = ans.Score
			rq.Pros = ans.Pros
			rq.Cons = ans.Cons
			rq.ReferenceAnswer = ans.ReferenceAnswer
			rq.Skipped = ans.Skipped
			if ans.Skipped {
				skippedCount++
			} else {
				answeredCount++
				totalScore += ans.Score
			}
			// 按标签归类模块得分
			for _, tag := range q.Tags {
				moduleMap[tag] = append(moduleMap[tag], ans.Score)
			}
		} else {
			rq.Skipped = true
			skippedCount++
		}
		questions[i] = rq
	}

	// 计算综合得分（全部题目参与，跳过=0）
	allCount := len(session.Questions)
	avgScore := 0
	if allCount > 0 {
		// 跳过和未作答记0分，totalScore 已是已答题目之和
		avgScore = totalScore / allCount
	}

	// 计算模块得分
	moduleScores := calcModuleScores(moduleMap)

	// 通过判定
	passStatus, passReason := model.CalcPassStatus(avgScore, allCount, skippedCount)

	// 调 DeepSeek 生成综合评价
	scoresSummary := buildScoresSummary(questions)
	aiRaw, err := GenerateReportSummary(session.Config.JobTitle, session.Config.Round, avgScore, scoresSummary)
	var aiSummary *model.AISummaryReport
	if err == nil && aiRaw != nil {
		weaknesses := make([]model.WeaknessDetail, len(aiRaw.Weaknesses))
		for i, w := range aiRaw.Weaknesses {
			weaknesses[i] = model.WeaknessDetail{
				Point:      w.Point,
				Suggestion: w.Suggestion,
				Resource:   w.Resource,
			}
		}
		aiSummary = &model.AISummaryReport{
			Strengths:  aiRaw.Strengths,
			Weaknesses: weaknesses,
			Roadmap:    aiRaw.Roadmap,
		}
	}

	report := &model.InterviewReport{
		InterviewID:   interviewID,
		UserID:        userID,
		JobTitle:      session.Config.JobTitle,
		Difficulty:    session.Config.Difficulty,
		Round:         session.Config.Round,
		Experience:    session.Config.Experience,
		TotalScore:    avgScore,
		Grade:         model.CalcGrade(avgScore),
		PassStatus:    passStatus,
		PassReason:    passReason,
		StartTime:     session.StartTime,
		EndTime:       now,
		TotalSeconds:  totalSeconds,
		AnsweredCount: answeredCount,
		SkippedCount:  skippedCount,
		TotalCount:    allCount,
		Questions:     questions,
		ModuleScores:  moduleScores,
		AISummary:     aiSummary,
		CreatedAt:     now,
		Mode:          session.Mode,
	}

	// 视频模式：生成表达能力评价并计算平均指标
	if session.Mode == "video" {
		// 收集 AnswerRecord 指针
		var answerRecords []*model.AnswerRecord
		for i := range session.Questions {
			if ans, ok := session.Answers[i]; ok {
				answerRecords = append(answerRecords, ans)
			}
		}
		report.AvgExpressionScore = model.CalcAvgExpressionScore(questions)
		report.AvgSpeechRate = model.CalcAvgSpeechRate(questions)
		// 调 DeepSeek 生成口头表达综合评价
		if summary, err := GenerateVideoExpressionSummary(session.Config.JobTitle, answerRecords); err == nil {
			report.ExpressionSummary = summary
		}
	}

	// 永久保存报告
	b, _ := json.Marshal(report)
	_ = repository.SetPermanent(reportKey(userID, interviewID), string(b))

	return report, nil
}

// GetReport 获取报告
func GetReport(userID, interviewID string) (*model.InterviewReport, error) {
	raw, err := repository.Get(reportKey(userID, interviewID))
	if err != nil {
		return nil, fmt.Errorf("报告不存在")
	}
	var report model.InterviewReport
	if err := json.Unmarshal([]byte(raw), &report); err != nil {
		return nil, err
	}
	return &report, nil
}

// CreateShareLink 生成分享链接 token（7天有效）
func CreateShareLink(userID, interviewID string) (string, error) {
	token := uuid.New().String()
	shareData := map[string]string{"userId": userID, "interviewId": interviewID}
	b, _ := json.Marshal(shareData)
	if err := repository.Set(shareKey(token), string(b), 7*24*time.Hour); err != nil {
		return "", err
	}
	return token, nil
}

// GetSharedReport 获取分享报告（脱敏）
func GetSharedReport(token string) (map[string]interface{}, error) {
	raw, err := repository.Get(shareKey(token))
	if err != nil {
		return nil, fmt.Errorf("分享链接无效或已过期")
	}
	var shareData map[string]string
	if err := json.Unmarshal([]byte(raw), &shareData); err != nil {
		return nil, err
	}
	report, err := GetReport(shareData["userId"], shareData["interviewId"])
	if err != nil {
		return nil, err
	}
	// 脱敏：隐藏用户答案
	result := map[string]interface{}{
		"interviewId":  report.InterviewID,
		"jobTitle":     report.JobTitle,
		"round":        report.Round,
		"difficulty":   report.Difficulty,
		"totalScore":   report.TotalScore,
		"grade":        report.Grade,
		"passStatus":   report.PassStatus,
		"moduleScores": report.ModuleScores,
		"aiSummary":    report.AISummary,
		"createdAt":    report.CreatedAt,
	}
	return result, nil
}

func calcModuleScores(moduleMap map[string][]int) []model.ModuleScore {
	var scores []model.ModuleScore
	for module, vals := range moduleMap {
		sum := 0
		for _, v := range vals {
			sum += v
		}
		avg := float64(sum) / float64(len(vals))
		level := "待提升"
		if avg >= 90 {
			level = "优秀"
		} else if avg >= 75 {
			level = "良好"
		} else if avg >= 60 {
			level = "及格"
		}
		scores = append(scores, model.ModuleScore{
			Module:   module,
			Count:    len(vals),
			AvgScore: avg,
			Level:    level,
		})
	}
	return scores
}

func buildScoresSummary(questions []model.ReportQuestion) string {
	var parts []string
	for _, q := range questions {
		if q.Skipped {
			parts = append(parts, fmt.Sprintf("题%d[%s]: 跳过(0分)", q.Index+1, strings.Join(q.Tags, "/")))
		} else {
			parts = append(parts, fmt.Sprintf("题%d[%s]: %d分", q.Index+1, strings.Join(q.Tags, "/"), q.Score))
		}
	}
	return strings.Join(parts, "；")
}
