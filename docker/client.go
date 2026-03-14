package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// DockerClient abstracts Docker Engine API operations for testability.
type DockerClient interface {
	ListContainers(ctx context.Context) ([]types.Container, error)
	ContainerLogs(ctx context.Context, id string, tail string) (io.ReadCloser, error)
	ContainerAction(ctx context.Context, id string, action string) error
	Close() error
}

// RealDockerClient wraps the official Docker client.
type RealDockerClient struct {
	cli *client.Client
}

// NewRealDockerClient creates a RealDockerClient, pinging Docker to confirm availability.
func NewRealDockerClient() (*RealDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err = cli.Ping(pingCtx); err != nil {
		return nil, err
	}

	return &RealDockerClient{cli: cli}, nil
}

func (c *RealDockerClient) ListContainers(ctx context.Context) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, container.ListOptions{All: true})
}

func (c *RealDockerClient) ContainerLogs(ctx context.Context, id string, tail string) (io.ReadCloser, error) {
	return c.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	})
}

func (c *RealDockerClient) ContainerAction(ctx context.Context, id string, action string) error {
	switch action {
	case "start":
		return c.cli.ContainerStart(ctx, id, container.StartOptions{})
	case "stop":
		return c.cli.ContainerStop(ctx, id, container.StopOptions{})
	case "restart":
		return c.cli.ContainerRestart(ctx, id, container.StopOptions{})
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (c *RealDockerClient) Close() error {
	return c.cli.Close()
}
