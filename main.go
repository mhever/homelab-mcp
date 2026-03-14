package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mhever/homelab-mcp/docker"
	"github.com/mhever/homelab-mcp/k8s"
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

	// Kubernetes tools (graceful skip if unavailable)
	k8sClient, err := k8s.NewRealK8sClient()
	if err != nil {
		log.Printf("Kubernetes unavailable, skipping k8s tools: %v", err)
	} else {
		k8s.RegisterTools(server, k8sClient)
		log.Println("Kubernetes tools registered")
	}

	log.Println("homelab-mcp server starting on stdio")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
