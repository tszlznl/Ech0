# Ech0 MCP 设计文档（详细版）

日期：2026-04-04  
状态：v1 已实现（Streamable HTTP on /mcp, 6 tools + 3 resources）  
适用范围：Ech0 后端 + API 集成层（不含前端 UI 大改）

## 1. 背景与目标

Ech0 已具备开放 API、Webhook、访问令牌等能力，但目前对 AI 工具生态的接入路径仍以“定制集成”为主。  
MCP（Model Context Protocol）提供了标准化协议层，可让 Ech0 作为统一能力提供方被 Cursor、Claude Desktop、IDE Agent 等宿主直接接入。

本设计目标：

- 为 Ech0 增加一个可部署、可审计、可扩展的 MCP Server 实现。
- 首版以低风险高价值场景为主：先读后写、默认最小权限。
- 对齐 Ech0 的自托管定位：本地可跑、私有可控、运维简单。
- 复用现有业务能力（API、Token、模型、服务层）避免重复造轮子。

非目标：

- 不在本阶段重构 Ech0 全量 API 规范。
- 不在本阶段引入复杂的多租户隔离体系。
- 不在本阶段实现全自动 AI 代理编排平台（仅提供协议能力）。

## 2. 设计原则

- **最小权限**：每个 MCP tool 必须声明并校验 scope。
- **显式授权**：所有写操作需要明确授权，不做隐式升级。
- **协议优先**：MCP 作为集成协议层，不绕开 Ech0 现有业务服务层。
- **可观测**：每次 `tools/call` 记录结构化审计日志。
- **向后兼容**：不破坏现有 Web/API 使用方式，MCP 为增量能力。
- **先简单可用**：首版聚焦核心场景，后续再扩展 prompts、事件流等高级能力。

## 3. 术语与角色

- **MCP Host**：AI 应用（如 Cursor、Claude Desktop）。
- **MCP Client**：Host 内部对接某个 MCP Server 的客户端。
- **MCP Server**：Ech0 暴露的 MCP 服务端实现。
- **Resource**：可读上下文数据（帖子、标签、统计、配置摘要等）。
- **Tool**：可执行动作（查询、发布、更新草稿、上传文件等）。
- **Prompt**：可参数化提示模板（首版可选）。

## 4. 总体架构

### 4.1 逻辑分层

1. **MCP Transport 层**
   - 首版支持 `stdio`（本地）与 `streamable-http`（远程）。
   - 负责 JSON-RPC 收发、会话初始化、错误映射。

2. **MCP Protocol 层**
   - 实现 `tools/list`、`tools/call`、`resources/list`、`resources/read`。
   - 管理工具注册表、输入 schema 校验、结果标准化封装。

3. **Ech0 Adapter 层**
   - 将 MCP 请求映射到 Ech0 现有 service 接口。
   - 处理 viewer/token 注入、业务错误转换、审计上下文补齐。

4. **Domain Service 层（复用现有）**
   - 使用 Ech0 当前业务服务与仓储逻辑，不在 MCP 层重复实现业务规则。

### 4.2 部署形态

- **本地模式（推荐开发）**：Ech0 MCP 子进程以 `stdio` 启动。
- **远程模式（推荐集成）**：Ech0 主进程挂载 `/mcp` endpoint，对外暴露 streamable-http。

建议策略：

- v1 默认启用本地模式。
- 远程模式需显式配置开启并绑定专用 token 策略。

## 5. 能力范围设计

## 5.1 v1（必须交付）

### A. Resources（只读）

- `ech0://posts/recent?limit=xx`
- `ech0://posts/{id}`
- `ech0://tags`
- `ech0://stats/overview`
- `ech0://profile/me`

说明：

- 仅返回当前调用身份有权访问的数据。
- 资源列表分页，避免单次返回过大内容。

### B. Tools（读写混合，先少量）

只读类：

- `search_posts(query, tags?, from?, to?, limit?)`
- `get_post(id)`
- `list_tags(limit?)`

写入类（需高权限 scope）：

- `create_post(content, visibility?, tags?, draft?)`
- `update_post(id, content?, tags?, visibility?)`
- `delete_post(id)`

约束：

- 所有写操作增加幂等键 `request_id`（可选但建议）。
- 写工具默认返回结构化结果 + 可读文本摘要。

