package main

import (
	"context"
	"log"
	"os"

	"github.com/mhever/homelab-mcp/docker"
	"github.com/mhever/homelab-mcp/system"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	// Docker tools (graceful skip if unavailable)
	dockerClient, err := docker.NewRealDockerClient()
	if err != nil {
		log.Printf("Docker unavailable, skipping docker tools: %v", err)
	} else {
		defer dockerClient.Close()
		docker.RegisterTools(server, dockerClient)
		log.Println("Docker tools registered")
	}

	log.Println("homelab-mcp server starting on stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
