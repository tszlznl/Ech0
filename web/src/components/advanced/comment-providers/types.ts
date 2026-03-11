export interface CommentProviderAdapter {
  mount: (el: HTMLElement, setting: App.Api.Setting.CommentSetting) => Promise<void> | void
  update?: (setting: App.Api.Setting.CommentSetting) => Promise<void> | void
  unmount?: () => Promise<void> | void
}

export type CommentProviderFactory = () => Promise<CommentProviderAdapter>
