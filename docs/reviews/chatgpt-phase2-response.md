# ChatGPT Phase 2 设计建议
来源：chatgpt.com（已登录会话）
日期：2026-09-17

# 政务 WAF Phase 2 可落地设计

## 1. Phase 2 目标范围（In / Out）

### 1.1 Phase 2 核心目标

Phase 2 建议只解决一件核心事情：

> 把 Phase 1 中“数据库里的 Policy / Rule”真正变成数据面节点可以安全运行、验证、切换和回滚的 Coraza 配置。

完整链路：

```text
管理员修改 Policy / Rule
        ↓
Policy Publish
        ↓
生成 Policy Revision
        ↓
Compiler 编译 SecLang
        ↓
生成不可变 Config Artifact
        ↓
生成 config_version
        ↓
Node desired_version 更新
        ↓
Agent 拉取配置
        ↓
SHA256 / 签名校验
        ↓
本地语法检查
        ↓
Stage
        ↓
Caddy + Coraza Validate
        ↓
Reload
        ↓
current_version 更新
        ↓
失败自动 rollback last_good_version
```

### 1.2 Phase 2 In Scope

建议纳入：

1. `waf-agent` 可运行最小版本。
2. Agent ↔ 控制面安全通信。
3. Agent 注册、身份认证、心跳。
4. 节点运行状态上报。
5. `Policy Revision` 不可变快照。
6. Policy → SecLang Compiler。
7. Config Artifact 生成。
8. Config Version 管理。
9. Agent 配置拉取。
10. 配置 SHA256 校验。
11. 配置签名校验。
12. Agent 本地 Validate。
13. Caddy + Coraza Reload。
14. Reload 失败自动回滚。
15. 手工指定版本回滚。
16. Node desired/current/last_good 状态管理。
17. 配置部署历史。
18. Agent/Compiler 操作审计。
19. 控制面宕机情况下，数据面继续使用本地最后成功配置。

### 1.3 Phase 2 Out Scope

暂时不要做：

* ClickHouse 攻击事件落库。
* 实时攻击事件检索。
* Kafka / Pulsar。
* SIEM 对接。
* 告警中心。
* CRS 规则在线编辑器。
* 修改 CRS 官方规则。
* 自动 CRS 调优。
* 灰度比例流量。
* 多集群跨地域配置同步。
* Agent 主动接管业务请求。
* Agent 代理业务流量。
* 控制面直接 Reload Caddy。
* 控制面连接数据面本地文件系统。
* Agent 直连 PostgreSQL。
* Policy Draft 每次修改立即下发。
* SecLang 任意脚本执行能力。

其中 EventStore / AlertStore 继续保留 interface/noop 即可。

---

# 2. waf-agent 最小职责与目录建议

## 2.1 Agent 设计原则

`waf-agent` 不应该变成第二个控制面。

Agent 只负责：

```text
控制面通信
+
配置生命周期
+
本地数据面生命周期
```

不要承担：

* Policy 业务逻辑。
* RBAC。
* 用户管理。
* Rule CRUD。
* PostgreSQL。
* Web UI。
* 攻击策略决策。

## 2.2 Agent 最小职责

建议 Agent Phase 2 只实现以下职责。

### A. Node Identity

读取：

```text
node_id
client certificate
private key
control_plane endpoint
cluster_id
```

### B. Heartbeat

周期性上报：

```text
node_id
agent_version
caddy_version
coraza_version
current_version
last_good_version
apply_state
last_error_code
uptime
```

### C. Desired State Polling

定时询问：

```text
desired_version 是否 > current_version
```

Agent 不接受 PostgreSQL 通知。

也不要求控制面主动连接 Agent。

### D. Artifact Download

下载：

```text
manifest.json
config.tar.gz
signature
```

### E. Artifact Verify

至少验证：

```text
SHA256
config_version
node/site scope
manifest
digital signature
```

### F. Local Stage

新配置先进入：

```text
staging/
```

绝对不能直接覆盖当前运行配置。

### G. Validate

Agent 本地完成两级验证：

```text
SecLang parse
        ↓
Caddy config validate
```

### H. Activate

验证成功后：

```text
staging
   ↓ atomic switch
current
   ↓
caddy reload
```

### I. Rollback

如果 reload 失败：

```text
current → failed
previous/last_good → current
reload
```

### J. Status Report

上报：

