package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicm9rZXItYWRtaW4ifQ.mGhRz0YMwev2jmkIz9cYCewZxSbMqXUwKvZ0fEegK9A"
	
	// 主题: persistent://tcenter/serverless/audit
	// API 格式: /admin/v2/persistent/{tenant}/{namespace}/{topic}/subscriptions
	url := "http://192.168.80.11:30257/admin/v2/persistent/tcenter/serverless/audit/subscriptions"
	
	fmt.Printf("查询主题: persistent://tcenter/serverless/audit\n")
	fmt.Printf("URL: %s\n\n", url)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	
	if resp.StatusCode == 200 {
		content := strings.TrimSpace(string(body))
		if content == "null" || content == "[]" || content == "" {
			fmt.Println("订阅数量: 0")
		} else {
			// 解析订阅列表
			count := strings.Count(content, ",") + 1
			if strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]") {
				// 去掉括号
				list := strings.Trim(content, "[]")
				if strings.TrimSpace(list) == "" {
					count = 0
				}
			}
			fmt.Printf("订阅数量: %d\n", count)
			fmt.Printf("订阅列表:\n%s\n", content)
		}
	} else {
		fmt.Printf("响应: %s\n", string(body))
	}
}
