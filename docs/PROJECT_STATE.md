# PROJECT_STATE

Date: 2026-09-17 (Asia/Shanghai)  
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
   AgentClient (future)
        ▼
  waf-agent nodes (data plane)
        │
  WAF Runtime (planned evaluation: Coraza+CRS and/or SamWaf — undecided)
        │
  Event/Alert stores (future; ClickHouse adapter reserved as noop)
```

Phase 1 implements only the control-plane backend. Data-plane engines, Agent protocol, ClickHouse, Kafka, Dameng, etc. are out of scope until ChatGPT issues PASS with an explicit task list.

## Current Phase

**Phase 1 — Control-plane baseline + AI collaboration workflow**

In scope now: inventory, quality gates, security review, `.gitignore`, collaboration docs (`PROJECT_STATE` / `AI_HANDOFF` / `NEXT_TASK`), minimal README/compose fixes, git baseline commit.

Not in scope: SamWaf / Coraza / CRS / Agent / ClickHouse / Kafka / Dameng integration; architecture redesign; large refactors; new business features.

## Current Implemented Modules (from real code)

### Repository layout

| Path | Status |
|------|--------|
| `waf-control/` | Implemented Go module (`github.com/gov-waf/waf-control`) |
| `go.work` | Present; `use ./waf-control` |
| `deploy/Dockerfile`, `deploy/docker-compose.yml` | Present (postgres + waf-control) |
| `waf-agent/`, `waf-policy/`, `waf-rules/` | Placeholder READMEs only |
| `docs/` | architecture, ER, API examples, OpenAPI, reviews |

### `waf-control/internal` (real packages)

| Package | Role |
|---------|------|
| `config` | Env-based config (`HTTP_ADDR`, `DATABASE_URL`, `JWT_*`, `ADMIN_*`, `LOG_LEVEL`) |
| `db` | pgx pool + golang-migrate runner |
| `httpx` | Unified envelope, errors, Recover / RequestID / AccessLog |
| `log` | slog helpers |
| `auth` | Login, JWT HS256, RBAC, users/roles/permissions, bcrypt hasher, SeedAdmin |
| `site` | Site CRUD (Handler → Service → Repository → PG) |
| `node` | Node CRUD |
| `policy` | Policy CRUD |
| `rule` | Rule CRUD |
| `audit` | Audit log service + middleware |
| `event` | Noop EventStore + list handler |
| `alert` | Noop AlertStore + list handler |

### Other real pieces

- `cmd/server/main.go` — chi router, migrations on boot, admin seed, health/ready, swagger
- `migrations/` — `0001_init`, `0002_seed` (permissions/roles; no plaintext admin password in SQL)
- `pkg/adapters/{coraza,agent,clickhouse}` — interfaces + noop only
- `pkg/pagination` — pagination helper
- Tests present: `internal/auth/hasher_test.go`, `internal/httpx/response_test.go`, `internal/site/service_test.go`

## Planned Architecture Decision

**SamWaf Runtime — evaluation only.**

- Do **not** integrate SamWaf (or Coraza/CRS/Agent/ClickHouse) in this phase.
- Coraza adapters already exist as noop placeholders from earlier planning; treating SamWaf as an alternative/additional runtime candidate requires ChatGPT architecture review.
- Decision owner: ChatGPT (PASS / REWORK). Implementation owner after PASS: `waf`.

## Core Architecture Principles

1. Do not modify Coraza core source.
2. Do not modify OWASP CRS official rule files.
3. Treat the WAF detection engine as replaceable (Coraza today as placeholder; SamWaf Runtime under evaluation only).
4. Keep Policy (business model) decoupled from SecLang / engine-specific artifacts.
5. Keep control plane and data plane strictly separated.
6. WAF nodes must not require the management plane to be online for traffic handling.
7. Control-plane / management failures must never impact live business traffic.
8. All deployable configurations must be versioned.
9. All deployable configurations must support rollback.
10. All significant management operations must be audit-logged.
11. Agents must not connect directly to PostgreSQL (or other control-plane databases).
12. Production secrets (JWT, DB passwords, admin bootstrap passwords, keys/PEMs) must come from environment or secret stores — never commit `.env`, keys, or PEMs.
13. GitHub is the shared source of truth; `waf` owns code/dev; ChatGPT owns architecture review (PASS / REWORK).
14. Do not start SamWaf / Coraza / CRS / Agent / ClickHouse / Kafka / Dameng integration until ChatGPT returns PASS with an explicit next-task list.
