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

**Phase 1 — Control-plane security hardening (REWORK IN PROGRESS)**

In scope: production config fail-fast, identity-only JWT + live RBAC, user/role transactions, audit error logging, trusted proxies, password min length, body size limit, swagger production gate, GitHub CI, security-boundary tests.

Not in scope: SamWaf / Coraza / CRS / Agent / Node data-plane / ClickHouse / Kafka / Dameng / frontend; WAF feature work; large refactors; sites.domain unique / node_group / Policy/Rule field models / upstream multi-node / config_version changes.

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
| `config` | Env-based config (`APP_ENV`, `HTTP_ADDR`, `DATABASE_URL`, `JWT_*`, `ADMIN_*`, `TRUSTED_PROXIES`, `SWAGGER_ENABLED`, `MAX_BODY_BYTES`, `LOG_LEVEL`) + production Validate |
| `db` | pgx pool + golang-migrate runner |
| `httpx` | Unified envelope, errors, Recover / RequestID / AccessLog / TrustedProxies / MaxBodyBytes / ClientIP |
| `log` | slog helpers |
| `auth` | Login, identity-only JWT HS256, live RBAC from DB, transactional user+roles, bcrypt, SeedAdmin |
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
- Tests: config validation, JWT identity, live RBAC middleware, user tx (fake repo), client IP, body limit, swagger gate, audit error path, hasher, site validate

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


## Control-Plane Security Principles (Phase 1 Hardening)

1. **Production fail-fast**: `APP_ENV=production` refuses empty/dev-default/too-short `JWT_SECRET` and empty/dev-default/too-short `ADMIN_PASSWORD`. Logs name the failed check; never print secret/password values.
2. **JWT is identity only**: claims are `user_id`, `username`, `iat`, `exp`. Embedded roles/permissions are never the source of truth.
3. **Live RBAC**: every protected request verifies JWT → loads user from DB → requires `enabled` → loads current roles/permissions → `RequirePermission`. Role revocation takes effect immediately for old tokens. `PermissionLoader` interface allows a future cache (no Redis yet).
4. **User+roles transactions**: CreateUser / UpdateUser multi-step writes are owned by the repository/tx layer (not handlers). Invalid role → ROLLBACK (no leftover user); update failure → original data unchanged.
5. **Audit never silent**: audit Write failures do not fail the business HTTP response, but must ERROR-log `request_id`, actor, method, path, resource, resource_id, error — never password/JWT/Authorization/secret/full sensitive body.
6. **Future audit delivery**: transactional outbox / async delivery (e.g. Kafka) is planned; **not implemented in this phase**.
7. **Trusted proxies**: only when `RemoteAddr` is in `TRUSTED_PROXIES` CIDRs are `X-Forwarded-For` / `X-Real-IP` parsed (correct client IP from chain); otherwise ignore headers.
8. **Password minimum length**: new user passwords ≥ 12 characters; production admin bootstrap ≥ 12 and not the documented default.
9. **Body size limit**: unified ~1 MiB limit on JSON mutating APIs; oversized → HTTP 413 envelope.
10. **Swagger production off**: development default on; production default off unless `SWAGGER_ENABLED=true`; production `/swagger/*` returns 404.
11. **CI gates**: GitHub Actions on PRs to main run gofmt / vet / test / race / build for `waf-control`.
12. **SamWaf Runtime remains evaluation-only** — no integration until ChatGPT PASS with an explicit task list.

