package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

// ListArticles GET /api/v1/community/articles
func ListArticles(c *gin.Context) {
	jobCategory := c.Query("jobCategory")
	sortBy := c.DefaultQuery("sortBy", "time")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	list, total, err := service.ListArticles(jobCategory, sortBy, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// GetArticle GET /api/v1/community/articles/:id
func GetArticle(c *gin.Context) {
	articleID := c.Param("id")
	article, err := service.GetArticle(articleID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, article)
}

// GenerateAIArticle POST /api/v1/community/articles/ai（需鉴权）
func GenerateAIArticle(c *gin.Context) {
	userID := c.GetString("userId")

	var body struct {
		Topic       string `json:"topic" binding:"required"`
		JobCategory string `json:"jobCategory" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	article, err := service.GenerateAIArticle(userID, body.Topic, body.JobCategory)
	if err != nil {
		if err.Error() == "今日AI生成次数已用完" {
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": err.Error()})
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, article)
}

// LikeArticle POST /api/v1/community/articles/:id/like（需鉴权）
func LikeArticle(c *gin.Context) {
	userID := c.GetString("userId")
	articleID := c.Param("id")

	if err := service.LikeArticle(userID, articleID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "操作成功"})
}

// CollectArticle POST /api/v1/community/articles/:id/collect（需鉴权）
func CollectArticle(c *gin.Context) {
	userID := c.GetString("userId")
	articleID := c.Param("id")

	if err := service.CollectArticle(userID, articleID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "收藏成功"})
}

// ListComments GET /api/v1/community/articles/:id/comments（需鉴权）
func ListComments(c *gin.Context) {
	articleID := c.Param("id")

	comments, err := service.ListComments(articleID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, comments)
}

// AddComment POST /api/v1/community/articles/:id/comments（需鉴权）
func AddComment(c *gin.Context) {
	userID := c.GetString("userId")
	articleID := c.Param("id")

	var body struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	comment, err := service.AddComment(userID, articleID, body.Content)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, comment)
}
