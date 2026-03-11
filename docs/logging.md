# Ech0 日志规范

## 目标
- 全项目统一使用 `internal/util/log`。
- 日志保持结构化，便于检索与系统日志面板消费。

## 基本规则
- 不再使用 `fmt.Print*`、`println`、标准库 `log.*` 记录业务日志。
- 推荐优先使用：
  - `logUtil.Info(...)`
  - `logUtil.Warn(...)`
  - `logUtil.Error(...)`
  - 或 `logUtil.GetLogger().Info/Warn/Error(...)`
- 错误对象统一使用 `zap.Error(err)`，不要用 `zap.String("error", err.Error())`。

## 字段建议
- 保留通用字段：`time`、`level`、`msg`。
- 业务字段推荐：
  - `module`: 模块名（如 `task`、`storage`、`oauth`）
  - `user_id`: 用户标识
  - `path`: 资源路径
  - `provider`: 第三方提供商
  - `error`: 使用 `zap.Error(err)` 自动生成

## 级别约定
- `Debug`: 高频调试信息，仅本地或临时排查。
- `Info`: 关键流程正常事件（启动、任务完成、状态变化）。
- `Warn`: 可恢复异常（重试、降级、回退）。
- `Error`: 需要关注的失败（任务失败、调用失败、数据写入失败）。

## 示例
```go
logUtil.Error(
  "publish webhook event failed",
  zap.String("module", "webhook"),
  zap.String("event", "resource_uploaded"),
  zap.Error(err),
)
```
