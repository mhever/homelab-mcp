# Orchestration Trace: Adding the Version Tool

This document traces the multi-model execution flow for adding a simple `version` tool to the `homelab-mcp` server. It demonstrates how Claude Code acts as the orchestrator, delegating code generation to a DeepSeek subagent via bash shims.

## Execution Flow

### 1. The Trigger
The workflow begins with the initial request to the orchestrator (Claude Code).
* **Actor:** User
* **Input:** "Add a new MCP tool called `version` that returns a fixed static string with the current server version."

### 2. Orchestrator Context Gathering & Planning
Claude Code intercepts the request, avoids writing the production code directly, and prepares the handoff.
* **Actor:** Claude (Orchestrator)
* **Actions:**
  * Reads `CLAUDE.md` to confirm orchestration rules and mandatory handoffs.
  * Scans existing source files (e.g., `main.go`, tool registration patterns) to gather context.
  * Formulates a specialized prompt for the Coder agent containing the context, the exact requirements, and formatting constraints.

### 3. Subagent Invocation
Claude executes the local bash shim to pass the task to the DeepSeek model.
* **Actor:** Claude -> bash shim (`tools/deepseek`)
* **Actions:**
  * Claude runs the shim, passing the prepared prompt as an argument or via stdin.
  * The shim sources `~/.profile` to load the `OPENROUTER_API_KEY`.
  * The shim makes the API call to OpenRouter.

### 4. Code Generation
DeepSeek processes the prompt and generates the Go code.
* **Actor:** DeepSeek V3 (Coder)
* **Actions:**
  * Generates the Go code for the `version` tool (implementing the `CallToolResult` return type with `IsError: false`).
  * Formats the output as plain text/code blocks.
  * The bash shim logs the raw prompt and DeepSeek's response to `/tmp/deepseek.log` for debugging and audit purposes.
  * The shim returns the generated code to Claude's stdout.

### 5. Integration & Verification
Claude receives the payload and applies it to the codebase.
* **Actor:** Claude (Orchestrator)
* **Actions:**
  * Parses DeepSeek's response and writes the new tool logic into the Go project.
  * Registers the `version` tool in the MCP server initialization block.
  * Executes `go build` to verify compilation.
  * Executes `go test` to validate the integration against existing interface-based abstractions.
  * Reports task completion to the user.