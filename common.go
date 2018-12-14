package go_tool

import (
	"crypto/md5"
	"fmt"
	"github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"sync"
	"time"
)

//引用的相关工程
// github.com/drone/routes
// github.com/BurntSushi/toml
// github.com/garyburd/redigo
// github.com/satori/go.uuid

//生成随机数
func RandInt(min int, max int) int {
	if min == max {
		return min
	}
	if min > max {
		t := min
		min = max
		max = t
	}
	rand.Seed(getRandSeek())
	return min + rand.Intn(max-min)
}

var randSeek = int64(1)
var randMax = int64(1000000)
var l sync.Mutex

//获取随机数种子值
func getRandSeek() int64 {
	l.Lock()
	if randSeek >= randMax {
		randSeek = 0
	}
	randSeek++
	l.Unlock()
	return time.Now().UnixNano() + randSeek
}

//生成GUID
func Guid() string {
	id := uuid.NewV4()
	return strings.ToUpper(id.String())
}

//获取字符串MD5
func Md5(s string) string {
	data := []byte(s)
	has := md5.Sum(data)
	md5Str := fmt.Sprintf("%X", has)
	return md5Str
}
