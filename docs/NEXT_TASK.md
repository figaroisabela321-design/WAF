# NEXT_TASK

**Review Result:** REWORK IN PROGRESS

**Date:** 2026-09-17 (Asia/Shanghai)

## Agent roles

| Role | Owner |
|------|-------|
| Code / implementation | `waf` (Grok Bot) |
| Architecture review / PASS-REWORK | ChatGPT |
| Shared source of truth | GitHub |

## Current work

Phase 1 Security Hardening implemented on branch `fix/phase1-security-hardening`. Waiting for ChatGPT review (PASS / REWORK). Do **not** invent SamWaf tasks.

## Explicit holds

- Do **NOT** start SamWaf integration.
- Do **NOT** start Coraza / CRS / Agent / Node data-plane / ClickHouse / Kafka / Dameng / frontend work.
- Do **NOT** invent next implementation tasks beyond ChatGPT instructions.
- Do **NOT** merge this PR until ChatGPT review completes.

## Waiting for

ChatGPT must return either:

1. **PASS** — with an explicit, ordered next-task list; or
2. **REWORK** — with concrete, scoped fix instructions.

## References

- `docs/PROJECT_STATE.md`
- `docs/AI_HANDOFF.md`
- `docs/reviews/phase1-security-hardening-gates.log`
- PR: fix/phase1-security-hardening → main
