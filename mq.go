package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Error   chan *amqp.Error
}

var MQ *RabbitMQ
var User string
var Pass string
var Host string
var Port string

var PingEachMinute int

var lastPing time.Time

// Get instance
func New() (*RabbitMQ, error) {

	var err error

	if MQ == nil {
		MQ.Conn, err = Connect()
		if err != nil {
			return MQ, err
		}
		MQ.Channel, err = MQ.Conn.Channel()
		if err != nil {
			return MQ, err
		}
		return New()
	}
	if PingEachMinute > 0 && time.Now().After(lastPing.Add(time.Duration(PingEachMinute)*time.Minute)) {
		lastPing = time.Now()
		if MQ.Conn.IsClosed() {
			MQ.Channel.Close()
			MQ.Conn.Close()
			MQ.Conn, err = Connect()
			if err != nil {
				return MQ, err
			}
			return New()
		}
	}

	return MQ, nil
}

// Connect to DB
func Connect() (*amqp.Connection, error) {
	conn, err := amqp.Dial(GetLInk())
	if err != nil {
		return conn, err
	}

	return conn, err
}

// Format link
func GetLInk() string {
	connLink := fmt.Sprintf("amqp://%s:%s@%s:%s/", User, Pass, Host, Port)

	return connLink
}
