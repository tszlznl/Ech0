# Ech0 MCP 接入指南

Ech0 内建 [MCP（Model Context Protocol）](https://modelcontextprotocol.io/) Server：在标准协议下把**帖子、标签、评论、文件、资料与统计**等能力以 **Tools** 与 **Resources** 暴露给 AI 工作流。传输层为 **Streamable HTTP**（JSON-RPC 2.0 over HTTP），与主服务同端口，通过 **Bearer Token**（Audience `mcp-remote`）与 **Scope** 做最小权限控制。

> 架构与实现细节见 [internal/mcp/README.md](../internal/mcp/README.md)。

## 接入方式

Ech0 的 MCP 端点采用 **Streamable HTTP**（JSON-RPC over HTTP），与 [MCP 规范](https://modelcontextprotocol.io/) 一致。若你的环境支持远程 MCP，并能携带 `Authorization: Bearer <token>`，将 `url` 指向本服务即可。

若运行环境只支持本地 stdio 进程而非远程 HTTP，可通过网关或代理转发到本端点。

## 快速开始

### 1. 创建 MCP 专用 Access Token

在 Ech0 管理后台 **设置 → 访问令牌** 中创建一个新 Token：

- **Audience**：选择 `mcp-remote`（MCP 专用 audience，区别于 `cli`、`integration` 等）
- **Scopes**：根据需要勾选（建议最小权限）
  - 只读场景：`echo:read`、`profile:read`
  - 读写场景：再加上 `echo:write`
  - 评论场景：`comment:read`（查看）、`comment:write`（发表）
  - 文件场景：`file:read`（查看）、`file:write`（删除）
  - 互联场景：`connect:read`（查看连接列表与对端信息）、`connect:write`（添加/删除连接）
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

## 可用 Tools

### Posts & Tags

| Tool | 说明 | 所需 Scope |
|------|------|-----------|
| `search_posts` | 搜索帖子（支持关键词、标签过滤、分页、排序） | `echo:read` |
| `get_post` | 根据 ID 获取帖子详情 | `echo:read` |
| `get_today_posts` | 获取今日帖子（支持时区参数） | `echo:read` |
| `list_tags` | 列出所有标签 | `echo:read` |
| `create_post` | 创建新帖子（返回新帖 ID） | `echo:write` |
| `update_post` | 更新已有帖子（返回帖子 ID） | `echo:write` |
| `delete_post` | 删除帖子（返回帖子 ID） | `echo:write` |
| `like_post` | 点赞帖子 | `echo:write` |
| `delete_tag` | 删除标签 | `echo:write` |

### Comments

| Tool | 说明 | 所需 Scope |
|------|------|-----------|
| `list_comments` | 列出指定帖子的公开评论 | `comment:read` |

### Files

| Tool | 说明 | 所需 Scope |
|------|------|-----------|
| `list_files` | 列出已上传文件（支持分页、搜索、存储类型过滤） | `file:read` |
| `get_file` | 获取文件元信息 | `file:read` |
| `delete_file` | 删除文件 | `file:write` |

### Connects

| Tool | 说明 | 所需 Scope |
|------|------|-----------|
| `list_connects` | 列出本实例已保存的对端连接（id + URL） | `connect:read` |
| `get_connects_info` | 聚合获取所有对端的公开信息（名称、logo、帖子数等，有 30 分钟缓存） | `connect:read` |
| `add_connect` | 添加远程 Ech0 实例连接 | `connect:write` |
| `delete_connect` | 删除已保存的连接 | `connect:write` |

## 可用 Resources

| Resource URI | 说明 | 所需 Scope |
|-------------|------|-----------|
| `ech0://posts/recent` | 最近的帖子（默认 20 条） | `echo:read` |
| `ech0://posts/{id}` | 按 ID 读取单篇帖子 | `echo:read` |
| `ech0://tags` | 所有标签及使用次数 | `echo:read` |
| `ech0://profile/me` | 当前用户资料 | `profile:read` |
| `ech0://comments/recent` | 最近的公开评论（默认 20 条） | `comment:read` |
| `ech0://stats/heatmap` | 过去一年的每日帖子数量热力图 | `echo:read` |
| `ech0://connect/self` | 当前实例的公开信息卡片（名称、URL、logo、帖子统计、版本） | `connect:read` |

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
