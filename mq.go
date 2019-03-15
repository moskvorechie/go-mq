package mq

import "github.com/streadway/amqp"

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Error   chan *amqp.Error
}

func (mq *RabbitMQ) Connect(connLink string) error {

	// Connect
	var err error
	mq.Conn, err = amqp.Dial(connLink)
	if err != nil {
		return err
	}

	// Open channel
	mq.Channel, err = mq.Conn.Channel()
	if err != nil {
		return err
	}

	return nil
}
