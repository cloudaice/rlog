package rlog

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/alphazero/redis"
)

const (
	RedisHost string = "localhost"
	RedisPort int    = 6379
	RedisDb   int    = 8
	Password  string = "123456"
)

func TestRlog(t *testing.T) {
	rlog := NewRlog()
	err := rlog.SetRedis(RedisHost, RedisPort, RedisDb, Password, "TEST_RLOG")
	if err != nil {
		t.Error("Can not set Redis")
	}
	log.SetOutput(rlog)

	spec := redis.DefaultSpec().Host(RedisHost).Port(RedisPort).Db(RedisDb).Password(Password)
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
