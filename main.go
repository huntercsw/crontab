package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	Conf    = new(CronTabConf)
	LogFile *os.File
	Logger  MyLog
	Etcd    = new(EtcdHandler)
)

func init() {
	var err error
	if err = Conf.Init(); err != nil {
		fmt.Println(fmt.Sprintf("load configuration error: %v", err))
		os.Exit(1)
	}
	if err = LogInit(Conf.Log.Path); err != nil {
		fmt.Println(fmt.Sprintf("log init error: %v", err))
		os.Exit(1)
	}
	if err = Etcd.Init(); err != nil {
		Logger.Error(fmt.Sprintf("etcd handler init error: %v", err))
		os.Exit(1)
	}
}

func main() {
	defer func() {
		LogFile.Close()
		Etcd.cli.Close()
	}()
	ctx, cancel := context.WithCancel(context.Background())
	go Etcd.Watch(ctx, "name")
	for i := 0; i < 10; i ++ {
		Etcd.kv.Put(ctx, "name", fmt.Sprintf("yinuo-%s", strconv.Itoa(i)))
		time.Sleep(time.Second)
	}
	cancel()
	time.Sleep(time.Second*2)
}
