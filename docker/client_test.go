package docker_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/mhever/homelab-mcp/docker"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDockerClient implements docker.DockerClient via testify/mock.
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) ListContainers(ctx context.Context) ([]dockertypes.Container, error) {
	args := m.Called(ctx)
	return args.Get(0).([]dockertypes.Container), args.Error(1)
}

func (m *MockDockerClient) ContainerLogs(ctx context.Context, id string, tail string) (io.ReadCloser, error) {
	args := m.Called(ctx, id, tail)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerClient) ContainerAction(ctx context.Context, id string, action string) error {
	args := m.Called(ctx, id, action)
	return args.Error(0)
}

func (m *MockDockerClient) Close() error {
	return nil
}

func textFromResult(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		return ""
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

func TestHandleDockerContainers_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockClient.On("ListContainers", mock.Anything).Return([]dockertypes.Container{
		{Names: []string{"/nginx"}},
		{Names: []string{"/redis"}},
	}, nil)

	result, _, err := docker.HandleDockerContainers(context.Background(), mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	text := textFromResult(t, result)
	assert.Contains(t, text, "nginx")
	assert.Contains(t, text, "redis")
	mockClient.AssertExpectations(t)
}

func TestHandleDockerContainers_Error(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockClient.On("ListContainers", mock.Anything).Return([]dockertypes.Container{}, assert.AnError)

	result, _, err := docker.HandleDockerContainers(context.Background(), mockClient)
	assert.NoError(t, err) // handlers never return Go errors
	assert.True(t, result.IsError)
	mockClient.AssertExpectations(t)
}

func TestHandleContainerLogs_Success(t *testing.T) {
	mockClient := new(MockDockerClient)

	// Build a Docker multiplexed log frame: 8-byte header + payload
	payload := []byte("log line\n")
	var buf bytes.Buffer
	buf.WriteByte(1) // stdout stream
	buf.Write([]byte{0, 0, 0})
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(payload)))
	buf.Write(sizeBytes)
	buf.Write(payload)

	mockClient.On("ContainerLogs", mock.Anything, "mycontainer", "50").
		Return(io.NopCloser(bytes.NewReader(buf.Bytes())), nil)

	args := docker.ContainerLogsArgs{Container: "mycontainer", Tail: "50"}
	result, _, err := docker.HandleContainerLogs(context.Background(), args, mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	text := textFromResult(t, result)
	assert.Contains(t, text, "log line")
	mockClient.AssertExpectations(t)
}

func TestHandleContainerLogs_DefaultTail(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockClient.On("ContainerLogs", mock.Anything, "mycontainer", "100").
		Return(io.NopCloser(bytes.NewReader([]byte{})), nil)

	args := docker.ContainerLogsArgs{Container: "mycontainer", Tail: ""}
	_, _, err := docker.HandleContainerLogs(context.Background(), args, mockClient)
	assert.NoError(t, err)

	mockClient.AssertCalled(t, "ContainerLogs", mock.Anything, "mycontainer", "100")
	mockClient.AssertExpectations(t)
}

func TestHandleContainerAction_Success(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockClient.On("ContainerAction", mock.Anything, "mycontainer", "restart").Return(nil)

	args := docker.ContainerActionArgs{Container: "mycontainer", Action: "restart"}
	result, _, err := docker.HandleContainerAction(context.Background(), args, mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	text := textFromResult(t, result)
	assert.Contains(t, text, "restart")
	assert.Contains(t, text, "mycontainer")
	mockClient.AssertExpectations(t)
}

func TestHandleContainerAction_InvalidAction(t *testing.T) {
	mockClient := new(MockDockerClient)

	args := docker.ContainerActionArgs{Container: "x", Action: "explode"}
	result, _, err := docker.HandleContainerAction(context.Background(), args, mockClient)
	assert.NoError(t, err) // handlers never return Go errors
	assert.True(t, result.IsError)

	mockClient.AssertNotCalled(t, "ContainerAction")
}

func TestHandleContainerLogs_Error(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockClient.On("ContainerLogs", mock.Anything, "mycontainer", "100").Return(io.NopCloser(bytes.NewReader([]byte{})), assert.AnError)

	args := docker.ContainerLogsArgs{Container: "mycontainer", Tail: ""}
	result, _, err := docker.HandleContainerLogs(context.Background(), args, mockClient)
	assert.NoError(t, err)
	assert.True(t, result.IsError)

	mockClient.AssertExpectations(t)
}

func TestHandleContainerAction_ClientError(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockClient.On("ContainerAction", mock.Anything, "mycontainer", "stop").Return(fmt.Errorf("daemon error"))

	args := docker.ContainerActionArgs{Container: "mycontainer", Action: "stop"}
	result, _, err := docker.HandleContainerAction(context.Background(), args, mockClient)
	assert.NoError(t, err)
	assert.True(t, result.IsError)

	mockClient.AssertExpectations(t)
}

func TestHandleDockerContainers_EmptyNames(t *testing.T) {
	mockClient := new(MockDockerClient)
	mockClient.On("ListContainers", mock.Anything).Return([]dockertypes.Container{
		{Names: []string{}, ID: "abc123def456789"},
	}, nil)

	result, _, err := docker.HandleDockerContainers(context.Background(), mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	text := textFromResult(t, result)
	assert.Contains(t, text, "abc123def456")
	mockClient.AssertExpectations(t)
}
