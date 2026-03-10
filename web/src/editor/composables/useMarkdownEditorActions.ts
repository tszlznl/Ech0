import type { MarkdownEditorAction } from '../types'

function ensureSelection(
  textArea: HTMLTextAreaElement,
  fallback: string,
): { start: number; end: number; selected: string } {
  const start = textArea.selectionStart ?? 0
  const end = textArea.selectionEnd ?? 0
  const selected = textArea.value.slice(start, end) || fallback
  return { start, end, selected }
}

function replaceSelection(
  textArea: HTMLTextAreaElement,
  replacement: string,
  start: number,
  end: number,
) {
  textArea.setRangeText(replacement, start, end, 'end')
}

function lineStartIndex(value: string, index: number) {
  const lineBreakIndex = value.lastIndexOf('\n', Math.max(0, index - 1))
  return lineBreakIndex === -1 ? 0 : lineBreakIndex + 1
}

function lineEndIndex(value: string, index: number) {
  const i = value.indexOf('\n', index)
  return i === -1 ? value.length : i
}

export function applyMarkdownAction(textArea: HTMLTextAreaElement, action: MarkdownEditorAction) {
  const value = textArea.value
  const start = textArea.selectionStart ?? 0
  const end = textArea.selectionEnd ?? 0

  switch (action) {
    case 'bold': {
      const picked = ensureSelection(textArea, '粗体文本')
      replaceSelection(textArea, `**${picked.selected}**`, picked.start, picked.end)
      break
    }
    case 'italic': {
      const picked = ensureSelection(textArea, '斜体文本')
      replaceSelection(textArea, `*${picked.selected}*`, picked.start, picked.end)
      break
    }
    case 'heading': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = line.startsWith('## ') ? line : `## ${line || '标题'}`
      replaceSelection(textArea, normalized, from, to)
      break
    }
    case 'quote': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = line.startsWith('> ') ? line : `> ${line || '引用'}`
      replaceSelection(textArea, normalized, from, to)
      break
    }
    case 'unorderedList': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = line.startsWith('- ') ? line : `- ${line || '列表项'}`
      replaceSelection(textArea, normalized, from, to)
      break
    }
    case 'orderedList': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = /^\d+\.\s/.test(line) ? line : `1. ${line || '列表项'}`
      replaceSelection(textArea, normalized, from, to)
      break
    }
    case 'codeBlock': {
      const picked = ensureSelection(textArea, '')
      const block = `\n\`\`\`\n${picked.selected}\n\`\`\`\n`
      replaceSelection(textArea, block, picked.start, picked.end)
      break
    }
    case 'link': {
      const picked = ensureSelection(textArea, '链接文本')
      const markdownLink = `[${picked.selected}](https://)`
      replaceSelection(textArea, markdownLink, picked.start, picked.end)
      break
    }
  }

  textArea.dispatchEvent(new Event('input', { bubbles: true }))
  textArea.focus()
}
