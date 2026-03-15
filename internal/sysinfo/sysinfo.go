package sysinfo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemInfo struct {
	OS              string
	Arch            string
	CPUModel        string
	CPUCores        int
	CPUUsagePercent float64
	TotalRAMMB      uint64
	AvailableRAMMB  uint64
	Uptime          time.Duration
}

func Collect() (*SystemInfo, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("collect cpu info: %w", err)
	}

	cpuModel := "unknown"
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	cpuCores, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("collect cpu cores: %w", err)
	}

	// Đo CPU usage trong 1 giây — cần interval để có số có nghĩa
	cpuPercents, err := cpu.Percent(time.Second, false) // false = tổng hợp tất cả cores
	if err != nil {
		return nil, fmt.Errorf("collect cpu usage: %w", err)
	}
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("collect memory info: %w", err)
	}

	uptimeSeconds, err := host.Uptime()
	if err != nil {
		return nil, fmt.Errorf("collect uptime: %w", err)
	}

	s := &SystemInfo{
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		CPUModel:        cpuModel,
		CPUCores:        cpuCores,
		CPUUsagePercent: cpuUsage,
		TotalRAMMB:      memStat.Total / 1024 / 1024,
		AvailableRAMMB:  memStat.Available / 1024 / 1024,
		Uptime:          time.Duration(uptimeSeconds) * time.Second,
	}

	log.Info().
		Str("os", s.OS).
		Str("arch", s.Arch).
		Str("cpu_model", s.CPUModel).
		Int("cpu_cores", s.CPUCores).
		Float64("cpu_usage_percent", s.CPUUsagePercent).
		Uint64("total_ram_mb", s.TotalRAMMB).
		Uint64("available_ram_mb", s.AvailableRAMMB).
		Str("uptime", s.Uptime.String()).
		Msg("system information collected")

	return s, nil
}
