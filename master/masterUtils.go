package main

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ETCD struct {
	cli *clientv3.Client
}

type MONGO struct {
	cli *mongo.Client
}

func (e *ETCD) ETCDInit() (err error) {
	if e.cli, err = clientv3.New(clientv3.Config{
		Endpoints: MasterConf.EtcdEndPoints,
		DialTimeout: 2 * time.Second,
	}); err != nil {
		return
	}
	return
}

func (m *MONGO) MONGOInit() (err error) {
	var (
		ctx context.Context
	)
	uri := "mongodb://" + MasterConf.MongoDB
	ctx, _ = context.WithTimeout(context.TODO(), 5*time.Second)
	if m.cli, err = mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMaxPoolSize(5).SetMinPoolSize(2)); err != nil {
		return
	}
	return
}