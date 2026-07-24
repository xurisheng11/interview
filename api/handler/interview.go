package handler

import (
	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

func CreateInterview(c *gin.Context) {
	userID := c.GetString("userId")
	var req service.CreateInterviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	session, err := service.CreateInterview(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"interviewId": session.InterviewID,
		"questions":   session.Questions,
	})
}

func GetInterviewList(c *gin.Context) {
	userID := c.GetString("userId")
	list, err := service.GetInterviewList(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, list)
}

func GetInterview(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("id")
	session, err := service.GetInterviewSession(userID, interviewID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, session)
}

func SubmitAnswer(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("id")
	var req service.SubmitAnswerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := service.SubmitAnswer(userID, interviewID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func PauseInterview(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("id")
	if err := service.PauseInterview(userID, interviewID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "面试已暂停", nil)
}

func CompleteInterview(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("id")
	report, err := service.GenerateReport(userID, interviewID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if report == nil {
		response.InternalError(c, "报告生成失败")
		return
	}
	response.Success(c, gin.H{"interviewId": interviewID, "totalScore": report.TotalScore})
}
