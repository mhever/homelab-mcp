package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mhever/homelab-mcp/system"
)

func main() {
	log.SetOutput(os.Stderr)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "homelab-mcp",
		Version: "0.1.0",
	}, nil)

	// System tools (always available)
	sysClient := &system.GopsutilClient{}
	system.RegisterTools(server, sysClient)

	log.Println("homelab-mcp server starting on stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
