# internal/storage — VFS 基础设施实现层

本包是 `pkg/storagex` 契约的具体实现，提供本地磁盘和对象存储两种后端，以及业务级 StorageService、工厂构建和数据迁移能力。

## 目录结构

```
internal/storage/
├── service.go        # StorageService: 业务级文件操作（Upload/Delete/Presign/Open/...）
├── types.go          # Category、Source 常量、元数据类型
├── local/
│   └── fs.go         # LocalFS: 本地磁盘 VFS 实现
├── objectfs/
│   └── fs.go         # ObjectFS: S3 兼容对象存储 VFS 实现
├── factory/
│   └── factory.go    # 工厂: 根据配置构建 StorageService 或 storagex.FS
├── migrate/
│   └── migrate.go    # 迁移工具: Copy(src, dst, prefix)
└── zip/
    └── zip.go        # ZIP 压缩/解压工具
```

## 架构

```
CommonService
      │
      ▼
StorageService         ← 业务级 API（Upload/Delete/Presign/ResolveURL/Open/Stat/Exists）
      │
      ▼
storagex.FS            ← 统一 VFS 接口
      │
  ┌───┴────┐
  ▼        ▼
LocalFS  ObjectFS
```

没有额外的 interface 中间层——`StorageService` 是具体 struct，直接持有 `storagex.FS`。

## StorageService

[service.go](service.go) 是业务代码的唯一存储入口，替代了旧的 `StoragePort` 接口和 `PortBridge` 适配层。

| 方法 | 说明 |
|---|---|
| `Upload(ctx, category, userID, fileName, contentType, reader)` | key 生成 + write + URL 解析 |
| `Delete(ctx, path)` | 按虚拟路径删除 |
| `Presign(ctx, category, userID, fileName, contentType, method, expiry)` | 生成预签名 URL |
| `ResolveURL(ctx, virtualPath)` | 虚拟路径 → 公开 URL |
| `Open(ctx, path)` | 读取文件 |
| `Stat(ctx, path)` | 获取文件元数据 |
| `Exists(ctx, path)` | 检查文件是否存在 |
| `Source()` | 返回后端标识 ("local" / "s3") |
| `VFS()` | 暴露底层 storagex.FS |

## 后端实现

### LocalFS (`local/`)

虚拟路径映射到本地目录树：

```
root = "data/files"（默认）
虚拟路径 /images/a.png → 物理路径 data/files/images/a.png
URL 解析 /images/a.png → /files/images/a.png
```

构造方式（functional options）：

```go
fs := local.NewLocalFS()                              // 使用默认值
fs := local.NewLocalFS(local.WithRoot("data/files"))  // 自定义 root
```

### ObjectFS (`objectfs/`)

虚拟路径映射到 S3 兼容对象存储的 key：

```
pathPrefix = "uploads"
虚拟路径 /images/a.png → key uploads/images/a.png
```

额外实现了 `storagex.Signer`（预签名 URL）和 `storagex.URLResolver`（CDN/直连 URL）。

```go
fs, err := objectfs.New(s3Setting)
fs, err := objectfs.New(s3Setting, objectfs.WithPathPrefix("v2"))
```

## 工厂 (`factory/`)

```go
svc, err := factory.Build(factory.BuildInput{
    Mode:     factory.ModeLocal,
    DataRoot: "data/files",
})

fs, err := factory.BuildFS(factory.BuildInput{
    Mode:      factory.ModeS3,
    S3Setting: s3Setting,
})
```

## 迁移工具 (`migrate/`)

```go
result, err := migrate.Copy(ctx, srcFS, dstFS, "/images", migrate.Options{
    Conflict: migrate.ConflictSkip,      // 跳过已存在 | ConflictOverwrite 覆盖
    DryRun:   false,                     // true 时只统计不实际复制
})
```

支持 Local→S3、S3→Local、Local→Local 任意方向迁移。
