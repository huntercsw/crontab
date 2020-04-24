package main

import (
	"encoding/json"
	"fmt"
)

const (
	JOB_PATH = "/cron/jobs/"

	JOB_KILL_PATH = "/cron/job/control/cancel/"

	JOB_LOCK_PATH = "/cron/job/lock/"
)

type Job struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	CronExpress string `json:"cronExpress"`
	Status      bool   `json:"status"`
	HostName	string `json:"hostName"`
}

type JsonResponse struct {
	ErrorCode int
	Data      interface{}
}

func (job *Job) JobInit(name, command, cronExp string) {
	job.Name, job.Command, job.CronExpress, job.Status, job.HostName = name, command, cronExp, false, ""
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
