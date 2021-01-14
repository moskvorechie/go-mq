package mq_test

import (
	"github.com/moskvorechie/go-mq/v3"
	"testing"
)

func TestNew(t *testing.T) {
	x, err := mq.New(mq.Config{
		User:           "rabbit",
		Pass:           "rabbit",
		Host:           "127.0.0.1",
		Port:           "30401",
		PingEachMinute: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if x.GetConn().IsClosed() {
		t.Fatal("conn closed")
	}
	err = x.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !x.GetConn().IsClosed() {
		t.Fatal("conn not closed")
	}
}
