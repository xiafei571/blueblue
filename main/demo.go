package main

import (
	"fmt"

	"github.com/xiafei571/blueblue/rabbitMQ"
)

func main() {
	rabbitmq := rabbitMQ.NewRabbitMQSimple("" + "blue")
	rabbitmq.PublishSimple("Hello World!")
	fmt.Println("发送成功！")

	recieve := rabbitMQ.NewRabbitMQSimple("" + "blue")
	recieve.ConsumeSimple()
}
