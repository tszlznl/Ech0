# Token Guide

## 目标

- 统一项目主题变量命名与分层，降低页面样式耦合。
- 支持通过修改 token 文件快速魔改主题。
- 避免业务页面直接绑定历史场景化 token 名。

## 分层规则

- `foundation`: 设计原子（字体、圆角、阴影、动效等）。
- `semantic`: 语义层（颜色/边框/文本语义）。
- `component`: 组件层（按钮、输入框、对话框、选择器等）。

## 文件结构

- `themes/tokens/foundation.scss`
- `themes/tokens/semantic.light.scss`
- `themes/tokens/semantic.dark.scss`
- `themes/tokens/component.light.scss`
- `themes/tokens/component.dark.scss`
- `themes/light.scss`、`themes/dark.scss` 仅作为薄装配入口

## 命名规范

- 颜色：`--color-*`
- 圆角：`--radius-*`
- 阴影：`--shadow-*`
- 字体：`--font-family-*`
- 组件：`--btn-*`、`--input-*`、`--dialog-*`、`--select-*`、`--switch-*`

## 禁止项

- 新增 `--panel-*` 前缀 token。
- 新增 `var(--font-sans|font-display|font-mono)` 旧字体 token 引用。
- 新增或引用以下 legacy 前缀：`text-color*`、`text-color-next*`、`bg-color*`、`border-color*`、`ring-color*`、`divide-color*`、`timeline*`、`widget*`、`echo*`、`dashboard*`、`button-primary*`、`main-color*`、`tag-editor*`、`editor*`、`connect*`、`heatmap*`。

## 使用建议

- 页面/业务组件优先使用 `semantic` token。
- `Base*` 组件只使用 `component` token。
- 仅在 token 文件中定义具体色值，页面不硬编码 hex。

## 质量检查

- 运行 `pnpm token:check` 检查禁用 token 是否回流。
- 提交前运行 `pnpm lint` 与核心页面暗黑/亮色切换回归。
