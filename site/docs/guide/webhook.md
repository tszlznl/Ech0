---
title: 事件推送
description: Webhook 配置、事件类型与验签
---

当系统里发生你关心的事件时，Ech0 可以向配置的 **HTTP URL** 发送 **POST** 请求（JSON），用于联动 Slack、Discord、自建服务等。

在 **系统设置 → Webhook** 中新建：填写**名称**、**URL**（`http` 或 `https`）、可选 **Secret**（用于验签），并启用。

---

## 安全限制

接收地址**不能**指向内网、localhost、`.local` 等（防 SSRF）。本地调试请用**公网可访问的隧道域名**。

---

## 常见事件类型（节选）

| Topic（示意） | 说明 |
| --- | --- |
| `user.created` / `user.updated` / `user.deleted` | 用户变更 |
| `echo.created` / `echo.updated` / `echo.deleted` | 动态发布与变更 |
| `comment.created` / `comment.status.updated` / `comment.deleted` | 评论 |
| `resource.uploaded` | 文件上传完成 |
| `system.backup` / `system.export` | 备份与导出 |
| `system.backup_schedule.updated` | 备份计划变更 |
| `inbox.clear` | 收件箱清空 |
| `ech0.update.check` | 版本检查 |

完整列表与字段以当前版本及 **`/swagger`** 为准。

---

## 请求长什么样

一般会带类似请求头：`Content-Type: application/json`、`User-Agent`、`X-Ech0-Event`、`X-Ech0-Event-ID`、`X-Ech0-Timestamp`；若配置了 Secret，会有 `X-Ech0-Signature`（如 `sha256=<hex>`）。

正文为 JSON，包含事件类型、载荷等。**验签方式**：用配置的 Secret 对**原始请求体字节**做 HMAC-SHA256，与头中的十六进制比对；建议同时做**时间窗口**与**事件 ID 去重**。

---

## 重试与接收建议

投递失败通常会按策略**自动重试**；你的接收端应**尽快返回 2xx**，复杂逻辑可先入队异步处理，避免超时导致反复重试。

---

## 与评论、备份

评论相关事件可与 [评论系统](/docs/guide/comment) 联动排查；备份类事件与 [数据管理](/docs/guide/datacontrol) 中的计划任务相关。
