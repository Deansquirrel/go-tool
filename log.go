package go_tool

import (
	"log"
	"os"
	"time"
)

func Log(s string) error {
	fileName := GetDateStr(time.Now()) + ".log"
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	log.Println(s)
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func LogFile(s string, fileName string) error{
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	_,err = f.WriteString(s + "\n")
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}