import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

function escapeHtml(input: string): string {
  return input
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}

const markdown: MarkdownIt = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: false,
  langPrefix: 'language-',
  highlight(str: string, lang: string): string {
    const language = lang?.trim()
    if (language && hljs.getLanguage(language)) {
      try {
        const rendered = hljs.highlight(str, {
          language,
          ignoreIllegals: true,
        }).value
        return `<pre><code class="hljs language-${language}">${rendered}</code></pre>`
      } catch {
        // 降级到默认转义输出
      }
    }

    return `<pre><code class="hljs">${escapeHtml(str)}</code></pre>`
  },
})

const originalLinkOpen =
  markdown.renderer.rules.link_open ??
  ((tokens: any[], idx: number, options: any, _env: any, self: any) =>
    self.renderToken(tokens, idx, options))

markdown.renderer.rules.link_open = (
  tokens: any[],
  idx: number,
  options: any,
  env: any,
  self: any,
) => {
  const token = tokens[idx]
  const hrefIndex = token.attrIndex('href')

  if (hrefIndex >= 0) {
    const href = token.attrs?.[hrefIndex]?.[1] ?? ''
    if (/^https?:\/\//i.test(href)) {
      token.attrSet('target', '_blank')
      token.attrSet('rel', 'noopener noreferrer')
    }
  }

  return originalLinkOpen(tokens, idx, options, env, self)
}

export function renderMarkdown(source: string): string {
  if (!source) return ''
  return markdown.render(source)
}
