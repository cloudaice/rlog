package rlog

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/cloudaice/rlog/redis"
)

const (
	RedisHost string = "localhost"
	RedisPort int    = 6379
)

func do(rlog *Rlog, t *testing.T) {
	err := rlog.SetRedis(RedisHost, RedisPort, NODB, NOPASS, "TEST_RLOG")
	if err != nil {
		t.Error("Can not set Redis")
	}
	log.SetOutput(rlog)

	spec := redis.DefaultSpec().Host(RedisHost).Port(RedisPort)
	client, err := redis.NewPubSubClientWithSpec(spec)
	if err != nil {
		t.Error("Can not create pubsub client")
	}
	e := client.Subscribe("TEST_RLOG")
	if e != nil {
		t.Error(e)
	}
	subChan := client.Messages("TEST_RLOG")
	log.Println("Hello World")
	select {
	case msg := <-subChan:
		if index := bytes.Index(msg, []byte("Hello World")); index == -1 {
			t.Error("No message recived")
		}
	case <-time.After(time.Second):
		t.Error("Time out")
	}

}

func TestRlog(t *testing.T) {
	rlog := NewRlog()
	do(rlog, t)
}

func TestAsyncRlog(t *testing.T) {
	rlog := NewAsyncRlog()
	do(rlog, t)
}
