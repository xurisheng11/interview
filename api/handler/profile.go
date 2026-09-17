package handler

import (
	"errors"
	"io"

	"interview-sim/pkg/response"
	"interview-sim/service"

	"github.com/gin-gonic/gin"
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
		Nickname       string `json:"nickname"`
		Avatar         string `json:"avatar"`
		Bio            string `json:"bio"`
		TargetPosition string `json:"targetPosition"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := service.UpdateProfile(userId, body.Nickname, body.Avatar, body.Bio, body.TargetPosition); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "更新成功"})
}

// UploadAvatar POST /api/v1/profile/avatar（multipart/form-data，字段名 file）
// 头像必须真上传：以前小程序只把手机本地临时路径 wxfile://tmp_xxx 存进资料字段，
// 图片字节从没离开过用户手机，换设备或微信清理临时目录后头像必然空白
func UploadAvatar(c *gin.Context) {
	userId := c.GetString("userId")

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择头像图片（表单字段名需为 file）")
		return
	}
	if file.Size > 2<<20 {
		response.BadRequest(c, "头像不能超过 2MB")
		return
	}
	src, err := file.Open()
	if err != nil {
		response.BadRequest(c, "头像读取失败")
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		response.InternalError(c, "头像读取失败: "+err.Error())
		return
	}

	// 带上旧头像地址是为了写完后顺手删掉，换一次头像不在桶里堆一份垃圾
	var oldAvatar string
	if profile, perr := service.GetProfile(userId); perr == nil {
		if v, ok := profile["avatar"].(string); ok {
			oldAvatar = v
		}
	}

	avatarURL, err := service.SaveAvatar(userId, data, file.Header.Get("Content-Type"), oldAvatar)
	if err != nil {
		if errors.Is(err, service.ErrAvatarStorageUnavailable) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	if err := service.UpdateProfileAvatar(userId, avatarURL); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"avatar": avatarURL})
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

// UpdateJobStatus PUT /api/v1/profile/job-status
func UpdateJobStatus(c *gin.Context) {
	userId := c.GetString("userId")

	var body struct {
		JobStatus  string `json:"jobStatus"`
		Experience string `json:"experience"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := service.UpdateJobStatus(userId, body.JobStatus, body.Experience); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "求职状态更新成功"})
}
