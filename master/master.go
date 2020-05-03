package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"
)

var MasterLogger MasterLog
var StopChan = make(chan struct{})
var ArgumentConfPath string
var Etcd ETCD
var Mongo MONGO

func init() {

}

func ProgramArgumentsInit() {
	flag.StringVar(
		&ArgumentConfPath,
		"config",
		"./master.conf",
		"specify the configuration of master, default: ./master.conf")
	flag.Parse()
}

func main() {
	var (
		err    error
		stopWQ = sync.WaitGroup{}
		ctx    context.Context
	)
	stopWQ.Add(1)
	ctx = context.Background()

	ProgramArgumentsInit()

	MasterConfigInit()
	if MasterConfigErr != nil {
		fmt.Println("Loading Master Configuration Error:", MasterConfigErr)
		return
	}

	if err = MasterLogger.LogInit(MasterConf.LogPath); err != nil {
		fmt.Println("master log init error:", err)
		return
	}

	if err = Etcd.ETCDInit(); err != nil {
		fmt.Println("etcd init error:", err)
		return
	}

	if err = Mongo.MONGOInit(); err != nil {
		fmt.Println("mongo init error", err)
		return
	}

	defer func() {
		MasterLogger.LogFile.Close()
		Etcd.cli.Close()
		Mongo.cli.Disconnect(ctx)
	}()

	WebServerInit()
	if WebServerInitErr != nil {
		fmt.Println("Web Server Init Error:", WebServerInitErr)
		return
	}

	go func() {
		defer stopWQ.Done()
		for {
			select {
			case <-StopChan:
				return
			default:
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	fmt.Println("Corntab Master Running")
	stopWQ.Wait()
}
