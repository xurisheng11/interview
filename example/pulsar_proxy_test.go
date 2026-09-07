package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func tryConnect(name string, opts pulsar.ClientOptions) {
	fmt.Printf("\n========== 测试: %s ==========\n", name)
	fmt.Printf("URL: %s\n", opts.URL)
	
	client, err := pulsar.NewClient(opts)
	if err != nil {
		fmt.Printf("❌ 创建客户端失败: %v\n", err)
		return
	}
	defer client.Close()
	
	fmt.Println("✅ 客户端创建成功")

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "test-topic",
	}, pulsar.WithContext(ctx))
	if err != nil {
		fmt.Printf("⚠️ 创建生产者失败: %v\n", err)
		return
	}
	defer producer.Close()
	
	msgID, err := producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: []byte("test message"),
	})
	if err != nil {
		fmt.Printf("⚠️ 发送消息失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 消息发送成功: %v\n", msgID)
}

func main() {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicm9rZXItYWRtaW4ifQ.mGhRz0YMwev2jmkIz9cYCewZxSbMqXUwKvZ0fEegK9A"

	// 测试 1: 不加密 + 带 token
	tryConnect("pulsar:// + JWT", pulsar.ClientOptions{
		URL:            "pulsar://192.168.80.11:30257",
		Authentication: pulsar.NewAuthenticationToken(token),
	})

	// 测试 2: TLS + JWT
	tryConnect("pulsar+ssl:// + JWT", pulsar.ClientOptions{
		URL:            "pulsar+ssl://192.168.80.11:30257",
		Authentication: pulsar.NewAuthenticationToken(token),
		TLSAllowInsecureConnection: true,
	})

	// 测试 3: 不加密 + 不认证
	tryConnect("pulsar:// 无认证", pulsar.ClientOptions{
		URL: "pulsar://192.168.80.11:30257",
	})

	// 测试 4: TLS + 不认证
	tryConnect("pulsar+ssl:// 无认证", pulsar.ClientOptions{
		URL: "pulsar+ssl://192.168.80.11:30257",
		TLSAllowInsecureConnection: true,
	})
}
