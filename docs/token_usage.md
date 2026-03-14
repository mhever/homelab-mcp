# Token Usage Report — homelab-mcp

Tracking token consumption across the multi-model agent orchestration pipeline used to build this project.

## Model Roles

| Role | Model | Provider | Cost Model |
|------|-------|----------|------------|
| Orchestrator | Claude Sonnet 4.6 | Claude Code (Pro subscription) | Included in subscription |
| Coder | DeepSeek V3 | OpenRouter | Pay-per-token ($5 hard cap) |
| Reviewer | DeepSeek R1 | OpenRouter | Pay-per-token |
| Librarian | Gemini 3.1 Flash-Lite | Google (free tier) | Free |
| Test Writer | Gemini 3.1 Flash-Lite | Google (free tier) | Free |

## Output Tokens (Actual Work Done)

This is the apples-to-apples comparison. Output tokens represent content each model actually generated: orchestration decisions, code, review feedback.

### Phase 2: Docker Tools

3 tools, 8 tests.

| Model | Role | Output Tokens |
|-------|------|--------------|
| Claude Sonnet 4.6 | Orchestrator | 10,584 |
| DeepSeek V3 | Coder | 16,000 |
| DeepSeek R1 | Reviewer | 6,960 |
| Gemini Flash-Lite | Librarian / Tests | Not tracked |
| **Total output** | | **~33,500** |

### Phase 3: Kubernetes Tools

4 tools, 13 new tests (21 total).

| Model | Role | Output Tokens |
|-------|------|--------------|
| Claude Sonnet 4.6 | Orchestrator | 20,411 |
| DeepSeek V3 | Coder | 41,000 |
| DeepSeek R1 | Reviewer | 18,000 |
| Gemini Flash-Lite | Librarian / Tests | Not tracked |
| **Total output** | | **~79,400** |

### Cumulative Output (Phase 2 + 3)

| Model | Output Tokens | Share |
|-------|--------------|-------|
| DeepSeek V3 (coder) | 57,000 | 50% |
| DeepSeek R1 (reviewer) | 24,960 | 22% |
| Claude (orchestrator) | 30,995 | 27% |
| **Total** | **~113,000** | |

### Phase 5: Polish

Makefile, systemd unit, CI workflow, README update.

| Model | Role | Output Tokens |
|-------|------|--------------|
| Claude Sonnet 4.6 | Orchestrator | ~3,000 |
| DeepSeek V3.2 | Coder (infra files) | ~2,500 |
| Gemini Flash-Lite | Librarian (README + todo) | ~1,500 |
| **Total output** | | **~7,000** |

## Cost

| Provider | Phase 2 | Phase 3 | Total |
|----------|---------|---------|-------|
| DeepSeek (OpenRouter) | ~$0.02 | ~$0.03 | **$0.05** |
| Gemini (free tier) | $0.00 | $0.00 | $0.00 |
| Claude (Pro subscription) | ~4% weekly quota | ~3% weekly quota | ~7% weekly quota |

7 tools and 21 tests built for 5 cents in API costs (plus Pro subscription).

## Context Overhead (Claude Only)

Claude Code re-reads the full context window (codebase, CLAUDE.md, plan.md, conversation history) on every interaction. This inflates Claude's raw token count significantly but does not represent generated content.

| Phase | Cache Read Tokens | Cache Creation Tokens | Total Context Overhead |
|-------|------------------|-----------------------|----------------------|
| Phase 2 | 2,342,616 | 230,448 | 2,573,064 |
| Phase 3 | 4,356,462 | 348,405 | 4,704,867 |
| **Total** | **6,699,078** | **578,853** | **7,277,931** |

Phase 3 context overhead was ~1.8x Phase 2 because client-go and Kubernetes types are significantly larger than Docker client types.

## Observations

- DeepSeek produces ~72% of all output tokens (code + reviews). Claude produces ~27% (orchestration). The orchestrator/worker split is working as designed.
- DeepSeek R1 (reviewer) consistently uses 30-40% of V3's (coder) token count per phase. Review is cheaper than generation.
- Phase 3 was ~2.4x Phase 2 in output tokens. Kubernetes tools required more code, more types, and more test cases than Docker tools.
- Claude's context overhead (7.3M tokens) dwarfs actual output (31K tokens) by 235:1. Prompt caching keeps this affordable within Pro quota limits.
- Total DeepSeek cost across both phases: 5 cents. The $5 OpenRouter hard cap provides cost safety with massive headroom remaining.
