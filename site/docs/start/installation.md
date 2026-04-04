---
title: 安装部署
description: 推荐用 Docker，其次 Compose / 脚本 / Helm
---

> **推荐顺序**：**Docker 单容器** → Docker Compose → 安装脚本（systemd）→ 克隆仓库后用 Helm → 直接运行二进制。下文与仓库根目录 `README.zh.md` 保持一致。

---

## Docker（推荐）

```bash
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

- 部署完成后浏览器访问 `http://<服务器IP>:6277`。
- **请把** `JWT_SECRET="Hello Echos"` **改成你自己的随机字符串**，不要长期使用示例值。
- **第一个注册的账号**会成为管理员（Owner）；当前版本默认只有高权限账号可以发帖。
- 数据持久化在上例的 `/opt/ech0/data`（可按需改挂载路径）。

---

## Docker Compose

新建目录，放入 `docker-compose.yml`（可参考仓库示例或下列最小示例），在该目录执行：

```bash
docker compose up -d
```

若你本机命令仍是 `docker-compose`（带横杠），则使用：

```bash
docker-compose up -d
```

示例：

```yaml
services:
  ech0:
    image: sn0wl1n/ech0:latest
    container_name: ech0
    ports:
      - "6277:6277"
    volumes:
      - ./data:/app/data
    environment:
      - JWT_SECRET=请改为随机强密码
    restart: unless-stopped
```

---

## 脚本安装（systemd）

与 README 相同，从仓库拉取安装脚本：

```bash
curl -fsSL "https://raw.githubusercontent.com/lin-snow/Ech0/main/scripts/ech0.sh" -o ech0.sh && bash ech0.sh
```

- 脚本通过 **systemd** 安装和管理服务，涉及服务启停时通常需要 **root**。
- 自定义安装目录可执行：`bash ech0.sh install /your/path/ech0`（以脚本 `--help` 为准）。

---

## Kubernetes（Helm）

项目未提供在线 Helm 仓库，需要**先克隆仓库**再在本地安装：

```bash
git clone https://github.com/lin-snow/Ech0.git
cd Ech0
helm install ech0 ./charts/ech0
```

可自定义 release 名与命名空间，例如：

```bash
helm install my-ech0 ./charts/ech0 --namespace my-namespace --create-namespace
```

`JWT_SECRET` 等可通过 `values.yaml` 或 `--set` 传入，详见 Chart 目录内说明。

---

## 二进制

从 [GitHub Releases](https://github.com/lin-snow/Ech0/releases) 下载对应平台压缩包，解压后：

```bash
./ech0 serve
```

常见架构包括 `linux/amd64`、`linux/arm64`、`linux/armv7`、`windows/amd64` 等。

---

## 首次使用

按向导完成初始化；用户数量、发帖权限等以当前版本 **设置界面** 为准。
