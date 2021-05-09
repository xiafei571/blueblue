package main

import (
	it "blueblue/init"
	rabbitMQ "blueblue/rabbitMQ"
	"fmt"
)

type Info struct {
	Name string
	Age  int
	Sex  string
}

func main() {
	config := it.InitConfigure()
	user := config.GetString("MQ.User")
	password := config.GetString("MQ.Password")
	ip := config.GetString("MQ.IP")
	vhost := config.GetString("MQ.Vhost")
	MQURL := "amqp://" + user + ":" + password + "@" + ip + "/" + vhost
	fmt.Println(MQURL)

	// infos := []Info{
	// 	{
	// 		Name: "Sophia",
	// 		Age:  23,
	// 		Sex:  "female",
	// 	},
	// 	{
	// 		Name: "Benjie",
	// 		Age:  24,
	// 		Sex:  "male",
	// 	},
	// }
	// data, err := json.Marshal(infos)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(data)
	// fmt.Println(string(data))

	// rabbitmq := rabbitMQ.NewRabbitMQSimple(MQURL, ""+"blue")
	// rabbitmq.PublishByte(data)
	// fmt.Println("发送成功！")

	recieve := rabbitMQ.NewRabbitMQSimple(MQURL, ""+"blue")
	recieve.ConsumeSimple()
}
