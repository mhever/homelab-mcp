## Workflow Orchestration

### 1. Mandatory Handoffs (Gemini / DeepSeek Integration)
- **Code Generation & Modification:** You MUST delegate all code writing, refactoring, and file edits to the `deepseek-coder` subagent. Do not apply code yourself.
- **Code Review:** You MUST pass all modified files to the `deepseek-reviewer` subagent.
- **Iteration:** If `deepseek-reviewer` flags issues, feed the exact output back to `deepseek-coder` for remediation. Repeat until the reviewer passes the code.

### 2. Execution Rules
- **Plan First:** For tasks requiring 3+ steps, write a checklist to `tasks/todo.md` before execution.
- **Halt on Failure:** If a step fails twice, STOP. Do not loop blindly. Re-evaluate the plan.
- **Verification:** Run tests and check logs to prove the implementation works before marking a task complete.
- **Autonomous Resolution:** If given a bug report, use `gemini-coder` to resolve it directly based on logs/errors. 

### 3. Task Management & Memory
- **Tracking:** Update `tasks/todo.md` as items are completed.
- **Lessons Learned:** Upon user correction, immediately log the required behavior change in `tasks/lessons.md`.
- **Review History:** Read `tasks/lessons.md` at the start of every session to prevent repeating mistakes.

### 4. Code Standards
- Keep changes localized. Only touch necessary files.
- Resolve root causes; implement zero temporary workarounds.

Current plan is in @plan.md

# CLAUDE.md - homelab-mcp Orchestration Guide

## Project Context
Building a Go-based MCP server to monitor and operate mixed home lab infrastructure (k3s, Docker, bare-metal). Development is performed on an Ubuntu 24.04 ThinkCentre m715q via SSH.

## Workflow Orchestration

### 1. Mandatory Handoffs (Agentic Specialist Stack)
- **Code Generation & Implementation:** Delegate all code writing, refactoring, and file edits to `deepseek-coder`. It is your high-velocity builder.
- **Critical Logic Review:** Pass all modified files or concurrency logic to `deepseek-reviewer`. Use its reasoning capabilities to audit for race conditions and architectural alignment.
- **Testing & Mocks:** Delegate the creation of `_test.go` files and mock implementations to `gemini-tester`. This preserves paid credits for logic while utilizing free tokens for boilerplate.
- **Documentation & Sync:** Use `gemini-librarian` to update `README.md`, `tasks/todo.md`, and MCP schemas. It is responsible for keeping the "paperwork" in sync with the code.

### 2. Execution & Iteration Loop
1. **Plan:** Write a checklist to `tasks/todo.md` for any task requiring 3+ steps.
2. **Build/Fix:** Call `deepseek-coder` to implement.
3. **Audit:** Call `deepseek-reviewer`. If it flags issues, feed the output back to `deepseek-coder` for remediation. Repeat until the reviewer passes.
4. **Boilerplate:** Once the logic is final, call `gemini-tester` for unit test coverage.
5. **Finalize:** Call `gemini-librarian` to update docs before marking the task complete.

### 3. Task Management & Memory
- **Halt on Failure:** If a subagent step fails twice, **STOP**. Re-evaluate the strategy based on the specific error logs.
- **Continuous Learning:** Immediately log all model-specific quirks, environment issues, or user corrections into `tasks/lessons.md`.
- **Session Init:** Read `tasks/lessons.md` at the start of every session to avoid repeating known mistakes (especially model-flag drift).

### 4. Code Standards
- **Localization:** Keep changes localized; only touch necessary files.
- **Root Cause:** Resolve root causes; implement zero temporary workarounds.
- **MCP Patterns:** Use `mcp.TextContent` for tool responses. Return errors as `CallToolResult` with `IsError: true` to avoid crashing the stdio transport.
- **Logging:** All logs MUST go to `os.Stderr`. `os.Stdout` is reserved for MCP JSON-RPC.

## Development Environment
- **Root Directory:** `~/projects/homelab-mcp`
- **Primary Model:** Sonnet 4.6 (Orchestrator)
- **Subagent Tools:** `deepseek` (OpenRouter shim) and `gemini` (local CLI)