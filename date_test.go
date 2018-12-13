package go_tool

import (
	"testing"
	"time"
)

func TestGetDateStr(t *testing.T) {
	got := GetDateStr(time.Now())
	if len(got) != 10 {
		t.Fatal("err")
	}
}

func TestGetDateTimeStr(t *testing.T) {
	got := GetDateTimeStr(time.Now())
	if len(got) != 19 {
		t.Fatal("err")
	}
}
