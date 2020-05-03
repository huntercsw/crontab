package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
	"time"
)

var (
	getAllJobsResponse *clientv3.GetResponse
	jobStatusChan      = make(chan struct{}, 100)
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
		row     *mvccpb.KeyValue
		job     *Job
		jobPlan *JobPlan
	)

	if getAllJobsResponse, err = Etcd.cli.Get(ctx, JOB_PATH, clientv3.WithPrefix()); err != nil {
		return
	}

	for _, row = range getAllJobsResponse.Kvs {
		job = new(Job)
		if err = json.Unmarshal(row.Value, job); err != nil {
			msg := fmt.Sprintf("%s json unmarshal to job obj error: %v", string(row.Value), err)
			WorkerLogger.Error.Println(msg)
			err = errors.New(msg)
			break
		}
		if jobPlan, err = NewJobPlan(job); err != nil {
			break
		}
		jobScheduler.JobMap[string(row.Key)[len(JOB_PATH):]] = jobPlan
	}
	return
}

func (jobScheduler *JobScheduler) jobDirWatcher(ctx context.Context) {
	var (
		jobEventChan  clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		event         *clientv3.Event
		err           error
		job           *Job
	)
	jobEventChan = Etcd.cli.Watch(
		ctx,
		JOB_PATH,
		clientv3.WithRev(getAllJobsResponse.Header.Revision+1),
		clientv3.WithPrefix(),
	)
	for watchResponse = range jobEventChan {
		job = new(Job)
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
				if jobPlan, err1 := NewJobPlan(job); err1 == nil {
					jobScheduler.JobMap[job.Name] = jobPlan
				}
			case mvccpb.DELETE:
				fmt.Println(string(event.Kv.Key)[len(JOB_PATH):])
				delete(jobScheduler.JobMap, string(event.Kv.Key)[len(JOB_PATH):])
			}
		}
		jobStatusChan <- struct{}{}
	}
}

func (jobScheduler *JobScheduler) jobSchedule() {
	if len(jobScheduler.JobMap) == 0 {
		time.Sleep(time.Second)
		return
	}

	var (
		jobPlan     *JobPlan
		current     = time.Now()
		jobMapIndex = 0
	)

	for _, jobPlan = range jobScheduler.JobMap {
		if jobPlan.NextTime.Before(current) || jobPlan.NextTime.Equal(current) {
			go jobPlan.Job.jobExec()
			jobPlan.NextTime = jobPlan.CronExpress.Next(current)
		}

		if jobMapIndex == 0 {
			jobScheduler.NextTime = jobPlan.NextTime
			jobMapIndex += 1
			continue
		}

		if jobPlan.NextTime.Before(jobScheduler.NextTime) {
			jobScheduler.NextTime = jobPlan.NextTime
		}
	}

	jobScheduler.ScheduleDuration = jobScheduler.NextTime.Sub(current)
}

func (job *Job) jobExec() {
	fmt.Println(time.Now().String(), *job)
}

func (jobScheduler *JobScheduler) workerRun(ctx context.Context, wg sync.WaitGroup) {
	defer wg.Done()

	var (
		timer = time.NewTimer(jobScheduler.ScheduleDuration)
	)

	for {
		select {
		case <-jobStatusChan:
		case <-timer.C:
		}
		jobScheduler.jobSchedule()
		timer.Reset(jobScheduler.ScheduleDuration)
	}
}

func (jobScheduler *JobScheduler) jobSchedulerInit(ctx context.Context, wg sync.WaitGroup) (err error) {
	if err = jobScheduler.jobMapInit(ctx); err != nil {
		return
	}

	jobScheduler.jobSchedule()

	go jobScheduler.jobDirWatcher(ctx)

	go jobScheduler.workerRun(ctx, wg)

	return
}

func JobHandlerInit(ctx context.Context, wg sync.WaitGroup) (err error) {
	var (
		jobScheduler = new(JobScheduler)
	)
	jobScheduler.JobMap = make(map[string]*JobPlan, 1)

	err = jobScheduler.jobSchedulerInit(ctx, wg)

	return
}

