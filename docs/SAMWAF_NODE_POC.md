# SamWaf Node Runtime POC (Phase 2A)

Date: 2026-09-17 23:26 Asia/Shanghai (UTC+8)  
Branch: `feat/samwaf-node-poc`  
Result: **`SAMWAF_RUNTIME_POC = BLOCKED`**

## Goal

Honest feasibility POC: can SamWaf become an embeddable data-plane Runtime behind gov-waf NodeConfig (Validate/Apply/LKG/events/heartbeat) **without** rewriting large fused core?

## Result

**BLOCKED.** Analysis of pinned upstream `d975b12` shows SQLite authority, process-wide `global`/`globalobj` singletons, and management web/API lifecycle are tightly fused with the proxy/WAF engine. Meeting gov-waf Runtime constraints would require a large core rewrite / sustained fork — out of scope for Phase 2A.

Forbidden fake-success paths were **not** used:

- Did not run stock SamWaf binary + SamWafWeb/API as “POC success”
- Did not hand-roll SQLi/XSS regex “WAF”
- Did not treat full process restart as only Apply
- Did not create a fake `waf-node` that claims Coraza/CRS without real integration

## What was delivered

| Artifact | Status |
|----------|--------|
| Upstream freeze outside repo | Done — `/workspace/upstream/SamWaf` @ `d975b12` |
| `docs/SAMWAF_INTEGRATION_ANALYSIS.md` | Done — 20 questions + GO/NO-GO |
| `docs/THIRD_PARTY_SAMWAF.md` | Done |
| `docs/SAMWAF_NODE_POC.md` | This file |
| `waf-node/` module + `gov-waf-node` binary | **Not created** (BLOCKED) |
| CP `config/publish` + node events/heartbeat | **Not implemented** |
| E2E curls (observe/protect SQLi/XSS) | **N/A** — no node binary |
| SamWaf Management UI used | **NO** |

## Curl / E2E results

**N/A — BLOCKED before implementation.** No listener on `:18081`, no management `:19090`, no poc-backend exercise in this PR.

## Quality gates

| Module | gofmt | vet | test | race | build |
|--------|-------|-----|------|------|-------|
| `waf-control` | PASS | PASS | PASS | PASS | PASS |
| `waf-node` | N/A | N/A | N/A | N/A | N/A |

Log: `docs/reviews/phase2a-samwaf-node-poc-gates.log`

## Alternatives (not started)

- A: SamWaf as embeddable Runtime — **blocked**
- B: Caddy + Coraza — documented only; **not started**
- C: Custom proxy + Coraza — documented only; **not started**

## Next

`Review Result: PENDING CHATGPT REVIEW` — do **not** start Phase 2B or Coraza upgrades until ChatGPT returns PASS/REWORK with an explicit task list.
