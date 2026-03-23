import { describe, expect, it } from 'vitest'
import { renderMarkdown } from '../../src/editor/core/markdown'

describe('renderMarkdown task list', () => {
  it('渲染已完成与未完成任务为 checkbox', () => {
    const source = ['- [ ] pending item', '- [x] completed item'].join('\n')

    const html = renderMarkdown(source)

    expect(html).toContain('type="checkbox"')
    expect(html).toContain('aria-label="Task item"')
    expect(html).toContain('task-list-item')
    expect(html).toContain('checked')
  })

  it('渲染松散列表中的任务项为 checkbox', () => {
    const source = ['- [ ] pending item', '', '- [X] completed item'].join('\n')

    const html = renderMarkdown(source)

    expect(html).toContain('task-list-item-checkbox')
    expect(html).toContain('<input class="task-list-item-checkbox"')
    expect(html).toContain('aria-label="Task item"')
    expect(html).toContain('checked')
  })
})
