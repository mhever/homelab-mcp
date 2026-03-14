package docker

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/mhever/homelab-mcp/internal/mcputil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxFrameSize = 10 * 1024 * 1024 // 10MB

type DockerContainersArgs struct{}

type ContainerLogsArgs struct {
	Container string `json:"container" jsonschema:"description=container name or ID,required"`
	Tail      string `json:"tail" jsonschema:"description=number of log lines to return (default 100)"`
}

type ContainerActionArgs struct {
	Container string `json:"container" jsonschema:"description=container name or ID,required"`
	Action    string `json:"action" jsonschema:"description=action: start stop restart,required"`
}

// RegisterTools registers the three Docker tools with the MCP server.
func RegisterTools(server *mcp.Server, client DockerClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_containers",
		Description: "List all Docker containers with status, uptime, and ports",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DockerContainersArgs) (*mcp.CallToolResult, any, error) {
		return HandleDockerContainers(ctx, client)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_container_logs",
		Description: "Get logs from a Docker container",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ContainerLogsArgs) (*mcp.CallToolResult, any, error) {
		return HandleContainerLogs(ctx, args, client)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_container_action",
		Description: "Perform an action (start, stop, restart) on a Docker container",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ContainerActionArgs) (*mcp.CallToolResult, any, error) {
		return HandleContainerAction(ctx, args, client)
	})
}

// HandleDockerContainers lists all Docker containers in a formatted table.
func HandleDockerContainers(ctx context.Context, client DockerClient) (*mcp.CallToolResult, any, error) {
	containers, err := client.ListContainers(ctx)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("ListContainers error: %v", err))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-20s %-20s %-20s %-20s\n", "CONTAINER", "IMAGE", "STATUS", "PORTS"))

	for _, c := range containers {
		name := c.ID
		if len(c.ID) >= 12 {
			name = c.ID[:12]
		}
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		var ports []string
		for _, port := range c.Ports {
			ports = append(ports, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		}

		sb.WriteString(fmt.Sprintf("%-20s %-20s %-20s %-20s\n",
			name, c.Image, c.Status, strings.Join(ports, ",")))
	}

	return mcputil.TextResult(sb.String())
}

// HandleContainerLogs retrieves logs from a Docker container, stripping the multiplexed stream header.
func HandleContainerLogs(ctx context.Context, args ContainerLogsArgs, client DockerClient) (*mcp.CallToolResult, any, error) {
	if args.Tail == "" {
		args.Tail = "100"
	}

	reader, err := client.ContainerLogs(ctx, args.Container, args.Tail)
	if err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("ContainerLogs error: %v", err))
	}
	defer reader.Close()

	var sb strings.Builder
	header := make([]byte, 8)

	for {
		_, err := io.ReadFull(reader, header)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return mcputil.ErrorResult(fmt.Sprintf("ReadHeader error: %v", err))
		}

		size := binary.BigEndian.Uint32(header[4:8])
		if size > maxFrameSize {
			return mcputil.ErrorResult("log frame too large")
		}
		payload := make([]byte, size)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return mcputil.ErrorResult(fmt.Sprintf("ReadPayload error: %v", err))
		}

		sb.Write(payload)
	}

	return mcputil.TextResult(sb.String())
}

// HandleContainerAction performs a start/stop/restart action on a container.
func HandleContainerAction(ctx context.Context, args ContainerActionArgs, client DockerClient) (*mcp.CallToolResult, any, error) {
	validActions := map[string]bool{"start": true, "stop": true, "restart": true}
	if !validActions[args.Action] {
		return mcputil.ErrorResult(fmt.Sprintf("invalid action: %s, must be one of: start, stop, restart", args.Action))
	}

	if err := client.ContainerAction(ctx, args.Container, args.Action); err != nil {
		return mcputil.ErrorResult(fmt.Sprintf("ContainerAction error: %v", err))
	}

	return mcputil.TextResult(fmt.Sprintf("Action '%s' completed for container '%s'", args.Action, args.Container))
}

