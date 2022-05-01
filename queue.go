package smq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
	"sync"
	"time"
)

// 基于redis的发布订阅实现异步队列服务

const CacheTaskName = "queue:task"
const CacheFailName = "queue:fail"

// Task 消息接口
type Task struct {
	Payload func() interface{}
	Handler Handler
}

// Handler 处理任务handler接口
type Handler func(payload interface{}) error

type Queue struct {
	client   *redis.Client
	handlers map[string]*Task
	log      Log
	lock     sync.Mutex
}

// New 创建一个queue对象
// log 基于Log，只记录info、error两种类型的日志
// message 为消息结构体，在此处主要用于注册callback方法
func New(client *redis.Client, log Log, message ...*Task) *Queue {
	q := &Queue{client: client, handlers: map[string]*Task{}, log: log}
	for _, m := range message {
		q.Register(fmt.Sprintf("%T", m.Payload()), m)
	}
	return q
}

// Register 注册callback方法
func (q *Queue) Register(name string, task *Task) {
	q.lock.Lock()
	q.handlers[name] = task
	q.lock.Unlock()
}

// IsRegister 判断是否已经注册
func (q *Queue) IsRegister(name string) bool {
	_, ok := q.handlers[name]
	return ok
}

// Run 处理任务消息
func (q *Queue) Run() {
	var i time.Duration = 0
	for {
		str, err := q.client.RPop(context.Background(), CacheTaskName).Result()
		if err == nil {
			if str != "" {
				payload := strings.SplitN(str, "|", 2)

				if len(payload) != 2 {
					q.log.Error(fmt.Sprintf("format error %s", str))
					continue
				}

				if !q.IsRegister(payload[0]) {
					q.log.Error(fmt.Sprintf("handler %s not exist; %s", payload[0], payload[1]))
					continue
				}

				var data = q.handlers[payload[0]].Payload()
				err = json.Unmarshal([]byte(payload[1]), data)
				if err == nil {
					err = q.handlers[payload[0]].Handler(data)
				}
				if err == nil {
					i = 0
				} else {
					q.log.Error(fmt.Sprintf("handle failed; %s %s", str, err.Error()))
					_ = q.client.LPush(context.Background(), CacheFailName, str).Err()
				}
				continue
			}
		}
		if err != redis.Nil { // ??
			q.log.Error(fmt.Sprintf("client error %s", err.Error()))
		}

		i++
		i = i % 3
		if i > 0 {
			time.Sleep(i * time.Second)
		}
	}
}

// Delivery 投递任务消息
func (q *Queue) Delivery(m interface{}) {
	var err error
	name := fmt.Sprintf("%T", m)
	if q.IsRegister(name) {
		data, err := json.Marshal(m)
		if err == nil {
			err = q.client.LPush(context.Background(), CacheTaskName, fmt.Sprintf("%s|%s", name, string(data))).Err()
		}
	} else {
		err = errors.New(fmt.Sprintf("payload[%s] has not handler", name))
	}

	if err != nil {
		q.log.Error(err.Error())
	}
}