```text
DOWNLOADING
VALIDATING
ACTIVATING
ACTIVE
FAILED
ROLLING_BACK
```

---

## 2.3 waf-agent 推荐目录

```text
waf-agent/
├── cmd/
│   └── waf-agent/
│       └── main.go
│
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   └── lifecycle.go
│   │
│   ├── client/
│   │   ├── controlplane.go
│   │   ├── heartbeat.go
│   │   └── config.go
│   │
│   ├── identity/
│   │   ├── identity.go
│   │   └── mtls.go
│   │
│   ├── config/
│   │   ├── manager.go
│   │   ├── downloader.go
│   │   ├── verifier.go
│   │   ├── staging.go
│   │   └── rollback.go
│   │
│   ├── runtime/
│   │   ├── runtime.go
│   │   ├── caddy.go
│   │   └── coraza.go
│   │
│   ├── state/
│   │   ├── state.go
│   │   └── store.go
│   │
│   └── version/
│       └── version.go
│
├── configs/
│   └── agent.example.yaml
│
├── Dockerfile
└── go.mod
```

其中建议：

```go
type Runtime interface {
    Validate(configPath string) error
    Reload(configPath string) error
    Version(ctx context.Context) (*RuntimeVersion, error)
}
```

Phase 2 实现：

```text
CaddyRuntime
```

以后可以支持：

```text
NginxRuntime
EnvoyRuntime
```

而不用修改 Agent 核心生命周期。

---

# 3. 控制面 ↔ Agent 通信协议

## 3.1 HTTP 还是 gRPC

Phase 2 建议：

> **HTTPS + JSON + mTLS，Agent Pull 模式。暂时不要引入 gRPC。**

原因不是 gRPC 不好，而是当前阶段 HTTP 更合适。

当前 Phase 1 已经具备：

```text
chi
OpenAPI
HTTP middleware
统一错误处理
审计体系
```

直接增加 `/agent/v1` 即可。

gRPC 会额外增加：

```text
protobuf
HTTP/2 运维要求
LB/Ingress 配置
调试工具
双协议维护
证书排障复杂度
```

但 Phase 2 并不需要真正的双向实时通信。

配置变更是低频事件：

```text
秒级同步足够
```

因此建议：

```text
Phase 2：HTTP Pull
Phase 3：有大量实时事件/流式数据需求后，再考虑 gRPC Stream
```

---

## 3.2 通信方向

坚持：

```text
Agent → Control Plane
```

控制面不要主动连接 Agent。

即：

```text
Agent
 ├── heartbeat →
 ├── desired config →
 ├── artifact download →
 └── deployment report →
```

这样政务网络环境里：

* 无须向 Agent 开入站管理端口。
* 更适合 NAT。
* 更适合安全域隔离。
* 管理面故障不会影响业务流量。

---

## 3.3 Agent 身份认证

管理用户继续：

```text
JWT + RBAC
```

Agent 不使用管理员 JWT。

建议：

```text
TLS
+
mTLS
```

Agent 每节点一套：

```text
client.crt
client.key
node_id
```

控制面保存：

```text
node_id
certificate_fingerprint
certificate_serial
certificate_expire_at
```

Phase 2 可以先采用“部署时预置证书”。

例如：

```text
/etc/gov-waf-agent/
├── agent.yaml
├── tls/
│   ├── ca.crt
│   ├── client.crt
│   └── client.key
```

后续再实现自动 CSR/证书轮换。

不要把长期共享 API Key 配在所有 Agent 上。

---

## 3.4 Heartbeat

建议：

```http
POST /agent/v1/heartbeat
```

周期：

```text
15 秒
```

控制面：

```text
45 秒无心跳 → STALE
90 秒无心跳 → OFFLINE
```

这两个状态应该由 `last_seen_at` 推导，而不是 Agent 自己声明。

Request：

```json
{
  "node_id": "node-001",
  "agent_version": "2.0.0",
  "caddy_version": "2.x",
  "coraza_version": "3.x",
  "current_version": 103,
  "last_good_version": 103,
  "apply_state": "ACTIVE",
  "last_error_code": "",
  "uptime_seconds": 82731
}
```

Response 可以顺便携带：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "desired_version": 104,
    "desired_generation": 12,
    "poll_after_seconds": 15
  }
}
```

这样大多数情况下甚至不需要额外请求查询 desired state。

---

## 3.5 获取 Desired Config

```http
GET /agent/v1/configs/desired?node_id=node-001&current_version=103
```

无变化：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "changed": false,
    "desired_version": 103
  }
}
```

