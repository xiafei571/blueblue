import (
	"fmt"

	"github.com/xiafei571/ble/rabbitMQ"
)

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("" +
		"blue")
	rabbitmq.PublishSimple("Hello World!")
	fmt.Println("发送成功！")
}