package model

import "time"

// ResumeRecord 简历记录（存 MongoDB）
type ResumeRecord struct {
	ID             string              `json:"id" bson:"_id"`
	UserID         string              `json:"userId" bson:"userId"`
	Filename       string              `json:"filename" bson:"filename"`
	FileSize       int64               `json:"fileSize" bson:"fileSize"`
	MIMEType       string              `json:"mimeType" bson:"mimeType"`
	UploadedAt     time.Time           `json:"uploadedAt" bson:"uploadedAt"`
	ParsedContent  ResumeContent       `json:"parsedContent" bson:"parsedContent"`
	AnalysisStatus string              `json:"analysisStatus" bson:"analysisStatus"` // pending/analyzing/done/failed
	Analysis       *ResumeAnalysis     `json:"analysis,omitempty" bson:"analysis,omitempty"`
}

// ResumeContent 解析出的简历结构化内容
type ResumeContent struct {
	RawText        string          `json:"rawText" bson:"rawText"`
	JobTitle       string          `json:"jobTitle" bson:"jobTitle"`
	WorkExperience []WorkEntry     `json:"workExperience" bson:"workExperience"`
	Projects       []ProjectEntry  `json:"projects" bson:"projects"`
	Skills         []string        `json:"skills" bson:"skills"`
}

// WorkEntry 工作经历条目
type WorkEntry struct {
	Company  string `json:"company" bson:"company"`
	Position string `json:"position" bson:"position"`
	Duration string `json:"duration" bson:"duration"`
	Desc     string `json:"desc" bson:"desc"`
}

// ProjectEntry 项目经历条目
type ProjectEntry struct {
	Name  string `json:"name" bson:"name"`
	Role  string `json:"role" bson:"role"`
	Stack string `json:"stack" bson:"stack"`
	Desc  string `json:"desc" bson:"desc"`
}

// ResumeAnalysis AI 分析结果
type ResumeAnalysis struct {
	TotalScore  int                  `json:"totalScore" bson:"totalScore"`
	Dimensions  []ResumeDimension    `json:"dimensions" bson:"dimensions"`
	Suggestions []string             `json:"suggestions" bson:"suggestions"`
	AnalyzedAt  time.Time            `json:"analyzedAt" bson:"analyzedAt"`
}

// ResumeDimension 评分维度
type ResumeDimension struct {
	Name  string `json:"name" bson:"name"`
	Score int    `json:"score" bson:"score"`
}

// ResumeListItem 列表展示用
type ResumeListItem struct {
	ID             string    `json:"id"`
	Filename       string    `json:"filename"`
	UploadedAt     time.Time `json:"uploadedAt"`
	AnalysisStatus string    `json:"analysisStatus"`
	TotalScore     *int      `json:"totalScore,omitempty"`
}
