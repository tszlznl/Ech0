# Ech0 MCP 接入指南

Ech0 内建 [MCP（Model Context Protocol）](https://modelcontextprotocol.io/) Server：在标准协议下把**帖子、标签、评论、文件、互联、资料与统计**等能力以 **Tools** 与 **Resources** 暴露给 AI 工作流。传输层为 **Streamable HTTP**（JSON-RPC 2.0 over HTTP），与主服务同端口，通过 **Bearer Token** 与 **Scope** 做最小权限控制。

> **Audience 要求**：MCP 端点**仅允许** Audience 为 **`mcp-remote`** 的 Access Token 访问。使用其他 audience（如 `public-client`、`cli`）的 Token 将被拒绝（HTTP 403）。

> 架构与实现细节见 [internal/mcp/README.md](../../internal/mcp/README.md)。

## 接入方式

Ech0 的 MCP 端点采用 **Streamable HTTP**（JSON-RPC over HTTP），与 [MCP 规范](https://modelcontextprotocol.io/) 一致。若你的环境支持远程 MCP，并能携带 `Authorization: Bearer <token>`，将 `url` 指向本服务即可。

若运行环境只支持本地 stdio 进程而非远程 HTTP，可通过网关或代理转发到本端点。

## 快速开始

### 1. 创建 MCP 专用 Access Token

在 Ech0 管理后台 **设置 → 访问令牌** 中创建一个新 Token：

- **Audience**：选择 `MCP (AI Agent)`（即 `mcp-remote`，MCP 专用 audience，区别于 `cli`、`integration` 等）
- **Scopes**：根据需要勾选（建议最小权限）
  - 只读场景：`echo:read`、`profile:read`
  - 读写场景：再加上 `echo:write`
  - 评论场景：`comment:read`（查看）、`comment:write`（发表）
  - 文件场景：`file:read`（查看）、`file:write`（删除 / 外部文件入库）
  - 互联场景：`connect:read`（查看连接列表与对端信息）、`connect:write`（添加/删除连接）
  - 管理场景：`admin:settings`（Webhook 管理等管理员操作）
- **有效期**：建议选择 8 小时或 1 个月（不建议永不过期）

创建后妥善保存 Token，它只会显示一次。

### 2. 配置 MCP Host

在你使用的 MCP 客户端配置中（具体文件名与入口因产品而异）添加远程服务，例如：

```json
{
  "mcpServers": {
    "ech0": {
      "url": "https://your-ech0-instance.com/mcp",
      "headers": {
        "Authorization": "Bearer <your-access-token>"
      }
    }
  }
}
```

如果是本地开发环境：

```json
{
  "mcpServers": {
    "ech0": {
      "url": "http://localhost:6277/mcp",
      "headers": {
        "Authorization": "Bearer <your-access-token>"
      }
    }
  }
}
```

## MCP Endpoint

- **地址**：`/mcp`（复用主服务端口，默认 6277）
- **协议**：MCP Streamable HTTP（JSON-RPC 2.0 over HTTP POST）
- **GET /mcp**：返回服务状态信息
- **POST /mcp**：处理 JSON-RPC 请求

## 能力总览

当前 MCP 共暴露 **26 个 Tool** 与 **9 个 Resource**，按业务域整理如下。

### Posts & Tags

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `search_posts` | 按关键词 / 标签 ID 搜索帖子，返回分页结果 `{items, total, page, page_size}` | `echo:read` |
| Tool | `get_post` | 按 UUID 获取单篇帖子（含内容、标签、点赞数、附件、扩展块） | `echo:read` |
| Tool | `get_today_posts` | 获取今日发布的帖子（支持 IANA 时区参数） | `echo:read` |
| Tool | `list_tags` | 列出全部标签（id、名称、使用次数） | `echo:read` |
| Tool | `create_post` | 创建帖子；支持 `content`、`echo_files`、`layout`、`extension`，至少提供其一 | `echo:write` |
| Tool | `update_post` | 更新帖子；`echo_files` / `extension` 提供时为**全量替换** | `echo:write` |
| Tool | `delete_post` | 永久删除帖子 | `echo:write` |
| Tool | `like_post` | 帖子点赞数 +1 | `echo:write` |
| Tool | `delete_tag` | 删除标签并解除与所有帖子的关联 | `echo:write` |
| Resource | `ech0://posts/recent` | 最近 20 条帖子（可附 `?limit=N`） | `echo:read` |
| Resource | `ech0://posts/{id}` | 按 UUID 读取单篇帖子 | `echo:read` |
| Resource | `ech0://tags` | 全部标签及使用次数 | `echo:read` |
| Resource | `ech0://stats/heatmap` | 过去 30 个日历日每日发帖数（热力图，UTC 日界） | `echo:read` |

### Comments

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_comments` | 列出指定帖子下已通过的公开评论（与 `GET /api/comments` 等价） | `comment:read` |
| Tool | `create_comment` | 以集成/AI 身份发表评论（与 `create_integration_comment` 相同，推荐在 Agent 中使用此名称） | `comment:write` |
| Tool | `create_integration_comment` | 同上；与 `POST /api/comments/integration` 共用同一套校验与落库逻辑（无验证码、无 form_token） | `comment:write` |
| Resource | `ech0://comments/recent` | 全站最近 20 条公开评论 | `comment:read` |
| Resource | `ech0://guide/integration-comment` | 集成评论说明：REST 端点、Audience（`mcp-remote` / `integration`）、请求体、curl 示例、与本 MCP 会话 Token 的对应关系 | `comment:read` |

### Files

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_files` | 分页列出已上传文件元数据；返回的 `id` 可作为 `create_post.echo_files[].file_id` 引用 | `file:read` |
| Tool | `get_file` | 获取单个文件元信息（名称、URL、尺寸等）；`id` 可用于 `echo_files` 引用 | `file:read` |
| Tool | `delete_file` | 永久删除文件 | `file:write` |
| Tool | `create_external_file` | 用外部 URL 注册文件记录（无需上传）；返回含 `id` 的文件元信息，可直接用于 `echo_files` | `file:write` |
| Resource | `ech0://guide/file-upload` | 文件上传指南：REST 上传端点、参数、curl 示例、以及如何将上传结果用于 `create_post` | `file:read` |

> **提示**：本地文件上传通过 REST API（`POST /api/files/upload`，multipart/form-data）完成。AI Agent 可读取 `ech0://guide/file-upload` 获取完整操作指南。已有 URL 的外部文件可直接使用 `create_external_file` 注册。

### Connects（实例互联）

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_connects` | 列出本实例已保存的对端连接 | `connect:read` |
| Tool | `get_connects_info` | 聚合获取所有对端的公开信息（有 30 分钟缓存） | `connect:read` |
| Tool | `add_connect` | 添加远程 Ech0 实例连接 | `connect:write` |
| Tool | `delete_connect` | 删除已保存的连接 | `connect:write` |
| Resource | `ech0://connect/self` | 本实例公开信息卡片（名称、URL、logo、帖子统计、版本） | `connect:read` |

### Agent

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `get_recent` | AI 生成的站点近况摘要（有缓存，首次可能需数秒） | `echo:read` |

### Webhooks

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_webhooks` | 列出所有已配置的 Webhook（不含 secret） | `admin:settings` |
| Tool | `create_webhook` | 创建 Webhook 端点 | `admin:settings` |
| Tool | `update_webhook` | 更新 Webhook（按 id，全量替换） | `admin:settings` |
| Tool | `delete_webhook` | 删除 Webhook | `admin:settings` |
| Tool | `test_webhook` | 向 Webhook 端点发送测试请求 | `admin:settings` |

### User

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Resource | `ech0://profile/me` | 当前 Token 对应用户的资料（id、username、email、avatar、admin） | `profile:read` |

## 安全说明

- MCP 使用与 Ech0 API 相同的 JWT 鉴权体系，每个 Tool/Resource 都有独立的 Scope 校验。
- 请求限流：默认 20 RPS / 40 Burst（按 IP）。
- 请求体大小限制：256 KB。
- Tool 执行超时：10 秒。
- 建议在生产环境使用 HTTPS 并配合反向代理。
- Token 遵循最小权限原则：只读场景不要赋予 `echo:write`。

## 协议兼容

- 协议版本：`2025-11-25`
- 支持方法：`initialize`、`tools/list`、`tools/call`、`resources/list`、`resources/read`
- 传输方式：Streamable HTTP（与 MCP 规范一致）

## 示例：使用 curl 测试

```bash
# Initialize
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'

# List tools
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'

# Search posts
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_posts","arguments":{"query":"hello","page":1,"page_size":10}}}'

# Create a post
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"create_post","arguments":{"content":"Hello from MCP!","tags":["mcp","test"]}}}'
```
