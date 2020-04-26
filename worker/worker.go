package main

import (
	"context"
	"flag"
	"fmt"
)

var (
	ConfPath string
	WorkerLogger = new(WorkerLog)
	Etcd = new(ETCD)
	Mongo = new(MONGO)
)

func ProgramArgumentsInit () {
	flag.StringVar(&ConfPath, "config", "./worker.conf", "specify the configuration of master")
	flag.Parse()
}

func main() {
	var (
		err error
	)

	ProgramArgumentsInit()

	WorkerConf.WorkerConfigInit()
	if WorkerConfigErr != nil {
		fmt.Println("crontab worker configuration error:", WorkerConfigErr)
		return
	}

	if err = WorkerLogger.LogInit(WorkerConf.LogPath); err != nil {
		fmt.Println("crontab worker logger initialization error:", err)
		return
	}

	if err = Etcd.EtcdInit(); err != nil {
		fmt.Println("crontab worker Etcd initialization error:", err)
		return
	}

	if err = Mongo.MongoInit(); err != nil {
		fmt.Println("crontab worker MongoDB initialization error:", err)
		return
	}

	defer func() {
		WorkerLogger.LogFile.Close()
		Etcd.cli.Close()
		Mongo.cli.Disconnect(context.TODO())
	}()

	fmt.Println("crontab worker started")
}