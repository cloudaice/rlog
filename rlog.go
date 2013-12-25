package rlog

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudaice/rlog/redis"
)

const (
	NODB   int    = 0
	NOPASS string = ""
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
	return &Rlog{
		async:  true,
		buffer: make(chan []byte, 1024),
	}
}

func (this *Rlog) SetRedis(host string, port int, db int, password string, channel string) (err error) {
	var spec *redis.ConnectionSpec
	switch {
	case db == NODB && password == NOPASS:
		spec = redis.DefaultSpec().Host(host).Port(port)
	case db == NODB && password != NOPASS:
		spec = redis.DefaultSpec().Host(host).Port(port).Password(password)
	case db != NODB && password == NOPASS:
		spec = redis.DefaultSpec().Host(host).Port(port).Db(db)
	case db != NODB && password != NOPASS:
		spec = redis.DefaultSpec().Host(host).Port(port).Db(db).Password(password)
	}

	client, e := redis.NewSynchClientWithSpec(spec)
	if e != nil {
		err = errors.New("Can not create redis connection")
		return err
	}
	this.client = client
	this.channel = channel
	this.conf = &RedisConf{
		Host:     host,
		Port:     port,
		Db:       db,
		Password: password,
	}
	go this.asyncWriter()
	return
}

func (this *Rlog) Write(p []byte) (n int, err error) {
	if !this.async {
		this.lock.Lock()
		defer this.lock.Unlock()
		count, err := this.client.Publish(this.channel, p)
		return int(count), err
	} else {
		select {
		case this.buffer <- p:
		case <-time.After(10 * time.Millisecond):
			fmt.Println(os.Stderr, "Write Buffer Error")
		}
		return 0, nil
	}
}

/*
  与redis重新建立连接
*/
func (this *Rlog) recon() {
	spec := redis.DefaultSpec().Host(this.conf.Host).Port(this.conf.Port).Db(this.conf.Db).Password(this.conf.Password)
	client, e := redis.NewSynchClientWithSpec(spec)
	if e != nil {
		fmt.Fprintln(os.Stderr, "Reconnect Error")
		return
	}
	this.client = client
}

func (this *Rlog) asyncWriter() {
	for p := range this.buffer {
		_, err := this.client.Publish(this.channel, p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Publish Error")
			this.recon()
		}
	}
}
