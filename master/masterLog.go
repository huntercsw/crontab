package main

import (
	"log"
	"os"
)

type MasterLog struct {
	Logger *log.Logger
	LogFile *os.File
	Info log.Logger
	Error log.Logger
}

func (master *MasterLog) LogInit(path string) (err error){
	master.LogFile, err = os.OpenFile(path, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 666)
	if err != nil {
		return
	}
	master.Logger = log.New(master.LogFile, "", log.Llongfile)
	master.Logger.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	master.Info, master.Error = *master.Logger, *master.Logger
	master.Info.SetPrefix("INFO\t")
	master.Error.SetPrefix("ERROR\t")
	return
}