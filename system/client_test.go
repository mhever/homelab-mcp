package system

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

type MockSystemClient struct {
	HostInfoResult       *host.InfoStat
	HostInfoErr          error
	CPUPercentResult     []float64
	CPUPercentErr        error
	VirtualMemoryResult  *mem.VirtualMemoryStat
	VirtualMemoryErr     error
	LoadAvgResult        *load.AvgStat
	LoadAvgErr           error
	DiskPartitionsResult []disk.PartitionStat
	DiskPartitionsErr    error
	DiskUsageResults     map[string]*disk.UsageStat
	DiskUsageErr         error
}

func (m *MockSystemClient) HostInfo(ctx context.Context) (*host.InfoStat, error) {
	return m.HostInfoResult, m.HostInfoErr
}

func (m *MockSystemClient) CPUPercent(ctx context.Context) ([]float64, error) {
	return m.CPUPercentResult, m.CPUPercentErr
}

func (m *MockSystemClient) VirtualMemory(ctx context.Context) (*mem.VirtualMemoryStat, error) {
	return m.VirtualMemoryResult, m.VirtualMemoryErr
}

func (m *MockSystemClient) LoadAvg(ctx context.Context) (*load.AvgStat, error) {
	return m.LoadAvgResult, m.LoadAvgErr
}

func (m *MockSystemClient) DiskPartitions(ctx context.Context) ([]disk.PartitionStat, error) {
	return m.DiskPartitionsResult, m.DiskPartitionsErr
}

func (m *MockSystemClient) DiskUsage(ctx context.Context, path string) (*disk.UsageStat, error) {
	if m.DiskUsageErr != nil {
		return nil, m.DiskUsageErr
	}
	if res, ok := m.DiskUsageResults[path]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("not found")
}

func TestHandleSystemOverview_Success(t *testing.T) {
	mock := &MockSystemClient{
		HostInfoResult: &host.InfoStat{
			Hostname:        "testhost",
			Uptime:          1234567,
			Platform:        "ubuntu",
			PlatformVersion: "24.04",
		},
		CPUPercentResult: []float64{12.3, 5.1, 8.7, 3.2},
		VirtualMemoryResult: &mem.VirtualMemoryStat{
			Total:       16 * 1024 * 1024 * 1024,
			Used:        6 * 1024 * 1024 * 1024,
			UsedPercent: 37.5,
		},
		LoadAvgResult: &load.AvgStat{
			Load1:  0.82,
			Load5:  0.65,
			Load15: 0.51,
		},
	}

	result, _, err := HandleSystemOverview(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected no error result, got error")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	expectedSubstrings := []string{"testhost", "14d 6h 56m", "12.3%", "6.0/16.0 GB", "0.82"}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(text, sub) {
			t.Errorf("expected output to contain %q, but it didn't: %s", sub, text)
		}
	}
}

func TestHandleSystemOverview_HostInfoError(t *testing.T) {
	mock := &MockSystemClient{
		HostInfoErr: fmt.Errorf("host info failed"),
	}

	result, _, err := HandleSystemOverview(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Errorf("expected error result, got success")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "HostInfo error") {
		t.Errorf("expected output to contain 'HostInfo error', but it didn't: %s", text)
	}
}

func TestHandleSystemDisk_Success(t *testing.T) {
	mock := &MockSystemClient{
		DiskPartitionsResult: []disk.PartitionStat{
			{Mountpoint: "/", Fstype: "ext4"},
			{Mountpoint: "/boot", Fstype: "ext4"},
			{Mountpoint: "/snap/foo", Fstype: "squashfs"},
		},
		DiskUsageResults: map[string]*disk.UsageStat{
			"/": {
				Total:       100 * 1024 * 1024 * 1024,
				Used:        40 * 1024 * 1024 * 1024,
				Free:        60 * 1024 * 1024 * 1024,
				UsedPercent: 40.0,
			},
			"/boot": {
				Total:       1024 * 1024 * 1024,
				Used:        200 * 1024 * 1024,
				Free:        824 * 1024 * 1024,
				UsedPercent: 19.5,
			},
		},
	}

	result, _, err := HandleSystemDisk(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected no error result, got error")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "/") {
		t.Errorf("expected output to contain '/', but it didn't")
	}
	if !strings.Contains(text, "/boot") {
		t.Errorf("expected output to contain '/boot', but it didn't")
	}
	if strings.Contains(text, "squashfs") || strings.Contains(text, "/snap/foo") {
		t.Errorf("expected output NOT to contain 'squashfs' or '/snap/foo', but it did: %s", text)
	}
}

func TestHandleSystemDisk_PartitionsError(t *testing.T) {
	mock := &MockSystemClient{
		DiskPartitionsErr: fmt.Errorf("partitions failed"),
	}

	result, _, err := HandleSystemDisk(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Errorf("expected error result, got success")
	}
}
