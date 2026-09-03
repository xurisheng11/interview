package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

	if len(text) > 8000 {
		text = text[:8000]
	}
	return text, nil
}

// extractPDF 先尝试纯 Go 库，再尝试 Python PyMuPDF，最后回退到 OCR
func extractPDF(data []byte) (string, error) {
	// 方法 1: 尝试纯 Go PDF 库
	text, err := extractPDFByGo(data)
	if err == nil && strings.TrimSpace(text) != "" {
		// 检查是否有有效内容
		if hasValidContent(text) {
			return text, nil
		}
	}

	// 方法 2: 尝试 Python PyMuPDF（最可靠，对 Canva 等自定义字体支持好）
	text, err = extractPDFByPython(data)
	if err == nil && strings.TrimSpace(text) != "" {
		if hasValidContent(text) {
			return text, nil
		}
	}

	// 方法 3: 回退到 OCR
	return extractPDFByOCR(data)
}

// hasValidContent 检查文本是否包含有效内容
func hasValidContent(text string) bool {
	// 统计中文字符
	chineseCount := 0
	readableCount := 0
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			chineseCount++
		}
		if r >= 32 && r < 127 || r >= 0x4E00 && r <= 0x9FFF {
			readableCount++
		}
	}
	return chineseCount > 5 || (readableCount > len(text)/2 && len(text) > 20)
}

// findPython 查找可用的 Python 解释器
func findPython() (string, error) {
	pythonPaths := []string{
		"C:\\Users\\xuris\\AppData\\Local\\Programs\\Python\\Python312\\python.exe",
		"python",
		"python3",
	}
	for _, p := range pythonPaths {
		cmd := exec.Command(p, "-c", "import fitz; print('ok')")
		if err := cmd.Run(); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("未找到 Python 或 PyMuPDF")
}

// extractPDFByGo 使用纯 Go PDF 库提取文本
func extractPDFByGo(data []byte) (string, error) {
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

	return strings.TrimSpace(sb.String()), nil
}

// extractPDFByPython 使用 Python PyMuPDF 提取文本（最可靠的方法）
func extractPDFByPython(data []byte) (string, error) {
	// 写入临时 PDF 文件
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "resume_*.pdf")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	pdfPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(pdfPath)

	if err := os.WriteFile(pdfPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入临时 PDF 失败: %w", err)
	}

	// 查找 Python
	pythonPath, err := findPython()
	if err != nil {
		return "", err
	}

	// 查找 Python 提取脚本
	scriptName := "pdf_extract_text.py"
	scriptPaths := []string{
		filepath.Join(getWorkDir(), "scripts", scriptName),
		filepath.Join(getExecDir(), "scripts", scriptName),
		"E:\\interview\\api\\scripts\\" + scriptName,
		"C:\\interview\\api\\scripts\\" + scriptName,
	}
	var scriptPath string
	for _, p := range scriptPaths {
		if _, err := os.Stat(p); err == nil {
			scriptPath = p
			break
		}
	}
	if scriptPath == "" {
		return "", fmt.Errorf("未找到 PDF 文本提取脚本")
	}

	// 调用 Python 脚本（带 120 秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonPath, scriptPath, pdfPath)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("PDF 文本提取超时（120秒），文件可能过大或页数过多")
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("Python 脚本执行失败: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("Python 脚本执行失败: %w", err)
	}

	var result struct {
		Success      bool   `json:"success"`
		Text         string `json:"text"`
		CharCount    int    `json:"charCount"`
		ChineseCount int    `json:"chineseCount"`
		Method       string `json:"method"`
		Error        string `json:"error"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("解析 Python 输出失败: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("Python 提取失败: %s", result.Error)
	}

	return result.Text, nil
}

// extractPDFByOCR 通过 PyMuPDF + Tesseract OCR 从 PDF 中提取文字
func extractPDFByOCR(data []byte) (string, error) {
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "resume_*.pdf")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	pdfPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(pdfPath)

	if err := os.WriteFile(pdfPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入临时 PDF 失败: %w", err)
	}

	// 查找 Python
	pythonPath, err := findPython()
	if err != nil {
		return "", fmt.Errorf("未找到 Python: %w", err)
	}

	// 查找 OCR 脚本
	scriptName := "pdf_ocr.py"
	scriptPaths := []string{
		filepath.Join(getWorkDir(), "scripts", scriptName),
		filepath.Join(getExecDir(), "scripts", scriptName),
		"E:\\interview\\api\\scripts\\" + scriptName,
		"C:\\interview\\api\\scripts\\" + scriptName,
	}
	var scriptPath string
	for _, p := range scriptPaths {
		if _, err := os.Stat(p); err == nil {
			scriptPath = p
			break
		}
	}

	var text string
	var ocrError error

	if scriptPath != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, pythonPath, scriptPath, pdfPath)
		output, err := cmd.Output()
		if err == nil {
			var result struct {
				FullText string `json:"fullText"`
				Success  bool   `json:"success"`
				Error    string `json:"error"`
			}
			if jsonErr := json.Unmarshal(output, &result); jsonErr == nil {
				if result.Success && result.FullText != "" {
					text = result.FullText
				} else if result.Error != "" {
					ocrError = fmt.Errorf("OCR 返回错误: %s", result.Error)
				}
			}
		} else {
			if ctx.Err() == context.DeadlineExceeded {
				ocrError = fmt.Errorf("OCR 识别超时（180秒），文件可能过大或页数过多")
			} else if exitErr, ok := err.(*exec.ExitError); ok {
				ocrError = fmt.Errorf("OCR 脚本执行失败 (exit %d): %s", exitErr.ExitCode(), string(exitErr.Stderr))
			} else {
				ocrError = fmt.Errorf("OCR 脚本执行失败: %w", err)
			}
		}
	} else {
		ocrError = fmt.Errorf("未找到 OCR 脚本")
	}

	if text != "" {
		return text, nil
	}

	if ocrError != nil {
		return "", fmt.Errorf("PDF 文字提取失败，OCR 识别也失败: %v。请：① 使用可复制文字的 PDF ② 上传 Word 格式简历", ocrError)
	}
	return "", fmt.Errorf("PDF 文字提取失败，请：① 使用可复制文字的 PDF ② 上传 Word 格式简历")
}

// getExecDir 返回可执行文件所在目录
func getExecDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// getWorkDir 返回当前工作目录
func getWorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
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

// CleanResumeText 清理简历文本，移除乱码和无用字符
func CleanResumeText(text string) string {
	// 移除零宽字符
	text = regexp.MustCompile(`[\u200B-\u200D\uFEFF]`).ReplaceAllString(text, "")
	// 规范化空白
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
