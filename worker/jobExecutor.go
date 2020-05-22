package main

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"os/exec"
)

type JobRuntimeInfo struct {
	CpuIdle         float64
	MemAvailable    uint64
	MemUsedPercent  string
	SwapFree        uint64
	SwapUsedPercent string
}

func (jobRuntimeInfo *JobRuntimeInfo) GetJobRuntimeInfo() (err error) {
	var (
		//cpuInfo  []cpu.TimesStat
		memInfo  *mem.VirtualMemoryStat
		swapInfo *mem.SwapMemoryStat
	)

	if _, err = cpu.Times(false); err != nil {
		goto RESPONSE
	} else {
		//fmt.Println(cpuInfo)
	}

	if memInfo, err = mem.VirtualMemory(); err != nil {
		goto RESPONSE
	} else {
		jobRuntimeInfo.MemAvailable, jobRuntimeInfo.MemUsedPercent =
			memInfo.Available/1024/1024, fmt.Sprintf("%.2f", memInfo.UsedPercent*100)
	}

	if swapInfo, err = mem.SwapMemory(); err != nil {
		goto RESPONSE
	} else {
		jobRuntimeInfo.SwapFree, jobRuntimeInfo.SwapUsedPercent =
			swapInfo.Free/1024/1024/1024, fmt.Sprintf("%.2f", swapInfo.UsedPercent*100)
	}

RESPONSE:
	return
}

func (job *Job) JobExecInLinux(ctx context.Context) {
	var (
		cmd        *exec.Cmd
	)

	cmd = exec.CommandContext(ctx, "/bin/bash", "-c", job.Command)

	// TODO: update etcd
	//executedJob.StartTime = time.Now()
	if stdOut, stdErr := cmd.CombinedOutput(); stdErr != nil {
		WorkerLogger.Error.Println(fmt.Sprintf("job[%s] error: %v", job.Name, stdErr))
		fmt.Println(stdErr)
	} else {
		fmt.Println(stdOut)
	}
	//executedJob.Pid = cmd.ProcessState.Pid()
	//executedJob.ExitCode = cmd.ProcessState.ExitCode()
}

func (job *Job) JobExecInWin(ctx context.Context) {
	fmt.Println("Running Win cmd:", job.Command)
	jobRuntimeInfo := new(JobRuntimeInfo)
	cmd := exec.CommandContext(ctx,"cmd", "/c", job.Command)
	result, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}
	jobRuntimeInfo.GetJobRuntimeInfo()
	fmt.Println("job runtime:",jobRuntimeInfo)
	fmt.Println(string(result))
}

