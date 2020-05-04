package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"net"
	"os"
	"runtime"
)

type SysInfo struct {
	HoseName      string
	IPAddress     string
	IPMaskAddr    string
	CpuLogicCores int
	MemTotal      uint64
	SwapTotal     uint64
	DiscTotal     uint64
	DiscFree      uint64
	DiscPercent   float64
}

type SysInstantaneousInfo struct {
	CpuIdle         float64
	MemAvailable     uint64
	MemUsedPercent  float64
	SwapFree        uint64
	SwapUsedPercent float64
}

func (sysInfo *SysInfo) SysInfoInit() (err error) {
	var (
		memInfo  *mem.VirtualMemoryStat
		swapInfo *mem.SwapMemoryStat
		discInfo *disk.UsageStat
	)

	if sysInfo.HoseName, err = os.Hostname(); err != nil {
		goto RESPONSE
	}

	if err = sysInfo.GetLocalIP(); err != nil {
		goto RESPONSE
	}

	sysInfo.CpuLogicCores = runtime.NumCPU()

	if memInfo, err = mem.VirtualMemory(); err != nil {
		goto RESPONSE
	} else {
		sysInfo.MemTotal = memInfo.Total
	}

	if swapInfo, err = mem.SwapMemory(); err != nil {
		goto RESPONSE
	} else {
		sysInfo.SwapTotal = swapInfo.Total
	}

	if discInfo, err = disk.Usage("/"); err != nil {
		goto RESPONSE
	} else {
		sysInfo.DiscTotal, sysInfo.DiscFree, sysInfo.DiscPercent = discInfo.Total, discInfo.Free, discInfo.UsedPercent
	}

RESPONSE:
	return
}

func (sysInfo *SysInfo) GetLocalIP() (err error) {
	var (
		addrs []net.Addr
		addr  net.Addr
	)

	addrs, err = net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr = range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		sysInfo.IPAddress = ipAddr.IP.String()
		sysInfo.IPMaskAddr = ipAddr.Mask.String()
	}
	return
}

func (sysInstantaneousInfo *SysInstantaneousInfo) SysInstantaneousInit() (err error) {
	var (
		cpuInfo  []cpu.TimesStat
		memInfo  *mem.VirtualMemoryStat
		swapInfo *mem.SwapMemoryStat
	)

	if cpuInfo, err = cpu.Times(false); err != nil {
		goto RESPONSE
	} else {
		fmt.Println(cpuInfo)
	}

	if memInfo, err = mem.VirtualMemory(); err != nil {
		goto RESPONSE
	} else {
		sysInstantaneousInfo.MemAvailable, sysInstantaneousInfo.MemUsedPercent =
			memInfo.Available, memInfo.UsedPercent
	}

	if swapInfo, err = mem.SwapMemory(); err != nil {
		goto RESPONSE
	} else {
		sysInstantaneousInfo.SwapFree, sysInstantaneousInfo.SwapUsedPercent = swapInfo.Free, swapInfo.UsedPercent
	}

	RESPONSE:
	return
}

