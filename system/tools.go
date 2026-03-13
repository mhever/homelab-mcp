package system

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mhever/homelab-mcp/internal/mcputil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SystemOverviewArgs struct{}
type SystemDiskArgs struct{}

func RegisterTools(server *mcp.Server, client SystemClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "system_overview",
		Description: "System overview: CPU, memory, disk, uptime, load average",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SystemOverviewArgs) (*mcp.CallToolResult, any, error) {
		return HandleSystemOverview(ctx, client)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "system_disk",
		Description: "Per-mount disk usage breakdown",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SystemDiskArgs) (*mcp.CallToolResult, any, error) {
		return HandleSystemDisk(ctx, client)
	})
}

func HandleSystemOverview(ctx context.Context, client SystemClient) (*mcp.CallToolResult, any, error) {
	info, err := client.HostInfo(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("HostInfo error: %v", err))
	}

	cpuPercents, err := client.CPUPercent(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("CPUPercent error: %v", err))
	}

	vm, err := client.VirtualMemory(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("VirtualMemory error: %v", err))
	}

	avg, err := client.LoadAvg(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("LoadAvg error: %v", err))
	}

	days := info.Uptime / (24 * 3600)
	hours := (info.Uptime % (24 * 3600)) / 3600
	minutes := (info.Uptime % 3600) / 60
	uptimeStr := fmt.Sprintf("%dd %dh %dm", days, hours, minutes)

	var cpuParts []string
	for _, p := range cpuPercents {
		cpuParts = append(cpuParts, fmt.Sprintf("%.1f%%", p))
	}

	usedGB := float64(vm.Used) / (1024 * 1024 * 1024)
	totalGB := float64(vm.Total) / (1024 * 1024 * 1024)

	var sb strings.Builder
	fmt.Fprintf(&sb, "Host: %s | Uptime: %s | OS: %s %s\n", info.Hostname, uptimeStr, info.Platform, info.PlatformVersion)
	fmt.Fprintf(&sb, "CPU: [%s] (%d cores)\n", strings.Join(cpuParts, " "), len(cpuPercents))
	fmt.Fprintf(&sb, "Memory: %.1f/%.1f GB (%.1f%%)\n", usedGB, totalGB, vm.UsedPercent)
	fmt.Fprintf(&sb, "Load: %.2f / %.2f / %.2f", avg.Load1, avg.Load5, avg.Load15)

	return mcputil.TextResult(sb.String())
}

func HandleSystemDisk(ctx context.Context, client SystemClient) (*mcp.CallToolResult, any, error) {
	partitions, err := client.DiskPartitions(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("DiskPartitions error: %v", err))
	}

	filterFstype := map[string]bool{
		"squashfs": true,
		"tmpfs":    true,
		"devtmpfs": true,
		"overlay":  true,
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-16s %-13s %-9s %-9s %-9s %-6s\n", "Mount", "Filesystem", "Size", "Used", "Avail", "Use%"))

	for _, p := range partitions {
		if filterFstype[p.Fstype] {
			continue
		}

		usage, err := client.DiskUsage(ctx, p.Mountpoint)
		if err != nil {
			log.Printf("DiskUsage error for %s: %v", p.Mountpoint, err)
			continue
		}

		sizeGB := float64(usage.Total) / (1024 * 1024 * 1024)
		usedGB := float64(usage.Used) / (1024 * 1024 * 1024)
		availGB := float64(usage.Free) / (1024 * 1024 * 1024)

		sb.WriteString(fmt.Sprintf("%-16s %-13s %-9s %-9s %-9s %-6s\n",
			p.Mountpoint,
			p.Fstype,
			fmt.Sprintf("%.1f GB", sizeGB),
			fmt.Sprintf("%.1f GB", usedGB),
			fmt.Sprintf("%.1f GB", availGB),
			fmt.Sprintf("%.1f%%", usage.UsedPercent)))
	}

	return mcputil.TextResult(sb.String())
}
