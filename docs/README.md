# 文档索引

本目录按用途分为 **使用说明**（部署与功能接入）与 **开发设计**（贡献代码时的约定与背景）。品牌与配图在 `imgs/`。

## 使用说明（`usage/`）

面向部署管理员与集成方：如何接入能力、如何迁移数据。

| 文档 | 说明 |
|------|------|
| [usage/mcp-usage.md](usage/mcp-usage.md) | MCP（Model Context Protocol）接入：Token、Host 配置、协议要点 |
| [usage/webhook-usage.md](usage/webhook-usage.md) | Webhook：事件、签名、管理接口与故障处理 |
| [usage/storage-migration.md](usage/storage-migration.md) | 存储迁移：本地与 S3、`key` 与路径规则、换桶与迁移注意事项 |

## 开发设计（`dev/`）

面向本仓库贡献者：UI 约定、国际化契约、日志规范，以及鉴权模型的设计记录。

| 文档 | 说明 |
|------|------|
| [dev/table-design-standard.md](dev/table-design-standard.md) | Panel 表格组件（含管理端列表）的布局与交互规范 |
| [dev/i18n-contract.md](dev/i18n-contract.md) | 前后端国际化约定（locale、API 错误字段、key 命名） |
| [dev/logging.md](dev/logging.md) | 日志库使用与字段约定 |
| [dev/access-token-scope-design.md](dev/access-token-scope-design.md) | Access Token 的 `typ` / scope / audience 设计背景（实现以代码为准） |

## 资源（`imgs/`）

Logo、架构示意图等静态资源；根目录 [README.md](../README.md) 中的预览图等链接指向此处。
