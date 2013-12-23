远程日志模块
------------

通过该模块实现Go语言的写远程日志

###功能

+ 利用Redis的subpub功能实现远程写日志的功能
+ 实现异步写和同步写两种方式


###安装方法

    go get github.com/cloudaice/rlog


###Example

+ 同步写方式

    package main
    
    import (
        "bytes"
        "fmt"
        "log"
    
        "github.com/cloudaice/rlog"
        "github.com/cloudaice/rlog/redis"
    )
    
    const (
        RedisHost string = "localhost"
        RedisPort int    = 6379
        RedisDb   int    = 8
        Password  string = "123456"
    )
    
    func main() {
        /*
            创建异步写远程对象
            remotelog := rlog.NewAsyncRlog() 
    
            创建同步写远程对象
            remotelog := rlog.NewRlog()
        */
        remotelog := rlog.NewAsyncRlog()
        err := remotelog.SetRedis(RedisHost, RedisPort, RedisDb, Password, "TEST_RLOG")
        if err != nil {
        	fmt.Println("Set Redis Error")
        }
        log.SetOutput(remotelog)
    
        spec := redis.DefaultSpec().Host(RedisHost).Port(RedisPort).Db(RedisDb).Password(Password)
        client, err := redis.NewPubSubClientWithSpec(spec)
        if err != nil {
        	fmt.Println("Subpub Redis Error")
        }
        e := client.Subscribe("TEST_RLOG")
        if e != nil {
        	fmt.Println("Sub Redis Error")
        }
        subChan := client.Messages("TEST_RLOG")
        log.Println("Hello World")
        msg := <-subChan
        if index := bytes.Index(msg, []byte("Hello World")); index == -1 {
        	fmt.Println("Not receive")
        } else {
        	fmt.Println("Receive")
        }
    }
