---
title: 开发指南
description: 克隆仓库后在本地跑前后端
---

克隆仓库后，请先阅读根目录 **README.zh.md** 与 **Makefile**。

---

## 后端（Go）

- **Go 版本**：不低于 `go.mod` 中声明的版本（当前为 1.26+）。
- **CGO**：若使用含 SQLite 的构建，需要本机 C 编译器（Windows 可用 MinGW-w64，macOS `brew install gcc`，Linux `build-essential`）。
- **Wire**：若修改了依赖注入，在相应包执行 `wire` 生成 `wire_gen.go`（见 `internal/di/` 等）。
- **代码风格**：可用 **golangci-lint**（`golangci-lint run`、`golangci-lint fmt`）。
- **热重载（可选）**：**Air**，`make air-install` 或 `go install github.com/air-verse/air@latest`。
- **接口文档**：**swag** 生成 Swagger；本地起服务后打开 `http://localhost:6277/swagger/index.html`。

启动：

```bash
make run    # 启动后端
make dev    # 若已安装 Air，则热重载
```

---

## 前端（`web/`）

- 使用 **pnpm** 安装依赖；Node 版本按团队约定（可用 fnm、nvm 管理）。

```bash
cd web
pnpm install
pnpm dev
```

开发服务器端口以终端输出为准（Vite 常见 **5173**），后端默认 **6277**。前端通过代理或环境配置请求后端，详见 `web` 内 Vite 配置。

---

## 联调顺序

1. 启动后端 `make run`。
2. 启动前端 `pnpm dev`。
3. 浏览器访问前端地址，接口指向本地后端。

---

## 更多

事件总线、日志、MCP 等见仓库内 `docs/` 与 README；本页只覆盖「把开发环境跑起来」的最短路径。
