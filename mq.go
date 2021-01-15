package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
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
func (mq *RabbitMQ) NewChannel(chReconnectOnFailure chan struct{}) (ch *amqp.Channel, err error) {
	ch, err = mq.conn.Channel()
	if err != nil {
		return
	}
	if chReconnectOnFailure != nil {
		mq.reconnectChanOnFailure(ch, chReconnectOnFailure)
	}
	return
}

// Restore connection if conn close
func (mq *RabbitMQ) RestoreConnections(ch *amqp.Channel, chReconnectOnFailure chan struct{}) (err error) {
	if !mq.conn.IsClosed() {
		return
	}
	mq.lastPing = time.Now()
	for k := 1; k <= 10; k++ {
		time.Sleep(time.Duration(300*k) * time.Millisecond)
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

		ch, err = mq.NewChannel(chReconnectOnFailure)
		if err != nil {
			log.Fatal(err)
		}

		return
	}
	log.Fatal("mq: reconnect conn failure 2")
	return
}

func (mq *RabbitMQ) reconnectChanOnFailure(ch *amqp.Channel, chExit chan struct{}) {
	var err error
	chNotifyClose := make(chan *amqp.Error)
	go func(ch *amqp.Channel) {
		for {
			select {
			case <-chNotifyClose:
				log.Println("mq: chan closed, trying reconnect")
				_ = mq.Close()
				log.Println("mq: reconnect chan ...")
				for k := 1; k <= 10; k++ {
					time.Sleep(time.Duration(300*k) * time.Millisecond)
					ch, err = mq.NewChannel(chExit)
					if err != nil {
						log.Printf("mq: bad chan reconnect, try %d", k)
						continue
					}
					log.Println("mq: reconnect chan success")
					return
				}
				log.Fatal("mq: reconnect chan failure")
			case <-chExit:
				log.Println("mq: reconnect chan exit")
				return
			}
		}
	}(ch)
	ch.NotifyClose(chNotifyClose)
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
