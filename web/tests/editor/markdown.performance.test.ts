import { describe, expect, it } from 'vitest'
import { performance } from 'node:perf_hooks'
import { renderMarkdown } from '../../src/editor/core/markdown'

const PERF_THRESHOLD_MS = 900
const ITERATIONS = 120

function createBenchmarkMarkdown(): string {
  const sections: string[] = []

  for (let i = 0; i < 40; i += 1) {
    sections.push(`## Section ${i + 1}`)
    sections.push(`Paragraph with link https://example.com/${i}`)
    sections.push(`- [ ] pending task ${i}`)
    sections.push(`- [x] done task ${i}`)
    sections.push('```ts')
    sections.push(`const n${i} = ${i}`)
    sections.push(`console.log(n${i})`)
    sections.push('```')
  }

  return sections.join('\n')
}

describe('renderMarkdown performance baseline', () => {
  it('stays under regression threshold for common mixed content', () => {
    const source = createBenchmarkMarkdown()

    const start = performance.now()
    for (let i = 0; i < ITERATIONS; i += 1) {
      renderMarkdown(source, { expandLabel: 'Expand', collapseLabel: 'Collapse' })
    }
    const elapsedMs = performance.now() - start

    expect(elapsedMs).toBeLessThan(PERF_THRESHOLD_MS)
  })
})
