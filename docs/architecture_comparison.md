# Architecture Comparison: Custom Multi-Model vs. Native Subagents

This document contrasts the custom multi-model orchestration pattern used in `homelab-mcp` (Claude orchestrating DeepSeek/Gemini via bash shims) against a native Claude Code subagent architecture. 

The core difference lies in **Agentic Tool-Use** versus **Single-Shot Generation**.

## High-Level Comparison

| Feature | Custom Setup (DeepSeek Shims) | Native Setup (Claude Subagents) |
| :--- | :--- | :--- |
| **Agent Paradigm** | Single-shot generation (Blind) | Autonomous agentic loop (Think → Act → Observe) |
| **Tool Access** | None (Relies entirely on prompt) | Native (Read, Grep, Glob, Bash) |
| **Context Gathering** | Orchestrator's responsibility | Subagent's responsibility |
| **Orchestrator Overhead** | High (~7.6M tokens in this project) | Low |
| **Cost** | Extremely low ($0.05 total for Coder) | High (Native API pricing per subagent loop) |
| **Vendor Lock-in** | None (Model-agnostic) | Complete (Tied to Anthropic ecosystem) |

## 1. Custom Setup: The "Blind" Single-Shot Coder
In the custom setup, subagents (like DeepSeek V3/R1) are invoked via simple bash shims. They have no access to the file system and no ability to execute commands.

* **The Process:** The subagent only knows what is explicitly provided in its prompt buffer. 
* **The "Context Tax":** Because the subagent cannot "look around," the main orchestrator (Claude) must act as a micro-manager. Claude has to run the `grep`s, read the interface files, grab the test mocks, and bundle it all into a massive, highly detailed text prompt. 
* **Result:** This explains the 7.6M token context overhead. The orchestrator is forced to build a precise mental model of the codebase and pass it as text.

## 2. Native Setup: Autonomous Agentic Subagents
Native Claude subagents spin up with a blank context window but possess built-in tool access.

* **The Process:** When tasked with adding a feature, a native subagent can autonomously run a `grep` command, read the output, open specific files, analyze existing structs, and *then* write the code.
* **Context Delegation:** The orchestrator simply hands off the high-level objective. The subagent gathers its own context through its internal loop, drastically reducing the orchestrator's token overhead and planning burden.

## 3. The Power of Personas
Regardless of the setup (custom or native), utilizing subagents enforces strict **personas**. Giving an agent a narrow, specific role (e.g., a rigid "Reviewer" vs. a creative "Builder") fundamentally alters its output. 

* **Builder:** Optimized for rapid code generation and fulfilling interface requirements.
* **Reviewer:** Prompted to be highly critical, focusing specifically on edge cases, unhandled errors, and performance (e.g., DeepSeek R1 catching unguarded slice access and unbounded memory allocation).
* **Librarian/Tester:** Focused purely on generating accurate tests or documentation without the distraction of implementation details.

Separating these personas prevents a single model from compromising its critical analysis in favor of confirming its own generated code.