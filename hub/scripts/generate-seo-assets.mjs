import { writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const SCRIPT_DIR = dirname(fileURLToPath(import.meta.url))
const HUB_ROOT = join(SCRIPT_DIR, '..')
const PUBLIC_DIR = join(HUB_ROOT, 'public')

const SITE_URL = (process.env.VITE_HUB_SITE_ORIGIN ?? process.env.HUB_SITE_ORIGIN ?? 'https://hub.ech0.app').replace(
  /\/+$/,
  '',
)

function sitemapXml(urls) {
  const body = urls
    .map(
      (url) => `  <url>
    <loc>${url.loc}</loc>
    <changefreq>${url.changefreq}</changefreq>
    <priority>${url.priority}</priority>
  </url>`,
    )
    .join('\n')
  return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${body}
</urlset>
`
}

async function main() {
  const urls = [
    { loc: `${SITE_URL}/`, changefreq: 'weekly', priority: '1' },
    { loc: `${SITE_URL}/explore`, changefreq: 'daily', priority: '0.95' },
  ]

  await writeFile(join(PUBLIC_DIR, 'sitemap.xml'), sitemapXml(urls), 'utf8')
  await writeFile(
    join(PUBLIC_DIR, 'robots.txt'),
    `User-agent: *\nAllow: /\n\nSitemap: ${SITE_URL}/sitemap.xml\n`,
    'utf8',
  )
}

main().catch((error) => {
  console.error('[hub:generate-seo-assets] failed:', error)
  process.exitCode = 1
})
