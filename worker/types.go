package main

import (
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

type Job struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	CronExpress string `json:"cronExpress"`
	Status      int    `json:"status"` // 1: RUNNING;  2: SLEEPING;  3: KILLING;  4: CLOSED
	HostName    string `json:"hostName"`
}

func (job *Job) JobInit() {
	job.Name, job.CronExpress, job.Status, job.HostName, job.Command = "", "", 0, "", ""
}

type JobScheduler struct {
	JobMap           sync.Map
	ScheduleDuration time.Duration
}
