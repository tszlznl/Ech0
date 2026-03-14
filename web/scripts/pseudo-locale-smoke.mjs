import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const root = process.cwd()
const zhPath = join(root, 'src/locales/messages/zh-CN.json')
const zh = JSON.parse(readFileSync(zhPath, 'utf8'))

const pseudo = (value) => {
  return `［${String(value)
    .replace(/[a-zA-Z]/g, (c) => `${c}${c}`)
    .replace(/\s+/g, ' ')}］`
}

const transform = (obj) => {
  if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
    const output = {}
    for (const [k, v] of Object.entries(obj)) {
      output[k] = transform(v)
    }
    return output
  }
  return pseudo(obj)
}

const pseudoLocale = transform(zh)
const sample = pseudoLocale?.commentManager?.title
if (!sample || !sample.startsWith('［') || !sample.endsWith('］')) {
  console.error('Pseudo locale generation failed.')
  process.exit(1)
}

console.log('Pseudo locale smoke check passed.')
