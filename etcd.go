package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"sync"
	"time"
)

type EtcdHandler struct {
	cli   *clientv3.Client
	kv    clientv3.KV
	lease clientv3.Lease
	watch clientv3.Watcher
}

func (etcd *EtcdHandler) Init() (err error) {
	etcdEndpoints := []string{}
	for _, endPoint := range Conf.Etcd.EndPoints {
		etcdEndpoints = append(etcdEndpoints, fmt.Sprintf("%s:%s", endPoint.Host, endPoint.Port))
	}
	if etcd.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 3 * time.Second,
	}); err != nil {
		return
	}

	etcd.kv = clientv3.NewKV(etcd.cli)
	etcd.lease = clientv3.NewLease(etcd.cli)
	etcd.watch = clientv3.NewWatcher(etcd.cli)
	return
}

func (etcd *EtcdHandler) Get(ctx context.Context, keyName string, opts ...clientv3.OpOption) (res *clientv3.GetResponse, err error) {
	if res, err = etcd.kv.Get(ctx, keyName, opts...); err != nil {
		Logger.Error(fmt.Sprintf("get key[%s] from etcd error: %v", keyName, err))
		return
	}
	return
}

func (etcd *EtcdHandler) Put(ctx context.Context, k string, v string, opts ...clientv3.OpOption) (rsp *clientv3.PutResponse, err error) {
	if rsp, err = etcd.kv.Put(ctx, k, v, opts...); err != nil {
		Logger.Error(fmt.Sprintf("put key[%s] value[%s] to etcd error: %v", k, v, err))
		return
	}
	return
}

func (etcd *EtcdHandler) PutWithLease(ctx context.Context, k string, v string, expirationTime int64) (err error) {
	var (
		leaseGrantRsp *clientv3.LeaseGrantResponse
		putRsp        *clientv3.PutResponse
	)
	if leaseGrantRsp, err = etcd.lease.Grant(ctx, expirationTime); err != nil {
		return
	}
	leaseId := leaseGrantRsp.ID
	if putRsp, err = etcd.kv.Put(ctx, k, v, clientv3.WithLease(leaseId)); err != nil {
		return
	} else {
		Logger.Info(fmt.Sprintf("insert with lease key[%s] succeed: %v", k, putRsp.Header))
		return
	}
}

func (etcd *EtcdHandler) PutWithLeaseKeepAlive(ctx context.Context, k string, v string, expirationTime int64) (err error) {
	// KeepAlive在当前函数的生命周期内将一直有效，或者当KeepAlive收到cancel信号的时候也将失效
	var leaseGrantRsp *clientv3.LeaseGrantResponse
	if leaseGrantRsp, err = etcd.lease.Grant(ctx, expirationTime); err != nil {
		return
	}
	leaseId := leaseGrantRsp.ID
	if _, err = etcd.lease.KeepAlive(ctx, leaseId); err != nil {
		return
	}
	/*
		var leaseKeepAliveRsp <-chan *clientv3.LeaseKeepAliveResponse
		if leaseKeepAliveRsp, err = etcd.lease.KeepAlive(ctx, leaseId); err != nil {
			for {
				select {
				case rsp := <-leaseKeepAliveRsp:
					if rsp != nil {			// 租约失效，channel被删除，将读取到nil
						fmt.Println(rsp)
					}
				}
			}
		}
	*/
	if _, err = etcd.kv.Put(ctx, k, v, clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}

func (etcd *EtcdHandler) FetchAllWithPrefix(ctx context.Context, k string) (kvs sync.Map, err error) {
	var (
		getRsp *clientv3.GetResponse
	)
	if getRsp, err = etcd.Get(ctx, k, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, items := range getRsp.Kvs {
		var key string
		if len(k) == len(items.Key) {
			key = k
		} else {
			key = string(items.Key)[len(k):]
		}
		kvs.Store(key, string(items.Value))
	}
	return
}

func (etcd *EtcdHandler) DeleteOne(ctx context.Context, k string) (key string, exist bool, err error) {
	var (
		delRsp *clientv3.DeleteResponse
	)
	if delRsp, err = etcd.kv.Delete(ctx, k, clientv3.WithPrevKV()); err != nil {
		return
	}
	if len(delRsp.PrevKvs) == 0 {
		exist = false
		return
	} else {
		exist = true
		key = string(delRsp.PrevKvs[0].Key)
		return
	}
}

func (etcd *EtcdHandler) DeleteWithPrefix(ctx context.Context, k string) (kvs []*mvccpb.KeyValue, exist bool, err error) {
	var (
		delRsp *clientv3.DeleteResponse
	)
	if delRsp, err = etcd.kv.Delete(ctx, k, clientv3.WithPrevKV(), clientv3.WithPrefix()); err != nil {
		return
	} else {
		kvs = delRsp.PrevKvs
		if len(kvs) == 0 {
			exist = false
			return
		} else {
			exist = true
			return
		}
	}
}

func (etcd *EtcdHandler) Watch(ctx context.Context, k string, opts ...interface{}) (err error) {
	defer func() {
		fmt.Println("watch goroutine done")
	}()
	var (
		watchChan  clientv3.WatchChan
		getRsp     *clientv3.GetResponse
		keyVersion int64
	)
	if getRsp, err = etcd.kv.Get(ctx, k); err != nil {
		return
	}
	if getRsp.Count == 0 {
		return errors.New(fmt.Sprintf("key[%s] dose not exist", k))
	}
	keyVersion = getRsp.Header.Revision + 1 //watch from next revision
	watchChan = etcd.watch.Watch(ctx, k, clientv3.WithRev(keyVersion))
	for watchResponse := range watchChan {
		for _, event := range watchResponse.Events {
			fmt.Println(event)
		}
	}
	return
}