## 5.2 v1.1（建议）

- `upload_file(name, mime_type, content_base64)`（小文件）
- `create_draft(...)`、`publish_draft(id)`
- `append_post_comment(post_id, content)`

## 5.3 v2（可选）

- Prompts 能力（写作模板、摘要模板、发布前检查模板）
- 与 Webhook 联动的增量事件消费（非 MCP 强制项）
- 跨实例能力（Connect 场景）统一查询网关

## 6. 认证与授权模型

## 6.1 Token 策略

推荐新建 MCP 专用 token 类型：

- `typ = mcp_access`
- `scope`：细粒度权限（如 `mcp:post:read`、`mcp:post:write`）
- `aud`：`mcp-local` 或 `mcp-remote`
- `exp`：默认短时效（建议 24h 或更短）
- `jti`：唯一 ID，用于审计与吊销

不建议直接复用站内 session token 给第三方 MCP 客户端。

## 6.2 Scope 建议词表（首版）

- `mcp:post:read`
- `mcp:post:write`
- `mcp:post:delete`
- `mcp:tag:read`
- `mcp:file:write`
- `mcp:profile:read`
- `mcp:admin:settings`（保留，默认不开放）

## 6.3 授权规则

- `tools/call` 前置做 scope 判定，拒绝后返回 MCP 错误 + HTTP 403（远程模式）。
- `resources/read` 同样做 scope 判定，防止“资源绕过工具权限”。
- 高危操作（删除、设置修改）需要额外安全检查（可选二次确认标记）。

## 7. 协议映射与错误语义

## 7.1 协议方法映射

- `initialize`：返回 server capabilities（tools/resources/prompts）
- `tools/list`：返回可见工具清单（按权限过滤）
- `tools/call`：执行工具并返回 `content[]` + `isError`
- `resources/list`：返回可见资源清单（可分页）
- `resources/read`：返回资源内容（支持 text/json）

## 7.2 错误分类

- **认证失败**：token 缺失/过期/签名错误
- **授权失败**：scope 不足、aud 不匹配
- **参数错误**：schema 校验失败
- **业务错误**：目标不存在、状态冲突、内容违规
- **系统错误**：数据库异常、上游依赖异常

返回建议：

- MCP 返回对 Agent 友好的结构化错误（`code`、`message`、`details`）。
- 日志中保留内部错误栈；响应仅返回安全可披露信息。

## 8. 数据与模型设计

新增或扩展实体建议：

1. `mcp_server_settings`
   - `enabled`
   - `transport_mode`（stdio/http）
   - `endpoint_path`
   - `allowed_origins`
   - `rate_limit_*`

2. `mcp_tool_registry`（可选，若支持动态开关）
   - `tool_name`
   - `enabled`
   - `required_scopes`
   - `risk_level`

3. `mcp_audit_logs`
   - `request_id`
   - `jti`
   - `user_id`
   - `tool_or_resource`
   - `decision`（allow/deny）
   - `latency_ms`
   - `error_code`
   - `created_at`

## 9. 配置设计

建议环境变量：

- `ECH0_MCP_ENABLE=true|false`
- `ECH0_MCP_TRANSPORT=stdio|http`
- `ECH0_MCP_HTTP_PATH=/mcp`
- `ECH0_MCP_ALLOWED_ORIGINS=https://xxx,https://yyy`
- `ECH0_MCP_RATE_LIMIT_RPS=20`
- `ECH0_MCP_MAX_INPUT_BYTES=262144`
- `ECH0_MCP_ENABLE_PROMPTS=false`

管理台设置（后续）：

- MCP 开关
- 工具白名单开关
- 默认 token 时效
- 审计日志查看入口

## 10. 安全与风控

## 10.1 基础安全

- 远程模式强制 TLS（反向代理层或内建）。
- 校验 Origin（防 DNS Rebinding）。
- 严格输入 schema，限制字符串长度和数组大小。
- 对 `upload_file` 做 mime/type/size 白名单。

## 10.2 执行安全

- 工具执行超时（例如 5s/10s 分级）。
- 单用户并发上限与整体限流。
- 对高风险工具（删除/批量更新）启用更严格审计。

## 10.3 数据安全

