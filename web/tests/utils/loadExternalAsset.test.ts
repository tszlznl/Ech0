import { describe, expect, it } from 'vitest'

import { loadExternalScript, loadExternalStyle } from '../../src/utils/loadExternalAsset'

const scriptSelector = (src: string) => `script[data-external-script="${src}"]`
const styleSelector = (href: string) => `link[data-external-style="${href}"]`

describe('loadExternalAsset', () => {
  it('deduplicates concurrent script loads and reuses loaded script', async () => {
    const src = `/others/scripts/APlayer.min.js?test=${Date.now()}`

    const promiseA = loadExternalScript(src)
    const promiseB = loadExternalScript(src)

    expect(promiseA).toBe(promiseB)
    expect(document.querySelectorAll(scriptSelector(src))).toHaveLength(1)

    const script = document.querySelector<HTMLScriptElement>(scriptSelector(src))
    expect(script).toBeTruthy()
    script?.dispatchEvent(new Event('load'))

    await expect(promiseA).resolves.toBeUndefined()
    await expect(loadExternalScript(src)).resolves.toBeUndefined()
    expect(document.querySelectorAll(scriptSelector(src))).toHaveLength(1)
  })

  it('deduplicates style loads and reuses loaded stylesheet', async () => {
    const href = `/others/styles/APlayer.min.css?test=${Date.now()}`

    const promiseA = loadExternalStyle(href)
    const promiseB = loadExternalStyle(href)

    expect(promiseA).toBe(promiseB)
    expect(document.querySelectorAll(styleSelector(href))).toHaveLength(1)

    const link = document.querySelector<HTMLLinkElement>(styleSelector(href))
    expect(link).toBeTruthy()
    link?.dispatchEvent(new Event('load'))

    await expect(promiseA).resolves.toBeUndefined()
    await expect(loadExternalStyle(href)).resolves.toBeUndefined()
    expect(document.querySelectorAll(styleSelector(href))).toHaveLength(1)
  })
})
