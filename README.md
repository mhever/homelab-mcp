# homelab-mcp

A Go-based [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server designed for home lab operations. Unlike generic Kubernetes servers, `homelab-mcp` is built for mixed-workload environments, exposing tools for **k3s**, **Docker**, and **Bare-metal system metrics** through a single interface.

This project is part of my professional portfolio, demonstrating Go development, systems engineering, and **agentic development workflows**.

## 🚀 Current Status: Phase 3 (Complete)

- [x] **Phase 0: Scaffold** - MCP SDK wiring & stdio transport.
- [x] **Phase 1: System Tools** - `system_overview` and `system_disk` via `gopsutil`.
- [x] **Phase 2: Docker Tools** - Container management & logs.
- [x] **Phase 3: Kubernetes Tools** - Pods, Events, and FluxCD status.

For the full implementation roadmap, see [plan.md](./plan.md).

## 🛠️ Features (Implemented)

### System Tools
| Tool | Description |
|------|-------------|
| `system_overview` | Returns CPU load, memory usage, and uptime metrics. |
| `system_disk` | Provides a breakdown of disk usage per mount point. |

### Docker Tools
| Tool | Description |
|------|-------------|
| `docker_containers` | Lists all containers with name, image, status, and ports. |
| `docker_container_logs` | Gets container logs; args: container name/ID, tail line count. |
| `docker_container_action` | Start/stop/restart a container; args: container name/ID, action. |

### Kubernetes Tools
| Tool | Description |
|------|-------------|
| k8s_cluster_overview | Node status and per-namespace pod count summary |
| k8s_pods | List/filter pods by namespace and status |
| k8s_pod_logs | Get logs from a pod (args: namespace, pod, container, tail) |
| k8s_events | List cluster events filtered by namespace and type |

## 🤖 Built with Agents

This repository is developed using a multi-agent orchestration pattern where Claude acts as the orchestrator and delegates to specialist subagents:

| Role | Model | Task |
|------|-------|------|
| **Orchestrator** | Claude Sonnet 4.6 (Claude Code) | Planning, routing, and verification |
| **Builder** | DeepSeek V3 (`deepseek-chat` via OpenRouter) | Code generation and file edits |
| **Reviewer** | DeepSeek R1 (`deepseek-r1` via OpenRouter) | Logic audits, race condition detection, security review |
| **Tester** | Gemini Flash-Lite (`gemini-3.1-flash-lite-preview`) | Unit test and mock generation |
| **Librarian** | Gemini Flash-Lite (`gemini-3.1-flash-lite-preview`) | README and documentation sync |

This "Hybrid AI" approach assigns each task to the model best suited for it, enabling high-velocity development while managing cost and quota across providers.

## 🏗️ Technical Architecture

- **Language:** Go 1.25+
- **Transport:** stdio (designed for SSH-remote invocation).
- **Graceful Degradation:** The server detects available services (Docker/k3s) at startup and only registers tools for active components.
- **Platform:** Optimized for Ubuntu 24.04 (ThinkCentre m715q).

## 📋 Prerequisites

- **Go:** 1.25 or higher.
- **Access:** User must be in the `docker` group for Docker tools.
- **Kubeconfig:** Accessible at `~/.kube/config`.

## 💻 Usage

### Direct Invocation (Test)
```bash
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | go run main.go
```

### Claude Desktop

Example configuration, with Claude Desktop running on Windows 11:

```json
"mcpServers": {
    "homelab": {
        "command": "ssh",
        "args": [
            "-T",
            "agent@thinkcentre.local",
            "/home/agent/projects/homelab-mcp/homelab-mcp"
        ],
        "env": {
            "PROGRAMDATA": "C:\\ProgramData",
            "SystemRoot": "C:\\Windows",
            "HOME": "C:\\Users\\${USERNAME}"
        }
    }
}
```

Without specifing the env vars, Claude Desktop runs the process in a "sandbox" type environment. Without proper env configuration, the ssh binary included with Windows 11 was not able to even start up.