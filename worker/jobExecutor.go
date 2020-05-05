package main

import (
	"fmt"
	"os/exec"
	"time"
)

var (
	executedJobMap = make(map[string]JobExecuted, 100)
)

func (executedJob *JobExecuted) JobExec() {
	var (
		cmd        *exec.Cmd
	)

	cmd = exec.CommandContext(executedJob.Ctx, "/bin/bash", "-c", executedJob.Job.Command)

	executedJob.StartTime = time.Now()
	if executedJob.OutPut, executedJob.Err = cmd.CombinedOutput(); executedJob.Err != nil {
		WorkerLogger.Error.Println(fmt.Sprintf("job[%s] error: %v", executedJob.Job.Name, executedJob.Err))
		return
	}
	executedJob.Pid = cmd.ProcessState.Pid()
	executedJob.ExitCode = cmd.ProcessState.ExitCode()
}
