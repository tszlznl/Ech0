# 表格设计标准（Panel 模块）

本文档用于统一 `Panel` 各模块（如 `Comment Manager`、`Webhook` 等）的表格设计与交互规范，减少样式漂移和重复调参。

## 1. 设计目标

- 桌面端：信息清晰、列间距适中、操作可达。
- 小屏端：优先保持同一张表，通过横向滚动浏览，不拆卡片。
- 交互一致：状态展示、操作按钮、滚动条提示统一。

## 2. 容器与滚动

- 表格外层统一使用：
  - `x-scrollbar overflow-x-auto rounded-lg border border-[var(--color-border-subtle)]`
- 所有可能横向溢出的表格都必须带 `x-scrollbar`，确保用户感知“可左右滑动”。
- `x-scrollbar` 样式定义在 `web/src/assets/main.css`，禁止各页面重复定义滚动条样式。

## 3. 表格基础结构

- 建议基础写法：
  - `table`：`w-full min-w-[xxx] table-fixed text-sm`
  - `thead tr`：`bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]`
  - `tbody tr`：`border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]`
- 说明：
  - `w-full`：默认填满容器，避免右侧空白。
  - `min-w-[xxx]`：为小屏保留横向滚动空间。
  - `table-fixed`：列宽可控，不因某列内容把整体拉散。

## 4. 列宽与间距规范

- 原则：**刚好够用 + 少量余量**，避免“列太宽导致看起来很松”。
- 建议列宽策略：
  - 文本短列（状态、开关、动作）：固定较小宽度（如 `72~96px`）。
  - URL/邮箱等中长列：中等宽度（如 `180~260px`）+ `truncate`。
  - 主文本列（名称/昵称）：`100~140px`。
- 内边距建议：
  - 表头/单元格 `px-1` 到 `px-2`，仅在确有拥挤时提升到 `px-3`。

## 5. 文案与换行

- 表头默认加 `whitespace-nowrap`（尤其 `Time`、`Actions` 等短列）。
- 操作按钮文案（如 `Set as hot`）必须防换行：
  - `.table-action { white-space: nowrap; flex-shrink: 0; }`
- 时间列建议防换行，避免日期被拆行影响可读性。

## 6. 状态与开关

- 状态使用统一 `status-pill`（tag）体系：
  - 成功：绿
  - 失败：红
  - 未知：黄
- 启停能力建议独立列展示，使用统一开关组件（`BaseSwitch`）。
- 若同列同时有 `Switch + Tag`，优先保留 `Switch`，避免信息冗余。

## 7. 操作列规范

- 图标操作统一使用 `BaseButton`（避免原生按钮风格不一致）。
- 推荐尺寸：
  - `class="h-8 w-8 !p-1.5"`（根据密度可微调）。
- 操作列水平间距建议 `gap-1` 或 `gap-2`，避免过稀。
- 所有操作按钮应提供 `title`，提升可理解性与可访问性。

## 8. 复选框对齐（如有）

- 复选框列（头部/行）应居中对齐：
  - 单元格加 `align-middle`
  - 内层容器用 `flex items-center justify-center`

## 9. 落地检查清单

- 是否使用了统一外层：`x-scrollbar + overflow-x-auto + border + rounded`。
- 表格是否 `w-full` 且设置了合理 `min-w`。
- 关键列是否存在无意义大宽度。
- `Actions`、`Time`、长文案按钮是否避免换行。
- 小屏是否通过横向滚动可完整访问所有列。
- 开关、状态、按钮是否与全局组件规范一致（`BaseSwitch` / `BaseButton`）。

