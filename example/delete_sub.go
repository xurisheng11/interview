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
	
	// 删除订阅: persistent://tcenter/serverless/audit 下的 "audit" 订阅
	// API: DELETE /admin/v2/persistent/{tenant}/{namespace}/{topic}/subscription/{subscriptionName}
	url := "http://192.168.80.11:30257/admin/v2/persistent/tcenter/serverless/audit/subscription/audit"
	
	fmt.Printf("删除订阅: persistent://tcenter/serverless/audit 的 'audit'\n")
	fmt.Printf("URL: %s\n\n", url)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", strings.TrimSpace(string(body)))
	
	if resp.StatusCode == 204 || resp.StatusCode == 200 {
		fmt.Println("\n✅ 订阅 'audit' 删除成功！")
	}
}
