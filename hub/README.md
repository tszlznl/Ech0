# Ech0 Hub

面向 **[hub.ech0.app](https://hub.ech0.app)** 的聚合前端：从一份静态 **实例清单 JSON** 读取多个 Ech0 部署的地址，在浏览器中并行请求各实例的帖子接口，合并为**统一时间线**浏览。

## 架构

| 维度 | 说明 |
|------|------|
| **CSR** | 内容由浏览器端请求数据后渲染，无服务端模板。 |
| **SPA** | Vue 3 + `vue-router` 单页应用，路由切换不整页刷新。 |
| **PWA** | `vite-plugin-pwa` 提供 Web App Manifest 与 Service Worker，可安装到桌面/主屏，并对静态资源做缓存（具体策略以 `vite.config.ts` 为准）。 |

实现演进中的详细任务清单见仓库根目录：`docs/superpowers/plans/2026-04-05-ech0-hub-csr-spa-pwa.md`。

## 如何登记到 Hub 列表

在本仓库用 **Issue 表单「登记到 Ech0 Hub」** 提交即可（需登录 GitHub）。模板会为 Issue 打上 **`hub`** 标签；仓库里需已存在同名标签（可在 **Settings → Labels** 新建）。提交后 Actions 会解析内容并**自动发起带 `hub` 标签的 PR** 修改 `hub/public/hub.json`，维护者审核合并后生效。请事先为实例配置好 **CORS**（允许 `https://hub.ech0.app`）。

## 实例清单 `public/hub.json`

Hub 启动后请求同源的 `/hub.json`（开发时由 Vite 从 `public/` 提供）。每条实例**只需** `id` 与 `url`：

```json
{
  "instances": [
    { "id": "my-instance", "url": "https://your-ech0-origin.example.com" }
  ]
}
```

- **`id`**：实例短标识，用于区分来源与展示。
- **`url`**：实例 API 根地址，**不要**末尾斜杠；聚合请求为 `{url}/api/echo/query`（与主项目 `internal/router` 一致）。探活使用同一主机上的 **`GET {url}/healthz`**（见 `internal/router/resource.go`，非 `/api` 前缀）。

## 探活与版本门槛

1. 对每个实例请求 **`GET {url}/healthz`**，解析统一响应中 **`code === 1`** 的 `data.version`（与 `internal/handler/common/common.go` 中 `Healthz` 一致）。  
2. 仅当版本 **≥ 4.4.0**（与主项目 `internal/model/common/common.go` 中 `Version` 所代表的语义一致）时，该实例才参与帖子聚合。  
3. 不满足或请求失败时，在页面上单独列出原因，不参与时间线。

## 正文 / 图片与 web 复用

- **Markdown 与图片**：通过 Vite `resolve.alias` 将 `@` 指向仓库 `web/src`，复用 `TheMdPreview`（底层 `MarkdownRenderer`）与 `TheImageGallery`；图集请求时传入实例 `baseUrl`，与主站 Hub 场景一致。  
- **全局样式与 i18n**：Hub 入口引入 `web` 的主题 SCSS、`virtual:uno.css`、`vue-i18n`（当前加载 `zh-CN` 文案），以驱动上述组件。  
- **Extension**：聚合侧**不展示**带 `extension` 的帖子（在 `src/composables/useHubMergeFeed.ts` 中过滤）；Hub 不再包含 Extension 卡片或相关封装组件。

## 跨域（CORS）

浏览器从 `hub.ech0.app` 访问各实例时，**`/api/*` 与 `/healthz`** 等均需对 Hub 来源放行 CORS（例如 `Access-Control-Allow-Origin: https://hub.ech0.app`）。若无法修改实例，需要在 Hub 同域做**反向代理**（同源则无 CORS 问题）。

## 开发与构建

```bash
pnpm install
pnpm dev
pnpm build
pnpm preview
```

- 本地开发默认 `http://localhost:5173`，请将实例 CORS 同时允许该来源以便调试。

## 技术栈

Vue 3、TypeScript、Vite、`vue-router`、`vite-plugin-pwa`、`vue-i18n`、`unocss`；并与同仓库 **`web/`** 共享部分展示组件（见上节）。

## 聚合逻辑（摘要）

1. `GET /hub.json` → 实例列表。  
2. `GET {url}/healthz` → 版本 ≥ 4.4.0 的实例进入候选。  
3. 对每个候选实例 `POST {url}/api/echo/query`，请求体字段与主项目 `EchoQueryDto` 一致；成功响应 `code === 1`，`data.items` 为帖子数组（见 `internal/model/common/result.go`）。  
4. 合并多源结果，按 `created_at` 降序；探活失败、拉取失败或部分源失败时，在页面上分别提示。

## 仓库位置

本目录为 Monorepo 中的 `hub/` 子项目，与主项目 Ech0（Go 后端）共享版本历史，但构建与部署独立。
