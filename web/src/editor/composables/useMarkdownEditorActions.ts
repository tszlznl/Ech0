// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

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
  textArea.focus()
  textArea.setSelectionRange(start, end)
  const changedByCommand =
    typeof document !== 'undefined' &&
    typeof document.execCommand === 'function' &&
    document.execCommand('insertText', false, replacement)

  if (!changedByCommand) {
    textArea.setRangeText(replacement, start, end, 'end')
  }
}

function insertSnippet(
  textArea: HTMLTextAreaElement,
  snippet: string,
  cursorOffset: number,
  start: number,
  end: number,
) {
  textArea.setRangeText(snippet, start, end, 'end')
  const cursor = start + cursorOffset
  textArea.setSelectionRange(cursor, cursor)
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
      if (start !== end) {
        const picked = ensureSelection(textArea, '')
        replaceSelection(textArea, `**${picked.selected}**`, picked.start, picked.end)
        break
      }
      insertSnippet(textArea, '****', 2, start, end)
      break
    }
    case 'italic': {
      if (start !== end) {
        const picked = ensureSelection(textArea, '')
        replaceSelection(textArea, `*${picked.selected}*`, picked.start, picked.end)
        break
      }
      insertSnippet(textArea, '**', 1, start, end)
      break
    }
    case 'heading': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = line.startsWith('## ') ? line : `## ${line}`
      replaceSelection(textArea, normalized, from, to)
      if (!line) {
        const cursor = from + 3
        textArea.setSelectionRange(cursor, cursor)
      }
      break
    }
    case 'quote': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = line.startsWith('> ') ? line : `> ${line}`
      replaceSelection(textArea, normalized, from, to)
      if (!line) {
        const cursor = from + 2
        textArea.setSelectionRange(cursor, cursor)
      }
      break
    }
    case 'unorderedList': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = line.startsWith('- ') ? line : `- ${line}`
      replaceSelection(textArea, normalized, from, to)
      if (!line) {
        const cursor = from + 2
        textArea.setSelectionRange(cursor, cursor)
      }
      break
    }
    case 'orderedList': {
      const from = lineStartIndex(value, start)
      const to = lineEndIndex(value, end)
      const line = value.slice(from, to)
      const normalized = /^\d+\.\s/.test(line) ? line : `1. ${line}`
      replaceSelection(textArea, normalized, from, to)
      if (!line) {
        const cursor = from + 3
        textArea.setSelectionRange(cursor, cursor)
      }
      break
    }
    case 'codeBlock': {
      const picked = ensureSelection(textArea, '')
      const block = `\n\`\`\`\n${picked.selected}\n\`\`\`\n`
      replaceSelection(textArea, block, picked.start, picked.end)
      if (!picked.selected) {
        const cursor = picked.start + 5
        textArea.setSelectionRange(cursor, cursor)
      }
      break
    }
    case 'link': {
      if (start !== end) {
        const picked = ensureSelection(textArea, '')
        const markdownLink = `[${picked.selected}]()`
        replaceSelection(textArea, markdownLink, picked.start, picked.end)
        break
      }
      insertSnippet(textArea, '[]()', 1, start, end)
      break
    }
  }

  textArea.dispatchEvent(new Event('input', { bubbles: true }))
  textArea.focus()
}
