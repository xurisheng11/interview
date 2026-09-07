package main

import (
	"context"
	"log"

	"github.com/apache/pulsar-client-go/pulsar"
)

func main() {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 创建消费者
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            "my-topic",
		SubscriptionName: "my-subscription",
		Type:             pulsar.Shared, // 可选: Exclusive, Failover, Shared, KeyShared
	})
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	// 接收消息
	for {
		msg, err := consumer.Receive(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Received message: %s", string(msg.Payload()))

		// 处理完后 ack
		consumer.Ack(msg)
	}
}
