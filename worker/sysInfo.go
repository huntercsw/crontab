package main

import (
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"net"
	"os"
	"runtime"
)

type SysInfo struct {
	HoseName      string
	IPAddress     string
	CpuLogicCores int
	MemTotal      uint64
	SwapTotal     uint64
	DiscTotal     uint64
	DiscFree      uint64
	DiscPercent   string
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
		sysInfo.MemTotal = memInfo.Total / 1024 / 1024 / 1024
	}

	if swapInfo, err = mem.SwapMemory(); err != nil {
		goto RESPONSE
	} else {
		sysInfo.SwapTotal = swapInfo.Total / 1024 / 1024 / 1024
	}

	if discInfo, err = disk.Usage("/"); err != nil {
		goto RESPONSE
	} else {
		sysInfo.DiscTotal, sysInfo.DiscFree, sysInfo.DiscPercent =
			discInfo.Total/1024/1024/1024, discInfo.Free/1024/1024/1024, fmt.Sprintf("%.2f", discInfo.UsedPercent*100)
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

		if ipAddr.IP.To4() != nil { // IPV4
			sysInfo.IPAddress = ipAddr.IP.String() // 192.168.1.1
			return nil
		}
	}
	return
}

