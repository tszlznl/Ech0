export interface HubInstance {
  id: string
  url: string
}

export interface HubConfig {
  instances: HubInstance[]
}