有变化：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "changed": true,
    "desired_version": 104,
    "artifact_id": "cfg_01...",
    "sha256": "17a2...",
    "signature": "base64...",
    "download_url": "/agent/v1/configs/104/artifact"
  }
}
```

---

## 3.6 Artifact 下载

```http
GET /agent/v1/configs/{config_version}/artifact
```

这里不要使用 JSON envelope 包二进制。

直接：

```http
Content-Type: application/gzip
ETag: "<sha256>"
X-Config-Version: 104
X-Artifact-SHA256: 17a2...
```

Artifact 下载支持：

```text
ETag
Range（可选）
超时
重试
```

---

## 3.7 Apply 状态上报

```http
POST /agent/v1/configs/{config_version}/report
```

例如：

```json
{
  "node_id": "node-001",
  "state": "VALIDATING",
  "message": ""
}
```

成功：

```json
{
  "node_id": "node-001",
  "state": "ACTIVE",
  "current_version": 104,
  "last_good_version": 104
}
```

失败：

```json
{
  "node_id": "node-001",
  "state": "FAILED",
  "error_code": "ENGINE_RELOAD_FAILED",
  "error_message": "caddy reload returned exit code 1"
}
```

---

## 3.8 Reload

不建议：

```text
Control Plane → Agent /reload
```

改为 desired-state 模型。

管理员调用：

```http
POST /api/v1/nodes/{node_id}/reload
```

控制面只做：

```text
desired_generation++
```

例如当前：

```text
version = 104
generation = 12
```

触发 Reload：

```text
version = 104
generation = 13
```

Agent 心跳发现：

```text
desired_generation != applied_generation
```

于是重新 Validate + Reload。

这样控制面永远不需要主动进入 Agent 网络。

---

## 3.9 Rollback

管理员接口：

```http
POST /api/v1/nodes/{node_id}/rollback
```

Request：

```json
{
  "target_version": 103
}
```

控制面只修改：

```text
desired_version = 103
```

Agent 正常执行：

```text
Download/Local Cache
→ Verify
→ Validate
→ Activate
```

不要设计一套特殊 rollback 协议。

从 Agent 的角度：

> rollback 本质就是部署一个旧版本。

这会让状态机简单很多。

---

## 3.10 推荐错误码

HTTP 状态码处理协议层错误。

业务错误使用稳定错误码。

建议至少：

```text
AGENT_UNAUTHORIZED
AGENT_CERT_INVALID
AGENT_NODE_MISMATCH

CONFIG_NOT_FOUND
CONFIG_VERSION_CONFLICT
CONFIG_SCOPE_MISMATCH
CONFIG_DOWNLOAD_FAILED
CONFIG_HASH_MISMATCH
CONFIG_SIGNATURE_INVALID

CONFIG_EXTRACT_FAILED
CONFIG_VALIDATE_FAILED

SECLANG_PARSE_FAILED
CADDY_VALIDATE_FAILED
ENGINE_RELOAD_FAILED

ROLLBACK_VERSION_NOT_FOUND
ROLLBACK_FAILED

LOCAL_STORAGE_FULL
LOCAL_PERMISSION_DENIED

