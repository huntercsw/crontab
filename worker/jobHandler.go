package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
)

var (
	getAllJobsResponse *clientv3.GetResponse
)

func (jobScheduler *JobScheduler) jobMapInit(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("get all jobs function panic: %v", r)
			WorkerLogger.Error.Println(msg)
			err = errors.New(msg)
		}
	}()

	var (
		row *mvccpb.KeyValue
		job = new(Job)
	)

	if getAllJobsResponse, err = Etcd.cli.Get(ctx, JOB_PATH, clientv3.WithPrefix()); err != nil {
		return
	}

	for _, row = range getAllJobsResponse.Kvs {
		if err = json.Unmarshal(row.Value, job); err != nil {
			msg := fmt.Sprintf("%s json unmarshal to job obj error: %v", string(row.Value), err)
			WorkerLogger.Error.Println(msg)
			err = errors.New(msg)
			break
		}
		jobScheduler.JobMap.Store(string(row.Key)[len(JOB_PATH):], *job)
	}
	return
}

func (jobScheduler *JobScheduler) jobMapAppend(job *Job) (err error) {
	_, ok := jobScheduler.JobMap.LoadOrStore(job.Name, job) //若key已存在，则返回true和key对应的value，不会修改原来的value
	if ok {
		err = errors.New(fmt.Sprintf("job[%s] has exist", job.Name))
	}
	return
}

func (jobScheduler *JobScheduler) jobMapUpdate(job *Job) (err error) {
	_, exist := jobScheduler.JobMap.Load(job.Name)
	if !exist {
		err = errors.New(fmt.Sprintf("job[%s] dose not exist", job.Name))
	} else {
		jobScheduler.JobMap.Store(job.Name, job)
	}
	return
}

func (jobScheduler *JobScheduler) jobMapDelete(job *Job) (err error) {
	_, exist := jobScheduler.JobMap.Load(job.Name)
	if !exist {
		err = errors.New(fmt.Sprintf("job[%s] dose not exist", job.Name))
	} else {
		jobScheduler.JobMap.Delete(job.Name)
	}
	return
}

func (jobScheduler *JobScheduler) jobDirWatcher(ctx context.Context, wg sync.WaitGroup) {
	defer wg.Done()
	var (
		jobEventChan  clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		event         *clientv3.Event
		err           error
		job           = new(Job)
	)
	jobEventChan = Etcd.cli.Watch(
		ctx,
		JOB_PATH,
		clientv3.WithRev(getAllJobsResponse.Header.Revision+1),
		clientv3.WithPrefix(),
	)
	for watchResponse = range jobEventChan {
		for _, event = range watchResponse.Events {
			switch event.Type {
			case mvccpb.PUT:
				if err = json.Unmarshal(event.Kv.Value, job); err != nil {
					WorkerLogger.Error.Printf(
						"jobDirWatcher jsonUnmarshal job[%s] to job object error: %v",
						string(event.Kv.Value),
						err,
					)
					job.JobInit()
				}
				jobScheduler.JobMap.Store(job.Name, *job)
			case mvccpb.DELETE:
				fmt.Println(string(event.Kv.Key)[len(JOB_PATH):])
				jobScheduler.JobMap.Delete(string(event.Kv.Key)[len(JOB_PATH):])
			}
		}
		jobScheduler.JobMap.Range(func(k, v interface{}) bool {
			fmt.Println(v)
			return true
		})
		job.JobInit()
	}
}

func (jobScheduler *JobScheduler) jobSchedulerInit(ctx context.Context, wg sync.WaitGroup) (err error) {
	if err = jobScheduler.jobMapInit(ctx); err != nil {
		return
	}
	jobScheduler.JobMap.Range(func(k, v interface{}) bool {
		fmt.Println(k, v)
		return true
	})

	go jobScheduler.jobDirWatcher(ctx, wg)

	return
}

func JobHandlerInit(ctx context.Context, wg sync.WaitGroup) (err error) {
	var (
		jobScheduler = new(JobScheduler)
	)

	err = jobScheduler.jobSchedulerInit(ctx, wg)

	return
}
