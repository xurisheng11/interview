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

	// 创建生产者
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "my-topic",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	// 发送消息
	msgID, err := producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: []byte("Hello Pulsar!"),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Message sent: %v", msgID)
}
