const scriptLoadPromises = new Map<string, Promise<void>>()
const styleLoadPromises = new Map<string, Promise<void>>()

const scriptSelector = (src: string) => `script[data-external-script="${src}"],script[src="${src}"]`
const styleSelector = (href: string) => `link[data-external-style="${href}"],link[href="${href}"]`
const DEFAULT_SCRIPT_TIMEOUT_MS = 8_000
const DEFAULT_SCRIPT_RETRIES = 0

type ScriptLoadOptions = {
  timeoutMs?: number
  retries?: number
}

const resolveFromReadyState = (
  node: HTMLScriptElement | HTMLLinkElement,
  url: string,
  isScript: boolean,
  timeoutMs?: number,
) => {
  const stateAttr = isScript ? 'data-external-script-ready' : 'data-external-style-ready'
  if (node.getAttribute(stateAttr) === 'true') {
    return Promise.resolve()
  }
  if (!isScript && (node as HTMLLinkElement).sheet) {
    node.setAttribute(stateAttr, 'true')
    return Promise.resolve()
  }

  return new Promise<void>((resolve, reject) => {
    let timer: ReturnType<typeof setTimeout> | null = null
    const cleanup = () => {
      node.removeEventListener('load', onLoad)
      node.removeEventListener('error', onError)
      if (timer) {
        clearTimeout(timer)
        timer = null
      }
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
    if (timeoutMs && timeoutMs > 0) {
      timer = setTimeout(() => {
        cleanup()
        reject(new Error(`Timed out loading external asset: ${url}`))
      }, timeoutMs)
    }
  })
}

const createExternalScript = (src: string) => {
  const script = document.createElement('script')
  script.src = src
  script.defer = true
  script.setAttribute('data-external-script', src)
  script.removeAttribute('data-external-script-failed')
  document.head.appendChild(script)
  return script
}

export const loadExternalScript = (src: string, options: ScriptLoadOptions = {}): Promise<void> => {
  if (typeof document === 'undefined') return Promise.resolve()
  const cached = scriptLoadPromises.get(src)
  if (cached) return cached

  const timeoutMs = options.timeoutMs ?? DEFAULT_SCRIPT_TIMEOUT_MS
  const retries = Math.max(0, options.retries ?? DEFAULT_SCRIPT_RETRIES)

  const promise = (async () => {
    let latestError: unknown = null

    for (let attempt = 0; attempt <= retries; attempt += 1) {
      let script = document.querySelector<HTMLScriptElement>(scriptSelector(src))

      if (script?.getAttribute('data-external-script-failed') === 'true') {
        script.remove()
        script = null
      }
      if (!script) {
        script = createExternalScript(src)
      }

      try {
        await resolveFromReadyState(script, src, true, timeoutMs)
        return
      } catch (error) {
        latestError = error
        script.setAttribute('data-external-script-failed', 'true')
        script.removeAttribute('data-external-script-ready')
        script.remove()
      }
    }

    throw latestError ?? new Error(`Failed to load external asset: ${src}`)
  })().catch((error) => {
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
