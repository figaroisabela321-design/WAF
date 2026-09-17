# SamWaf Integration Analysis (Phase 2A)

Date: 2026-09-17 23:26 Asia/Shanghai (UTC+8)  
Branch: `feat/samwaf-node-poc`  
POC decision: **`SAMWAF_RUNTIME_POC = BLOCKED`**

## 1. Upstream pin (frozen for entire POC)

| Field | Value |
|-------|-------|
| Clone path | `/workspace/upstream/SamWaf` (outside product repo) |
| Remote | `https://github.com/samwafgo/SamWaf.git` |
| HEAD SHA | `d975b12ec0a4757ca0e9698accd373dfee5f7c71` |
| Describe | `v1.3.25-beta.6-2-gd975b12` |
| Nearest tag | `v1.3.25-beta.6` |
| LICENSE | Apache License 2.0 (`LICENSE`) |
| Go module | `SamWaf` |
| Go version | `go 1.25` / toolchain `go1.25.9` (box has `go1.24.4` — separate issue if ever building upstream) |
| Coraza | `github.com/corazawaf/coraza/v3 v3.3.3` |
| CRS (bundled) | OWASP CRS **4.9.0-dev** (`cmd/samwaf/exedata/owasp/coreruleset/`, dataset version file `1.0.20241128`) |
| Mid-POC pull | **Forbidden** — pin above for entire Phase 2A |

## 2. Packages reviewed

`cmd/samwaf`, `wafenginecore`, `wafowasp`, `wafproxy`, `wafconfig`, `wafinit`, `wafdb`, `wafssl`, `model`, `service/waf_service`, `global`, `globalobj`, `router`, `api`, plus `wafmangeweb`, `wafupdate`, `wafacme`.

## 3. Answers to the 20 feasibility questions

### Q1 — How does the data-plane listener start?

`WafEngine.StartWaf()` → `LoadAllHost()` → `StartAllProxyServer()` / `StartProxyServer()`. Listeners are `http.Server` (and HTTP/3 holders) bound per host port from the host model (`Port` / `PortListensJSON` / `BindMorePort`). There is **no** standalone “bind this Host/upstream from JSON config” API independent of DB-loaded hosts.

### Q2 — How does reverse proxy work?

`wafproxy` (`reverseproxy.go`, `httputil.go`) is used from `wafenginecore` after WAF checks in `ServeHTTP`. Upstream comes from `Hosts.Remote_*` / load-balance tables also loaded via DB-backed host state.

### Q3 — Host / site model?

`model.Hosts` is a large GORM entity (domain, port, SSL, remote IP/port, defense JSON, CC, captcha, cache, path rules, etc.). Authority is **local SQLite (or MySQL/Postgres dialect) via `global.GWAF_LOCAL_DB`**, not an external control-plane NodeConfig.

### Q4 — Config reload without full process restart?

Yes, **inside SamWaf’s own channel/DB model**: `ReloadAllHostZeroGap()` re-reads all hosts from `GWAF_LOCAL_DB`, atomically replaces routing, diffs ports. Driven by `GWAF_CHAN_ENGINE` / host channels from management APIs — **not** by an external Validate/Apply Runtime interface.

### Q5 — Engine lifecycle?

Cold start path in `cmd/samwaf/main.go` `run()`: config → DB init → create `WafEngine` into `globalobj` → `StartWaf()` → optional tunnel/app engines → **`webmanager.StartLocalServer()`** → task scheduler → message loops. Stop: `Graceful()` → `CloseWaf()` / tunnel / apps / hostguard / scheduler / close DBs. Lifecycle is a **monolithic process service**, not an embeddable library object.

### Q6 — Coraza / CRS integration?

Real: `wafowasp` builds `coraza.NewWAF` from `data/owasp/coraza.conf` + CRS + SamWaf before/after overlays. Modes: `On` / `DetectionOnly` / `Off` (`GCONFIG_OWASP_MODE`, defaults DetectionOnly). This is usable **only after** full SamWaf data-dir / OWASP manager init — not as a free-standing adapter behind gov-waf NodeConfig.

### Q7 — Observe vs protect?

Mapped to Coraza `SecRuleEngine DetectionOnly` vs `On`. Host also has `LogOnlyMode` / `GUARD_STATUS`. Switching is via SamWaf system config / OWASP overrides stored and applied through SamWaf’s config+DB path.

