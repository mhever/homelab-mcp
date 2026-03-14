package k8s_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mhever/homelab-mcp/k8s"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockK8sClient implements k8s.K8sClient via testify/mock.
type MockK8sClient struct {
	mock.Mock
}

func (m *MockK8sClient) GetNodes(ctx context.Context) ([]k8s.NodeInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]k8s.NodeInfo), args.Error(1)
}

func (m *MockK8sClient) GetPods(ctx context.Context, namespace string) ([]k8s.PodInfo, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]k8s.PodInfo), args.Error(1)
}

func (m *MockK8sClient) GetPodLogs(ctx context.Context, namespace, pod, container string, tail int) (string, error) {
	args := m.Called(ctx, namespace, pod, container, tail)
	return args.String(0), args.Error(1)
}

func (m *MockK8sClient) GetEvents(ctx context.Context, namespace string) ([]k8s.EventInfo, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]k8s.EventInfo), args.Error(1)
}

// textContent extracts the text from the first content item of a CallToolResult.
func textContent(t *testing.T, result *mcp.CallToolResult) string {
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

func TestHandleClusterOverview_Success(t *testing.T) {
	mockClient := new(MockK8sClient)
	mockClient.On("GetNodes", mock.Anything).Return([]k8s.NodeInfo{
		{Name: "node1", Status: "Ready", Roles: "control-plane", Age: "10d", Version: "v1.34.3+k3s1"},
	}, nil)
	mockClient.On("GetPods", mock.Anything, "").Return([]k8s.PodInfo{
		{Namespace: "kube-system", Name: "coredns", Ready: "1/1", Status: "Running", Restarts: 0, Age: "10d"},
		{Namespace: "default", Name: "nginx", Ready: "1/1", Status: "Running", Restarts: 0, Age: "5d"},
	}, nil)

	result, _, err := k8s.HandleClusterOverview(context.Background(), mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)

	output := textContent(t, result)
	assert.True(t, strings.Contains(output, "node1"))
	assert.True(t, strings.Contains(output, "kube-system"))
}

func TestHandleClusterOverview_NodesError(t *testing.T) {
	mockClient := new(MockK8sClient)
	mockClient.On("GetNodes", mock.Anything).Return(nil, errors.New("nodes unavailable"))

	result, _, err := k8s.HandleClusterOverview(context.Background(), mockClient)
	assert.NoError(t, err) // tools never return a Go error
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandleK8sPods_Success(t *testing.T) {
	mockClient := new(MockK8sClient)
	mockClient.On("GetPods", mock.Anything, "").Return([]k8s.PodInfo{
		{Namespace: "default", Name: "web-abc", Ready: "1/1", Status: "Running"},
		{Namespace: "default", Name: "db-xyz", Ready: "1/1", Status: "Running"},
	}, nil)

	result, _, err := k8s.HandleK8sPods(context.Background(), k8s.PodsArgs{}, mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	output := textContent(t, result)
	assert.True(t, strings.Contains(output, "web-abc"))
	assert.True(t, strings.Contains(output, "db-xyz"))
}

func TestHandleK8sPods_StatusFilter(t *testing.T) {
	mockClient := new(MockK8sClient)
	mockClient.On("GetPods", mock.Anything, "").Return([]k8s.PodInfo{
		{Namespace: "default", Name: "web-abc", Status: "Running"},
		{Namespace: "default", Name: "db-xyz", Status: "Running"},
		{Namespace: "default", Name: "pending-pod", Status: "Pending"},
	}, nil)

	result, _, err := k8s.HandleK8sPods(context.Background(), k8s.PodsArgs{Status: "Running"}, mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	output := textContent(t, result)
	assert.True(t, strings.Contains(output, "web-abc"))
	assert.True(t, strings.Contains(output, "db-xyz"))
	assert.False(t, strings.Contains(output, "pending-pod"))
}

func TestHandleK8sPodLogs_Success(t *testing.T) {
	mockClient := new(MockK8sClient)
	mockClient.On("GetPodLogs", mock.Anything, "default", "mypod", "", 100).Return("log line 1\nlog line 2\n", nil)

	result, _, err := k8s.HandleK8sPodLogs(context.Background(), k8s.PodLogsArgs{
		Namespace: "default",
		Pod:       "mypod",
		Tail:      100,
	}, mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	output := textContent(t, result)
	assert.True(t, strings.Contains(output, "log line 1"))
}

func TestHandleK8sPodLogs_DefaultTail(t *testing.T) {
	mockClient := new(MockK8sClient)
	// Tail=0 in args should default to 100 inside the handler.
	mockClient.On("GetPodLogs", mock.Anything, "default", "mypod", "", 100).Return("some log\n", nil)

	result, _, err := k8s.HandleK8sPodLogs(context.Background(), k8s.PodLogsArgs{
		Namespace: "default",
		Pod:       "mypod",
		Tail:      0,
	}, mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	mockClient.AssertExpectations(t)
}

func TestHandleK8sEvents_Success(t *testing.T) {
	mockClient := new(MockK8sClient)
	mockClient.On("GetEvents", mock.Anything, "").Return([]k8s.EventInfo{
		{Namespace: "kube-system", Type: "Warning", Reason: "BackOff", Object: "Pod/coredns-abc", Message: "Back-off restarting"},
	}, nil)

	result, _, err := k8s.HandleK8sEvents(context.Background(), k8s.EventsArgs{}, mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	output := textContent(t, result)
	assert.True(t, strings.Contains(output, "BackOff"))
	assert.True(t, strings.Contains(output, "Pod/coredns-abc"))
}

func TestHandleK8sEvents_TypeFilter(t *testing.T) {
	mockClient := new(MockK8sClient)
	mockClient.On("GetEvents", mock.Anything, "").Return([]k8s.EventInfo{
		{Namespace: "kube-system", Type: "Warning", Reason: "BackOff", Object: "Pod/coredns"},
		{Namespace: "kube-system", Type: "Warning", Reason: "FailedMount", Object: "Pod/nginx"},
		{Namespace: "kube-system", Type: "Normal", Reason: "Started", Object: "Pod/web"},
	}, nil)

	result, _, err := k8s.HandleK8sEvents(context.Background(), k8s.EventsArgs{Type: "Warning"}, mockClient)
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	output := textContent(t, result)
	assert.True(t, strings.Contains(output, "BackOff"))
	assert.True(t, strings.Contains(output, "FailedMount"))
	assert.False(t, strings.Contains(output, "Started"))
}
