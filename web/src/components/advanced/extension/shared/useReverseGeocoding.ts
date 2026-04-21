import { ref, watch, type Ref } from 'vue'

const NOMINATIM_ENDPOINT = 'https://nominatim.openstreetmap.org/reverse'
const DEBOUNCE_MS = 1000

const cache = new Map<string, string>()

const formatCoord = (lat: number, lng: number) => `${lat.toFixed(4)}, ${lng.toFixed(4)}`

const cacheKey = (lat: number, lng: number) => `${lat.toFixed(6)},${lng.toFixed(6)}`

export function useReverseGeocoding(lat: Ref<number | null>, lng: Ref<number | null>) {
  const displayName = ref<string>('')
  const isFetching = ref<boolean>(false)
  let timer: ReturnType<typeof setTimeout> | null = null
  let abortCtrl: AbortController | null = null

  const resolve = async (la: number, lo: number) => {
    const key = cacheKey(la, lo)
    if (cache.has(key)) {
      displayName.value = cache.get(key) as string
      return
    }

    abortCtrl?.abort()
    abortCtrl = new AbortController()
    isFetching.value = true
    try {
      const url = `${NOMINATIM_ENDPOINT}?lat=${la}&lon=${lo}&format=json`
      const res = await fetch(url, {
        headers: { Accept: 'application/json' },
        signal: abortCtrl.signal,
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = (await res.json()) as { display_name?: string }
      const name = data?.display_name?.trim() || formatCoord(la, lo)
      cache.set(key, name)
      displayName.value = name
    } catch (err) {
      if ((err as Error).name !== 'AbortError') {
        displayName.value = formatCoord(la, lo)
      }
    } finally {
      isFetching.value = false
    }
  }

  watch(
    [lat, lng],
    ([la, lo]) => {
      if (timer) clearTimeout(timer)
      if (la === null || lo === null || Number.isNaN(la) || Number.isNaN(lo)) {
        displayName.value = ''
        return
      }
      timer = setTimeout(() => {
        void resolve(la, lo)
      }, DEBOUNCE_MS)
    },
    { immediate: false },
  )

  return { displayName, isFetching }
}
