你是资深安全产品架构师。请基于下面已完成的「政务 WAF Phase 1 控制面」现状，给出 Phase 2 可落地设计（中文、结构化、可直接指导开发）。

## Phase 1 已完成
- 仓库布局：gov-waf / waf-control（Go）+ waf-agent/waf-policy/waf-rules 占位 + deploy Docker
- 栈：Go + chi + PostgreSQL + JWT/RBAC + golang-migrate + OpenAPI + Docker
- 分层：Handler(DTO) → Service → Repository(interface) → PG
- 业务 CRUD：Site / Node / Policy / Rule
- 系统：登录、用户、角色、权限、操作审计
- 健康检查：/health /ready；统一响应信封 {code,message,data}
- 预留接口（noop）：CorazaEngine、AgentClient、EventStore(ClickHouse)、AlertStore
- 明确未做：Coraza 检测、真实 Agent 协议、ClickHouse 落库、SecLang 编译

## 约束
1. 控制面与数据面分离；Agent 不直连 PostgreSQL
2. 不修改 Coraza 核心 / 不改 OWASP CRS 官方文件
3. Policy 与 SecLang 解耦；配置需版本化与可回滚
4. 管理面异常不得影响业务流量
5. 面向政务私有化部署

## 请输出（按编号）
1. Phase 2 目标范围（In / Out）
2. waf-agent 最小职责与目录建议
3. 控制面↔Agent 通信协议（鉴权、心跳、配置拉取、Reload、回滚、错误码）建议用 HTTP 还是 gRPC，给出关键 API 草图
4. Policy → SecLang 编译流水线步骤与产物落盘结构
5. Coraza + Caddy 集成边界（控制面仍不碰流量）
6. 配置版本号设计（config_version）与节点状态机
7. 与现有 Phase 1 表/接口的增量改动清单（尽量少破坏）
8. 建议的 2 周迭代任务拆分（按天/人天粗估）
9. 主要风险与验收标准

请具体、可执行，避免空泛口号。不要生成前端页面。
