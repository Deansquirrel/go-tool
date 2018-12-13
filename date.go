package go_tool

import (
	"time"
)

func GetDateStr(t time.Time) string {
	return t.Format("2006-01-02")
}

func GetDateTimeStr(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}