package handler

import (
	"encoding/base64"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"interview-sim/pkg/response"
	"interview-sim/service"
)

// UploadFaceFrame 视频面试情绪抓帧：小程序每题自动拍一张照片上传，
// 后端调腾讯云人脸属性识别情绪/微笑，结果按题目存进会话（识别失败静默返回，不阻断面试）
func UploadFaceFrame(c *gin.Context) {
	userID := c.GetString("userId")
	interviewID := c.Param("id")

	index, err := strconv.Atoi(c.PostForm("questionIndex"))
	if err != nil || index < 0 {
		response.BadRequest(c, "questionIndex 参数错误")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "缺少图片文件（表单字段名需为 file）")
		return
	}
	if file.Size > 2<<20 {
		response.BadRequest(c, "图片超过 2MB 限制")
		return
	}
	src, err := file.Open()
	if err != nil {
		response.BadRequest(c, "图片读取失败")
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		response.BadRequest(c, "图片读取失败")
		return
	}

	info, err := service.SaveFaceFrame(userID, interviewID, index, base64.StdEncoding.EncodeToString(data))
	if err != nil {
		// 抓帧分析失败对面试无影响，返回 200 + 空结果，前端静默处理
		response.Success(c, gin.H{"skipped": true, "reason": err.Error()})
		return
	}
	response.Success(c, info)
}
