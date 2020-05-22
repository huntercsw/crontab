package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorhill/cronexpr"
	"runtime"
	"sync"
	"time"
)

const (
	JOB_PATH            = "/cron/jobs/"
	JOB_KILL_PATH       = "/cron/job/control/cancel/"
	JOB_LOCK_PATH       = "/cron/job/lock/"
	JOB_STATUS_RUNNING  = 1
	JOB_STATUS_SLEEPING = 2
	JOB_STATUS_KILLING  = 3
	JOB_STATUS_CLOSED   = 4
)

var (
	jobDirChanged = make(chan struct{}, 100)
)

type Job struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	CronExpress string `json:"cronExpress"`
	Status      int    `json:"status"` // 1: RUNNING;  2: SLEEPING;  3: KILLING;  4: CLOSED
	HostName    string `json:"hostName"`
	NextTime    time.Time
}

type JobHash struct {
	JobMap   map[string]*Job
	NextScan *time.Time
}

func (job *Job) getNextExecTime() (err error) {
	var (
		cronExpr *cronexpr.Expression
	)
	if cronExpr, err = cronexpr.Parse(job.CronExpress); err != nil {
		WorkerLogger.Error.Println(fmt.Sprintf("cronExpression of job[%s] analisys error: %v", job.Name, err))
		return
	}
	job.NextTime = cronExpr.Next(time.Now())
	return
}

func (jobHash *JobHash) jobHashInit(ctx context.Context, sysInfo *SysInfo) (getAllJobsResponse *clientv3.GetResponse, err error) {
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
		isFirst = true
	)

	if getAllJobsResponse, err = Etcd.cli.Get(ctx, JOB_PATH, clientv3.WithPrefix()); err != nil {
		return
	}

	jobHash.JobMap = make(map[string]*Job)
	for _, row = range getAllJobsResponse.Kvs {
		job = new(Job)
		if err = json.Unmarshal(row.Value, job); err != nil {
			msg := fmt.Sprintf("%s json unmarshal to job obj error: %v", string(row.Value), err)
			WorkerLogger.Error.Println(msg)
			err = errors.New(msg)
			break
		}

		if job.HostName != sysInfo.HoseName && job.HostName != sysInfo.IPAddress {
			continue
		}

		if job.CronExpress == "" {
			continue
		}

		if err = job.getNextExecTime(); err != nil {
			break
		}

		if isFirst {
			jobHash.NextScan = &job.NextTime
			isFirst = false
			fmt.Println(jobHash.NextScan)
		} else {
			if job.NextTime.Before(*jobHash.NextScan) {
				jobHash.NextScan = &job.NextTime
			}
		}
		jobHash.JobMap[string(row.Key)[len(JOB_PATH):]] = job
	}
	return
}

func (jobHash *JobHash) jobDirWatcher(ctx context.Context, getAllJobResponse *clientv3.GetResponse) {
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
		clientv3.WithRev(getAllJobResponse.Header.Revision+1),
		clientv3.WithPrefix(),
	)

	for watchResponse = range jobEventChan {
		for _, event = range watchResponse.Events {
			switch event.Type {
			case mvccpb.PUT:
				job = new(Job)
				if err = json.Unmarshal(event.Kv.Value, job); err != nil {
					WorkerLogger.Error.Printf(
						"jobDirWatcher jsonUnmarshal job[%s] to job object error: %v",
						string(event.Kv.Value),
						err,
					)
				}
				if err1 := job.getNextExecTime(); err1 == nil {
					jobHash.JobMap[job.Name] = job
				}
			case mvccpb.DELETE:
				fmt.Println(string(event.Kv.Key)[len(JOB_PATH):])
				delete(jobHash.JobMap, string(event.Kv.Key)[len(JOB_PATH):])
			}
		}
		jobDirChanged <- struct{}{}
	}
}

func (jobHash *JobHash) scanJobHashAndRanJob(ctx context.Context) {
	var (
		nextScan = jobHash.NextScan
	)

	for _, job := range jobHash.JobMap {
		if job.NextTime.Before(*jobHash.NextScan) || job.NextTime.Equal(*jobHash.NextScan) {
			go func() {
				if runtime.GOOS == "windows" {
					job.JobExecInWin(ctx)
				} else {
					job.JobExecInLinux(ctx)
				}
			}()

			if err := job.getNextExecTime(); err != nil {
				break
			}
		}

		if job.NextTime.Before(*nextScan) {
			nextScan = &job.NextTime
		}
	}

	if nextScan.Before(*jobHash.NextScan) {
		jobHash.NextScan = nextScan
	}
}

func JobSchedulerInit(ctx context.Context, wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		jobHash           = new(JobHash)
		sysInfo           = new(SysInfo)
		getAllJobResponse *clientv3.GetResponse
		timer             *time.Timer
	)

	sysInfo.SysInfoInit()

	if getAllJobResponse, err = jobHash.jobHashInit(ctx, sysInfo); err != nil {
		return
	}
	timer = time.NewTimer(jobHash.NextScan.Sub(time.Now()))

	go jobHash.jobDirWatcher(ctx, getAllJobResponse)

	go func() {
		for {
			select {
			case <-jobDirChanged:
			case <-timer.C:
			}
			jobHash.scanJobHashAndRanJob(ctx)
			timer.Reset(jobHash.NextScan.Sub(time.Now()))
		}
	}()

	wg.Wait()
	return
}

