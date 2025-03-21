# Ech0 - 极致轻量的自托管Memo工具

<p align="center">
  <img alt="Ech0" src="./docs/imgs/FluentEmojiClipboard.svg" width="320">
</p>

**单文件架构**的极简备忘录系统，专注核心记录功能，10.8MB内存即可流畅运行。私人数据完全掌控，支持RSS订阅与永久存档。

![界面预览](./docs/imgs/screenshot.png)

---

## 核心优势

☁️ **原子级轻量**：内存占用仅10.8MB，单SQLite文件存储架构  
🚀 **极速部署**：无需配置，从安装到使用只需1条命令  
✍️ **零干扰写作**：纯净Markdown编辑器，支持快捷键操作  
📦 **数据主权**：所有内容存储于本地SQLite文件，支持RSS订阅导出  
🎉 **永久免费**：MIT协议开源，无追踪/无订阅/无服务依赖
🌍 **跨端适配**：完美兼容桌面/移动浏览器  

---

## 10秒部署

### docker部署

```shell
docker run -d \
  --name ech0 \
  -p 1314:1314 \
  -v /opt/ech0/data:/app/data \
  sn0wl1n/ech0:v2.2.1
```

> 💡 部署完成后访问 ip:1314 即可使用
> 📍 首次使用注册的账号会被设置为管理员（目前仅管理员支持发布内容）

### docker-componse部署

创建一个新目录并将 `docker-compose.yml` 文件放入其中
在该目录下执行以下命令启动服务：

```shell
docker-compose up -d
```

---

<p align="center">「少即是多」—— 路德维希·密斯·凡德罗</p>
