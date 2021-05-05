import "github.com/xiafei571/ble/rabbitMQ"

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("" +
		"blue")
	rabbitmq.ConsumeSimple()
}