# SMQ
> 封装一个简单的消息队列

## 使用

```bash
go get -u github.com/soonio/smq
```


## 用法示例

```go
package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/soonio/smq"
	"time"
)

type Log struct{}

func (l *Log) Info(msg string) {
	fmt.Println("[INFO]" + msg)
}

func (l *Log) Error(msg string) {
	fmt.Println("[ERROR]" + msg)
}

type Demo struct {
	Id       uint64 `json:"id"`
	Datetime string `json:"datetime"`
	Index    int    `json:"index"`
}

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1",
		Password: "123456", // no password set
		DB:       0,           // use default DB
	})
	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	queue := smq.New(client, &Log{}, &smq.Task{
		Payload: func() interface{} {
			return &Demo{}
		},
		Handler: func(payload interface{}) error {
			if v, ok := payload.(*Demo); ok {
				fmt.Printf("业务处理 %+v %s \n", v, time.Now().Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	})

	go queue.Run()

	go func() {
		i := 0
		for i < 10 {
			queue.Delivery(&Demo{
				Index:    i,
				Id:       10086,
				Datetime: time.Now().Format("2006-01-02 15:04:05"),
			})
			i++
			time.Sleep(time.Second)
		}
	}()

	select {}
}

```