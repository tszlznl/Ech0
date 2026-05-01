// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import chalk from 'chalk'

// ASCII Art Banner
const banner = `
███████╗     ██████╗    ██╗  ██╗     ██████╗
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝

` as const

const gradientColors = [
  chalk.hex('#f38ba8'), // Catppuccin Pink
  chalk.hex('#fab387'), // Catppuccin Peach
  chalk.hex('#f9e2af'), // Catppuccin Yellow
  chalk.hex('#a6e3a1'), // Catppuccin Green
  chalk.hex('#94e2d5'), // Catppuccin Teal
  chalk.hex('#89b4fa'), // Catppuccin Blue
  chalk.hex('#cba6f7'), // Catppuccin Mauve
  chalk.hex('#f5c2e7'), // Catppuccin Flamingo
  chalk.hex('#eba0ac'), // Catppuccin Maroon
] as const

function printGradientBanner(text: string): string {
  const lines = text.trim().split('\n')
  return lines
    .map((line, index) => {
      const colorFn = gradientColors[index % gradientColors.length]
      return colorFn ? colorFn(line) : line
    })
    .join('\n')
}

function printWelcome(): void {
  // 只打印渐变 Banner
  console.log() // 添加一个空行
  console.log(printGradientBanner(banner))
  console.log() // 添加一个空行
}

printWelcome()

export { printWelcome }
