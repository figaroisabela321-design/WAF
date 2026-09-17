# NEXT_TASK

**Status:** PENDING — waiting for ChatGPT architecture review (PASS / REWORK)

**Date:** 2026-09-17 (Asia/Shanghai)

## Agent roles

| Role | Owner |
|------|-------|
| Code / implementation | `waf` (Grok Bot) |
| Architecture review / PASS-REWORK | ChatGPT |
| Shared source of truth | GitHub |

## Explicit holds

- Do **NOT** start SamWaf integration.
- Do **NOT** start Coraza / CRS / Agent / ClickHouse / Kafka / Dameng integration.
- Do **NOT** invent next implementation tasks.
- Do **NOT** redesign architecture or perform large refactors while status is PENDING.

## Waiting for

ChatGPT must return either:

1. **PASS** — with an explicit, ordered next-task list the `waf` agent may execute; or  
2. **REWORK** — with concrete, scoped fix instructions for the `waf` agent.

Until then, `NEXT_TASK` remains empty of implementation work.

## Current baseline references

- `docs/PROJECT_STATE.md`
- `docs/AI_HANDOFF.md`
- `docs/reviews/phase1-baseline-check.log`
