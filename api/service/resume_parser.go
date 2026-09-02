package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
)

// ExtractTextFromFile 从 PDF 或 Word 文件中提取纯文本
// data: 文件内容; mimeType: application/pdf / application/msword / application/vnd.openxmlformats-...
func ExtractTextFromFile(data []byte, mimeType string) (string, error) {
	var text string
	var err error

	switch mimeType {
	case "application/pdf":
		text, err = extractPDF(data)
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		text, err = extractDOCX(data)
	case "application/msword":
		text, err = "", fmt.Errorf("不支持 .doc 格式，请将简历另存为 .docx 或 PDF 格式后重试")
	default:
		// 尝试当作纯文本读取
		text = string(data)
		if !utf8.ValidString(text) {
			text = ""
		}
		if text == "" {
			err = fmt.Errorf("无法识别的文件格式: %s", mimeType)
		}
	}

	if err != nil {
		return "", fmt.Errorf("文件解析失败: %w", err)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("无法从文件中提取文本，请确认文件内容不为空")
	}
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "")
	}

	// 限制长度，避免 prompt 过长
	if len(text) > 8000 {
		text = text[:8000]
	}
	return text, nil
}

// extractPDF 使用纯 Go 库提取 PDF 文本
func extractPDF(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	r, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("PDF 格式错误: %w", err)
	}

	var sb strings.Builder
	numPages := r.NumPage()
	for i := 1; i <= numPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(content)
		sb.WriteString("\n")
	}

	text := sb.String()
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("PDF 中未提取到文本内容，可能为扫描件或图片型 PDF，建议：① 使用文字型 PDF（可直接选中文本的 PDF）② 将 Word 简历另存为 PDF ③ 或上传 Word 格式的简历")
	}
	return text, nil
}

// extractDOCX 使用纯 Go zip 库提取 DOCX 文本
func extractDOCX(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("DOCX 格式错误: %w", err)
	}

	var sb strings.Builder
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			xmlData, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// 解析 XML，提取所有 <w:t> 标签内的文本
			dec := xml.NewDecoder(bytes.NewReader(xmlData))
			for {
				token, err := dec.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}
				if se, ok := token.(xml.StartElement); ok {
					if se.Name.Local == "t" {
						for _, attr := range se.Attr {
							if attr.Name.Local == "xml:space" && attr.Value == "preserve" {
								// 保留空白
							}
						}
						var text string
						if err := dec.DecodeElement(&text, &se); err == nil {
							sb.WriteString(text)
						}
					}
				}
			}
		}
	}

	text := sb.String()
	if text == "" {
		return "", fmt.Errorf("DOCX 中未提取到文本内容")
	}
	return text, nil
}

// ExtractTextFromReader 从 io.Reader 中提取文本
func ExtractTextFromReader(reader io.Reader, mimeType string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return ExtractTextFromFile(data, mimeType)
}
