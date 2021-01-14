package mq_test

import (
	"github.com/moskvorechie/go-mq/v2"
	"testing"
)

func TestNew(t *testing.T) {

	mq.User = "rabbit"
	mq.Pass = "rabbit"
	mq.Host = "127.0.0.1"
	mq.Port = "30401"
	mq.PingEachMinute = 1

	x, err := mq.New()
	if err != nil {
		t.Fatal(err)
	}

	if x.Conn.IsClosed() {
		t.Fatal("conn closed")
	}

	mq.Close()

	x, err = mq.New()
	if err != nil {
		t.Fatal(err)
	}

	if x.Conn.IsClosed() {
		t.Fatal("conn closed")
	}
}
