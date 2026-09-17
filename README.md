# 政务 WAF 安全防护平台

基于 **Coraza + OWASP CRS** 的政务 Web 应用防火墙控制面 / 数据面分离架构。

> **Phase 1（本仓库当前范围）**：控制面后端基础能力 —— HTTP API、PostgreSQL、JWT/RBAC、站点/节点/策略/规则 CRUD、审计日志、OpenAPI。  
> **不包含**：Coraza 检测引擎、真实 Agent 通信、ClickHouse 事件存储（均以接口 + noop 适配器预留）。

---

## 产品愿景与架构

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  管理控制台 UI   │────▶│  waf-control     │────▶│  PostgreSQL     │
│  (未来)         │     │  控制面 API       │     │  配置 / RBAC    │
└─────────────────┘     └────────┬─────────┘     └─────────────────┘
                                 │
                    AgentClient  │  (Phase 2)
                                 ▼
                        ┌──────────────────┐
                        │  waf-agent 节点   │
                        │  Coraza + CRS    │
                        └────────┬─────────┘
                                 │ events
                                 ▼
                        ┌──────────────────┐
                        │  ClickHouse      │
                        │  (Phase 2/3)     │
                        └──────────────────┘
```

| 阶段 | 内容 |
|------|------|
| **Phase 1** | 控制面后端：站点/节点/策略/规则、认证授权、审计、迁移、Docker Compose |
| **Phase 2** | 数据面 Agent、Coraza 引擎接入、配置下发与心跳 |
| **Phase 3** | ClickHouse 事件/告警、运营大屏、策略编排增强 |

相关模块占位说明见：`waf-agent/`、`waf-policy/`、`waf-rules/`。

---

## 目录结构

```
gov-waf/
├── README.md
├── go.work
├── .env.example
├── .gitignore
├── waf-control/          # Phase 1 控制面 Go 服务
│   ├── cmd/server/
│   ├── internal/         # config, httpx, db, auth, site, node, policy, rule, audit, event, alert
│   ├── pkg/adapters/     # coraza / agent / clickhouse noop
│   ├── migrations/
│   └── docs/openapi.yaml
├── waf-agent/README.md
├── waf-policy/README.md
├── waf-rules/README.md
├── deploy/
│   ├── Dockerfile
│   └── docker-compose.yml
└── docs/
    ├── architecture.md
    ├── er-diagram.md
    └── api-examples.md
```

ER 图见 [docs/er-diagram.md](docs/er-diagram.md)。更多架构说明见 [docs/architecture.md](docs/architecture.md)。

---

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HTTP_ADDR` | `:8080` | HTTP 监听地址 |
| `DATABASE_URL` | `postgres://waf:waf@localhost:5432/waf?sslmode=disable` | PostgreSQL 连接串 |
| `JWT_SECRET` | `dev-jwt-secret-change-me` | JWT HS256 密钥 |
| `JWT_EXPIRE_HOURS` | `24` | Token 过期小时数 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `ADMIN_USERNAME` | `admin` | 首次启动种子管理员用户名 |
| `ADMIN_PASSWORD` | `Admin@123` | 首次启动种子管理员密码（bcrypt，仅当不存在 admin 角色用户时写入） |

参考 `.env.example`。

---

## 默认管理员

| 字段 | 值 |
|------|-----|
| 用户名 | `admin` |
| 密码 | `Admin@123` |
| 角色 | `admin`（拥有全部权限码） |

权限码：`site:read/write`、`node:read/write`、`policy:read/write`、`rule:read/write`、`user:read/write`、`audit:read`。

---

## 快速启动（推荐）

```bash
cd /workspace/gov-waf
docker compose -f deploy/docker-compose.yml up --build
```

服务启动后：

- API: http://localhost:8080  
- Swagger UI: http://localhost:8080/swagger/  
- OpenAPI YAML: http://localhost:8080/swagger/openapi.yaml  

进程内会自动执行 SQL 迁移并种子管理员。

---

## 本地开发（无 Docker 跑 API）

```bash
# 1. 启动 Postgres（示例）
docker run -d --name waf-pg -e POSTGRES_USER=waf -e POSTGRES_PASSWORD=waf \
  -e POSTGRES_DB=waf -p 5432:5432 postgres:16-alpine

# 2. 运行控制面
cd /workspace/gov-waf/waf-control
export DATABASE_URL='postgres://waf:waf@localhost:5432/waf?sslmode=disable'
export JWT_SECRET='dev-jwt-secret-change-me'
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD='Admin@123'
go run ./cmd/server
```

---

## curl 示例

```bash
# 健康检查
curl -s http://localhost:8080/health

# 登录
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123"}' | jq -r '.data.token')

# 创建站点
curl -s -X POST http://localhost:8080/api/v1/sites \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"政务门户","domain":"portal.gov.example","upstream_host":"10.0.0.10","upstream_port":80,"protocol":"http","protection_mode":"observe"}'

# 站点列表
curl -s 'http://localhost:8080/api/v1/sites?page=1&page_size=20' \
  -H "Authorization: Bearer $TOKEN"
```

更多示例：[docs/api-examples.md](docs/api-examples.md)。

---

## 统一响应信封

成功：

```json
{"code":0,"message":"ok","data":{...}}
```

失败：

```json
{"code":40001,"message":"name is required","data":null}
```

列表：

```json
{"code":0,"message":"ok","data":{"items":[],"total":0,"page":1,"page_size":20}}
```

---

## 测试与构建

```bash
cd /workspace/gov-waf/waf-control
go test ./...
go build -o /tmp/waf-control ./cmd/server
```

---

## 未来扩展（接口预留）

- `pkg/adapters/coraza` — `CorazaEngine`  
- `pkg/adapters/agent` — `AgentClient`  
- `pkg/adapters/clickhouse` / `internal/event` — `EventStore`  
- `internal/alert` — `AlertStore`  

Phase 1 仅提供 noop 实现与空列表 API（`/api/v1/events`、`/api/v1/alerts`）。
