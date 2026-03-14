# AGENTS.md — Orchestration Setup for homelab-mcp

This document describes how this project was built: which AI models handle which roles, how they hand off to each other, the tooling that connects them, and the real lessons learned from running this in production.

The `.claude/agents/` directory shows *what* each agent does. This document explains *why* the setup is structured this way.

---

## The Core Idea

Claude Code (Sonnet 4.6) acts as the **orchestrator**: it reads requirements, writes the plan, routes tasks to specialists, and verifies results. It does not write production code itself.

Three specialist subagents handle the actual work, each chosen for cost-efficiency at their specific task:

| Role | Model | Why This Model |
|------|-------|----------------|
| **Orchestrator** | Claude Sonnet 4.6 | Best reasoning for routing and verification; Sonnet preferred over Opus to preserve weekly quota |
| **Builder** | DeepSeek V3.2 (`deepseek/deepseek-V3.2`) | Fastest and cheapest for large code generation with file context; handles 10k+ token I/O without breaking the bank |
| **Reviewer** | DeepSeek R1 (`deepseek/deepseek-r1`) | Reasoning model that shows chain-of-thought; catches race conditions and subtle logic bugs that standard models miss |
| **Tester** | Gemini Flash-Lite (`gemini-3.1-flash-lite-preview`) | Free tier; well-suited for mechanical test boilerplate where inference quality matters less than volume |
| **Librarian** | Gemini Flash-Lite (`gemini-3.1-flash-lite-preview`) | Free tier; doc sync is low-stakes and low-complexity |

---

## The Iteration Loop

Every feature phase follows this sequence:

```
1. Plan     → Orchestrator writes tasks/todo.md checklist
2. Build    → deepseek-coder implements the code
3. Audit    → deepseek-reviewer reads the output, reports findings
4. Fix      → If findings exist, feed reviewer output back to deepseek-coder
              Repeat until reviewer passes
5. Test     → gemini-tester writes _test.go files
6. Docs     → gemini-librarian updates README and todo.md
7. Verify   → Orchestrator runs go build ./... and go test ./...
```

Steps 3–4 form a tight remediation loop. The reviewer output is passed verbatim to the coder with instructions to fix each finding. In practice, one round of review + one round of fixes is usually sufficient.

---

## The DeepSeek Shim

