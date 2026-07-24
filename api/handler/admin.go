package handler

import (
	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

// AdminListUsers GET /api/v1/admin/users
func AdminListUsers(c *gin.Context) {
	result, err := service.AdminListUsers()
	if err != nil {
		response.InternalError(c, "获取用户列表失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

// AdminGetUser GET /api/v1/admin/users/:id
func AdminGetUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := service.AdminGetUser(userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

// AdminResetPassword PUT /api/v1/admin/users/:id/password
func AdminResetPassword(c *gin.Context) {
	targetUserID := c.Param("id")
	var req service.AdminResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminResetPassword(targetUserID, &req); err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "密码重置成功", nil)
}

// AdminSetRole PUT /api/v1/admin/users/:id/role
func AdminSetRole(c *gin.Context) {
	targetUserID := c.Param("id")
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := service.AdminSetRole(targetUserID, body.Role); err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessMsg(c, "角色更新成功", nil)
}

// AdminDeleteUser DELETE /api/v1/admin/users/:id
func AdminDeleteUser(c *gin.Context) {
	targetUserID := c.Param("id")
	operatorID, _ := c.Get("userId")
	if err := service.AdminDeleteUser(operatorID.(string), targetUserID); err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessMsg(c, "用户已删除", nil)
}

// AdminMigrateUsers POST /api/v1/admin/migrate-users
// 一次性迁移接口：把 Redis 中的老用户补录进 users:all，用完可以不再调用
func AdminMigrateUsers(c *gin.Context) {
	count, err := service.AdminMigrateUsers()
	if err != nil {
		response.InternalError(c, "迁移失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "迁移完成", gin.H{"migratedCount": count})
}
