package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
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

func (job *Job) CronExpressionAnalysis() (cronExpr *cronexpr.Expression, err error) {
	if cronExpr, err = cronexpr.Parse(job.CronExpress); err != nil {
		WorkerLogger.Error.Println(fmt.Sprintf("cronExpression of job[%s] analisys error: %v", job.Name, err))
	}
	return
}

type JobPlan struct {
	Job         *Job
	CronExpress *cronexpr.Expression
	NextTime    time.Time
}

func NewJobPlan(job *Job) (jobPlan *JobPlan, err error) {
	var (
		cronExpr *cronexpr.Expression
	)
	jobPlan = new(JobPlan)

	if cronExpr, err = cronexpr.Parse(job.CronExpress); err != nil {
		WorkerLogger.Error.Println(fmt.Sprintf("cronExpression of job[%s] analysis error: %v", job.Name, err))
		return
	}

	jobPlan.Job, jobPlan.CronExpress, jobPlan.NextTime = job, cronExpr, cronExpr.Next(time.Now())

	return
}

type JobScheduler struct {
	JobMap           map[string]*JobPlan
	ScheduleDuration time.Duration
	NextTime         time.Time
}
