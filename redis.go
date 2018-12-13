package go_tool

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

type MyRedis struct{
	Server string
	Auth string
	MaxIdle int
	MaxActive int
	IdleTimeout int
}

var pool *redis.Pool

//创建连接池
func (myRedis *MyRedis)newPool() *redis.Pool{

	return &redis.Pool{
		MaxIdle:myRedis.MaxIdle,
		MaxActive:myRedis.MaxActive,
		IdleTimeout:time.Duration(1000 * 1000 * 1000 * myRedis.IdleTimeout),
		Dial:func()(redis.Conn,error){
			c,err := redis.Dial("tcp",myRedis.Server)
			if err != nil {
				return nil,err
			}
			_,err = c.Do("auth",myRedis.Auth)
			if err != nil {
				c.Close()
				return nil,err
			}
			return c,nil
		},
		TestOnBorrow: func(c redis.Conn,t time.Time) error {
			if time.Since(t) < time.Minute{
				return nil
			}
			_,err := c.Do("PING")
			return err
		},
	}
}

//插入
func (myRedis *MyRedis)Set(db string,key string, value string)(result string,err error){
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	defer conn.Close()
	_,err = conn.Do("SELECT",db)
	if err != nil {
		return
	}
	result,err = redis.String(conn.Do("SET",key,value))
	return
}

//查询
func (myRedis *MyRedis)Get(db string,key string) (result string,err error){
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	defer conn.Close()
	_,err = conn.Do("SELECT",db)
	if err != nil {
		return
	}
	result,err = redis.String(conn.Do("GET",key))
	return
}

//检查是否存在
func (myRedis *MyRedis)IsExists(db string,key string)(result bool,err error){
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	defer conn.Close()
	_,err = conn.Do("SELECT",db)
	if err != nil {
		return
	}
	result,err = redis.Bool(conn.Do("EXISTS",key))
	return
}

//删除
func (myRedis *MyRedis)Del(db string,key string)(err error){
	if pool == nil {
		pool = myRedis.newPool()
	}
	conn := pool.Get()
	defer conn.Close()
	_,err = conn.Do("SELECT",db)
	if err != nil {
		return
	}
	_,err = conn.Do("DEL",key)
	return
}
