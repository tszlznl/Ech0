#!/usr/bin/env bash
# 本地提交前可运行：后端 Go 格式化与 lint、前端 web 格式化与 lint、以及 web 的 i18n 校验。
# 等价入口：项目根执行 make check
# 依赖：项目根已安装 golangci-lint；web 目录已 pnpm install。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "=== 后端：格式化 (golangci-lint fmt，同 make fmt) ==="
make fmt

echo "=== 后端：Lint (golangci-lint run，同 make lint) ==="
make lint

echo "=== 前端 web：格式化 (prettier --write src/) ==="
pnpm -C web format

echo "=== 前端 web：Lint (eslint . --fix) ==="
pnpm -C web lint

echo "=== 前端 web：i18n 校验 (key / unused / hardcoded / pseudo-smoke) ==="
pnpm -C web run i18n:check

echo "=== dev_check 全部完成 ==="
