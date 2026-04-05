import { fetchHealthz, meetsHubMinVersion, getMinSupportedVersion } from './healthz'
import type { HubInstance } from '../types/hub'

export interface ProbeFailure {
  instance: HubInstance
  reason: string
}

export interface ProbeResult {
  eligible: HubInstance[]
  probeFailures: ProbeFailure[]
}

/**
 * 并行探活：必须 healthz 成功且 version ≥ 4.4.0 才参与聚合。
 */
export async function probeInstances(
  instances: HubInstance[],
  signal?: AbortSignal,
): Promise<ProbeResult> {
  const probeFailures: ProbeFailure[] = []
  const eligible: HubInstance[] = []

  const settled = await Promise.allSettled(
    instances.map(async (inst) => {
      const h = await fetchHealthz(inst.url, signal)
      if (!h.ok) {
        return { kind: 'fail' as const, instance: inst, reason: h.message }
      }
      if (!meetsHubMinVersion(h.version)) {
        return {
          kind: 'fail' as const,
          instance: inst,
          reason: `Version ${h.version} is below the Hub minimum (${getMinSupportedVersion()})`,
        }
      }
      return { kind: 'ok' as const, instance: inst }
    }),
  )

  for (let i = 0; i < settled.length; i++) {
    const r = settled[i]!
    if (r.status === 'rejected') {
      const inst = instances[i]!
      probeFailures.push({
        instance: inst,
        reason: r.reason instanceof Error ? r.reason.message : String(r.reason),
      })
      continue
    }
    const val = r.value
    if (val.kind === 'fail') {
      probeFailures.push({ instance: val.instance, reason: val.reason })
    } else {
      eligible.push(val.instance)
    }
  }

  return { eligible, probeFailures }
}
