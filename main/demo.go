package main

import (
	"fmt"

	"github.com/xiafei571/blueblue/rabbitMQ"
)

func main() {
	config := initConfigure()
	MQURL := "amqp://" + config.rabbitMQ.User + "+:" + config.rabbitMQ.Password + "@" + config.rabbitMQ.IP + "//" + config.rabbitMQ.Vhost
	fmt.Println(MQURL)
	rabbitmq := rabbitMQ.NewRabbitMQSimple(MQURL, ""+"blue")
	rabbitmq.PublishSimple("Hello World!")
	fmt.Println("发送成功！")

	recieve := rabbitMQ.NewRabbitMQSimple(MQURL, ""+"blue")
	recieve.ConsumeSimple()
}
