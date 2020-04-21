package main

import (
	"context"
	"fmt"
	"os"
)

var (
	Conf    = new(CronTabConf)
	LogFile *os.File
	Logger  MyLog
	Etcd    = new(EtcdHandler)
	Mongo   = new(MongoDB)
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
	if err = Mongo.Init(); err != nil {
		Logger.Error(fmt.Sprintf("mongoDB init error: %v", err))
		os.Exit(1)
	}
}

func main() {
	defer func() {
		LogFile.Close()
		Etcd.cli.Close()
		Mongo.cli.Disconnect(context.Background())
	}()
	ctx, cancel := context.WithCancel(context.Background())
	//if objIDs, err := Mongo.InsertManyRecord(
	//	ctx,
	//	"test",
	//	"test_collection",
	//	"yinuo",
	//	"/usr/local/bin/yinuo",
	//	"",
	//	"hello"); err != nil{
	//	fmt.Println(err)
	//} else {
	//	for _, objID := range objIDs {
	//		fmt.Println(objID.(primitive.ObjectID))
	//	}
	//}
	if rsp, err := Mongo.Fetch(ctx); err != nil {
		fmt.Println(err)
	} else {
		for _, data := range rsp {
			fmt.Println(data)
		}
	}
	cancel()
}
