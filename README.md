# homelab-mcp

A Go-based [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server designed for home lab operations. Unlike generic Kubernetes servers, `homelab-mcp` is built for mixed-workload environments, exposing tools for **k3s**, **Docker**, and **Bare-metal system metrics** through a single interface.

This project is part of my professional portfolio, demonstrating Go development, systems engineering, and **agentic development workflows**.

## 🚀 Current Status: Phase 2 (Complete)

- [x] **Phase 0: Scaffold** - MCP SDK wiring & stdio transport.
- [x] **Phase 1: System Tools** - `system_overview` and `system_disk` via `gopsutil`.
- [x] **Phase 2: Docker Tools** - Container management & logs.
- [ ] **Phase 3: Kubernetes Tools** - Pods, Events, and FluxCD status.

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

## 🤖 Built with Agents

This repository is developed using a multi-agent orchestration pattern:
- **Orchestrator:** Claude 4.6 Opus (via Claude Code) handles high-level planning and decision-making.
- **Workers:** Gemini 3.1 Flash-Lite subagents (via `gemini-cli`) handle isolated code generation, audits, and analysis.

This "Hybrid AI" approach allows for high-velocity development while maintaining strict cost and quota management.

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