# PROJECT_STATE

Date: 2026-09-17 23:26 Asia/Shanghai (UTC+8)  
Agents: `waf` (Grok Bot, code/dev) · ChatGPT (architecture review / PASS-REWORK) · GitHub (shared source of truth)

## Project Goal

Build a government-facing, privately deployable Web Application Firewall platform with a clear control-plane / data-plane split: unified site/node/policy/rule management, RBAC + audit on the control plane, and replaceable WAF runtime engines on the data plane.

## Target Architecture

```
Management Console (future UI)
        │
        ▼
  waf-control (Go control plane API)
        │
   PostgreSQL (config / RBAC / audit)
        │
   Node management API / Agent (future)
        ▼
  waf-node (data plane) — NOT created this round
        │
  WAF Runtime (SamWaf embed evaluation: BLOCKED; alternatives B/C pending ChatGPT)
        │
  Event/Alert stores (future; ClickHouse adapter reserved as noop)
```

## Current Phase

**Phase 2A — SamWaf Node Runtime POC (ANALYSIS COMPLETE → BLOCKED)**

In scope this round: freeze SamWaf upstream outside repo; mandatory analysis; honest GO/NO-GO; docs; push PR; waf-control quality gates.

Out of scope: Phase 2B; Coraza upgrades; implementing alternative B (Caddy+Coraza) or C (custom proxy+Coraza); merging to main; fake success via stock SamWaf UI.

## SamWaf Runtime POC decision

**`SAMWAF_RUNTIME_POC = BLOCKED`**

Pinned upstream: `d975b12ec0a4757ca0e9698accd373dfee5f7c71` (`v1.3.25-beta.6-2-gd975b12`), Apache-2.0, Coraza v3.3.3, CRS 4.9.0-dev.

Blocking fusion: SQLite (`GWAF_LOCAL_DB`) as host authority + process-wide `global`/`globalobj` singletons + management web (`wafmangeweb` :26666) + engine lifecycle in `cmd/samwaf`. Cannot meet embeddable Runtime (Validate/Apply/LKG/CP-offline with PG authority) without large core rewrite.

Details: `docs/SAMWAF_INTEGRATION_ANALYSIS.md`, `docs/SAMWAF_NODE_POC.md`, `docs/THIRD_PARTY_SAMWAF.md`.

## Current Implemented Modules (from real code)

### Repository layout

| Path | Status |
|------|--------|
| `waf-control/` | Implemented Go module (`github.com/gov-waf/waf-control`) |
| `go.work` | Present; `use ./waf-control` |
| `deploy/Dockerfile`, `deploy/docker-compose.yml` | Present (postgres + waf-control) |
| `waf-agent/`, `waf-policy/`, `waf-rules/` | Placeholder READMEs only |
| `waf-node/` | **Not created** (POC BLOCKED) |
| `docs/` | architecture, ER, API examples, OpenAPI, SamWaf analysis, reviews |

### `waf-control` (unchanged this round)

Phase 1 security hardening remains on `main` (merged PR #1). This branch adds analysis docs only; no CP publish/events endpoints added (would be meaningless without node runtime).

## Core Architecture Principles

1. Do not modify Coraza core source.
2. Do not modify OWASP CRS official rule files.
3. Treat the WAF detection engine as replaceable.
4. Keep Policy (business model) decoupled from SecLang / engine-specific artifacts.
5. Keep control plane and data plane strictly separated.
6. WAF nodes must not require the management plane to be online for traffic handling.
7. Control-plane / management failures must never impact live business traffic.
8. All deployable configurations must be versioned.
9. All deployable configurations must support rollback.
10. All significant management operations must be audit-logged.
11. Agents/nodes must not connect directly to PostgreSQL.
12. Production secrets from environment/secret stores — never commit `.env`, keys, or PEMs.
13. GitHub is the shared source of truth; `waf` owns code/dev; ChatGPT owns architecture review (PASS / REWORK).
14. Do not start Phase 2B / Coraza upgrades / alternative B or C until ChatGPT returns PASS with an explicit next-task list after reviewing this BLOCKED POC.

## Control-Plane Security Principles (Phase 1 — still in force)

Production fail-fast `APP_ENV`, identity-only JWT + live RBAC, user/role transactions, audit ERROR logging, trusted-proxy XFF chain, password min length, body size limit, swagger production gate, CI gates for `waf-control`.
