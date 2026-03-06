# storagex — 统一虚拟文件系统契约

`storagex` 是 Ech0 的**纯领域层**存储抽象，定义了与后端无关的统一 VFS 接口。业务代码只依赖本包，不直接接触本地磁盘或对象存储。

## 核心设计

```
业务层
  │
  ▼
storagex.FS          ← 统一操作面
  │
  ├─ LocalFS         ← data/ 作为 root
  └─ ObjectFS        ← bucket 作为 root
```

所有文件操作使用**虚拟路径**（如 `/images/a.png`），由后端实现负责映射到物理位置。

## 包内文件

| 文件 | 职责 |
|---|---|
| `fs.go` | `FS` 接口（Open/Write/Delete/Stat/List/Exists）、`Signer`、`URLResolver`、`FileInfo`、`WriteOptions` |
| `path.go` | 路径规范：`NormalizePath`、`ValidateSegment`、`JoinPath`、`TrimVirtualPath` |
| `category.go` | 文件分类：`Category` 常量、`NormalizeCategory`、`CategoryDir`（category → 虚拟目录映射） |
| `key.go` | Key 生成：`KeyGenerator` 接口、`RandomKeyGenerator`、`StaticKeyGenerator` |
| `errors.go` | 标准错误：`ErrInvalidPath`、`ErrNotFound`、`ErrAlreadyExists`、`ErrUnsupported` |
| `config.go` | 后端配置：`ObjectStorageConfig`（S3/R2/MinIO 连接参数，存储层自治、不依赖业务 model） |

## FS 接口

```go
type FS interface {
    Open(ctx context.Context, path string) (io.ReadCloser, error)
    Write(ctx context.Context, path string, r io.Reader, opts WriteOptions) error
    Delete(ctx context.Context, path string) error
    Stat(ctx context.Context, path string) (*FileInfo, error)
    List(ctx context.Context, prefix string) ([]FileInfo, error)
    Exists(ctx context.Context, path string) (bool, error)
}
```

可选能力通过独立接口声明，后端按需实现：

- **`Signer`** — 生成预签名 URL（对象存储场景）
- **`URLResolver`** — 将虚拟路径解析为公开访问 URL

## 路径规范

- 虚拟路径始终以 `/` 开头：`/images/a.png`
- 禁止 `..` 遍历、空段、特殊字符
- `NormalizePath` 会自动清理 `//`、尾 `/`，并校验每个 segment
- `TrimVirtualPath` 去掉前导 `/`，生成对象存储 key

## Key 生成

`KeyGenerator` 负责从上传请求（category + userID + filename）生成唯一虚拟路径：

- **`RandomKeyGenerator`** — `/images/1_1700000000_abc123.png`（带用户 ID + 时间戳 + 随机后缀）
- **`StaticKeyGenerator`** — `/audios/music.mp3`（固定路径，用于单例文件如背景音乐）
