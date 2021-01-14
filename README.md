# Библиотека для подключения к RabbitMQ

Пример использования:

```go
package main

import (
	"github.com/moskvorechie/go-mq/v3"
	"github.com/streadway/amqp"
	"log"
)

func main() {
	mx, err := mq.New(mq.Config{
		User:               "rabbit",
		Pass:               "rabbit",
		Host:               "127.0.0.1",
		Port:               "30401",
		PingEachMinute:     1,
		ReconnectOnFailure: true,
	})
	ch, err := mx.GetChan().Consume("test-go", "test-go", false, false, false, false, amqp.Table{})
	if err != nil {
		panic(err)
	}
	for {
		select {
		case message, ok := <-ch:
			if !ok {
				return
			}
			log.Println(message)
		}
	}
}
```