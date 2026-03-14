package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mhever/homelab-mcp/internal/mcputil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ClusterOverviewArgs struct{}

type PodsArgs struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"namespace to filter (empty for all)"`
	Status    string `json:"status,omitempty" jsonschema:"filter by status e.g. Running Pending Error"`
}

type PodLogsArgs struct {
	Namespace string `json:"namespace" jsonschema:"pod namespace"`
	Pod       string `json:"pod" jsonschema:"pod name"`
	Container string `json:"container,omitempty" jsonschema:"container name (optional for single-container pods)"`
	Tail      int    `json:"tail,omitempty" jsonschema:"number of log lines (default 100)"`
}

type EventsArgs struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"namespace to filter (empty for all namespaces)"`
	Type      string `json:"type,omitempty" jsonschema:"filter by type: Normal or Warning"`
}

type FluxStatusArgs struct{}

// RegisterTools registers all five Kubernetes tools with the MCP server.
func RegisterTools(server *mcp.Server, client K8sClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_cluster_overview",
		Description: "Overview of Kubernetes cluster: node status and pod counts by namespace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ClusterOverviewArgs) (*mcp.CallToolResult, any, error) {
		return HandleClusterOverview(ctx, client)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_pods",
		Description: "List Kubernetes pods with status, restarts, and age; optionally filter by namespace or status",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args PodsArgs) (*mcp.CallToolResult, any, error) {
		return HandleK8sPods(ctx, args, client)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_pod_logs",
		Description: "Get logs from a Kubernetes pod",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args PodLogsArgs) (*mcp.CallToolResult, any, error) {
		return HandleK8sPodLogs(ctx, args, client)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_events",
		Description: "List recent Kubernetes events; optionally filter by namespace or type",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args EventsArgs) (*mcp.CallToolResult, any, error) {
		return HandleK8sEvents(ctx, args, client)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_flux_status",
		Description: "Show FluxCD GitRepository and Kustomization reconciliation status",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FluxStatusArgs) (*mcp.CallToolResult, any, error) {
		return HandleFluxStatus(ctx, client)
	})
}

// HandleClusterOverview returns node status and a per-namespace pod count summary.
func HandleClusterOverview(ctx context.Context, client K8sClient) (*mcp.CallToolResult, any, error) {
	nodes, err := client.GetNodes(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("GetNodes error: %v", err))
	}

	pods, err := client.GetPods(ctx, "")
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("GetPods error: %v", err))
	}

	var sb strings.Builder
	sb.WriteString("NODES\n")
	sb.WriteString(fmt.Sprintf("%-14s%-8s%-15s%-5s%s\n", "NAME", "STATUS", "ROLES", "AGE", "VERSION"))
	for _, node := range nodes {
		sb.WriteString(fmt.Sprintf("%-14s%-8s%-15s%-5s%s\n", node.Name, node.Status, node.Roles, node.Age, node.Version))
	}

	sb.WriteString("\nPOD SUMMARY (by namespace)\n")
	namespaceCounts := make(map[string]int)
	for _, pod := range pods {
		namespaceCounts[pod.Namespace]++
	}

	namespaces := make([]string, 0, len(namespaceCounts))
	for ns := range namespaceCounts {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	for _, ns := range namespaces {
		sb.WriteString(fmt.Sprintf("  %-14s%d pods\n", ns, namespaceCounts[ns]))
	}

	return mcputil.TextResult(sb.String())
}

// HandleK8sPods lists pods with optional namespace and status filtering.
func HandleK8sPods(ctx context.Context, args PodsArgs, client K8sClient) (*mcp.CallToolResult, any, error) {
	pods, err := client.GetPods(ctx, args.Namespace)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("GetPods error: %v", err))
	}

	var filtered []PodInfo
	for _, pod := range pods {
		if args.Status != "" && !strings.EqualFold(pod.Status, args.Status) {
			continue
		}
		filtered = append(filtered, pod)
	}

	if args.Status != "" && len(filtered) == 0 {
		return mcputil.TextResult(fmt.Sprintf("No pods found matching status %q", args.Status))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-14s%-24s%-7s%-9s%-10s%s\n", "NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE"))
	for _, pod := range filtered {
		sb.WriteString(fmt.Sprintf("%-14s%-24s%-7s%-9s%-10d%s\n",
			pod.Namespace, pod.Name, pod.Ready, pod.Status, pod.Restarts, pod.Age))
	}

	return mcputil.TextResult(sb.String())
}

// HandleK8sPodLogs returns logs for a specific pod.
func HandleK8sPodLogs(ctx context.Context, args PodLogsArgs, client K8sClient) (*mcp.CallToolResult, any, error) {
	if args.Tail < 0 {
		return mcputil.ErrorResult("tail must be a non-negative integer")
	}
	if args.Tail == 0 {
		args.Tail = 100
	}

	logs, err := client.GetPodLogs(ctx, args.Namespace, args.Pod, args.Container, args.Tail)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("GetPodLogs error: %v", err))
	}

	return mcputil.TextResult(logs)
}

// HandleK8sEvents lists events with optional namespace and type filtering.
func HandleK8sEvents(ctx context.Context, args EventsArgs, client K8sClient) (*mcp.CallToolResult, any, error) {
	events, err := client.GetEvents(ctx, args.Namespace)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("GetEvents error: %v", err))
	}

	var filtered []EventInfo
	for _, event := range events {
		if args.Type != "" && !strings.EqualFold(event.Type, args.Type) {
			continue
		}
		filtered = append(filtered, event)
	}

	if args.Type != "" && len(filtered) == 0 {
		return mcputil.TextResult(fmt.Sprintf("No events found matching type %q", args.Type))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-14s%-9s%-17s%-22s%-6s%s\n", "NAMESPACE", "TYPE", "REASON", "OBJECT", "AGE", "MESSAGE"))
	for _, event := range filtered {
		msg := event.Message
		if len(msg) > 80 {
			msg = msg[:77] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-14s%-9s%-17s%-22s%-6s%s\n",
			event.Namespace, event.Type, event.Reason, event.Object, event.Age, msg))
	}

	return mcputil.TextResult(sb.String())
}

// HandleFluxStatus returns FluxCD GitRepository and Kustomization resource status.
func HandleFluxStatus(ctx context.Context, client K8sClient) (*mcp.CallToolResult, any, error) {
	fluxResources, err := client.GetFluxStatus(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("GetFluxStatus error: %v", err))
	}

	if len(fluxResources) == 0 {
		return mcputil.TextResult("No FluxCD resources found")
	}

	var sb strings.Builder
	// Column widths: KIND=18, NAMESPACE=16, NAME=26, READY=8, REASON=20, AGE=6, MESSAGE truncated to 60
	sb.WriteString(fmt.Sprintf("%-18s%-16s%-26s%-8s%-20s%-6s%s\n",
		"KIND", "NAMESPACE", "NAME", "READY", "REASON", "AGE", "MESSAGE"))

	for _, flux := range fluxResources {
		msg := flux.Message
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-18s%-16s%-26s%-8s%-20s%-6s%s\n",
			flux.Kind, flux.Namespace, flux.Name, flux.Ready, flux.Reason, flux.Age, msg))
	}

	return mcputil.TextResult(sb.String())
}
