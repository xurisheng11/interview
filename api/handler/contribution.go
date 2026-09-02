package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

// SubmitContributedQuestion POST /api/v1/contributions/questions
func SubmitContributedQuestion(c *gin.Context) {
	userID := c.GetString("userId")

	var req service.ContributeQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 过滤空题目
	var validQs []service.ContributeQuestionItem
	for _, q := range req.Questions {
		if strings.TrimSpace(q.Content) != "" {
			validQs = append(validQs, q)
		}
	}
	if len(validQs) == 0 {
		response.BadRequest(c, "至少需要提交1道有效题目")
		return
	}
	req.Questions = validQs

	q, err := service.SubmitContributedQuestion(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 给贡献者加积分（提交即奖励）
	_ = service.AwardCredit(userID, service.CreditRuleSubmit)

	// 重新获取最新积分
	credits, _ := service.GetUserCredits(userID)

	response.Success(c, gin.H{
		"question": q,
		"message":  "提交成功！感谢您的贡献，每道题奖励 2 积分",
		"credits": credits,
	})
}

// GetCompanyContributedQuestions GET /api/v1/contributions/company/:company/questions
func GetCompanyContributedQuestions(c *gin.Context) {
	company := c.Param("company")
	jobTitle := c.Query("jobTitle")

	if company == "" {
		response.BadRequest(c, "公司名称不能为空")
		return
	}

	// 检查用户积分是否足够
	userID := c.GetString("userId")
	credits, _ := service.GetUserCredits(userID)

	// verified 题目需要积分，approved 可以免费看（降级体验）
	questions, err := service.GetCompanyContributedQuestions(company, jobTitle)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 分离 verified（高质量）和 approved（普通）题目
	var verified, approved []map[string]interface{}
	for _, q := range questions {
		item := map[string]interface{}{
			"id":             q.ID,
			"company":        q.Company,
			"jobTitle":       q.JobTitle,
			"content":        q.Content,
			"questionType":   q.QuestionType,
			"difficulty":     q.Difficulty,
			"tags":           q.Tags,
			"year":           q.Year,
			"round":          q.Round,
			"helpfulCount":   q.HelpfulCount,
			"useCount":       q.UseCount,
			"createdAt":      q.CreatedAt,
		}
		if q.Status == "verified" {
			verified = append(verified, item)
		} else {
			approved = append(approved, item)
		}
	}

	response.Success(c, gin.H{
		"verified": verified,
		"approved": approved,
		"credits":  credits,
		"total":    len(verified) + len(approved),
	})
}

// GetUserContributions GET /api/v1/contributions/me
func GetUserContributions(c *gin.Context) {
	userID := c.GetString("userId")
	contributions, err := service.GetUserContributions(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	credits, _ := service.GetUserCredits(userID)
	response.Success(c, gin.H{
		"contributions": contributions,
		"credits":       credits,
	})
}

// GetContributedQuestion GET /api/v1/contributions/questions/:id
func GetContributedQuestion(c *gin.Context) {
	questionID := c.Param("id")
	q, err := service.GetContributedQuestion(questionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	// 标记使用次数（面试时调用题目，增加贡献者积分）
	_ = service.OnQuestionUsed(questionID)

	response.Success(c, q)
}

// VoteContributedQuestion POST /api/v1/contributions/questions/:id/vote
func VoteContributedQuestion(c *gin.Context) {
	userID := c.GetString("userId")
	questionID := c.Param("id")

	var body struct {
		Vote string `json:"vote" binding:"required"` // "helpful" | "useless"
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := service.VoteContributedQuestion(userID, questionID, body.Vote); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 投票即给投票者加积分（鼓励参与验证）
	_ = service.AwardCredit(userID, 1)

	credits, _ := service.GetUserCredits(userID)
	response.Success(c, gin.H{"message": "投票成功", "credits": credits})
}

// ReportContributedQuestion POST /api/v1/contributions/questions/:id/report
func ReportContributedQuestion(c *gin.Context) {
	userID := c.GetString("userId")
	questionID := c.Param("id")

	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := service.ReportContributedQuestion(userID, questionID, body.Reason); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "举报已收到，感谢反馈"})
}

// ListContributedQuestionsForReview GET /api/v1/admin/contributions (管理员)
func ListContributedQuestionsForReview(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := service.ListContributedQuestionsForReview(status, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
	})
}

// ApproveContributedQuestion PUT /api/v1/admin/contributions/:id/approve
func ApproveContributedQuestion(c *gin.Context) {
	questionID := c.Param("id")

	if err := service.ApproveContributedQuestion(questionID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 审核通过后，给贡献者额外奖励
	cq, _ := service.GetContributedQuestion(questionID)
	if cq != nil {
		_ = service.AwardCredit(cq.ContributorID, service.CreditRuleVerified)
	}

	response.Success(c, gin.H{"message": "审核通过"})
}

// RejectContributedQuestion PUT /api/v1/admin/contributions/:id/reject
func RejectContributedQuestion(c *gin.Context) {
	questionID := c.Param("id")

	if err := service.RejectContributedQuestion(questionID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "已驳回"})
}

// GetUserCredits GET /api/v1/contributions/credits
func GetUserCredits(c *gin.Context) {
	userID := c.GetString("userId")
	credits, err := service.GetUserCredits(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"credits": credits})
}

// SearchQuestionsFallback GET /api/v1/contributions/search
func SearchQuestionsFallback(c *gin.Context) {
	company := c.Query("company")
	jobTitle := c.Query("jobTitle")

	if company == "" {
		response.BadRequest(c, "公司名称不能为空")
		return
	}

	result, err := service.SearchQuestionsFallback(company, jobTitle)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}
