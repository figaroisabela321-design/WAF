# 架构说明

## 控制面 vs 数据面

| | 控制面 (waf-control) | 数据面 (waf-agent) |
|--|---------------------|-------------------|
| 职责 | 配置、RBAC、审计、编排 | 流量检测、拦截、事件上报 |
| 存储 | PostgreSQL | 本地配置缓存 + ClickHouse 事件 |
| Phase 1 | ✅ 已实现 | ❌ 接口预留 |

## 分层

```
Handler (DTO) → Service (业务/校验) → Repository (接口) → PG 实现
```

外部能力通过适配器接口注入：

- `CorazaEngine` / `AgentClient` / `EventStore`

## 中间件链

1. Recover  
2. Request-ID  
3. Access Log（slog JSON + request_id）  
4. Audit（POST/PUT/DELETE `/api/v1/*`）  
5. JWT（除 `/health` `/ready` `/swagger` `/api/v1/auth/login`）  
6. RBAC permission codes  

## 迁移

启动时用 `golang-migrate` + `embed.FS` 自动执行 `migrations/*.sql`。管理员密码在 Go 启动时 bcrypt 种子（不写在 SQL 里）。
