import { readdirSync, readFileSync, statSync } from 'node:fs'
import { extname, join } from 'node:path'

const root = process.cwd()
const supportedExt = new Set(['.vue', '.ts', '.tsx', '.js', '.jsx', '.scss', '.css'])
const legacyPrefixes = [
  'text-color',
  'text-color-next',
  'bg-color',
  'border-color',
  'ring-color',
  'divide-color',
  'timeline',
  'widget',
  'echo',
  'dashboard',
  'button-primary',
  'main-color',
  'tag-editor',
  'editor',
  'connect',
  'heatmap',
]
const legacyPrefixPattern = legacyPrefixes
  .map((prefix) => prefix.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
  .join('|')
const forbiddenPatterns = [
  { name: 'panel token prefix', regex: /\b(?:var\(--panel-|--panel-)/g },
  { name: 'legacy font token', regex: /var\(--font-(?:sans|display|mono)\)/g },
  {
    name: 'legacy token prefix',
    regex: new RegExp(`\\b(?:var\\(--|--)(?:${legacyPrefixPattern})(?:-[a-z0-9]+)*\\b`, 'g'),
  },
]

const ignored = [/\/src\/themes\/custom\.scss$/]
const violations = []

const walkFiles = (dir, relativeBase = '') => {
  const current = join(root, dir)
  const entries = readdirSync(current)
  for (const entry of entries) {
    const rel = join(relativeBase, entry)
    const abs = join(current, entry)
    const stats = statSync(abs)
    if (stats.isDirectory()) {
      walkFiles(join(dir, entry), rel)
      continue
    }
    if (!supportedExt.has(extname(entry))) continue
    if (ignored.some((rule) => rule.test(abs))) continue

    const content = readFileSync(abs, 'utf8')
    for (const rule of forbiddenPatterns) {
      const matches = content.match(rule.regex)
      if (matches && matches.length > 0) {
        violations.push({ file: join('src', rel), rule: rule.name, count: matches.length })
      }
    }
  }
}

walkFiles('src')

if (violations.length > 0) {
  console.error('Found forbidden token usage:\n')
  for (const item of violations) {
    console.error(`- ${item.file}: ${item.rule} (${item.count})`)
  }
  process.exit(1)
}

console.log('Token usage check passed.')
