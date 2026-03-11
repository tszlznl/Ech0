import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

const CODE_BLOCK_COLLAPSE_THRESHOLD = 18
const CODE_BLOCK_COLLAPSED_LINES = 10

function escapeHtml(input: string): string {
  return input
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}

function getLineCount(code: string): number {
  if (!code) return 0
  return code.split(/\r?\n/).length
}

function renderCodeBlock(code: string, rendered: string, language?: string): string {
  const lineCount = getLineCount(code)
  const isCollapsible = lineCount >= CODE_BLOCK_COLLAPSE_THRESHOLD
  const languageClass = language ? ` language-${language}` : ''
  const pre = `<pre><code class="hljs${languageClass}">${rendered}</code></pre>`

  if (!isCollapsible) {
    return pre
  }

  return `<div class="code-block code-block--collapsible code-block--collapsed" style="--code-max-lines:${CODE_BLOCK_COLLAPSED_LINES};"><button type="button" class="code-block-toggle" data-expand-label="展开" data-collapse-label="收起" aria-expanded="false">展开</button>${pre}</div>`
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
        return renderCodeBlock(str, rendered, language)
      } catch {
        // 降级到默认转义输出
      }
    }

    return renderCodeBlock(str, escapeHtml(str))
  },
})

type LinkOpenRule = NonNullable<MarkdownIt['renderer']['rules']['link_open']>

const originalLinkOpen: LinkOpenRule =
  markdown.renderer.rules.link_open ??
  ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options))

markdown.renderer.rules.link_open = (...args: Parameters<LinkOpenRule>) => {
  const [tokens, idx, options, env, self] = args
  const token = tokens[idx]
  const href = token.attrGet('href') ?? ''

  if (/^https?:\/\//i.test(href)) {
    token.attrSet('target', '_blank')
    token.attrSet('rel', 'noopener noreferrer')
  }

  return originalLinkOpen(tokens, idx, options, env, self)
}

export function renderMarkdown(source: string): string {
  if (!source) return ''
  return markdown.render(source)
}
