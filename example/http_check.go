package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicm9rZXItYWRtaW4ifQ.mGhRz0YMwev2jmkIz9cYCewZxSbMqXUwKvZ0fEegK9A"
	
	testURLs := []struct {
		name string
		url  string
	}{
		{"Admin API v2 - clusters", "http://192.168.80.11:30257/admin/v2/clusters"},
		{"Admin API v2 - tenants", "http://192.168.80.11:30257/admin/v2/tenants"},
		{"Admin API v2 - namespaces", "http://192.168.80.11:30257/admin/v2/namespaces"},
		{"Admin API v2 - health", "http://192.168.80.11:30257/admin/v2/brokers/health"},
		{"Protocol version", "http://192.168.80.11:30257/version"},
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, test := range testURLs {
		fmt.Printf("\n========== 测试: %s ==========\n", test.name)
		fmt.Printf("URL: %s\n", test.url)

		req, _ := http.NewRequest("GET", test.url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应: %s\n", string(body))
	}
}
