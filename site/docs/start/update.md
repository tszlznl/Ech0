---
title: 版本更新
description: Docker、Compose、Helm 升级步骤
---

> **v3 → v4**：不能原地升级。请先在 v3 **导出快照**，再部署 v4，在 v4 面板做 **v3 迁移** 导入。

---

## Docker

```bash
docker stop ech0
docker rm ech0
docker pull sn0wl1n/ech0:latest
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

请把 `JWT_SECRET` 换成与你环境一致的值（与初次部署相同），数据卷路径勿改错。

---

## Docker Compose

```bash
cd /path/to/compose
docker compose pull && docker compose up -d --force-recreate
```

若使用旧命令 `docker-compose`，则：

```bash
docker-compose pull && docker-compose up -d --force-recreate
```

可选清理悬空镜像：`docker image prune -f`

---

## Kubernetes（Helm）

1. 进入克隆下来的仓库目录，拉最新代码：`git pull`  
2. 升级 release：

```bash
helm upgrade ech0 ./charts/ech0
```

若使用了自定义 release 名与命名空间：

```bash
helm upgrade my-ech0 ./charts/ech0 --namespace my-namespace
```