### Q8 — Attack / access events?

Web logs and sys logs enqueue to `GQEQUE_LOG_DB` and persist in `GWAF_LOCAL_LOG_DB`. Notify plugins (webhook, kafka, etc.) exist under `wafnotify`. There is **no** gov-waf Event V1 → `POST /internal/v1/node-events` shape; extracting events without DB/queue globals requires new code.

### Q9 — SQLite coupling?

**Hard.** `InitCoreDb` / `InitLogDb` / `InitStatsDb` set process globals. `LoadAllHost`, `ensureGlobalHost`, `ReloadAllHostZeroGap`, and essentially all `service/waf_service` CRUD use `global.GWAF_LOCAL_DB`. Engine start **cannot** proceed without DB.

### Q10 — Web UI / management coupling?

**Hard at process level.** Default port `GWAF_LOCAL_SERVER_PORT=26666`; `wafmangeweb.StartLocalServer()` started from main. Gin routers under `router/` + `api/` (~89 API files) expose full product UI/API (login, RBAC, host CRUD). Not optional for stock binary; skipping it means forking main and still leaving globals/DB.

### Q11 — Global singleton?

**Yes, pervasive.** `global` package: DB handles, channels, OWASP pointers, runtime flags, QPS, DNS, SSL flags, etc. `globalobj`: single `*WafEngine`, tunnel engine, task registry/scheduler, plugin manager, app engine. ~473 `global.` references across ~60 files in `wafenginecore` alone.

### Q12 — Multi-instance in one process?

**No.** Process-wide singletons and fixed management port make multiple independent Runtime instances in one process infeasible without rewriting globals to instance structs.

### Q13 — Clean Close?

`CloseWaf()` stops proxy servers, clears routing/certs/transports, enqueues sys log (needs log queue/DB). Full process graceful stop also closes DBs, tasks, tunnel, apps. Partial embed Close without those globals is undefined.

### Q14 — Does Apply require process restart?

Within SamWaf: host/engine reload can be zero-gap **if** config is already in SamWaf DB and channels fire. For gov-waf push of NodeConfig JSON: **no supported path** — would require writing SQLite rows then signaling channels, or restarting the whole SamWaf process (forbidden as sole Apply strategy for this POC).

### Q15 — Outbound network dependencies?

Present: `wafupdate` (update.samwaf.com style fetches), ACME (`wafacme` / lego), DNS reverse lookup (`GWAF_RUNTIME_DNS_SERVER` default `119.29.29.29`), IP geolocation DBs, optional cloud remote DB, notify webhooks. Must be disabled or risk-accepted for air-gapped gov deploy; not cleanly factored as a single “offline mode” switch for embed.

### Q16 — Management plane vs data plane separation inside SamWaf?

**Fused.** Same process owns management HTTP, SQLite authority, WAF listeners, tasks, updates. Control-plane offline resilience exists only as “local SQLite LKG”, which **conflicts** with gov-waf rule: PostgreSQL is CP authority; node must not treat SamWaf SQLite as config authority.

### Q17 — Can control plane avoid importing SamWaf types?

Only if a separate `waf-node` binary wraps SamWaf. That wrapper still cannot honestly implement Runtime Validate/Apply/LKG without either (a) driving SamWaf DB+channels or (b) large core rewrite. So the boundary is achievable at import level but **not** at behavioral Runtime level without rewrite.

### Q18 — Minimal reusable surface?

| Tier | Modules |
|------|---------|
| Reusable ideas / reference | Coraza+CRS wiring patterns in `wafowasp`; reverse proxy patterns in `wafproxy`; zero-gap reload ideas |
| Light-mod (still costly) | Disable update/ACME via config; bind management to localhost |
| High-coupling (blockers) | `global`, `globalobj`, `wafdb`, `service/waf_service`, `api`+`router`, `wafmangeweb`, `cmd/samwaf` lifecycle, host load from DB |
| Prefer disable | `wafupdate`, remote telemetry/env report, ACME if unused, tunnel/app engines, AI detector, Vue UI |

### Q19 — Rewrite cost estimate?

To get an embeddable Runtime with external JSON config, no management UI, no SQLite authority, multi-instance-safe lifecycle:

