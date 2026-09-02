package service

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"code.sajari.com/docconv/v2"
)

// ExtractTextFromFile 从 PDF 或 Word 文件中提取纯文本
// data: 文件内容; mimeType: application/pdf / application/msword / application/vnd.openxmlformats-...
func ExtractTextFromFile(data []byte, mimeType string) (string, error) {
	res, err := docconv.Convert(bytes.NewReader(data), mimeType, true)
	if err != nil {
		return "", fmt.Errorf("文件解析失败: %w", err)
	}

	text := strings.TrimSpace(res.Body)
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

// ExtractTextFromReader 从 io.Reader 中提取文本（内部读取全部字节后调用 Convert）
func ExtractTextFromReader(reader io.Reader, mimeType string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return ExtractTextFromFile(data, mimeType)
}
