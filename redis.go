package go_tool

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

type MyRedis struct {
	server      string
	auth        string
	maxIdle     int
	maxActive   int
	idleTimeout int
}

var pool *redis.Pool

func NewRedis(server string,auth string,maxIdle int,maxActive int,idleTimeout int) *MyRedis{
	redis := &MyRedis{
		server:server,
		auth:auth,
		maxIdle:maxIdle,
		maxActive:maxActive,
		idleTimeout:idleTimeout,
	}
	redis.newPool()
	return redis
}


func (myRedis *MyRedis)GetConfigJson()(string,error){
	configStr,err := json.Marshal(myRedis)
	if err != nil {
		return "",err
	}
	return string(configStr),nil
}

//创建连接池
func (myRedis *MyRedis) newPool() *redis.Pool {

	return &redis.Pool{
		MaxIdle:     myRedis.maxIdle,
		MaxActive:   myRedis.maxActive,
		IdleTimeout: time.Duration(1000 * 1000 * 1000 * myRedis.idleTimeout),
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", myRedis.server)
			if err != nil {
				return nil, err
			}
			_, err = c.Do("auth", myRedis.auth)
			if err != nil {
				errLs := c.Close()
				if errLs != nil {
					fmt.Println(errLs)
				}
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

//插入
func (myRedis *MyRedis) Set(db string, key string, value string) (result string, err error) {
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	defer func() {
		errLs := conn.Close()
		if errLs != nil {
			fmt.Println(errLs)
		}
	}()
	_, err = conn.Do("SELECT", db)
	if err != nil {
		return
	}
	result, err = redis.String(conn.Do("SET", key, value))
	return
}

//查询
func (myRedis *MyRedis) Get(db string, key string) (result string, err error) {
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	//defer func(){
	//	err = conn.Close()
	//	if err != nil {
	//		global.MyLog(err.Error())
	//	}
	//}()
	defer func() {
		errLs := conn.Close()
		if errLs != nil {
			fmt.Println(errLs)
		}
	}()
	_, err = conn.Do("SELECT", db)
	if err != nil {
		return
	}
	result, err = redis.String(conn.Do("GET", key))
	return
}

//检查是否存在
func (myRedis *MyRedis) IsExists(db string, key string) (result bool, err error) {
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	defer func() {
		errLs := conn.Close()
		if errLs != nil {
			fmt.Println(errLs)
		}
	}()
	_, err = conn.Do("SELECT", db)
	if err != nil {
		return
	}
	result, err = redis.Bool(conn.Do("EXISTS", key))
	return
}

//删除
func (myRedis *MyRedis) Del(db string, key string) (err error) {
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	defer func() {
		errLs := conn.Close()
		if errLs != nil {
			fmt.Println(errLs)
		}
	}()
	_, err = conn.Do("SELECT", db)
	if err != nil {
		return
	}
	_, err = conn.Do("DEL", key)
	return
}
