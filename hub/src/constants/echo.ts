/** 与 web/src/enums/enums 中 ImageLayout 取值一致，避免 Hub 直接依赖 web 的 enum 编译选项冲突 */
export const ImageLayout = {
  WATERFALL: 'waterfall',
  GRID: 'grid',
  HORIZONTAL: 'horizontal',
  CAROUSEL: 'carousel',
  STACK: 'stack',
} as const
