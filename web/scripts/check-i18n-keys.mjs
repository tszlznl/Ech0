import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const root = process.cwd()
const localesDir = join(root, 'src/locales/messages')

const flatten = (obj, prefix = '', result = new Map()) => {
  for (const [key, value] of Object.entries(obj)) {
    const next = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      flatten(value, next, result)
    } else {
      result.set(next, String(value))
    }
  }
  return result
}

const readLocale = (name) => {
  const content = readFileSync(join(localesDir, name), 'utf8')
  return flatten(JSON.parse(content))
}

const base = readLocale('zh-CN.json')
const target = readLocale('en-US.json')

const missingInTarget = []
for (const key of base.keys()) {
  if (!target.has(key)) missingInTarget.push(key)
}

const missingInBase = []
for (const key of target.keys()) {
  if (!base.has(key)) missingInBase.push(key)
}

if (missingInTarget.length > 0 || missingInBase.length > 0) {
  console.error('i18n key mismatch detected:')
  if (missingInTarget.length > 0) {
    console.error(`- Missing in en-US (${missingInTarget.length})`)
    missingInTarget.forEach((k) => console.error(`  - ${k}`))
  }
  if (missingInBase.length > 0) {
    console.error(`- Missing in zh-CN (${missingInBase.length})`)
    missingInBase.forEach((k) => console.error(`  - ${k}`))
  }
  process.exit(1)
}

console.log('i18n key parity check passed.')
