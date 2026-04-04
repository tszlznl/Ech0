# Ech0 MCP 接入指南

Ech0 内建了 MCP（Model Context Protocol）Server，允许 AI 应用（如 Cursor、Claude Desktop 等）通过标准化协议访问你的 Ech0 实例。

## 快速开始

### 1. 创建 MCP 专用 Access Token

在 Ech0 管理后台 **设置 → 访问令牌** 中创建一个新 Token：

- **Audience**：选择 `mcp-remote`
- **Scopes**：根据需要勾选（建议最小权限）
  - 只读场景：`echo:read`、`profile:read`
  - 读写场景：再加上 `echo:write`
- **有效期**：建议选择 8 小时或 1 个月（不建议永不过期）

创建后妥善保存 Token，它只会显示一次。

### 2. 配置 MCP Host

在你的 MCP 客户端（如 Cursor `mcp.json`）中添加：

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

| Tool | 说明 | 所需 Scope |
|------|------|-----------|
| `search_posts` | 搜索帖子（支持关键词、标签过滤、分页） | `echo:read` |
| `get_post` | 根据 ID 获取帖子详情 | `echo:read` |
| `list_tags` | 列出所有标签 | `echo:read` |
| `create_post` | 创建新帖子（支持 Markdown、标签、私密设置） | `echo:write` |
| `update_post` | 更新已有帖子 | `echo:write` |
| `delete_post` | 删除帖子 | `echo:write` |

## 可用 Resources

| Resource URI | 说明 | 所需 Scope |
|-------------|------|-----------|
| `ech0://posts/recent` | 最近的帖子（默认 20 条） | `echo:read` |
| `ech0://tags` | 所有标签及使用次数 | `echo:read` |
| `ech0://profile/me` | 当前用户资料 | `profile:read` |

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