DeepSeek is accessed via [OpenRouter](https://openrouter.ai), not a native SDK. The connection is a small bash shim at `tools/deepseek` (installed system-wide at `/usr/local/bin/deepseek`).

```bash
#!/bin/bash
# Usage: deepseek "deepseek/deepseek-chat" "Your Prompt"

. ~/.profile || true   # Load env vars — see "Lessons Learned" below

MODEL=$1
PROMPT=$2

JSON_PROMPT=$(echo "$PROMPT" | jq -aRs .)

echo "" >> /tmp/deepseek.log
date >> /tmp/deepseek.log
echo "$PROMPT" >> /tmp/deepseek.log

RESPONSE=$(curl -s -X POST "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$MODEL\",
    \"messages\": [{\"role\": \"user\", \"content\": $JSON_PROMPT}]
  }")
echo "$RESPONSE" >> /tmp/deepseek.log

echo "$RESPONSE" | jq -r '.choices[0].message.content // .error.message'
```

**What it does:** wraps a single OpenRouter chat completion call. Takes the model ID and prompt as positional arguments, escapes the prompt as JSON via `jq`, and extracts the response text (or error message) on the way out.

**The log:** every call appends the prompt and raw response to `/tmp/deepseek.log`. This is invaluable for debugging when a subagent produces unexpected output — you can see exactly what was sent and what came back.

**Installation:** copy `tools/deepseek` to `/usr/local/bin/deepseek` and make it executable:
```bash
sudo cp tools/deepseek /usr/local/bin/deepseek
sudo chmod +x /usr/local/bin/deepseek
export OPENROUTER_API_KEY="sk-or-..."   # add to ~/.profile for persistence
```

---

## Agent Definitions

Agent behaviour is defined in `.claude/agents/`. Claude Code loads these automatically when it launches subagents. Each file uses a YAML frontmatter block to declare the agent's name, description, and allowed tools, followed by the system prompt.

```
.claude/agents/
├── deepseek-coder.md      # Builder: runs deepseek-chat via the shim
├── deepseek-reviewer.md   # Reviewer: runs deepseek-r1 via the shim
├── gemini-tester.md       # Tester: runs gemini CLI for test generation
└── gemini-librarian.md    # Librarian: runs gemini CLI for doc sync
```

The agents do not call the AI APIs directly — they invoke the shim (for DeepSeek) or the `gemini` CLI (for Gemini), parse the output, and apply it using their available file tools (`Write`, `Edit`, `Bash`).

---

## Lessons Learned

These are real problems that surfaced during development, not hypothetical concerns.

### 1. Environment variables aren't inherited by subagent shells

**Problem:** `$OPENROUTER_API_KEY` was set in `~/.bashrc`, but subagents run non-interactive shells that never source it. Every DeepSeek call silently failed with an empty bearer token.

**Fix:** Moved the export to `~/.profile` (sourced by login shells) and added `. ~/.profile || true` at the top of the shim so it self-heals even in non-login contexts. The `|| true` prevents the shim from failing if `~/.profile` doesn't exist.

**Takeaway:** Never rely on `~/.bashrc` for credentials used in automated pipelines. Use `~/.profile`, a dedicated credentials file, or a secrets manager.

### 2. Gemini's free tier is extremely quota-sensitive

**Problem:** `gemini-3.1-pro-preview` and `gemini-3-flash-preview` both hit `limit: 0` or `limit: 5` errors almost immediately on the free tier. This blocked the tester and librarian roles entirely on first attempt.

**Fix:** The only model that consistently works within free tier limits is `gemini-3.1-flash-lite-preview`. The agent definitions use hard negative constraints ("DO NOT use X") to prevent the model from drifting to other variants when interpreting prompts.

**Takeaway:** Free tier quota enforcement is per-model, not per-API-key. Pin the model explicitly and add negative constraints to prevent drift.

### 3. Unused imports need a real fix, not a blank identifier hack

**Problem:** `deepseek-coder` generated `docker/tools.go` with `import "github.com/docker/docker/api/types"` but only used the type implicitly (via method return values). It added `var _ = types.Container{}` to silence the compiler.

**Fix:** Removed the import and the hack entirely. In Go, you don't need to import a type to use its fields when the value comes from a method call — the compiler resolves it from the callee's package.

**Takeaway:** Blank identifier import suppressors are a code smell. Review generated code for this pattern before accepting it.

### 4. DeepSeek R1 is worth the cost for security-critical review

**Problem:** Standard generation models (including DeepSeek V3) don't reliably catch subtle bugs like unguarded slice indexing (`c.Names[0]` with no bounds check), unbounded memory allocation from untrusted input, or blocking contexts in startup paths.

**Result:** In the Phase 2 review, R1 found: an unguarded `Names[0]` access that panics on containers with empty name slices, a `make([]byte, size)` with no upper bound on the `size` from the Docker stream header (potential OOM), and a `context.Background()` ping that blocks indefinitely if Docker is unresponsive. All three were real bugs.

**Takeaway:** Use a reasoning model for the review step. The chain-of-thought output also doubles as documentation for *why* each finding matters.

### 5. Bundle context aggressively for the builder

**Problem:** DeepSeek V3 produces much better output when given the full context upfront: existing code it should match, patterns it should follow, constraints it must not violate. Under-specified prompts produce code that compiles but diverges from project conventions.

**Fix:** The `deepseek-coder` subagent prompts include: the existing file to model against, the `mcputil` helper signatures, the exact import paths, and explicit constraints (e.g., "no raw Go errors from handlers", "all logs to stderr").

**Takeaway:** For code generation, more context is almost always better. The token cost of including existing code in the prompt is lower than the cost of a review-fix iteration.

---

## Cost Profile

All figures are approximate and based on Phase 2 usage:

| Step | Model | Typical tokens | Cost estimate |
|------|-------|---------------|---------------|
| Build (3 files) | deepseek-V3.2 | ~8k | ~$0.006 |
| Review | deepseek-r1 | ~7k | ~$0.01 |
| Fix iteration | deepseek-V3.2 | ~8k | ~$0.006 |
| Tests | gemini-flash-lite | ~14k | free |
| Docs | gemini-flash-lite | ~8k | free |

The orchestrator (Claude Sonnet 4.6) uses the Claude Code subscription, so its cost is fixed regardless of usage volume.
Using "npx ccusage@latest daily --breakdown --compact" the Phase 2 Sonnet orchestrator used 11k tokens.

**OpenRouter tip:** Set a per-key credit limit on your OpenRouter account. Agentic loops can run away if a subagent retries on failure without a circuit breaker.
