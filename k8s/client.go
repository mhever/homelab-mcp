package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient abstracts Kubernetes API operations for testability.
type K8sClient interface {
	GetNodes(ctx context.Context) ([]NodeInfo, error)
	GetPods(ctx context.Context, namespace string) ([]PodInfo, error)
	GetPodLogs(ctx context.Context, namespace, pod, container string, tail int) (string, error)
	GetEvents(ctx context.Context, namespace string) ([]EventInfo, error)
}

// RealK8sClient wraps the official Kubernetes client.
type RealK8sClient struct {
	clientset *kubernetes.Clientset
}

// NewRealK8sClient creates a RealK8sClient, verifying cluster connectivity.
func NewRealK8sClient() (*RealK8sClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		return nil, fmt.Errorf("kubernetes connectivity check failed: %w", err)
	}

	return &RealK8sClient{clientset: clientset}, nil
}

func (c *RealK8sClient) GetNodes(ctx context.Context) ([]NodeInfo, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var result []NodeInfo
	for _, node := range nodes.Items {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}

		var roles []string
		for key := range node.Labels {
			if strings.HasPrefix(key, "node-role.kubernetes.io/") {
				roles = append(roles, strings.TrimPrefix(key, "node-role.kubernetes.io/"))
			}
		}
		rolesStr := "<none>"
		if len(roles) > 0 {
			sort.Strings(roles)
			rolesStr = strings.Join(roles, ",")
		}

		var internalIP string
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				internalIP = addr.Address
				break
			}
		}

		result = append(result, NodeInfo{
			Name:       node.Name,
			Status:     status,
			Roles:      rolesStr,
			Age:        formatAge(node.CreationTimestamp.Time),
			Version:    node.Status.NodeInfo.KubeletVersion,
			InternalIP: internalIP,
		})
	}

	return result, nil
}

func (c *RealK8sClient) GetPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var result []PodInfo
	for _, pod := range pods.Items {
		var readyCount int
		totalCount := len(pod.Spec.Containers)
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				readyCount++
			}
		}

		var restarts int
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += int(cs.RestartCount)
		}

		result = append(result, PodInfo{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Ready:     fmt.Sprintf("%d/%d", readyCount, totalCount),
			Status:    string(pod.Status.Phase),
			Restarts:  restarts,
			Age:       formatAge(pod.CreationTimestamp.Time),
		})
	}

	return result, nil
}

func (c *RealK8sClient) GetPodLogs(ctx context.Context, namespace, pod, container string, tail int) (string, error) {
	if tail <= 0 {
		tail = 100
	}
	tailInt64 := int64(tail)
	logOpts := &corev1.PodLogOptions{
		TailLines: &tailInt64,
	}
	if container != "" {
		logOpts.Container = container
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(pod, logOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to open log stream: %w", err)
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("failed to read log stream: %w", err)
	}

	return string(data), nil
}

// eventWithTime pairs an EventInfo with its raw timestamp for sorting.
type eventWithTime struct {
	info EventInfo
	ts   time.Time
}

func (c *RealK8sClient) GetEvents(ctx context.Context, namespace string) ([]EventInfo, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var withTimes []eventWithTime
	for _, event := range events.Items {
		ts := event.LastTimestamp.Time
		if ts.IsZero() {
			ts = event.CreationTimestamp.Time
		}
		withTimes = append(withTimes, eventWithTime{
			info: EventInfo{
				Namespace: event.Namespace,
				Type:      event.Type,
				Reason:    event.Reason,
				Object:    fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
				Message:   event.Message,
				Age:       formatAge(ts),
			},
			ts: ts,
		})
	}

	// Sort most recent first; zero timestamps go last.
	sort.Slice(withTimes, func(i, j int) bool {
		ti, tj := withTimes[i].ts, withTimes[j].ts
		if ti.IsZero() {
			return false
		}
		if tj.IsZero() {
			return true
		}
		return ti.After(tj)
	})

	result := make([]EventInfo, len(withTimes))
	for i, e := range withTimes {
		result[i] = e.info
	}
	return result, nil
}

// formatAge returns a human-readable duration string using the largest applicable unit.
func formatAge(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d >= time.Minute:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
}
