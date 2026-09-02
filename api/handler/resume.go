package handler

import (
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"interview-sim/model"
	"interview-sim/pkg/response"
	"interview-sim/repository"
	"interview-sim/service"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

func UploadResume(c *gin.Context) {
	userID := c.GetString("userId")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的简历文件")
		return
	}
	defer file.Close()

	// 扩展名校验
	ext := strings.ToLower(header.Filename)
	if !strings.HasSuffix(ext, ".pdf") && !strings.HasSuffix(ext, ".doc") && !strings.HasSuffix(ext, ".docx") {
		response.BadRequest(c, "只支持 PDF 或 Word 文件（.pdf / .doc / .docx）")
		return
	}

	// 大小校验
	if header.Size > maxFileSize {
		response.BadRequest(c, "文件大小不能超过 10 MB")
		return
	}

	// MIME 类型校验
	mimeType := header.Header.Get("Content-Type")
	allowedMimes := map[string]bool{
		"application/pdf":                                                                 true,
		"application/msword":                                                              true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":          true,
		"application/octet-stream": true, // 部分 Word 文件会返回此类型，宽松处理
	}
	if !allowedMimes[mimeType] {
		// 允许无 Content-Type 但扩展名合法的请求
		if mimeType != "" && !strings.Contains(mimeType, "pdf") && !strings.Contains(mimeType, "word") {
			response.BadRequest(c, "文件格式不正确，仅支持 PDF 或 Word 文件")
			return
		}
	}

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		response.InternalError(c, "读取文件失败")
		return
	}

	// 提取文本
	rawText, err := service.ExtractTextFromFile(data, mimeType)
	if err != nil {
		response.BadRequest(c, "无法读取文件内容："+err.Error()+"，请确认文件可正常打开")
		return
	}

	// AI 解析结构化内容
	content, err := service.ParseResumeText(rawText)
	if err != nil {
		response.InternalError(c, "简历解析失败："+err.Error())
		return
	}

	// 保存记录
	resume := &model.ResumeRecord{
		ID:            uuid.New().String(),
		UserID:        userID,
		Filename:      header.Filename,
		FileSize:      header.Size,
		MIMEType:      mimeType,
		UploadedAt:    time.Now(),
		ParsedContent: *content,
		AnalysisStatus: "pending",
	}
	if err := repository.SaveResume(resume); err != nil {
		response.InternalError(c, "保存简历失败："+err.Error())
		return
	}

	// 异步触发 AI 分析
	service.AnalyzeResume(resume.ID)

	response.Success(c, gin.H{
		"id":             resume.ID,
		"filename":       resume.Filename,
		"analysisStatus": resume.AnalysisStatus,
	})
}

func ListResumes(c *gin.Context) {
	userID := c.GetString("userId")
	ids, err := repository.GetUserResumeIDs(userID)
	if err != nil {
		response.InternalError(c, "获取简历列表失败")
		return
	}

	items := make([]model.ResumeListItem, 0, len(ids))
	for _, id := range ids {
		r, err := repository.GetResume(id)
		if err != nil || r == nil {
			continue
		}
		item := model.ResumeListItem{
			ID:             r.ID,
			Filename:       r.Filename,
			UploadedAt:     r.UploadedAt,
			AnalysisStatus: r.AnalysisStatus,
		}
		if r.Analysis != nil {
			item.TotalScore = &r.Analysis.TotalScore
		}
		items = append(items, item)
	}
	response.Success(c, items)
}

func GetResume(c *gin.Context) {
	userID := c.GetString("userId")
	resumeID := c.Param("id")

	r, err := repository.GetResume(resumeID)
	if err != nil {
		response.InternalError(c, "获取简历失败")
		return
	}
	if r == nil {
		response.NotFound(c, "简历不存在")
		return
	}
	if r.UserID != userID {
		response.Forbidden(c, "无权访问此简历")
		return
	}
	response.Success(c, r)
}

func DeleteResume(c *gin.Context) {
	userID := c.GetString("userId")
	resumeID := c.Param("id")

	r, err := repository.GetResume(resumeID)
	if err != nil {
		response.InternalError(c, "操作失败")
		return
	}
	if r == nil {
		response.NotFound(c, "简历不存在")
		return
	}
	if r.UserID != userID {
		response.Forbidden(c, "无权删除此简历")
		return
	}

	if err := repository.DeleteResume(userID, resumeID); err != nil {
		response.InternalError(c, "删除失败："+err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功", nil)
}

func RetryAnalyzeResume(c *gin.Context) {
	userID := c.GetString("userId")
	resumeID := c.Param("id")

	r, err := repository.GetResume(resumeID)
	if err != nil || r == nil {
		response.NotFound(c, "简历不存在")
		return
	}
	if r.UserID != userID {
		response.Forbidden(c, "无权操作此简历")
		return
	}
	if r.AnalysisStatus == "failed" || r.AnalysisStatus == "pending" || r.AnalysisStatus == "analyzing" {
		// 重新触发分析
		service.AnalyzeResume(resumeID)
		response.SuccessMsg(c, "分析已重新开始", gin.H{"status": "pending"})
		return
	}
	response.BadRequest(c, "当前状态无需重新分析")
}

func CreateResumeInterview(c *gin.Context) {
	userID := c.GetString("userId")
	resumeID := c.Param("id")

	session, err := service.CreateResumeInterview(userID, resumeID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "无权") {
			response.Forbidden(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "未完成") || strings.Contains(err.Error(), "分析") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"interviewId": session.InterviewID,
		"source":      session.Source,
		"resumeId":    session.ResumeID,
	})
}
