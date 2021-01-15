package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type Config struct {
	User           string
	Pass           string
	Host           string
	Port           string
	PingEachMinute int
}

type RabbitMQ struct {
	conn          *amqp.Connection
	chNotifyClose chan *amqp.Error
	cfg           Config
	lastPing      time.Time
}

func New(cfg Config) (mq *RabbitMQ, err error) {
	mq = &RabbitMQ{
		cfg: cfg,
	}
	mq.conn, err = mq.connect()
	if err != nil {
		return
	}
	return
}

func (mq *RabbitMQ) NewChannel(asyncReconnectOnFailure bool) (ch *amqp.Channel, err error) {

	ch, err = mq.conn.Channel()
	if err != nil {
		return
	}
	if asyncReconnectOnFailure {
		mq.reconnectOnFailure(ch)
	}
	return
}

func (mq *RabbitMQ) reconnectOnFailure(ch *amqp.Channel) {
	var err error
	chNotifyClose := make(chan *amqp.Error)
	go func(ch *amqp.Channel) {
		for {
			<-chNotifyClose
			log.Println("mq: channel closed")
			_ = mq.Close()
			log.Println("mq: reconnect..")
			ch, err = mq.NewChannel(true)
			if err != nil {
				log.Fatal("mq: bad reconnect")
			}
			log.Println("mq: reconnect success")
			continue
		}
	}(ch)
	ch.NotifyClose(chNotifyClose)
}

// Connect to MQ
func (mq *RabbitMQ) GetConn() *amqp.Connection {
	return mq.conn
}

func (mq *RabbitMQ) Close() (err error) {
	err = mq.conn.Close()
	if err != nil {
		return
	}
	return
}

// Connect to MQ
func (mq *RabbitMQ) connect() (*amqp.Connection, error) {
	conn, err := amqp.Dial(mq.getLInk())
	if err != nil {
		return conn, err
	}

	return conn, err
}

// Format link
func (mq *RabbitMQ) getLInk() string {
	connLink := fmt.Sprintf("amqp://%s:%s@%s:%s/", mq.cfg.User, mq.cfg.Pass, mq.cfg.Host, mq.cfg.Port)

	return connLink
}

// Restore connection if conn close
func (mq *RabbitMQ) Up() (err error) {
	if mq.cfg.PingEachMinute > 0 && time.Now().After(mq.lastPing.Add(time.Duration(mq.cfg.PingEachMinute)*time.Minute)) {
		mq.lastPing = time.Now()
		if mq.conn.IsClosed() {
			_ = mq.conn.Close()
			mq.conn, err = mq.connect()
			if err != nil {
				return err
			}
		}
	}
	return
}
