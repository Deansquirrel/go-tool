package go_tool

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	err := Log("Log Test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogFile(t *testing.T) {
	err := LogFile("Log Test",GetDateStr(time.Now()) + ".log")
	if err != nil {
		t.Fatal(err)
	}
}