- 审计日志脱敏（不落完整敏感 payload）。
- 禁止在错误响应中泄露内部路径、SQL、密钥信息。
- 支持 token 快速吊销（基于 `jti` 黑名单或版本号）。

## 11. 可观测性与运维

指标建议：

- `mcp_requests_total{method,tool,result}`
- `mcp_request_latency_ms{method,tool}`
- `mcp_auth_fail_total{reason}`
- `mcp_rate_limited_total`
- `mcp_tool_timeout_total{tool}`

日志建议：

- 每次请求记录 `trace_id/request_id`，串联到 Ech0 现有日志体系。
- 对拒绝决策明确 `reason`（如 `scope_missing`、`origin_invalid`）。

运维建议：

- 默认关闭远程模式。
- 提供“最小开箱模板配置”与“生产安全模板配置”。

## 12. 与现有 Ech0 能力的映射

- Ech0 Open API：作为 MCP 工具后端实现基础。
- Access Token：扩展为 MCP 专用 scope 与 audience。
- Webhook：后续用于触发型自动化联动。
- Busen：可作为 MCP 执行事件的内部分发机制（异步场景）。

复用优先级：

1. 复用 service 层业务逻辑。
2. 复用鉴权与 viewer 上下文机制。
3. MCP 层仅做协议转换和能力编排。

## 13. 实施里程碑

### 里程碑 A：协议骨架（1-2 周）

- 建立 MCP server 基础框架与 transport 适配。
- 打通 `initialize`、`tools/list`、`resources/list`。
- 加入基础鉴权中间件与日志链路。

### 里程碑 B：v1 核心能力（1-2 周）

- 完成 v1 必须资源与工具。
- 接入 scope 校验与错误映射。
- 增加审计日志落库。

### 里程碑 C：安全加固与可运维（1 周）

- 限流、超时、Origin 校验、输入大小限制。
- 增加指标上报与告警基线。
- 文档与示例配置发布。

### 里程碑 D：v1.1 扩展（可选）

- 文件上传、草稿发布等中风险工具。
- Prompts（可选）与体验优化。

## 14. 测试与验收

## 14.1 单元测试

- 工具入参 schema 校验。
- scope 计算与拒绝分支。
- 错误映射（业务错误 -> MCP 错误）一致性。

## 14.2 集成测试

- stdio 模式端到端（tools/list + tools/call）。
- http 模式端到端（鉴权 + 限流 + CORS/Origin）。
- 典型业务流程：查询 -> 创建 -> 更新 -> 删除。

## 14.3 安全测试

- 无 token、过期 token、越权 scope。
- 超大输入、非法 mime、恶意 payload。
- 重放请求与速率攻击场景。

## 14.4 验收标准（DoD）

- 至少 6 个核心工具可用，且权限边界正确。
- 资源读取支持分页与权限过滤。
- 关键路径具备审计日志与基础监控指标。
- 文档提供本地接入和远程接入完整示例。
- 默认配置下无高危开放项（远程模式默认关闭）。

## 15. 风险与应对

- **风险：工具过多导致权限复杂度升高**  
  应对：v1 控制工具数，先保证高频场景闭环。

- **风险：MCP 层与 API 层规则不一致**  
  应对：强制走 service 层，禁止 MCP 直连 repository。

- **风险：远程暴露带来攻击面**  
  应对：默认关闭远程；上线前必须完成安全基线检查。

- **风险：集成方误用高权限 token**  
  应对：提供“最小权限模板 token”并在 UI 明确风险提示。

## 16. 开放问题（评审时确认）

- 是否首版即暴露远程 http transport，还是仅本地 stdio？
- 是否在 v1 就做 `upload_file`，还是延后到 v1.1？
- MCP token 与现有 access token 是“新类型并行”还是“同类型扩展”？
- 审计日志存储周期与清理策略如何定义（7/30/90 天）？

## 17. 结论

对 Ech0 来说，支持 MCP 的核心价值是把“开放 API 能力”升级为“AI 生态原生能力”。  
建议采用“先读后写、默认最小权限、协议层薄封装”的策略，以较低风险在 3-5 周内交付一个可用且可运维的 MCP v1。

该方案保持 Ech0 的轻量自托管定位，同时为后续 AI 自动化场景（智能写作、内容运营、跨平台分发）预留清晰扩展路径。
