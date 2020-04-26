package main

import (
	"log"
	"os"
)

type WorkerLog struct {
	Logger *log.Logger
	LogFile *os.File
	Info log.Logger
	Error log.Logger
}

func (worker *WorkerLog) LogInit(path string) (err error){
	worker.LogFile, err = os.OpenFile(path, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 666)
	if err != nil {
		return
	}
	worker.Logger = log.New(worker.LogFile, "", log.Llongfile)
	worker.Logger.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	worker.Info, worker.Error = *worker.Logger, *worker.Logger
	worker.Info.SetPrefix("INFO\t")
	worker.Error.SetPrefix("ERROR\t")
	return
}