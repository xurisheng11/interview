package handler

import (
	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/repository"
	"interview-sim/service"
)

func Register(c *gin.Context) {
	var req service.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := service.Register(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessMsg(c, "注册成功", result)
}

func Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := service.Login(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func GetMe(c *gin.Context) {
	userID, _ := c.Get("userId")
	user, err := repository.GetUserByID(userID.(string))
	if err != nil || user == nil {
		response.NotFound(c, "用户不存在")
		return
	}
	response.Success(c, user.ToDTO())
}
