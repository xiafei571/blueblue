package main

import (
	"encoding/json"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/xiafei571/blueblue/rabbitMQ"
)

type Info struct {
	Name string
	Age  int
	Sex  string
}

func main() {
	config := initConfigure()
	user := config.GetString("MQ.User")
	password := config.GetString("MQ.Password")
	ip := config.GetString("MQ.IP")
	vhost := config.GetString("MQ.Vhost")
	MQURL := "amqp://" + user + ":" + password + "@" + ip + "/" + vhost
	fmt.Println(MQURL)

	infos := []Info{
		{
			Name: "Sophia",
			Age:  23,
			Sex:  "female",
		},
		{
			Name: "Benjie",
			Age:  24,
			Sex:  "male",
		},
	}
	data, err := json.Marshal(infos)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
	fmt.Println(string(data))

	rabbitmq := rabbitMQ.NewRabbitMQSimple(MQURL, ""+"blue")
	rabbitmq.PublishByte(data)
	fmt.Println("发送成功！")

	recieve := rabbitMQ.NewRabbitMQSimple(MQURL, ""+"blue")
	recieve.ConsumeSimple()
}

func initConfigure() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		fmt.Println("err:", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(" Config file not found; ignore error if desired")
		} else {
			panic("Config file was found but another error was produced")
		}
	}
	// 监控配置和重新获取配置
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	return v
}
