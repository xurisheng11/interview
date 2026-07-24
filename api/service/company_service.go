package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"interview-sim/repository"
)

// CompanyIntelResult 公司面试情报结果
type CompanyIntelResult struct {
	Company    string              `json:"company"`
	JobTitle   string              `json:"jobTitle"`
	Process    string              `json:"process"`
	Questions  []CompanyQuestion   `json:"questions"`
	Tips       []string            `json:"tips"`
	Difficulty string              `json:"difficulty"`
	CreatedAt  int64               `json:"createdAt"`
}

// CompanyQuestion 公司面试题
type CompanyQuestion struct {
	Content  string   `json:"content"`
	Type     string   `json:"type"` // technical / behavioral
	Tags     []string `json:"tags"`
	Answer   string   `json:"answer"` // AI 参考答案（Markdown 格式）
}

func companyIntelKey(company, jobTitle string) string {
	key := strings.ToLower(strings.TrimSpace(company) + ":" + strings.TrimSpace(jobTitle))
	return "company:intel:v2:" + key
}

// GetCompanyIntel 获取公司面试情报（带缓存，24h有效）
func GetCompanyIntel(company, jobTitle string) (*CompanyIntelResult, error) {
	cacheKey := companyIntelKey(company, jobTitle)

	// 先查缓存
	cached, err := repository.Get(cacheKey)
	if err == nil && cached != "" {
		var result CompanyIntelResult
		if jsonErr := json.Unmarshal([]byte(cached), &result); jsonErr == nil {
			return &result, nil
		}
	}

	jobDesc := jobTitle
	if jobDesc == "" {
		jobDesc = "通用（所有岗位）"
	}

	// ---- 第一步：获取基础情报 + 题目列表（不含答案，内容少不会截断）----
	prompt1 := fmt.Sprintf(`你是一位资深的求职顾问，请根据公开信息介绍【%s】公司的面试情况。
岗位方向：%s

请严格按照以下 JSON 格式返回，不要包含任何其他文字：
{
  "company": "%s",
  "jobTitle": "%s",
  "process": "面试流程介绍（描述面试轮数、每轮特点，150字左右）",
  "questions": [
    {"content": "面试题目内容", "type": "technical", "tags": ["标签1", "标签2"]},
    {"content": "行为面试题内容", "type": "behavioral", "tags": ["标签"]}
  ],
  "tips": ["面试技巧1", "面试技巧2", "面试技巧3", "面试技巧4", "面试技巧5"],
  "difficulty": "初级或中级或高级"
}

要求：
1. questions 包含15道题，技术题10道、行为题5道，每道题只需content/type/tags三个字段
2. tips 5条针对该公司的具体建议
3. 内容真实客观`, company, jobDesc, company, jobTitle)

	raw1, err := Chat(prompt1)
	if err != nil {
		return nil, fmt.Errorf("AI 查询失败: %w", err)
	}
	raw1 = cleanJSON(raw1)

	var result CompanyIntelResult
	if err := json.Unmarshal([]byte(raw1), &result); err != nil {
		fixed := tryFixJSON(raw1)
		if fixErr := json.Unmarshal([]byte(fixed), &result); fixErr != nil {
			return nil, fmt.Errorf("解析基础情报失败: %w", err)
		}
	}
	result.CreatedAt = time.Now().Unix()

	// 写缓存 24h
	if b, err := json.Marshal(result); err == nil {
		_ = repository.Set(cacheKey, string(b), 24*time.Hour)
	}

	return &result, nil
}

// GetCompanyQuestionAnswer 获取单道公司面试题的参考答案（带缓存）
func GetCompanyQuestionAnswer(company, jobTitle, content string) (string, error) {
	// 用题目内容做 cache key
	cacheKey := "company:qa:" + company + ":" + content
	if len(cacheKey) > 200 {
		cacheKey = cacheKey[:200]
	}

	if cached, err := repository.Get(cacheKey); err == nil && cached != "" {
		return cached, nil
	}

	jobDesc := jobTitle
	if jobDesc == "" {
		jobDesc = "通用"
	}

	prompt := fmt.Sprintf(`请为以下面试题提供参考答案。

公司：%s
岗位：%s
题目：%s

要求：
1. 使用 Markdown 格式
2. 包含核心要点、思路分析和示例（如适用）
3. 长度 200-400 字

只返回答案内容，不要其他文字。`, company, jobDesc, content)

	answer, err := Chat(prompt)
	if err != nil {
		return "", fmt.Errorf("AI 生成答案失败: %w", err)
	}

	_ = repository.Set(cacheKey, answer, 24*time.Hour)
	return answer, nil
}
func fetchAnswers(questions []CompanyQuestion, company, jobDesc string) []CompanyQuestion {
	batchSize := 5
	for i := 0; i < len(questions); i += batchSize {
		end := i + batchSize
		if end > len(questions) {
			end = len(questions)
		}
		batch := questions[i:end]

		// 构建该批题目的 JSON
		type qInput struct {
			Index   int    `json:"index"`
			Content string `json:"content"`
		}
		inputs := make([]qInput, len(batch))
		for j, q := range batch {
			inputs[j] = qInput{Index: i + j, Content: q.Content}
		}
		inputJSON, _ := json.Marshal(inputs)

		prompt := fmt.Sprintf(`请为以下【%s】公司（岗位：%s）的面试题提供参考答案。

题目列表：%s

请严格按照以下 JSON 数组格式返回，不要包含任何其他文字：
[{"index":0,"answer":"参考答案，Markdown格式，包含核心要点，150-300字"}]

每道题必须返回对应的 index 和 answer。`, company, jobDesc, string(inputJSON))

		raw, err := Chat(prompt)
		if err != nil {
			continue
		}
		raw = cleanJSON(raw)

		var answers []struct {
			Index  int    `json:"index"`
			Answer string `json:"answer"`
		}
		if err := json.Unmarshal([]byte(raw), &answers); err != nil {
			continue
		}
		for _, a := range answers {
			if a.Index >= 0 && a.Index < len(questions) {
				questions[a.Index].Answer = a.Answer
			}
		}
	}
	return questions
}

// tryFixJSON 尝试修复截断的 JSON 字符串
func tryFixJSON(s string) string {
	// 找到最后一个完整的 } 或 ]
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '}' || s[i] == ']' {
			return s[:i+1]
		}
	}
	return s
}
