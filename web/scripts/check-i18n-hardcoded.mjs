import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const root = process.cwd()
const targetFiles = [
  'src/views/panel/modules/TheCommentManager.vue',
  'src/service/request/index.ts',
  'src/stores/user.ts',
]

const chinesePattern = /[\u4E00-\u9FFF]+/g
const tCallPattern = /\bt\(\s*['"`][^'"`]+['"`]/g

const violations = []

for (const relativePath of targetFiles) {
  const abs = join(root, relativePath)
  const content = readFileSync(abs, 'utf8')
  const chineseMatches = content.match(chinesePattern) || []
  if (chineseMatches.length === 0) continue

  const tMatches = content.match(tCallPattern) || []
  if (tMatches.length === 0) {
    violations.push(relativePath)
  }
}

if (violations.length > 0) {
  console.error('Potential hardcoded text found (no i18n t() usage in file):')
  violations.forEach((v) => console.error(`- ${v}`))
  process.exit(1)
}

console.log('Hardcoded text heuristic check passed.')
