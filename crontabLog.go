package main

import (
	"log"
	"os"
)

type MyLog struct {
	Logger *log.Logger
}

func (myLog *MyLog) Info(msg string) {
	myLog.Logger.SetPrefix("Info\t")
	myLog.Logger.Println(msg)
}

func (myLog *MyLog) Error(msg string) {
	myLog.Logger.SetPrefix("Error\t")
	myLog.Logger.Println(msg)
}

func LogInit(path string) (err error){
	LogFile, err = os.OpenFile(path, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 666)
	if err != nil {
		return
	}
	Logger.Logger = log.New(LogFile, "", log.Llongfile)
	Logger.Logger.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	return
}