package handler

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"interview-sim/pkg/response"
	"interview-sim/service"
)

// Transcribe 语音转文字：小程序录音结束后上传音频，腾讯云一句话识别转写为文本
func Transcribe(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "缺少音频文件（表单字段名需为 file）")
		return
	}
	if file.Size > 5<<20 {
		response.BadRequest(c, "音频超过 5MB 限制")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.BadRequest(c, "音频文件读取失败")
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil || len(data) == 0 {
		response.BadRequest(c, "音频内容为空")
		return
	}

	text, err := service.TranscribeAudio(data, filepath.Ext(file.Filename))
	if err != nil {
		if strings.Contains(err.Error(), "未配置") {
			response.Fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"text": text})
}