- Replace `global`/`globalobj` with per-instance deps → touches **engine + services + OWASP + queues**
- Replace DB-backed `LoadHost` with NodeConfig apply path
- Detach event pipeline from SQLite queues
- New main for `gov-waf-node` that never starts Gin management

Order-of-magnitude: **large core rewrite / sustained fork** (weeks–months), not a Phase 2A POC.

### Q20 — GO / NO-GO for Phase 2A embeddable Runtime?

**NO-GO → BLOCKED.** SamWaf cannot become an embeddable Runtime meeting gov-waf constraints without rewriting large fused core (SQLite + Global Singleton + Management Web + lifecycle).

## 4. Dependency graph (simplified)

```
cmd/samwaf/main
  ├─ wafconfig.LoadAndInitConfig → global config flags
  ├─ wafdb.InitCoreDb/LogDb/StatsDb → global.GWAF_LOCAL_*_DB (SQLite)
  ├─ wafowasp manager → global.GWAF_OWASP*
  ├─ globalobj.WafEngine.StartWaf
  │    ├─ ensureGlobalHost / LoadAllHost ──► GWAF_LOCAL_DB
  │    └─ StartAllProxyServer ──► wafproxy + ServeHTTP detections
  ├─ wafmangeweb.StartLocalServer ──► gin api/router (UI/login/RBAC)
  ├─ waftask scheduler, wafupdate, wafacme, notify, hostguard
  └─ channel loops (GWAF_CHAN_*) mutating engine from management ops
```

## 5. Lifecycle vs gov-waf Runtime target

| gov-waf Runtime need | SamWaf today |
|----------------------|--------------|
| Validate(NodeConfig) | No — validates via its own models/DB |
| Apply without CP | Only via local SQLite LKG |
| CurrentVersion / Status | Internal engine status ≠ config_version LKG files |
| Close cleanly | Process Graceful; not library Close |
| CP never imports SamWaf types | Possible only with separate binary + still no honest Apply |
| Management UI unused | Stock path starts it |
| Real Coraza/CRS | Yes, inside fused product |
| CP offline traffic OK | Yes for SamWaf-as-product; wrong config authority for gov-waf |

## 6. Network / DB / license notes

- **DB:** SQLite (wxsqlite3) default; MySQL/Postgres dialects exist but still local “SamWaf authority”, not gov-waf CP PostgreSQL.
- **Network:** update, DNS, ACME, optional remote DB, notifiers.
- **License:** Apache-2.0 — attribution OK; keep headers if code is reused later. See `docs/THIRD_PARTY_SAMWAF.md`.

## 7. Blocking modules / files / chain

1. `global/global.go` + `globalobj/globalobj.go` — process singletons  
2. `wafdb/localdb.go` + `service/waf_service/*` — SQLite authority  
3. `wafenginecore/wafworker.go` `LoadAllHost`/`LoadHost` — DB → listeners  
4. `cmd/samwaf/main.go` + `wafmangeweb/localserver.go` — fused management lifecycle  
5. `api/*` + `router/*` — management surface inseparable from product assumptions  

**Chain:** NodeConfig Apply ⇒ must mutate host runtime ⇒ today only via DB+channels+globals ⇒ implies management stack or fork rewrite.

## 8. Risks if forced integration (not done)

- Fake success via stock SamWaf binary + UI (explicitly forbidden)
- SQLite becomes shadow control plane
- Management port exposure
- Outbound update/ACME in gov networks
- Process restart as only Apply
- Go 1.25 toolchain mismatch with gov-waf go 1.24

## 9. Alternatives (document only — do NOT switch in this PR)

| Option | Summary |
|--------|---------|
| **A — SamWaf** | Blocked as embeddable Runtime without large rewrite; could remain a **separate appliance** later, not gov-waf node library |
| **B — Caddy + Coraza** | Aligns with earlier Phase 2 ChatGPT sketch; still requires ChatGPT PASS before implementation |
| **C — Custom proxy + Coraza** | Smallest control of lifecycle; more engineering than Caddy plugin path |

**This PR does not implement B or C.**

## 10. POC coding status

- `waf-node/` module: **not created** (honest BLOCKED)
- No SamWaf vendor/copy into product repo
- No merge to main; awaiting ChatGPT review
