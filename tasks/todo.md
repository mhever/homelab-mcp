# Phase 3: Kubernetes Tools

- [x] Create k8s/types.go -- NodeInfo, PodInfo, EventInfo structs
- [x] Create k8s/client.go -- K8sClient interface + RealK8sClient impl (InCluster + kubeconfig fallback)
- [x] Create k8s/tools.go -- RegisterTools with k8s_cluster_overview, k8s_pods, k8s_pod_logs, k8s_events
- [x] Create k8s/client_test.go -- mock-based tests (8 tests)
- [x] Update main.go -- graceful k8s degradation + signal-aware shutdown context
- [x] go mod tidy + go build ./...
- [x] go test ./... (21/21 pass)
- [x] deepseek-reviewer audit -- all 5 findings resolved
