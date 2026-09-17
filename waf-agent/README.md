# waf-agent（数据面 Agent）

Phase 2+ 组件。负责：

- 接收控制面下发的站点 / 策略 / 规则配置
- 运行 Coraza + OWASP CRS 检测
- 心跳上报、配置版本协商
- 将安全事件异步写入 ClickHouse（或经控制面转发）

Phase 1 控制面通过 `pkg/adapters/agent.AgentClient` 接口对接；当前为 noop。
