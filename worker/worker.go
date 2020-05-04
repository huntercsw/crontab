package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
)

var (
	ConfPath     string
	WorkerLogger = new(WorkerLog)
	Etcd         = new(ETCD)
	Mongo        = new(MONGO)
)

func ProgramArgumentsInit() {
	flag.StringVar(
		&ConfPath,
		"config",
		"./worker.conf",
		"specify the configuration of master, default: ./worker.conf",
		)
	flag.Parse()
}

func main() {
	var (
		err error
		ctx context.Context
		wg sync.WaitGroup
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
	defer WorkerLogger.LogFile.Close()

	if err = Etcd.EtcdInit(); err != nil {
		fmt.Println("crontab worker Etcd initialization error:", err)
		return
	}
	defer Etcd.cli.Close()

	if err = Mongo.MongoInit(); err != nil {
		fmt.Println("crontab worker MongoDB initialization error:", err)
		return
	}
	defer Mongo.cli.Disconnect(context.TODO())

	ctx = context.Background()
	wg.Add(1)
	if err = JobHandlerInit(ctx, wg); err != nil {
		fmt.Println("JobHandlerInit error", err)
		return
	}

	fmt.Println("crontab worker started")

	sysInfo, sysInstantaneousInfo := new(SysInfo), new(SysInstantaneousInfo)
	sysInfo.SysInfoInit()
	fmt.Println(sysInfo)
	sysInstantaneousInfo.SysInstantaneousInit()
	fmt.Println(sysInstantaneousInfo)

	wg.Wait()
}
