package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"interview-sim/model"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

// ListQuestions GET /api/v1/questions
func ListQuestions(c *gin.Context) {
	jobTitle := c.Query("jobTitle")
	difficulty := c.Query("difficulty")
	qType := c.Query("type")
	keyword := c.Query("keyword")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	list, total, err := service.ListQuestions(jobTitle, difficulty, qType, keyword, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// GetQuestion GET /api/v1/questions/:id
func GetQuestion(c *gin.Context) {
	questionID := c.Param("id")
	q, err := service.GetQuestion(questionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, q)
}

// PracticeQuestion POST /api/v1/questions/:id/practice（需鉴权）
func PracticeQuestion(c *gin.Context) {
	userID := c.GetString("userId")
	questionID := c.Param("id")

	var body struct {
		Answer string `json:"answer" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := service.PracticeQuestion(userID, questionID, body.Answer)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// CollectQuestion POST /api/v1/questions/:id/collect（需鉴权）
func CollectQuestion(c *gin.Context) {
	userID := c.GetString("userId")
	questionID := c.Param("id")

	if err := service.CollectQuestion(userID, questionID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "收藏成功"})
}

// UncollectQuestion DELETE /api/v1/questions/:id/collect（需鉴权）
func UncollectQuestion(c *gin.Context) {
	userID := c.GetString("userId")
	questionID := c.Param("id")

	if err := service.UncollectQuestion(userID, questionID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "取消收藏成功"})
}

// CreateQuestion POST /api/v1/questions（需鉴权，管理员）
func CreateQuestion(c *gin.Context) {
	var q model.QuestionItem
	if err := c.ShouldBindJSON(&q); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := service.CreateQuestion(&q); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, q)
}

// UpdateQuestion PUT /api/v1/questions/:id（需鉴权，管理员）
func UpdateQuestion(c *gin.Context) {
	questionID := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := service.UpdateQuestion(questionID, updates); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "更新成功"})
}

// DeleteQuestion DELETE /api/v1/questions/:id（需鉴权，管理员）
func DeleteQuestion(c *gin.Context) {
	questionID := c.Param("id")

	if err := service.DeleteQuestion(questionID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
