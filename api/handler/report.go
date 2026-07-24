package handler

import (
	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

func GetReport(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("interviewId")
	report, err := service.GetReport(userID, interviewID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, report)
}

func CreateShare(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("interviewId")
	token, err := service.CreateShareLink(userID, interviewID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"shareToken": token, "shareUrl": "/report/share/" + token, "expiresIn": "7天"})
}

func GetSharedReport(c *gin.Context) {
	token := c.Param("token")
	report, err := service.GetSharedReport(token)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, report)
}
