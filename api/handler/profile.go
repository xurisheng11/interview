package handler

import (
	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

// GetProfile GET /api/v1/profile
func GetProfile(c *gin.Context) {
	userId := c.GetString("userId")
	profile, err := service.GetProfile(userId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, profile)
}

// UpdateProfile PUT /api/v1/profile
func UpdateProfile(c *gin.Context) {
	userId := c.GetString("userId")

	var body struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Bio      string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := service.UpdateProfile(userId, body.Nickname, body.Avatar, body.Bio); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "更新成功"})
}

// ChangePassword PUT /api/v1/profile/password
func ChangePassword(c *gin.Context) {
	userId := c.GetString("userId")

	var body struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(body.NewPassword) < 8 {
		response.BadRequest(c, "新密码长度不能少于8位")
		return
	}

	if err := service.ChangePassword(userId, body.OldPassword, body.NewPassword); err != nil {
		if err.Error() == "原密码错误" {
			response.BadRequest(c, "原密码错误")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "密码修改成功"})
}

// GetStats GET /api/v1/profile/stats
func GetStats(c *gin.Context) {
	userId := c.GetString("userId")
	stats, err := service.GetStats(userId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

// GetScoreTrend GET /api/v1/profile/trend
func GetScoreTrend(c *gin.Context) {
	userId := c.GetString("userId")
	trend, err := service.GetScoreTrend(userId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, trend)
}

// GetCollections GET /api/v1/profile/collections
func GetCollections(c *gin.Context) {
	userId := c.GetString("userId")
	collections, err := service.GetCollections(userId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, collections)
}
