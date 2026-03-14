# Agentic Lessons Learned: homelab-mcp

This file serves as the persistent memory for the orchestration layer. All agents must review this file at the start of a session to prevent regression and optimize token usage.

## 🤖 Multi-Agent Orchestration & Model Rules

### Gemini (Free Tier Optimization)
- **Constraint:** The Free Tier for Gemini is extremely sensitive. 
- **Rule:** NEVER use `gemini-3.1-pro-preview` or standard `gemini-3-flash-preview` (Risk: `limit: 0` or `limit: 5` errors).
- **Rule:** ALWAYS use `--model gemini-3.1-flash-lite-preview`. 
- **Behavior:** Flash-Lite requires "Negative Constraints" (e.g., "DO NOT use X") in its prompt to prevent model drift.

### DeepSeek (OpenRouter / Reasoning)
- **Rule:** Use `deepseek/deepseek-chat` (V3) for high-speed boilerplate and implementation. It is cost-effective for large I/O.
- **Rule:** Use `deepseek/deepseek-r1` (R1) for the `gemini-reviewer` role. It excels at finding race conditions and logic flaws in Go.
- **FinOps:** Maintain a credit limit on OpenRouter keys to prevent runaway loops from depleting the balance.

### Claude (Orchestrator)
- **Efficiency:** Sonnet 4.6 is the preferred orchestrator over Opus to preserve weekly quotas.
- **Context Management:** Bundle tasks (e.g., "Implement + Review + Test") into a single prompt to reduce the number of round-trips and context re-summarization.

## 💻 Environment & Infrastructure (ThinkCentre m715q)

### Docker & Permissions
- **Issue:** The MCP server requires direct access to `/var/run/docker.sock`.
- **Lesson:** The `agent` user must be a member of the `docker` group. If tools fail with 'permission denied', verify membership with `groups`.
- **Note:** SSH sessions may need to be restarted after group changes for permissions to take effect.

### Go / MCP Patterns
- **Logging:** All logs must go to `os.Stderr`. `os.Stdout` is strictly reserved for the MCP JSON-RPC transport.
- **Error Handling:** Tools should return `IsError: true` within the `CallToolResult` rather than returning a Go error to the server, which would crash the transport.
- **Output:** Prefer formatted plain text or Markdown tables over raw JSON in tool responses; LLMs reason better with structured text.

## 🛠️ Performance Tuning
- **Latency:** When running local LLM inference on the ThinkCentre, MCP tool timeouts should be increased to account for CPU contention.