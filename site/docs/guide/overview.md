---
title: 产品概览
description: Ech0 是什么、适合谁、能做什么
---

# Ech0 是什么？

Ech0 是**自托管的个人微博客**：内容在一条**时间线**上展示，可以分享、评论，数据在你自己的服务器上。

和 Memos 这类「先快速记下想法」的工具相比，Ech0 更侧重**记录之后**——把时间线公开或半公开地发出去，让别人能持续看到和互动；不是双链笔记，也不是团队协作文档。

**更适合你**，若你想：在自己的机器或域名上发短文、链接和媒体；保留数据主权，同时需要 RSS、评论等轻量能力。

**不太适合**，若你需要：Obsidian 式知识库、Notion 式团队空间，或只做私密备忘、完全不发布时间线。

---

## 部署与文档索引

- 安装：[安装部署](/docs/start/installation)（**推荐 Docker**，与 README 一致）
- 升级：[版本更新](/docs/start/update)
- 多实例：[互联聚合](/docs/guide/federation)（Connect + `/hub`）
- 登录：[统一登录](/docs/guide/sso)（OAuth + Passkey）
- 评论：[评论系统](/docs/guide/comment)
- 附件与数据：[对象存储](/docs/guide/s3)、[数据管理](/docs/guide/datacontrol)

---

## 能力一览（概要）

- **写作**：Markdown 时间线，图片、链接，以及音乐/视频等扩展卡片。
- **数据**：默认本地；附件可接 **S3 兼容**存储；支持快照与迁移。
- **互动**：内建评论；**RSS**；**Connect / Hub** 浏览多实例。
- **账号**：OAuth、Passkey；自动化用 **访问令牌** 调 API。
- **扩展**：Webhook、Agent 摘要等（见各篇文档）。

## 许可

**AGPL-3.0** 开源，无广告、无强制云服务。