INTERNAL_ERROR
```

不要把：

```text
caddy exited with status 1
```

作为 error_code。

它应该进入：

```text
error_message / error_detail
```

---

# 4. Policy → SecLang 编译流水线

这是 Phase 2 最关键的控制面模块。

建议增加：

```text
waf-policy/compiler
```

而不是直接让 Handler 拼字符串。

## 4.1 编译输入必须是 Snapshot

不要直接：

```text
SELECT 当前 policy
+
SELECT 当前 rules
→ compile
```

否则规则修改后历史版本无法复现。

应先产生：

```text
PolicyRevision
```

Revision 内保存完整输入快照。

示意：

```json
{
  "policy": {
    "mode": "block",
    "paranoia_level": 1,
    "inbound_threshold": 5,
    "outbound_threshold": 4
  },
  "rules": [...],
  "exclusions": [...],
  "crs_version": "4.x"
}
```

Revision 创建后：

```text
不可修改
```

---

## 4.2 编译流水线

推荐严格分为以下步骤。

### Step 1：Load Revision

```text
policy_revision_id
```

读取不可变快照。

### Step 2：Schema Validate

检查：

```text
Policy 字段合法性
Mode 枚举
阈值
Rule ID
Rule Phase
Operator
Action
变量
Transform
```

### Step 3：Semantic Validate

例如：

```text
rule_id 是否冲突
引用 Rule 是否存在
自定义 Rule ID 是否落在规定 ID 区间
SecRuleRemoveById 是否指向合法 Rule
Policy 是否引用不存在的 Site
```

推荐给自定义规则划一个独立 ID 段，例如：

```text
1000000+
```

具体范围可由项目统一约定。

### Step 4：Normalize

生成内部 IR：

```go
type CompiledPolicy struct {
    Mode              string
    ParanoiaLevel     int
    InboundThreshold  int
    OutboundThreshold int
    Exclusions        []Exclusion
    Rules             []CompiledRule
}
```

所有列表进行稳定排序。

例如：

```text
RuleID ASC
Variable ASC
Action ASC
```

保证相同输入一定输出相同文件。

### Step 5：Render SecLang

不要修改 CRS 官方文件。

生成自己的覆盖文件。

推荐：

```text
00-coraza-base.conf
10-crs-setup.conf
20-policy.conf
30-exclusions.conf
40-custom-rules.conf
```

例如：

```text
Include /opt/owasp-crs/crs-setup.conf
Include /opt/owasp-crs/rules/*.conf
```

官方 CRS：

```text
只读挂载
```

不把 CRS 文件复制后修改。

### Step 6：静态 Parse

控制面执行 Compiler 级 SecLang parse。

目的：

```text
提前发现明显错误
```

但这不能替代 Agent Validate。

### Step 7：生成 Runtime Fragment

如果 Caddy 需要对应的 Coraza 配置引用，同时输出：

```text
runtime/caddy-waf.conf
```

### Step 8：Manifest

生成：

```json
{
  "schema_version": 1,
  "config_version": 104,
  "policy_revision_id": "...",
  "created_at": "...",
  "compiler_version": "2.0.0",
  "crs_version": "4.x",
  "files": [
    {
      "path": "seclang/20-policy.conf",
      "sha256": "..."
    }
  ]
}
```

### Step 9：整体 Digest

建议：

```text
artifact_sha256 =
SHA256(config.tar.gz)
```

### Step 10：Sign

控制面：

```text
Sign(artifact_sha256)
```

推荐 Ed25519。

Agent 内置或配置：

```text
public key
```

于是 Artifact 即使在内部文件服务器被替换也会校验失败。

### Step 11：Artifact Immutable

一旦生成：

```text
config_version=104
```

以后绝不覆盖 104。

有任何修改：

```text
生成 105
```

---

## 4.3 Artifact 结构

Artifact：

```text
config-104.tar.gz

manifest.json

seclang/
├── 00-coraza-base.conf
├── 10-crs-setup.conf
├── 20-policy.conf
├── 30-exclusions.conf
└── 40-custom-rules.conf

runtime/
└── caddy-waf.conf
```

不要把私钥、数据库信息、JWT 放入 Artifact。

---

## 4.4 控制面落盘

MVP 可以先使用本地磁盘：

```text
/var/lib/gov-waf/artifacts/
└── 0000000104/
    ├── manifest.json
    ├── config.tar.gz
    ├── config.tar.gz.sha256
    └── config.tar.gz.sig
```

数据库只存：

```text
metadata
path
hash
signature
```

而不是把 tar.gz 存 PostgreSQL BYTEA。

未来可以替换为：

```text
MinIO
S3 compatible storage
```

Repository interface 不变。

---

## 4.5 Agent 本地结构

推荐：

```text
/var/lib/gov-waf-agent/
├── configs/
│   ├── 103/
│   └── 104/
│
├── staging/
│   └── 105.tmp/
│
├── current -> configs/104
├── previous -> configs/103
└── state.json
```

目录切换采用：

```text
rename / symlink atomic switch
```

严禁：

```text
逐个覆盖 current/*.conf
```

否则中途进程崩溃会产生半套配置。

---

# 5. Coraza + Caddy 集成边界

推荐边界：

```text
                       CONTROL PLANE
             ┌──────────────────────────┐
             │ Policy / Rule / Compiler │
             │ Config Artifact          │
             │ Desired State            │
             └─────────────┬────────────┘
                           │ HTTPS/mTLS
                           ▼
                         AGENT
             ┌──────────────────────────┐
             │ Download                 │
             │ Verify                   │
             │ Validate                 │
             │ Local Config Lifecycle   │
             │ Reload                   │
             └─────────────┬────────────┘
                           │ localhost/process
                           ▼
                  Caddy + Coraza
                           │
                           ▼
                    Business Traffic
```

关键要求：

> 控制面进程里不要加载真实生产 Coraza 引擎处理请求。

控制面的 `CorazaEngine` 可以用于：

```text
离线语法验证
Compiler test
```

但不能用于：

```text
业务 HTTP 请求检测
```

---

## 5.1 Caddy 配置建议

建议区分：

```text
Base Config
Generated WAF Config
```

Base Config：

```text
/etc/caddy/Caddyfile
```

属于部署资产。

例如固定：

```text
import /var/lib/gov-waf-agent/current/runtime/caddy-waf.conf
```

Phase 2 Compiler 只生成：

```text
caddy-waf.conf
```

不要每次生成完整 `/etc/caddy/Caddyfile`。

这样：

```text
网络监听
TLS
证书
上游代理
日志
```

与：

```text
WAF Policy
```

解耦。

---

## 5.2 Reload 流程

Agent：

```text
1. 下载到 staging
2. SHA256
3. Signature
4. SecLang parse
5. 生成 staging runtime
6. caddy validate
7. current → previous
8. staging → configs/{version}
9. current → configs/{version}
10. caddy reload
11. health verify
12. success → last_good_version 更新
```

如果第 10/11 步失败：

```text
current → previous
caddy reload
```

只有回滚成功后，节点才能重新进入：

```text
ACTIVE
```

否则：

```text
DEGRADED
```

---

# 6. config_version 设计与节点状态机

## 6.1 不要把 Policy Version 当 config_version

建议明确区分三个概念：

```text
Policy
    ↓
Policy Revision
    ↓
Config Version
```

例如：

```text
Policy:
id = p-1001

PolicyRevision:
revision_id = pr-xxx

Config:
config_version = 104
```

Policy Revision 描述：

> 业务配置快照。

Config Version 描述：

> 数据面真正可部署的完整 Artifact。

这两个不能混用。

---

## 6.2 config_version

推荐：

```sql
BIGINT
```

由 PostgreSQL sequence 生成。

例如：

```text
101
102
103
104
```

接口展示可格式化：

```text
cfg-0000000104
```

不要采用：

```text
timestamp
20260916143822
```

也不建议：

```text
v1.2.3
```

因为部署版本不是软件版本。

每个 config_version：

```text
唯一
不可变
全局单调递增
```

这样一个节点永远可以简单判断：

```text
desired != current
```

注意：

> 不能用 `desired > current` 判断是否需要更新。

因为 rollback 时：

```text
desired = 103
current = 104
```

仍然需要执行。

正确判断：

```text
desired_version != current_version
```

---

## 6.3 Node 建议保存三个版本

必须有：

```text
desired_version
current_version
last_good_version
```

例如：

```text
desired_version   = 105
current_version   = 104
last_good_version = 104
```

105 部署失败：

```text
desired_version   = 105
current_version   = 104
last_good_version = 104
state             = FAILED
```

自动 rollback 后：

```text
current_version = 104
last_good_version = 104
```

不要失败后把 current_version 写成 105。

---

## 6.4 再增加 generation

建议加入：

```text
desired_generation
applied_generation
```

解决：

```text
同版本重新 Reload
```

判断条件：

```text
desired_version != current_version
OR
desired_generation != applied_generation
```

---

## 6.5 节点状态不要只有一个 state

建议拆成两个维度。

### Connectivity

由控制面计算：

```text
ONLINE
STALE
OFFLINE
```

### Apply State

Agent 上报：

```text
UNCONFIGURED
IDLE
DOWNLOADING
VERIFYING
VALIDATING
ACTIVATING
ACTIVE
FAILED
ROLLING_BACK
DEGRADED
```

不要出现：

```text
OFFLINE_DOWNLOADING
ONLINE_ACTIVE
```

这种组合枚举。

---

## 6.6 状态机

主流程：

```text
UNCONFIGURED
     │
     ▼
DOWNLOADING
     │
     ▼
VERIFYING
     │
     ▼
VALIDATING
     │
     ▼
ACTIVATING
     │
     ▼
ACTIVE
```

异常：

```text
DOWNLOADING ─────┐
VERIFYING ───────┤
VALIDATING ──────┤
ACTIVATING ──────┤
                 ▼
               FAILED
```

如果已经存在 last_good：

```text
FAILED
   ↓
ROLLING_BACK
   ↓
ACTIVE
```

回滚也失败：

```text
ROLLING_BACK
     ↓
DEGRADED
```

此时原则仍然是：

> 不主动退出 Caddy，不主动断流量。

保持当前仍能运行的旧进程。

---

# 7. Phase 1 表/接口增量改动

原则：

> 原 CRUD 不推翻，新加“Revision + Deployment”层。

这样对 Phase 1 破坏最小。

## 7.1 Policy 表

原表尽量保持。

增加：

```text
published_revision_id nullable
updated_at
```

Draft 的 Policy CRUD 仍然操作现有表。

新增：

```http
POST /api/v1/policies/{id}/publish
```

Publish 时产生 Revision。

---

## 7.2 新增 policy_revisions

建议：

```sql
policy_revisions
----------------
id UUID PK
policy_id UUID
revision_no BIGINT
snapshot JSONB
snapshot_sha256 VARCHAR
created_by UUID
created_at TIMESTAMPTZ
```

原则：

```text
INSERT ONLY
```

不要 UPDATE revision。

---

## 7.3 新增 config_versions

```sql
config_versions
---------------
config_version BIGINT PK
policy_revision_id UUID
status VARCHAR

artifact_path VARCHAR
artifact_sha256 VARCHAR
artifact_signature TEXT

compiler_version VARCHAR
crs_version VARCHAR

error_code VARCHAR
error_message TEXT

created_by UUID
created_at TIMESTAMPTZ
```

status：

```text
COMPILING
READY
FAILED
```

---

## 7.4 新增 node_runtime_state

不要把十几个运行字段硬塞进现有 `nodes`。

新增：

```sql
node_runtime_state
------------------
node_id UUID PK

desired_version BIGINT
current_version BIGINT
last_good_version BIGINT

desired_generation BIGINT DEFAULT 0
applied_generation BIGINT DEFAULT 0

apply_state VARCHAR

agent_version VARCHAR
caddy_version VARCHAR
coraza_version VARCHAR

last_error_code VARCHAR
last_error_message TEXT

last_seen_at TIMESTAMPTZ
updated_at TIMESTAMPTZ
```

---

## 7.5 新增 node_deployments

用于完整部署历史：

```sql
node_deployments
----------------
id UUID PK
node_id UUID
config_version BIGINT

trigger_type VARCHAR
status VARCHAR

started_at TIMESTAMPTZ
finished_at TIMESTAMPTZ

error_code VARCHAR
error_message TEXT
```

trigger_type：

```text
PUBLISH
MANUAL
RELOAD
ROLLBACK
AUTO_ROLLBACK
```

它非常重要。

否则 Node 表只能看到“现在是什么”，无法回答：

```text
昨天 16:21 为什么回滚？
谁触发的？
哪个配置失败？
```

---

## 7.6 Node 增量字段

现有 Node 表仅增加身份类字段即可：

```text
certificate_fingerprint
certificate_serial
certificate_expire_at
```

以及可能已有的：

```text
enabled
```

运行状态尽量放 `node_runtime_state`。

---

## 7.7 Rule 表

Phase 2 尽量不破坏。

但建议补：

```text
rule_type
```

枚举：

```text
CUSTOM
EXCLUSION
CRS_OVERRIDE
```

如果已有规则字段结构不方便，不急着大迁移。

`PolicyRevision.snapshot` 在 Publish 时冻结 Rule 内容即可。

---

## 7.8 新增管理 API

建议：

```text
POST /api/v1/policies/{id}/publish

GET  /api/v1/config-versions
GET  /api/v1/config-versions/{version}

POST /api/v1/nodes/{id}/deploy
POST /api/v1/nodes/{id}/reload
POST /api/v1/nodes/{id}/rollback

GET  /api/v1/nodes/{id}/runtime-status
GET  /api/v1/nodes/{id}/deployments
```

Node Deploy：

```json
{
  "config_version": 104
}
```

Rollback：

```json
{
  "target_version": 103
}
```

---

## 7.9 Agent API 独立

不要塞到管理员 API 下：

```text
/api/v1/...
```

建议独立：

```text
/agent/v1/heartbeat
/agent/v1/configs/desired
/agent/v1/configs/{version}/artifact
/agent/v1/configs/{version}/report
```

中间件也独立：

```text
Admin:
JWT → RBAC → Audit

Agent:
mTLS → NodeIdentity → AgentAudit
```

---

# 8. 建议的两周迭代任务

假设：

```text
2 名 Go 开发
+
1 名测试/DevOps 兼职
```

总投入约：

```text
22～26 人天
```

这比较现实。

## Day 1

架构和数据模型冻结。

工作：

```text
PolicyRevision
ConfigVersion
NodeRuntimeState
NodeDeployment
API contract
Agent state machine
```

投入：

```text
2 人天
```

输出：

```text
migration
OpenAPI 草案
状态机文档
```

---

## Day 2

控制面 Revision / Publish。

实现：

```text
policy publish
rule snapshot
snapshot canonicalization
snapshot SHA256
```

投入：

```text
2 人天
```

---

## Day 3

Compiler 基础框架。

实现：

```text
Schema Validate
Semantic Validate
Internal IR
deterministic sorting
```

投入：

```text
2～2.5 人天
```

---

## Day 4

SecLang Renderer。

实现：

```text
policy config
CRS setup override
exclusion
custom rule
manifest
```

投入：

```text
2.5 人天
```

要求当天必须完成：

```text
同一 Snapshot 编译 100 次 SHA256 完全一致
```

---

## Day 5

Artifact。

实现：

```text
tar.gz
sha256
Ed25519 signature
ArtifactStore interface
LocalArtifactStore
```

投入：

```text
2 人天
```

第一周结束应做到：

```text
Policy
→ Publish
→ Revision
→ Compile
→ cfg-104.tar.gz
```

---

## Day 6

Agent 骨架 + mTLS。

实现：

```text
config
identity
control-plane client
heartbeat
local state
```

投入：

```text
2.5 人天
```

---

## Day 7

Agent Artifact 生命周期。

实现：

```text
desired state
download
sha256
signature
staging
local cache
```

投入：

```text
2.5 人天
```

---

## Day 8

Caddy / Coraza runtime。

实现：

```text
SecLang validation
caddy validate
atomic switch
caddy reload
```

投入：

```text
3 人天
```

这是整个迭代风险最高的一天。

建议直接在和生产一致的容器环境里验证。

---

## Day 9

Rollback + Deployment History。

实现：

```text
reload failure
automatic rollback
manual rollback
node deployment record
error reporting
```

投入：

```text
2.5 人天
```

必须测试：

```text
104 OK
105 invalid
→ 自动恢复 104
```

---

## Day 10

集成测试和故障注入。

测试：

```text
Control Plane down
PostgreSQL down
Agent restart
Caddy restart
Artifact corrupted
signature invalid
network timeout
disk full
invalid SecLang
invalid Caddy config
reload failure
rollback failure
concurrent deploy
```

投入：

```text
3～4 人天
```

最终交付：

```text
Docker Compose 集成环境
+
E2E Test
+
Phase 2 API
+
运行手册
```

---

# 9. 主要风险与验收标准

## 9.1 风险：Compiler 输出不确定

危险：

相同 Policy 两次编译：

```text
SHA256 A
SHA256 B
```

这样版本审计基本失效。

措施：

```text
固定排序
固定换行符
固定编码 UTF-8
禁止写编译时间到 SecLang 文件
manifest 中时间不参与配置内容 digest，或明确规范 digest 计算范围
```

验收：

```text
相同 Revision 连续编译 100 次：
SecLang 内容完全一致
```

---

## 9.2 风险：配置覆盖导致半配置状态

禁止：

```text
cp *.conf /current
```

必须：

```text
stage
validate
rename
reload
```

验收：

部署过程中 kill -9 Agent：

```text
Caddy 继续使用旧版本。
```

---

## 9.3 风险：控制面异常影响业务

必须模拟：

```text
waf-control kill
PostgreSQL kill
网络断开
```

验收：

```text
现有 HTTP 请求不受影响。
Coraza 继续使用 last_good config。
Agent 不关闭 Caddy。
```

这是 Phase 2 的最高优先级验收项。

---

## 9.4 风险：错误配置导致 Caddy 起不来

必须：

```text
Validate before Reload
```

而且：

```text
Reload 失败 ≠ 停止 Caddy
```

验收：

下发错误 SecLang：

```text
节点保持旧配置 ACTIVE
错误状态回传控制面
业务请求继续正常处理
```

---

## 9.5 风险：自动回滚逻辑错误

测试：

```text
v100 good
v101 good
v102 bad
```

应得到：

```text
desired = 102
current = 101
last_good = 101
state = FAILED/ACTIVE after rollback
```

而不是：

```text
current = 102
```

---

## 9.6 风险：配置重复执行

Agent 请求可能重复。

所有操作必须幂等。

例如 Agent 连续收到：

```text
desired=104
generation=5
```

如果已经：

```text
current=104
applied_generation=5
```

则：

```text
NOOP
```

不得重复 reload。

---

## 9.7 风险：控制面和 Agent 并发修改版本

控制面只维护：

```text
desired
```

Agent 只维护并上报：

```text
current
last_good
applied_generation
```

不要让两边同时修改同一个语义字段。

---

## 9.8 风险：CRS 升级破坏历史配置

`config_version` 必须记录：

```text
coraza compatibility
crs_version
compiler_version
```

如果节点运行：

```text
CRS 4.x
```

不要允许下发要求：

```text
CRS 5.x
```

的 Artifact。

Agent Validate 前检查兼容性。

---

## 9.9 风险：Agent 被伪造

验收必须包括：

```text
无证书请求 → 401/403
错误证书 → 401/403
node-A 证书冒充 node-B → 拒绝
过期证书 → 拒绝
```

Agent identity：

```text
mTLS certificate
          ↓
       node_id
```

二者必须绑定。

---

# Phase 2 最终验收闭环

建议把 Phase 2 Definition of Done 最终压缩成以下一条 E2E：

```text
① 创建 Site
② 创建 Policy
③ 创建/关联 Rule
④ Publish Policy
⑤ 生成 immutable PolicyRevision
⑥ Compiler 生成 SecLang
⑦ 生成 config_version = N
⑧ 生成并签名 Artifact
⑨ Node desired_version = N
⑩ Agent heartbeat 获取 desired N
⑪ Agent 下载 Artifact
⑫ SHA256 + signature 校验
⑬ SecLang validate
⑭ Caddy validate
⑮ atomic activate
⑯ Caddy reload
⑰ current_version = N
⑱ last_good_version = N
⑲ 控制面显示节点 ACTIVE
```

然后再做一次失败链路：

```text
① 发布 config N+1
② 人为制造错误 SecLang
③ Agent Validate 失败
④ 不 Reload 错误配置
⑤ current_version 仍为 N
⑥ last_good_version 仍为 N
⑦ 业务流量正常
⑧ 控制面显示 CONFIG_VALIDATE_FAILED
```

再做一次 Reload 失败链路：

```text
N       = good
N + 1   = validate 成功但 reload 失败

Agent：
activate N+1
→ reload failed
→ restore N
→ reload N
→ current=N
→ last_good=N
→ 上报 AUTO_ROLLBACK
```

这三个场景全部通过后，我会认为 Phase 2 的“控制面 → 数据面配置闭环”真正成立。

---

# 推荐最终模块关系

```text
gov-waf/
│
├── waf-control/
│   ├── policy
│   ├── rule
│   ├── node
│   ├── deployment
│   └── agent-api
│
├── waf-policy/
│   ├── model
│   ├── revision
│   ├── validator
│   ├── compiler
│   └── renderer
│
├── waf-rules/
│   ├── custom
│   └── metadata
│
├── waf-agent/
│   ├── client
│   ├── identity
│   ├── config
│   ├── runtime
│   └── state
│
└── deploy/
    ├── control-plane
    ├── agent
    ├── caddy
    └── crs
```

其中最重要的依赖方向应保持为：

```text
Policy Domain
    ↓
Compiler
    ↓
Immutable Artifact
    ↓
Agent
    ↓
Caddy / Coraza
    ↓
Traffic
```

不要出现反向依赖：

```text
Coraza → PostgreSQL
Caddy → Control Plane
Agent → PostgreSQL
Control Plane → Business Traffic
```

这会成为后续 Phase 3 做 ClickHouse、攻击事件、实时监控和告警时仍然稳定的底层边界。
