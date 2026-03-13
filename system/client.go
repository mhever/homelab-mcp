package system

import (
	"context"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemClient interface {
	HostInfo(ctx context.Context) (*host.InfoStat, error)
	CPUPercent(ctx context.Context) ([]float64, error)
	VirtualMemory(ctx context.Context) (*mem.VirtualMemoryStat, error)
	LoadAvg(ctx context.Context) (*load.AvgStat, error)
	DiskPartitions(ctx context.Context) ([]disk.PartitionStat, error)
	DiskUsage(ctx context.Context, path string) (*disk.UsageStat, error)
}

type GopsutilClient struct{}

func (c *GopsutilClient) HostInfo(ctx context.Context) (*host.InfoStat, error) {
	return host.InfoWithContext(ctx)
}

// CPUPercent returns per-CPU usage. Interval=0 returns data since last call
// (or since boot on first call), avoiding a blocking sleep on every invocation.
func (c *GopsutilClient) CPUPercent(ctx context.Context) ([]float64, error) {
	return cpu.PercentWithContext(ctx, 0, true)
}

func (c *GopsutilClient) VirtualMemory(ctx context.Context) (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemoryWithContext(ctx)
}

func (c *GopsutilClient) LoadAvg(ctx context.Context) (*load.AvgStat, error) {
	return load.AvgWithContext(ctx)
}

func (c *GopsutilClient) DiskPartitions(ctx context.Context) ([]disk.PartitionStat, error) {
	return disk.PartitionsWithContext(ctx, false)
}

func (c *GopsutilClient) DiskUsage(ctx context.Context, path string) (*disk.UsageStat, error) {
	return disk.UsageWithContext(ctx, path)
}
