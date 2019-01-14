package go_tool

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func Log(s string) error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		dir = ""
	} else {
		dir = dir + "\\"
	}
	fileName := dir + "" + GetDateStr(time.Now()) + ".log"
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

func LogFile(s string, fileName string) error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		dir = ""
	} else {
		dir = dir + "\\"
	}
	fileName = dir + fileName
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	_, err = f.WriteString(s + "\n")
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}
