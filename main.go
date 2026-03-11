package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	log.SetOutput(os.Stderr)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "homelab-mcp",
		Version: "0.1.0",
	}, nil)

	type PingInput struct{}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ping",
		Description: "Check if the homelab MCP server is running",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in PingInput) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("pong — homelab-mcp is alive at %s", time.Now().Format(time.RFC3339))},
			},
		}, nil, nil
	})

	log.Println("homelab-mcp server starting on stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
