package handler

import (
	"fmt"
	"time"

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

// DownloadWordReport 生成并下载 Word 版面试报告（.docx）
func DownloadWordReport(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("interviewId")
	report, err := service.GetReport(userID, interviewID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	data, err := service.BuildWordReport(report)
	if err != nil {
		response.InternalError(c, "报告生成失败: "+err.Error())
		return
	}
	filename := fmt.Sprintf("interview-report-%s.docx", time.Now().Format("20060102-1504"))
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", data)
}
