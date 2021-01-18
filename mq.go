package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

type Config struct {
	User      string
	Pass      string
	Host      string
	Port      string
	Heartbeat time.Duration
}

type RabbitMQ struct {
	conn          *amqp.Connection
	chNotifyClose chan *amqp.Error
	cfg           Config
	lastPing      time.Time
	sync.RWMutex
}

func New(cfg Config) (mq *RabbitMQ, err error) {
	mq = &RabbitMQ{
		cfg: cfg,
	}
	mq.conn, err = mq.connect()
	if err != nil {
		return
	}
	if cfg.Heartbeat > 0 {
		mq.conn.Config.Heartbeat = cfg.Heartbeat
	}
	return
}

// If set chReconnectOnFailure channel will try reconnect on failure
func (mq *RabbitMQ) NewChannel() (ch *amqp.Channel, err error) {
	ch, err = mq.conn.Channel()
	if err != nil {
		return ch, err
	}
	return ch, err
}

// Restore connection if conn close
func (mq *RabbitMQ) TryRestoreConnections(ch *amqp.Channel) {
	mq.Lock()
	defer mq.Unlock()
	var err error
	if !mq.conn.IsClosed() {
		return
	}
	for k := 1; k <= 20; k++ {
		time.Sleep(time.Duration(500*k) * time.Millisecond)
		_ = ch.Close()
		_ = mq.conn.Close()
		mq.conn, err = mq.connect()
		if err != nil {
			log.Printf("mq: bad conn reconnect, try %d", k)
			continue
		}
		if mq.conn.IsClosed() {
			log.Fatal("mq: reconnect conn failure 1")
		}
		log.Printf("mq: conn reconnect success")

		ch, err = mq.NewChannel()
		if err != nil {
			log.Fatal(err)
		}

		return
	}
	log.Fatal("mq: reconnect conn failure 2")
	return
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
