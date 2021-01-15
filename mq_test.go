package mq_test

import (
	"github.com/moskvorechie/go-mq/v3"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	mx, err := mq.New(mq.Config{
		User:               "rabbit",
		Pass:               "rabbit",
		Host:               "127.0.0.1",
		Port:               "30401",
		PingEachMinute:     1,
		ReconnectOnFailure: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ch, err := mx.NewChannel(true)
	if err != nil {
		t.Fatal(err)
	}
	err = ch.Close()
	if err != nil {
		t.Fatal(err)
	}

	if mx.GetConn().IsClosed() {
		t.Fatal("conn closed")
	}
	err = mx.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !mx.GetConn().IsClosed() {
		t.Fatal("conn not closed")
	}

}

func TestLong(t *testing.T) {

	chClose := make(chan bool)
	mx, err := mq.New(mq.Config{
		User:               "rabbit",
		Pass:               "rabbit",
		Host:               "127.0.0.1",
		Port:               "30401",
		PingEachMinute:     1,
		ReconnectOnFailure: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ch, err := mx.NewChannel(true)
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}

	// Create channel
	err = ch.ExchangeDeclare("test-go", "direct", true, false, false, false, amqp.Table{})
	if err != nil {
		t.Fatal(err)
	}

	// Create queue
	_, err = ch.QueueDeclare("test-go", false, false, false, false, amqp.Table{})
	if err != nil {
		t.Fatal(err)
	}

	// Bind queue to exchange
	err = ch.QueueBind("test-go", "test-go", "test-go", false, amqp.Table{})
	if err != nil {
		t.Fatal(err)
	}

	// Listen 10 messages
	wg.Add(1)
	go func(mx *mq.RabbitMQ) {
		defer wg.Done()
		ch, err := ch.Consume("test-go", "test-go", false, false, false, false, amqp.Table{})
		if err != nil {
			t.Fatal(err)
		}
		for {
			select {
			case message, ok := <-ch:
				if !ok {
					return
				}
				log.Println(message.Body)
				err = message.Ack(false)
				if err != nil {
					t.Fatal(err)
				}
			case _, ok := <-chClose:
				if !ok {
					return
				}
			}
		}
	}(mx)

	// Send 10 messages
	wg.Add(1)
	go func(mx *mq.RabbitMQ) {
		defer wg.Done()
		total := 0
		for {

			min := 1
			max := 30
			wait := rand.Intn(max-min) + min

			time.Sleep(time.Duration(wait) * time.Second)
			total++
			if total > 1000 {
				break
			}

			err = ch.Publish("test-go", "test-go", false, false, amqp.Publishing{
				DeliveryMode: 2,
				Timestamp:    time.Now(),
				AppId:        "test-go",
				Body:         []byte("test-go"),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		close(chClose)
	}(mx)

	wg.Wait()
}
