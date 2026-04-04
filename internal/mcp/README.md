# internal/mcp

Ech0 内建的 MCP（Model Context Protocol）Server 实现。

通过 `/mcp` 端点（复用主服务 6277 端口）对外暴露 **Streamable HTTP**，任意支持该传输方式的 MCP Host 均可统一调用 Tools、读取 Resources；鉴权与 Scope 与 REST API 共用同一套 JWT。

## 架构

```
MCP-compatible client / Host
    │
    ▼
┌──────────────────────────────────────────────┐
│  Gin Router  /mcp  (POST + GET)              │
│  ├─ middleware.RateLimit                     │
│  ├─ middleware.OriginGuard                   │
│  ├─ middleware.JWTAuthMiddleware             │
│  └─ Handler.ServeEndpoint()                  │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Server                                      │
│  ├─ JSON-RPC 2.0 解析与分发                  │
│  ├─ initialize / tools/* / resources/*       │
│  ├─ 内置 scope 校验（per tool/resource）      │
│  ├─ tool 执行超时（10s context deadline）     │
│  └─ 结构化审计日志（zap）                     │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Registry                                    │
│  ├─ Tool 注册表（name → handler + scopes）    │
│  └─ Resource 注册表（uri → handler + scopes） │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Adapter（按业务域拆分）                      │
│  ├─ adapter_echo.go    → EchoService         │
│  ├─ adapter_user.go    → UserService         │
│  ├─ adapter_comment.go → CommentService      │
│  ├─ adapter_file.go    → FileService         │
│  └─ adapter_common.go  → CommonService       │
│  （不直连 Repository，强制走 Service 层）      │
└──────────────────────────────────────────────┘
```

## 文件职责

| 文件 | 职责 |
|------|------|
| `jsonrpc.go` | JSON-RPC 2.0 基础类型：Request、Response、RPCError、错误码常量 |
| `capability.go` | MCP 协议版本、ServerCapabilities、InitializeResult、ServerInfo |
| `tools.go` | Tool 相关类型：ToolDefinition、ToolCallParams、ToolCallResult、ContentItem |
| `resources.go` | Resource 相关类型：ResourceDefinition、ResourceReadParams、ResourceReadResult |
| `registry.go` | Tool/Resource 注册表，支持精确匹配与 URI 前缀匹配 |
| `adapter.go` | Adapter 结构体、构造函数、RegisterAll 入口、通用参数/结果 helper |
| `adapter_echo.go` | Echo 域：帖子 CRUD + 点赞/今日/标签 tools，posts/tags resources |
| `adapter_user.go` | User 域：profile/me resource |
| `adapter_comment.go` | Comment 域：list_comments tool，recent comments resource |
| `adapter_file.go` | File 域：list/get/delete file tools |
| `adapter_common.go` | Common 域：heatmap resource |
| `server.go` | MCP Server 核心：请求解析、方法分发、scope 校验、超时控制、审计日志 |
| `handler.go` | Gin 桥接层：组装 Registry → Adapter → Server，暴露 `ServeEndpoint()` |
| `server_test.go` | 单元测试：协议握手、tool 调用、scope 拒绝、resource 读取、错误处理 |

## 请求处理流程

1. HTTP 请求进入 `/mcp`，经过限流、Origin 校验、JWT 鉴权
2. `Handler.ServeEndpoint()` 将 `gin.Context` 转交 `Server.ServeHTTP()`
3. `Server` 解析 JSON-RPC，按 `method` 分发到对应处理函数
4. `tools/call` 和 `resources/read` 会查 `Registry` 获取 handler 与所需 scopes
5. 从 `viewer.Context` 提取当前 token 的 scopes，做细粒度权限校验
6. 调用 `Adapter` 中注册的业务函数，Adapter 转发到 Ech0 Service 层
7. 结果封装为 MCP 标准格式返回

## 扩展新 Tool / Resource

1. 新建 `adapter_<domain>.go`（如 `adapter_file.go`）
2. 实现注册函数（如 `registerFileTools(reg)`）和业务 handler
3. 在 `adapter.go` 的 `RegisterAll()` 中添加一行调用
4. 声明 `InputSchema`（JSON Schema）和所需 `scopes`

不需要修改 `server.go`、`registry.go` 或路由代码。

## 相关文档

- [MCP 接入指南](../../docs/mcp-usage.md) — Token 创建、Host 配置、curl 示例
