package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"

	"interview-sim/model"
)

// TestBuildWordReport_OOXMLValid 生成的 docx 必须是合法 zip，且 document.xml 是合法 XML
func TestBuildWordReport_OOXMLValid(t *testing.T) {
	report := &model.InterviewReport{
		InterviewID:   "itw-test",
		CompanyName:   "某某科技 & Co.", // 含 & 验证转义
		JobTitle:      "后端工程师",
		Round:         "round2",
		Difficulty:    "medium",
		Mode:          "video_call",
		TotalScore:    82,
		Grade:         "良好",
		PassStatus:    "pass",
		PassReason:    "表现良好",
		StartTime:     time.Now().Add(-time.Hour),
		EndTime:       time.Now(),
		TotalSeconds:  3000,
		TotalCount:    5,
		AnsweredCount: 4,
		SkippedCount:  1,
		ModuleScores:  []model.ModuleScore{{Module: "Go", Count: 3, AvgScore: 85.4, Level: "良好"}},
		Questions: []model.ReportQuestion{
			{Content: "Q1 内容", UserAnswer: "回答 <标签>", Score: 85, Pros: []string{"优点1"}, Cons: []string{"缺点1"}, ReferenceAnswer: "参考答案", ExpressionFeedback: "语速适中"},
			{Content: "Q2 内容", Skipped: true},
		},
		AISummary: &model.AISummaryReport{
			Strengths:  []string{"基础扎实"},
			Weaknesses: []model.WeaknessDetail{{Point: "表达欠组织", Suggestion: "用 STAR 法", Resource: "书籍X"}},
			Roadmap:    "先补基础再刷项目",
		},
		FaceSummary:        &model.FaceSummary{Samples: 4, DominantName: "自然", TenseRatio: 0.25, SmileRatio: 0.5, NervousLevel: "轻度紧张"},
		AvgExpressionScore: 78,
		AvgSpeechRate:      210.5,
		ExpressionSummary:  "整体表达流畅",
	}

	data, err := BuildWordReport(report)
	if err != nil {
		t.Fatalf("BuildWordReport 失败: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("docx 不是合法 zip: %v", err)
	}
	names := map[string]bool{}
	var docXml string
	for _, f := range zr.File {
		names[f.Name] = true
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			docXml = string(b)
		}
	}
	for _, want := range []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml"} {
		if !names[want] {
			t.Errorf("docx 缺少部件 %s", want)
		}
	}
	if docXml == "" {
		t.Fatal("document.xml 为空")
	}

	// XML 必须可解析（标签配对/转义正确）
	dec := xml.NewDecoder(strings.NewReader(docXml))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("document.xml 非法 XML: %v", err)
		}
	}

	// 关键内容抽查：转义、横幅、得分卡、表格
	for _, want := range []string{"某某科技 &amp; Co.", "回答 &lt;标签&gt;", `w:fill="1F4E79"`, "<w:tbl>", "82", "良好"} {
		if !strings.Contains(docXml, want) {
			t.Errorf("document.xml 应包含 %q", want)
		}
	}
}
