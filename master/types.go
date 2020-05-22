package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorhill/cronexpr"
)

const (
	JOB_PATH = "/cron/jobs/"

	JOB_KILL_PATH = "/cron/job/control/cancel/"

	JOB_LOCK_PATH = "/cron/job/lock/"

	JOB_STATUS_RUNNING = 1

	JOB_STATUS_SLEEPING = 2

	JOB_STATUS_KILLING = 3

	JOB_STATUS_CLOSED = 4
)

type Job struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	CronExpress string `json:"cronExpress"`
	Status      int    `json:"status"`			// 1: RUNNING;  2: SLEEPING;  3: KILLING;  4: CLOSED
	HostName    string `json:"hostName"`
}

type JsonResponse struct {
	ErrorCode int
	Data      interface{}
}

func (job *Job) JobInit(name, command, cronExp, hostName string) {
	job.Name, job.Command, job.CronExpress, job.Status, job.HostName = name, command, cronExp, JOB_STATUS_CLOSED, hostName
}

func (job *Job) CronExpressionAnalysis() (cronExpr *cronexpr.Expression, err error) {
	if job.CronExpress != "" {
		if cronExpr, err = cronexpr.Parse(job.CronExpress); err != nil {
			MasterLogger.Error.Println(fmt.Sprintf("cronExpression of job[%s] analisys error: %v", job.Name, err))
		}
	}
	return
}

func (jr *JsonResponse) NewResponse(code int, data interface{}) (rsp []byte) {
	var err error
	jr.ErrorCode = code
	jr.Data = data
	if rsp, err = json.Marshal(jr); err != nil {
		MasterLogger.Error.Println(fmt.Sprintf("response marshal error: %v", err))
	}
	return
}
