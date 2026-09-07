package main

import (
	"context"
	"fmt"
	"log"

	"github.com/apache/pulsar-client-go/pulsar"
)

func main() {
	// Pulsar 连接配置
	brokerURL := "pulsar://192.168.80.11:30257"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicm9rZXItYWRtaW4ifQ.mGhRz0YMwev2jmkIz9cYCewZxSbMqXUwKvZ0fEegK9A"

	// 创建客户端（带 JWT 认证）
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:      brokerURL,
		Authentication: pulsar.NewAuthenticationToken(token),
	})
	if err != nil {
		log.Fatalf("连接 Pulsar 失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ 连接成功！")

	// 创建消费者测试
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            "test-topic",
		SubscriptionName: "test-subscription",
		Type:             pulsar.Shared,
	})
	if err != nil {
		fmt.Printf("⚠️ 订阅 test-topic 失败（可能主题不存在）: %v\n", err)
	} else {
		fmt.Println("✅ 订阅功能正常！")
		consumer.Close()
	}

	// 创建生产者测试
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "test-topic",
	})
	if err != nil {
		fmt.Printf("⚠️ 创建生产者失败（可能主题不存在）: %v\n", err)
	} else {
		fmt.Println("✅ 生产者功能正常！")
		
		// 发送测试消息
		ctx := context.Background()
		msgID, err := producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: []byte("test message"),
		})
		if err != nil {
			fmt.Printf("⚠️ 发送消息失败: %v\n", err)
		} else {
			fmt.Printf("✅ 消息发送成功，MessageID: %v\n", msgID)
		}
		producer.Close()
	}

	fmt.Println("\n🎉 Pulsar 连接测试完成！")
}
