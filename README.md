# homelab-mcp

Architectural TL;DR: A Go-based MCP server for mixed-workload home labs (k3s, Docker, Linux) built as an architectural study in multi-model agent orchestration. It utilizes a "blind subagent" pattern where Claude orchestrates DeepSeek (V3/R1) and Gemini to manage context flow and observability. The entire implementation was delivered with high transparency for a total DeepSeek API cost of approximately $0.08, using Claude paid subscription and Gemini free tier.

A Go-based [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server designed for home lab operations. Unlike generic Kubernetes servers, `homelab-mcp` is built for mixed-workload environments, exposing tools for **k3s**, **Docker**, and **Bare-metal system metrics** through a single interface.

This project is part of my professional portfolio, demonstrating Go development, systems engineering, and **agentic development workflows**.

Note: this setup was deliberately over-scoped for a project of this size. The goal was hands-on experience with multi-vendor orchestration, not optimal delivery speed.

## 🚀 Current Status: Phase 5 (Complete)

- [x] **Phase 0: Scaffold** - MCP SDK wiring & stdio transport.
- [x] **Phase 1: System Tools** - `system_overview` and `system_disk` via `gopsutil`.
- [x] **Phase 2: Docker Tools** - Container management & logs.
- [x] **Phase 3: Kubernetes Tools** - Pods, Events, and cluster overview.
- [x] **Phase 4: FluxCD Tool** - GitRepository and Kustomization reconciliation status.
- [x] **Phase 5: Polish** - Makefile, systemd service unit, CI workflow, and documentation.

For the full implementation roadmap, see [plan.md](./docs/plan.md).

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
| `k8s_cluster_overview` | Node status and per-namespace pod count summary |
| `k8s_pods` | List/filter pods by namespace and status |
| `k8s_pod_logs` | Get logs from a pod (args: namespace, pod, container, tail) |
| `k8s_events` | List cluster events filtered by namespace and type |
| `k8s_flux_status` | Show FluxCD GitRepository and Kustomization reconciliation status |

## 🤖 Built with Agents

More on the agents: [agents.md](./docs/agents.md)
Token consumption: [token_usage.md](./docs/token_usage.md)

This repository is developed using a multi-agent orchestration pattern where Claude acts as the orchestrator and delegates to specialist subagents:

| Role | Model | Task |
|------|-------|------|
| **Orchestrator** | Claude Sonnet 4.6 (Claude Code) | Planning, routing, and verification |
| **Builder** | DeepSeek V3.2 (`deepseek/deepseek-v3.2` via OpenRouter) | Code generation and file edits |
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

## 🚀 Quick Start

```bash
make build
sudo make install
# Enable as a service:
sudo cp deploy/homelab-mcp.service /etc/systemd/system/
sudo systemctl enable --now homelab-mcp
```

The project includes a `Makefile` with targets for `build`, `test`, `lint`, `install`, and `deploy`. A ready-to-use systemd service unit is provided at `deploy/homelab-mcp.service`.

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

Claude Desktop runs the process in a sandbox type environment. Without proper env configuration, the ssh binary included with Windows 11 was not able to even start up.
