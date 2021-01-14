package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type Config struct {
	User               string
	Pass               string
	Host               string
	Port               string
	PingEachMinute     int
	ReconnectOnFailure bool
}

type RabbitMQ struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	chClose  chan *amqp.Error
	cfg      Config
	lastPing time.Time
}

func New(cfg Config) (mq *RabbitMQ, err error) {
	mq = &RabbitMQ{
		cfg: cfg,
	}
	err = mq.connectCh()
	if cfg.ReconnectOnFailure {
		mq.reconnectOnFailure()
	}
	return
}

func (mq *RabbitMQ) reconnectOnFailure() {
	go mq.listenClose()
	mq.channel.NotifyClose(mq.chClose)
}

func (mq *RabbitMQ) listenClose() {
	for {
		<-mq.chClose
		log.Println("mq: channel closed")
		_ = mq.Close()
		log.Println("mq: reconnect..")
		err := mq.connectCh()
		if err != nil {
			log.Fatal("mq: bad reconnect")
		}
		log.Println("mq: reconnect success")
		continue
	}
}

func (mq *RabbitMQ) connectCh() (err error) {
	mq.conn, err = mq.connect()
	if err != nil {
		return
	}
	mq.channel, err = mq.conn.Channel()
	if err != nil {
		return
	}

	return
}

// Connect to MQ
func (mq *RabbitMQ) GetConn() *amqp.Connection {
	return mq.conn
}

func (mq *RabbitMQ) GetChan() *amqp.Channel {
	return mq.channel
}

func (mq *RabbitMQ) Close() (err error) {
	err = mq.channel.Close()
	if err != nil {
		return
	}
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
			_ = mq.channel.Close()
			_ = mq.conn.Close()
			err = mq.connectCh()
			if err != nil {
				return err
			}
		}
	}
	return
}
