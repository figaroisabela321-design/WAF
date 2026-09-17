# AI_HANDOFF

## Meta

| Field | Value |
|-------|-------|
| Date | 2026-09-17 (Asia/Shanghai) |
| Round | Phase 1 Security Hardening |
| Implementing agent | `waf` (Grok Bot) — code/dev |
| Review agent | ChatGPT — architecture review / PASS-REWORK |
| Source of truth | GitHub `https://github.com/figaroisabela321-design/WAF.git` |
| Branch | `fix/phase1-security-hardening` |
| Baseline | `b719ae750560bc73d209518d8a4c57f76e2c2b37` (main) |
| Review Result | REWORK IN PROGRESS — wait for ChatGPT |

## What Was Done

1. Production config fail-fast (`APP_ENV`, Validate for JWT_SECRET / ADMIN_PASSWORD).
2. JWT identity-only (`user_id`, `username`, `iat`, `exp`) + live RBAC via `PermissionLoader` / `JWTAuth`.
3. CreateUser / UpdateUser transactional writes in repository layer (`CreateWithRoles` / `UpdateWithRoles`).
4. Audit Write failures ERROR-logged (never silent); HTTP business response unchanged.
5. `TRUSTED_PROXIES` CIDR-aware client IP (ignore spoofed XFF when untrusted).
6. Password minimum length 12 for CreateUser / UpdateUser; production admin ≥ 12.
7. Unified ~1 MiB JSON body limit → HTTP 413 envelope.
8. Swagger off by default in production (`SWAGGER_ENABLED`); `/swagger/*` → 404.
9. GitHub Actions CI workflow authored at `docs/github-workflows/ci.yml` (copy to `.github/workflows/` requires PAT `workflow` scope).
10. Docs: `PROJECT_STATE` security principles, this handoff, `NEXT_TASK` REWORK IN PROGRESS.
11. Compose: `APP_ENV=development` + development-only warning; README production warning.

No SamWaf / Coraza / CRS / Agent / ClickHouse / Kafka / Dameng / frontend / WAF feature work.

## Quality Gates (real results)

Module root: `waf-control`. Captured 2026-09-17 22:28:07 CST.

| Check | Result |
|-------|--------|
| `gofmt` (`test -z "$(gofmt -l .)"`) | PASS |
| `go vet ./...` | PASS |
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go build ./...` | PASS |

Full log: `docs/reviews/phase1-security-hardening-gates.log`

## Runtime Verification

Docker Compose plugin **not installed** on this box (`docker: 'compose' is not a docker command`). `docker.sock` requires sudo; existing container `waf-pg` (Postgres 16) on `:5432` was reused. Binary fallback:

```text
go build -o /tmp/waf-control-new ./cmd/server
APP_ENV=development HTTP_ADDR=:28080 DATABASE_URL=postgres://waf:waf@127.0.0.1:5432/waf?sslmode=disable ...
```

| Scenario | Result |
|----------|--------|
| GET /health (:28080) | PASS |
| GET /ready (:28080) | PASS |
| POST /api/v1/auth/login | PASS |
| Protected GET /api/v1/sites | PASS |
| Disable user → old token | HTTP 401 `user disabled` PASS |
| Revoke roles → old token | HTTP 403 `permission denied: site:read` PASS |
| Swagger development | HTTP 200 PASS |
| Swagger production (:28081, APP_ENV=production) | HTTP 404 PASS |
| Production fail-fast empty JWT_SECRET | exit 1, named check, no secret printed PASS |
| Production fail-fast ADMIN_PASSWORD=Admin@123 | exit 1, named check, no password printed PASS |

## Security Fixes

- Production refuses empty / documented-dev-default / too-short JWT and admin password.
- JWT no longer embeds roles/permissions as authorization source of truth.
- Live DB RBAC on every protected request (disabled user / revoked role take effect immediately).
- User+roles create/update atomic in repository transactions.
- Audit write failures logged with request_id/actor/method/path/resource (no secrets).
- Trusted-proxy gated XFF / X-Real-IP parsing.
- Password min length 12 for API user create/update.
- 1 MiB body limit on mutating JSON APIs.
- Swagger disabled by default in production.
- CI workflow added for PR gates.

## Config Changes

| Variable | Notes |
|----------|-------|
| `APP_ENV` | `development` (default) / `production` |
| `TRUSTED_PROXIES` | Comma-separated CIDRs |
| `SWAGGER_ENABLED` | Dev default on; prod default off |
| `MAX_BODY_BYTES` | Default 1048576 |
| Compose | `APP_ENV=development`; development-only comment |

## Database Migration

NONE (no schema change).

## Tests Added / Updated

- `internal/config/config_test.go` — production validation matrix + swagger defaults + empty secrets
- `internal/auth/jwt_test.go` — identity-only claims
- `internal/auth/middleware_test.go` — valid/invalid/expired/disabled/role-revoked/missing user
- `internal/auth/service_tx_test.go` — fake repo tx COMMIT/ROLLBACK (see Known Limitations)
- `internal/httpx/client_ip_test.go` — trusted proxy matrix
- `internal/httpx/body_limit_test.go` — 413 envelope
- `internal/httpx/swagger_gate_test.go` — prod 404
- `internal/audit/middleware_test.go` — error log path without failing HTTP

## Known Limitations

1. **User tx tests use fake repository**, not live Postgres. Contract proves service→repo transactional API (`CreateWithRoles` / `UpdateWithRoles`). Prefer adding postgres integration tests when CI has a DB service.
2. **Docker Compose** could not be started here (compose plugin missing). Verified via existing `waf-pg` + local binary on `:28080` / `:28081`.
3. **govulncheck** not wired in CI (optional; enable once toolchain pin is stable).
4. **Audit outbox / async delivery** documented as future work — not implemented.
5. **Permission cache** interface ready (`PermissionLoader`); no Redis/cache yet.
6. Host `:8080` was occupied by another service; verification used `:28080` / `:28081`.
7. **GitHub Actions under `.github/workflows/`** could not be pushed: Classic PAT lacks `workflow` scope. Intended workflow is at `docs/github-workflows/ci.yml` — install manually or with a token that has `workflow` scope.

## Decisions Needed (ChatGPT)

1. PASS or REWORK on Phase 1 security hardening.
2. Confirm SamWaf Runtime remains evaluation-only.
3. Explicit next-task list after PASS (do not invent SamWaf tasks in NEXT_TASK).
4. Whether CI should add a Postgres service for integration tests of user transactions.
5. Whether to enable govulncheck in CI now.

## Files Changed (summary)

- `waf-control/internal/config/*`
- `waf-control/internal/auth/*` (jwt, middleware, service, repo, tests)
- `waf-control/internal/httpx/*` (client IP, body limit, decode, middleware)
- `waf-control/internal/audit/middleware.go` (+ test)
- `waf-control/cmd/server/main.go`
- Handlers (site/node/policy/rule) → `DecodeJSON`
- `docs/github-workflows/ci.yml` (intended Actions workflow; PAT lacked `workflow` scope for `.github/workflows/`)
- `deploy/docker-compose.yml`, `.env.example`, `README.md`
- `docs/PROJECT_STATE.md`, `docs/AI_HANDOFF.md`, `docs/NEXT_TASK.md`, gates log

## Notes for Reviewer

- Do not merge until ChatGPT review.
- No changes to sites.domain unique, node_group, Policy/Rule models, upstream multi-node, config_version/policy_version.
