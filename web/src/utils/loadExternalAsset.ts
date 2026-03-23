const scriptLoadPromises = new Map<string, Promise<void>>()
const styleLoadPromises = new Map<string, Promise<void>>()

const scriptSelector = (src: string) => `script[data-external-script="${src}"],script[src="${src}"]`
const styleSelector = (href: string) => `link[data-external-style="${href}"],link[href="${href}"]`

const resolveFromReadyState = (node: HTMLScriptElement | HTMLLinkElement, url: string, isScript: boolean) => {
  const stateAttr = isScript ? 'data-external-script-ready' : 'data-external-style-ready'
  if (node.getAttribute(stateAttr) === 'true') {
    return Promise.resolve()
  }
  if (!isScript && (node as HTMLLinkElement).sheet) {
    node.setAttribute(stateAttr, 'true')
    return Promise.resolve()
  }

  return new Promise<void>((resolve, reject) => {
    const cleanup = () => {
      node.removeEventListener('load', onLoad)
      node.removeEventListener('error', onError)
    }

    const onLoad = () => {
      node.setAttribute(stateAttr, 'true')
      cleanup()
      resolve()
    }

    const onError = () => {
      cleanup()
      reject(new Error(`Failed to load external asset: ${url}`))
    }

    node.addEventListener('load', onLoad, { once: true })
    node.addEventListener('error', onError, { once: true })
  })
}

export const loadExternalScript = (src: string): Promise<void> => {
  if (typeof document === 'undefined') return Promise.resolve()
  const cached = scriptLoadPromises.get(src)
  if (cached) return cached

  let script = document.querySelector<HTMLScriptElement>(scriptSelector(src))
  if (!script) {
    script = document.createElement('script')
    script.src = src
    script.defer = true
    script.setAttribute('data-external-script', src)
    document.head.appendChild(script)
  }

  const promise = resolveFromReadyState(script, src, true).catch((error) => {
    scriptLoadPromises.delete(src)
    throw error
  })
  scriptLoadPromises.set(src, promise)
  return promise
}

export const loadExternalStyle = (href: string): Promise<void> => {
  if (typeof document === 'undefined') return Promise.resolve()
  const cached = styleLoadPromises.get(href)
  if (cached) return cached

  let link = document.querySelector<HTMLLinkElement>(styleSelector(href))
  if (!link) {
    link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = href
    link.setAttribute('data-external-style', href)
    document.head.appendChild(link)
  }

  const promise = resolveFromReadyState(link, href, false).catch((error) => {
    styleLoadPromises.delete(href)
    throw error
  })
  styleLoadPromises.set(href, promise)
  return promise
}
