package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

func (job *Job) JobGetHandler(ctx context.Context) (count int64, kvs []*mvccpb.KeyValue, err error) {
	var getRsp *clientv3.GetResponse
	key := JOB_PATH + job.Name
	if getRsp, err = Etcd.cli.Get(ctx, key); err != nil {
		msg := fmt.Sprintf("JobGetHandler, get job by name from etcd error: %v", err)
		MasterLogger.Error.Println(msg)
		err = errors.New(fmt.Sprintf("get job[%s] by name from etcd error", job.Name))
		return
	}
	count, kvs = getRsp.Count, getRsp.Kvs
	return
}

func (job *Job) JobPostHandler(ctx context.Context) (err error) {
	var (
		count int64
		putRsp *clientv3.PutResponse
	)
	key := JOB_PATH + job.Name
	if count, _, err = job.JobGetHandler(ctx); err != nil {
		return
	}
	if count > 0 {
		err = errors.New(fmt.Sprintf("job[%s] has exist", key))
		return
	}
	value, _ := json.Marshal(job)
	if putRsp, err = Etcd.cli.KV.Put(ctx, key, string(value)); err != nil {
		msg := fmt.Sprintf("JobPostHandler, put job to etcd error: %v", err)
		MasterLogger.Error.Println(msg)
		err = errors.New("put job to etcd error")
		return
	}
	MasterLogger.Info.Println(fmt.Sprintf("add job[%s] to etcd, version: %d", key, putRsp.Header.Revision))
	return
}

func (job *Job) JobPutHandler(ctx context.Context) (err error) {
	var (
		count int64
		putRsp *clientv3.PutResponse
	)
	key := JOB_PATH + job.Name
	if count, _, err = job.JobGetHandler(ctx); err != nil {
		return
	}
	if count == 0 {
		err = errors.New(fmt.Sprintf("job[%s] dose not exist", key))
		return
	}
	value, _ := json.Marshal(job)
	if putRsp, err = Etcd.cli.Put(ctx, key, string(value), clientv3.WithPrevKV()); err != nil {
		msg := fmt.Sprintf("JobPuttHandler, update job[%s] error: %v", key, err)
		MasterLogger.Error.Println(msg)
		err = errors.New(fmt.Sprintf("edit job[%s] error", key))
		return
	}
	MasterLogger.Info.Println(fmt.Sprintf("update job[%s] from %s to %s", key, putRsp.PrevKv.Value, string(value)))
	return
}

func (job *Job) JobDeleteHandler(ctx context.Context) (err error) {
	var (
		delRsp *clientv3.DeleteResponse
	)
	key := JOB_PATH + job.Name
	if delRsp, err = Etcd.cli.Delete(ctx, key, clientv3.WithPrevKV()); err != nil {
		msg := fmt.Sprintf("delete job[%s] from etcd error: %v", key, err)
		MasterLogger.Error.Println(msg)
		err = errors.New(fmt.Sprintf("delete job[%s] error", key))
		return
	}
	if len(delRsp.PrevKvs) == 0{
		err = errors.New(fmt.Sprintf("job[%s] dose not exist", key))
		return
	}
	MasterLogger.Info.Println(fmt.Sprintf("delete job %s which value is %s", key, delRsp.PrevKvs[0].Value))
	return
}

func (job *Job) JobListHandler(ctx context.Context) (jobs []Job, err error) {
	var (
		getRsp *clientv3.GetResponse
		key = JOB_PATH + job.Name
	)

	if getRsp, err = Etcd.cli.Get(ctx, key, clientv3.WithPrefix()); err != nil {
		msg := fmt.Sprintf("get all jobs error: %v", err)
		MasterLogger.Error.Println(msg)
		err = errors.New("get all jobs error")
		return
	}
	for _, v := range getRsp.Kvs {
		j := new(Job)
		if err1 := json.Unmarshal(v.Value, j); err1 != nil {
			msg := fmt.Sprintf("%s json unmarshal to Job object error: %v", string(v.Value), err1)
			MasterLogger.Error.Println(msg)
			err = errors.New(msg)
			return
		}
		jobs = append(jobs, *j)
	}
	return
}

func (job *Job) JobKillHandler(ctx context.Context) (err error) {
	var (
		transaction clientv3.Txn
		txnRsp *clientv3.TxnResponse
		key string
	)

	if job.Status != JOB_STATUS_RUNNING {
		err = errors.New(fmt.Sprintf("job[%s] is not running", job.Name))
		return
	}

	key = JOB_PATH + job.HostName + "/" + job.Name
	transaction = Etcd.cli.Txn(ctx)
	transaction.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "kill")).
		Else(clientv3.OpGet(key))
	if txnRsp, err = transaction.Commit(); err != nil {
		MasterLogger.Error.Println(fmt.Sprintf("put kill signal of job [%s] to etcd error: %v", job.Name, err))
		err = errors.New(fmt.Sprintf("put kill signal of job[%s] to etcd error", job.Name))
		return
	}
	if !txnRsp.Succeeded {
		err = errors.New(fmt.Sprintf("kill signal of job[%s] has published", job.Name))
		return
	}
	return
}