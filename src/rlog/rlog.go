package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/alphazero/redis"
)

type RedisConf struct {
	Host     string
	Port     int
	Db       int
	Password string
}

type Rlog struct {
	async   bool
	channel string
	buffer  chan []byte
	lock    sync.Mutex
	client  redis.Client
	conf    *RedisConf
}

//创建同步对象
func NewRlog() *Rlog {
	return &Rlog{
		async: false,
	}
}

//创建异步对象
func NewAsyncRlog() *Rlog {
	RlogHandler := &Rlog{
		async:  true,
		buffer: make(chan []byte, 1024),
	}
}

func (this *Rlog) SetRedis(host string, port int, db int, password string, channel string) (err error) {
	spec := redis.DefaultSpec().Host(host).Port(port).Db(db).Password(password)
	client, e := redis.NewSynchClientWithSpec(spec)
	if e != nil {
		err = errors.New("Can not create redis connection")
		return err
	}
	this.client = client
	this.channel = channel
	this.conf = &RedisConf{
		Host:     host,
		port:     port,
		Db:       db,
		Password: password,
	}
	go this.asyncWriter()
}

func (this *Rlog) Write(p []byte) (n int, err error) {
	if !this.async {
		this.lock.Lock()
		defer this.lock.Unlock()
		count, err := this.client.Publish(this.channel, p)
		return int(count), err
	}
}

func (this *Rlog) recon() {
	spec := redis.DefaultSpec().Host(this.conf.Host).Port(this.conf.Port).Db(this.conf.Db).Password(this.conf.Password)
	client, e := redis.NewSynchClientWithSpec(spec)
	if e != nil {
		fmt.Fprintln(os.Stdout, "Reconnect Error")
		return
	}
	this.client = client
}

func (this *Rlog) asyncWriter() {
	for v := range this.buffer {
		_, err := this.client.Publish(this.channel, p)
		if err != nil {
			this.recon()
		}
	}
}
