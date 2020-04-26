package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/gorhill/cronexpr"
	"sync"
	"time"
)

type Job struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	CronExpress string `json:"cronExpress"`
	Status      int    `json:"status"`
	HostName    string `json:"hostName"`
}

func cronTest (cronExpression string) {
	var (
		err error
		cronResult *cronexpr.Expression
	)

	if cronResult, err = cronexpr.Parse(cronExpression); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(cronResult.Next(time.Now()))
	}
}

func multiCronJob (ctx context.Context, wg *sync.WaitGroup) {
	var (
		etcd *clientv3.Client
		getRsp *clientv3.GetResponse
		err error
		cronJobs = make(map[string]*cronexpr.Expression)
	)
	if etcd, err = clientv3.New(clientv3.Config{Endpoints:[]string{"192.168.1.151:2379"}}); err != nil {
		fmt.Println(err)
		return
	}
	if getRsp, err = etcd.Get(ctx, "/cron/jobs/", clientv3.WithPrefix()); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(getRsp.Count)
	}
	for _, v := range getRsp.Kvs {
		job := new(Job)
		json.Unmarshal(v.Value, job)
		if cron, err1 := cronexpr.Parse(job.CronExpress); err1 != nil {
			fmt.Println(err1)
			return
		} else {
			cronJobs[string(v.Key)] = cron
		}
	}
	fmt.Println(cronJobs)

	for jobName, cronExpression := range cronJobs{
		go func(ctx1 context.Context, jobName string, cronExpression *cronexpr.Expression, wg *sync.WaitGroup) {
			defer wg.Done()
			nextTime := cronExpression.Next(time.Now())
			CycleExit:
			for {
				select {
				case <-ctx.Done():
					fmt.Println(fmt.Sprintf("%s receive signal cancel from ctx", jobName))
					break CycleExit
				default:
					now := time.Now()
					if nextTime.Before(now) || nextTime.Equal(now) {
						fmt.Println(time.Now().String(), now, jobName)
						nextTime = cronExpression.Next(time.Now())
					} else {
					}
					time.Sleep(time.Millisecond*1000)
				}
			}

		}(ctx, jobName, cronExpression, wg)
	}

}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(5)
	multiCronJob(ctx, &wg)
	time.Sleep(time.Second*10)
	cancel()
	wg.Wait()
}
