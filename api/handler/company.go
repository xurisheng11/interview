package handler

import (
	"github.com/gin-gonic/gin"
	"interview-sim/pkg/response"
	"interview-sim/service"
)

// GetCompanyIntel 获取公司面试情报
func GetCompanyIntel(c *gin.Context) {
	company := c.Query("company")
	jobTitle := c.Query("jobTitle")

	if company == "" {
		response.BadRequest(c, "请输入公司名称")
		return
	}

	result, err := service.GetCompanyIntel(company, jobTitle)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetCompanyQuestionAnswer 获取单道题目的参考答案
func GetCompanyQuestionAnswer(c *gin.Context) {
	company := c.Query("company")
	jobTitle := c.Query("jobTitle")
	content := c.Query("content")

	if company == "" || content == "" {
		response.BadRequest(c, "参数缺失")
		return
	}

	answer, err := service.GetCompanyQuestionAnswer(company, jobTitle, content)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"answer": answer})
}
