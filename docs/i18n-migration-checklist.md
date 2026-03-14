# i18n 迁移完成清单

## 目标范围

- 前端界面文案国际化（Vue）
- 后端接口消息国际化（Go）
- 系统内置内容与模板国际化
- CI/CD i18n 守卫

## 已完成项

- [x] 定义统一语言协商策略：`user setting > explicit param/header > Accept-Language > default`
- [x] 建立前后端契约：`error_code`、`message_key`、`message_params`
- [x] 后端接入 `go-i18n` 与请求级本地化中间件
- [x] 后端错误响应支持 `message_key` + 参数化本地化
- [x] 前端接入 `vue-i18n`（含初始化与持久化）
- [x] 高优先级页面硬编码文案迁移为 `t()` key
- [x] 系统通知/内置内容支持按站点默认语言本地化
- [x] 增加 i18n 校验脚本：key 对齐、硬编码检测、伪本地化冒烟
- [x] 增加 CI 工作流执行 i18n 守卫
- [x] 输出契约文档：`docs/i18n-contract.md`

## 本轮收尾（命名统一与去重）

- [x] 新增统一通用 key 命名空间：`commonUi.*`
- [x] 收敛重复近义 key（`apply/cancel/edit/done/add/actions/none`）
- [x] 设置页改为优先复用 `commonUi.*`
- [x] 删除已无引用的重复 key（中英文语言包同步）

## 验证结果

- [x] `pnpm run type-check`
- [x] `pnpm run i18n:check`
- [x] IDE lints（本次修改文件）无新增问题

## 后续建议（可选）

- 将剩余非设置页中的通用按钮词（如 `query`, `clear`）继续收敛到公共命名空间
- 增加“未使用 key 检测”脚本，防止长期累积冗余 key
- 对关键页面补充伪本地化截图回归（按钮溢出/换行）
