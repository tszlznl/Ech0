<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import HubActiveCreators from '../components/HubActiveCreators.vue'
import HubHomeFoliageCanvas from '../components/HubHomeFoliageCanvas.vue'
import { fetchInstancesConnectBundle, type InstanceConnectSummary } from '../services/connectApi'
import { loadHubConfig } from '../services/loadHubConfig'
import { probeInstances } from '../services/probeInstances'
import { getHubSubmitIssueUrl } from '../utils/hubLinks'
import '../styles/hub-shell.css'

const connectSummaries = ref<InstanceConnectSummary[]>([])

onMounted(async () => {
  try {
    const instances = await loadHubConfig()
    if (instances.length === 0) return
    const probed = await probeInstances(instances)
    if (probed.eligible.length === 0) return
    const { summaries } = await fetchInstancesConnectBundle(probed.eligible)
    connectSummaries.value = summaries
  } catch {
    /* home: omit noisy errors; explore shows full diagnostics */
  }
})
</script>

<template>
  <div class="hub-shell min-h-screen w-full">
    <main class="mx-auto w-full max-w-[min(100%,34rem)] px-5 pb-28 pt-16 sm:pt-24">
      <section
        class="hub-home-hero flex flex-col items-center gap-9 text-center sm:gap-11"
      >
        <h1
          class="hub-home-title mx-auto max-w-[min(100%,26rem)] text-center font-normal leading-[1.45] tracking-[0.02em] text-[var(--color-text-primary)] text-[clamp(0.8125rem,2.65vw,1.1875rem)]"
        >
          Ech0 Hub — where echoes meet and ideas resonate.
        </h1>
        <HubHomeFoliageCanvas class="w-full" />
        <p
          class="hub-home-lede mx-auto max-w-[34ch] text-[0.8125rem] font-normal leading-[1.75] text-[var(--color-text-secondary)] sm:text-[0.84375rem]"
        >
          Discover voices from across the web, softly converging into a single stream of resonance.
        </p>
        <p
          class="hub-home-caption -mt-1 max-w-[30ch] text-[0.75rem] leading-[1.65] text-[var(--color-text-muted)]"
        >
          <em>Many corners, one echoing space.</em>
        </p>
        <div class="flex flex-col items-center gap-4 pt-1">
          <RouterLink
            :to="{ name: 'explore' }"
            class="inline-flex items-center justify-center rounded-full bg-[var(--color-accent)] px-6 py-2.5 text-xs font-medium tracking-[0.06em] text-[var(--color-bg-canvas)] no-underline shadow-sm transition hover:opacity-90"
          >
            Explore the feed
          </RouterLink>
          <a
            :href="getHubSubmitIssueUrl()"
            target="_blank"
            rel="noopener noreferrer"
            class="text-xs tracking-[0.04em] text-[var(--color-text-muted)] underline-offset-4 transition hover:text-[var(--color-text-primary)]"
          >
            Join Hub
          </a>
        </div>
      </section>

      <HubActiveCreators v-if="connectSummaries.length > 0" :summaries="connectSummaries" />
    </main>
  </div>
</template>
