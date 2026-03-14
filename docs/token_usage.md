# Token Usage Report — homelab-mcp

Tracking token consumption across the multi-model agent orchestration pipeline used to build this project.

All Claude figures are Sonnet 4.6 only (orchestrator). Opus tokens from separate chat sessions are excluded.

## Model Roles

| Role | Model | Provider | Cost Model |
|------|-------|----------|------------|
| Orchestrator | Claude Sonnet 4.6 | Claude Code (Pro subscription) | Included in subscription |
| Coder (Phase 2-3) | DeepSeek V3 | OpenRouter | Pay-per-token ($5 hard cap) |
| Coder (Phase 4-5) | DeepSeek V3.2 | OpenRouter | Pay-per-token ($5 hard cap) |
| Reviewer | DeepSeek R1 | OpenRouter | Pay-per-token |
| Librarian | Gemini 3.1 Flash-Lite | Google (free tier) | Free |
| Test Writer | Gemini 3.1 Flash-Lite | Google (free tier) | Free |

## Output Tokens Per Phase (Actual Work Done)

Output tokens represent content each model actually generated: orchestration decisions, code, review feedback. These are per-phase deltas, not cumulative.

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

4 tools, 13 new tests.

| Model | Role | Output Tokens |
|-------|------|--------------|
| Claude Sonnet 4.6 | Orchestrator | 9,827 |
| DeepSeek V3 | Coder | 25,000 |
| DeepSeek R1 | Reviewer | 11,040 |
| Gemini Flash-Lite | Librarian / Tests | Not tracked |
| **Total output** | | **~45,900** |

### Phase 4: FluxCD Tool

1 tool (k8s_flux_status), 4 new tests. Coder switched from DeepSeek V3 to V3.2.

| Model | Role | Output Tokens |
|-------|------|--------------|
| Claude Sonnet 4.6 | Orchestrator | 14,795 |
| DeepSeek V3.2 | Coder | 32,900 |
| DeepSeek R1 | Reviewer | 10,000 |
| Gemini Flash-Lite | Librarian / Tests | Not tracked |
| **Total output** | | **~57,700** |

### Phase 5: Polish

Makefile, systemd unit, CI workflow, README, docs.

| Model | Role | Output Tokens |
|-------|------|--------------|
| Claude Sonnet 4.6 | Orchestrator | 3,734 |
| DeepSeek V3.2 | Coder | 1,000 |
| DeepSeek R1 | Reviewer | 0 |
| Gemini Flash-Lite | Librarian | Not tracked |
| **Total output** | | **~4,700** |

## Cumulative (All Phases)

| Model | Output Tokens | Share |
|-------|--------------|-------|
| DeepSeek V3 (coder, Phase 2-3) | 41,000 | 29% |
| DeepSeek V3.2 (coder, Phase 4-5) | 33,900 | 24% |
| DeepSeek R1 (reviewer) | 28,000 | 20% |
| Claude Sonnet 4.6 (orchestrator) | 38,940 | 27% |
| **Total** | **~141,800** | |

## DeepSeek Cost (OpenRouter)

Total tokens consumed on OpenRouter across all phases: ~103K (V3 41K + V3.2 34K + R1 28K).

Total OpenRouter spend: **~$0.08** out of a $5 hard cap.

10 tools and 25 tests built for 8 cents.

## Context Overhead (Claude Sonnet Only)

Claude Code re-reads the full context window (codebase, CLAUDE.md, plan.md, conversation history) on every interaction. This inflates Claude's raw token count but does not represent generated content.

| Phase | Output Tokens | Cache Read | Cache Creation | Total Context Overhead |
|-------|--------------|------------|----------------|----------------------|
| Phase 2 | 10,584 | 2,342,616 | 230,448 | 2,573,064 |
| Phase 3 | 9,827 | 2,013,846 | 117,957 | 2,131,803 |
| Phase 4 | 14,795 | 1,941,823 | 146,911 | 2,088,734 |
| Phase 5 | 3,734 | 754,527 | 32,479 | 787,006 |
| **Total** | **38,940** | **7,052,812** | **527,795** | **7,580,607** |

Context overhead to output ratio: 195:1. Prompt caching is what makes this affordable on a Pro subscription.

## Observations

- DeepSeek handles 73% of all output tokens (code + reviews). Claude handles 27% (orchestration). The worker/orchestrator split works as designed.
- DeepSeek R1 (reviewer) uses 27-44% of the coder's token count per phase. Review is consistently cheaper than generation.
- Phase 4 (FluxCD, 1 tool) generated more output tokens than Phase 3 (4 tools). The dynamic client + unstructured parsing for CRDs required more code and more back-and-forth than standard client-go operations.
- Phase 5 (polish) was the cheapest phase by far. Infrastructure files (Makefile, CI, systemd) are small and don't need review loops.
- Total DeepSeek cost: 8 cents. The $5 hard cap on OpenRouter was never close to being tested.

## Lesson: Gemini Free Tier vs DeepSeek

Gemini Flash-Lite on the free tier served as librarian and test writer. In practice, DeepSeek V3/V3.2 through OpenRouter would have been a better choice for both roles. Flash-Lite's output quality required more correction cycles, and the free tier quota limits added friction. At $0.08 total for all DeepSeek usage across the entire project, cost is not a meaningful differentiator.

The multi-model setup was still valuable as a learning exercise: it demonstrated how to build vendor-agnostic orchestration with bash shims, how different models behave on the same tasks, and how to manage cost across providers. But for future projects, consolidating on DeepSeek (or a Claude Haiku-tier model) for all worker roles would be simpler and produce better results.