import { CommentProvider } from '@/enums/enums'
import type { CommentProviderAdapter, CommentProviderFactory } from './types'

const factories: Record<string, CommentProviderFactory> = {
  [CommentProvider.TWIKOO]: async () => (await import('./twikoo')).createTwikooAdapter(),
  [CommentProvider.WALINE]: async () => (await import('./waline')).createWalineAdapter(),
  [CommentProvider.ARTALK]: async () => (await import('./artalk')).createArtalkAdapter(),
  [CommentProvider.GISCUS]: async () => (await import('./giscus')).createGiscusAdapter(),
}

export async function createCommentAdapter(provider: string): Promise<CommentProviderAdapter | null> {
  const factory = factories[provider]
  if (!factory) return null
  return factory()
}
