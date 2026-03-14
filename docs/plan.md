# homelab-mcp: Home Lab Operations MCP Server

## Context

Building a Go MCP server for portfolio positioning that exposes tools to monitor and operate a mixed home lab infrastructure (k3s + Docker + bare-metal metrics) through a single interface. This differentiates from the dozens of existing k8s-only MCP servers by being opinionated about a real mixed-workload use case.

The server runs on an Ubuntu 24 ThinkCentre m715q home server with direct access to k3s kubeconfig, Docker socket, and /proc.

## Tech Stack

- **Language:** Go
- **MCP SDK:** Official Go SDK v1.4.0 (`github.com/modelcontextprotocol/go-sdk/mcp`)
- **Transport:** stdio (invoked via SSH from Claude Code/Desktop)
- **Key deps:** client-go, docker/client, gopsutil/v4

## Development Environment

**Development via Claude Code SSH remote:** `claude --ssh agent@thinkcentre.local`
All development, building, and testing happens directly on the Ubuntu server — no cross-compilation needed. Go, Docker, kubectl, and /proc are all available natively.

## Project Location

`~/projects/homelab-mcp` on the m715q server (module: `github.com/mhever/homelab-mcp`)

## Architecture

- Interface-based client abstraction per domain (K8sClient, DockerClient, SystemClient) for testability
- Graceful degradation: if Docker/k8s aren't available at startup, those tools are simply not registered
- All logging to stderr (stdout reserved for MCP stdio transport)
- FluxCD via dynamic client (unstructured) to avoid heavy fluxcd/pkg transitive deps
- Output as formatted plain text, not JSON — LLMs read it better
- Errors returned as `CallToolResult` with `IsError: true`, not Go errors

## Tools (10 total)

### Kubernetes (5)
| Tool | Args | Description |
|------|------|-------------|
| `k8s_cluster_overview` | none | Node status, resource usage, pod counts by namespace |
| `k8s_pods` | namespace?, status? | List/filter pods with status, restarts, age |
| `k8s_pod_logs` | namespace, pod, container?, tail? | Get pod logs |
| `k8s_events` | namespace?, type? | Recent cluster events |
| `k8s_flux_status` | none | GitRepository & Kustomization reconciliation status |

### Docker (3)
| Tool | Args | Description |
|------|------|-------------|
| `docker_containers` | none | List containers with status, uptime, ports |
| `docker_container_logs` | container, tail? | Get container logs |
| `docker_container_action` | container, action | Start/stop/restart a container |

### System (2)
| Tool | Args | Description |
|------|------|-------------|
| `system_overview` | none | CPU, memory, disk, uptime, load average |
| `system_disk` | none | Per-mount disk usage breakdown |

## File Structure

```
homelab-mcp/
├── main.go                    # Server setup, client init, graceful degradation
├── go.mod / go.sum
├── .gitignore
├── Makefile
├── README.md
├── system/
│   ├── client.go              # SystemClient interface + gopsutil impl
│   ├── tools.go               # RegisterTools + 2 handlers
│   └── client_test.go
├── docker/
│   ├── client.go              # DockerClient interface + real impl
│   ├── tools.go               # RegisterTools + 3 handlers
│   └── client_test.go
├── k8s/
│   ├── client.go              # K8sClient interface + real impl (incl. dynamic client)
│   ├── types.go               # Shared data types
│   ├── tools.go               # RegisterTools + 5 handlers
│   └── client_test.go
├── deploy/
│   └── homelab-mcp.service    # systemd unit file
└── .github/
    └── workflows/
        └── ci.yml             # lint, test, build, release
```

## Implementation Phases

### Phase 0: Scaffold
1. Create project dir, `go mod init`, `.gitignore`
2. `main.go` with MCP server + stdio transport + one dummy "ping" tool
3. Validate SDK wiring works
4. Git init + initial commit

### Phase 1: System Tools
1. `system/client.go` — SystemClient interface + gopsutil implementation
2. `system/tools.go` — `system_overview` and `system_disk` tools
3. `system/client_test.go` — mock-based tests
4. Wire into main.go, remove ping tool

### Phase 2: Docker Tools
1. `docker/client.go` — DockerClient interface + Docker Engine API impl
2. `docker/tools.go` — 3 container tools
3. `docker/client_test.go` — mock-based tests
4. Wire into main.go with graceful skip if Docker unavailable

### Phase 3: Kubernetes Tools
1. `k8s/client.go` — K8sClient interface + client-go impl (InClusterConfig fallback to kubeconfig)
2. `k8s/types.go` — PodInfo, NodeInfo, EventInfo structs
3. `k8s/tools.go` — 4 core k8s tools
4. `k8s/client_test.go` — mock-based tests
5. Wire into main.go with graceful skip

### Phase 4: FluxCD Tool
1. Add dynamic client to k8s client for GitRepository/Kustomization GVRs
2. Add `k8s_flux_status` tool parsing unstructured conditions
3. Tests with mock unstructured data

### Phase 5: Polish
1. `Makefile` (build, test, lint, deploy targets)
2. `README.md` with tool docs, setup instructions, MCP client config examples
3. `.github/workflows/ci.yml` (lint, test, cross-compile, release on tag)
4. `deploy/homelab-mcp.service` systemd unit

## Key Patterns

**Tool registration (SDK v1.4.0 generics pattern):**
```go
type PodLogsArgs struct {
    Namespace string `json:"namespace" jsonschema:"description=pod namespace"`
    Pod       string `json:"pod" jsonschema:"description=pod name"`
    Tail      int    `json:"tail" jsonschema:"description=number of lines (default 100)"`
}

mcp.AddTool(server, &mcp.Tool{
    Name:        "k8s_pod_logs",
    Description: "Get logs from a Kubernetes pod",
}, func(ctx context.Context, req *mcp.CallToolRequest, args PodLogsArgs) (*mcp.CallToolResult, any, error) {
    // handler
})
```

**Error helper:**
```go
func errorResult(msg string) (*mcp.CallToolResult, any, error) {
    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: msg}},
        IsError: true,
    }, nil, nil
}
```

**Deployment via SSH stdio:**
```json
{
  "mcpServers": {
    "homelab": {
      "command": "ssh",
      "args": ["agent@thinkcentre.local", "/usr/local/bin/homelab-mcp"]
    }
  }
}
```

## Pitfalls to Watch

- **Stderr for logging:** `log.SetOutput(os.Stderr)` — stdout is MCP's stdio channel
- **Docker socket perms:** User needs `docker` group membership
- **Kubeconfig path:** Set `KUBECONFIG` in systemd unit if not using default `~/.kube/config`
- **client-go version:** Pin to match k3s cluster's k8s version
## Prerequisites (on m715q, before Phase 0)

1. Ensure Go is installed (`go version`) — if not, install latest Go
2. Ensure `~/projects` directory exists
3. Verify Docker access: `docker ps` works for the dev user
4. Verify kubectl access: `kubectl get nodes` works
5. Verify git is installed and configured

## Verification

1. **Phase 0:** `echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | go run main.go` — should return tool list
2. **Each phase:** `go test ./...` passes
3. **Final:** `go build -o homelab-mcp .`, then configure in Claude Code MCP settings and verify all tools respond
4. **Integration:** Use Claude Code to ask "what's running on my home server?" and confirm it calls the right tools
