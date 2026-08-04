<template>
  <div
    data-testid="aiwelink-home"
    class="aiwelink-home"
    :class="{ 'is-dark': dark }"
    :data-intro-stage="stage"
    @click="handleAnchorClick"
  >
    <ParticleNetwork
      class="compose-layer network-layer"
      :interactive="stage === 'ready'"
      @ready="markCanvasReady"
    />

    <div class="homepage-content">
      <HomepageNavigation
        class="compose-layer navigation-layer"
        :authenticated="authenticated"
        :dashboard-path="dashboardPath"
        :dark="dark"
        :doc-url="docUrl"
        @toggle-theme="$emit('toggle-theme')"
      />

      <main>
        <HomepageHero
          class="compose-layer hero-layer"
          :authenticated="authenticated"
          :dashboard-path="dashboardPath"
        />
        <HomepageUseCases class="compose-layer use-cases-layer" />
        <HomepagePricing class="compose-layer pricing-layer" />
        <HomepageModels class="compose-layer models-layer" />
        <HomepageFinalCta
          class="compose-layer cta-layer"
          :authenticated="authenticated"
          :dashboard-path="dashboardPath"
        />
      </main>
    </div>

    <Transition name="intro-exit">
      <HomepageIntro v-if="stage !== 'ready'" :stage="stage" />
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted } from 'vue'

import { useHomepageIntro } from '@/composables/useHomepageIntro'
import HomepageFinalCta from './HomepageFinalCta.vue'
import HomepageHero from './HomepageHero.vue'
import HomepageIntro from './HomepageIntro.vue'
import HomepageModels from './HomepageModels.vue'
import HomepageNavigation from './HomepageNavigation.vue'
import HomepagePricing from './HomepagePricing.vue'
import HomepageUseCases from './HomepageUseCases.vue'
import ParticleNetwork from './ParticleNetwork.vue'

defineProps<{
  authenticated: boolean
  dashboardPath: string
  dark: boolean
  docUrl: string
}>()

defineEmits<{
  'toggle-theme': []
}>()

const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
const { stage, markContentReady } = useHomepageIntro({ reducedMotion })

let canvasReady = false
let documentReady = false
let readinessReported = false
let unmounted = false
let revealObserver: IntersectionObserver | null = null

function reportReadiness() {
  if (readinessReported || !canvasReady || !documentReady) return
  readinessReported = true
  markContentReady()
}

function markCanvasReady() {
  canvasReady = true
  reportReadiness()
}

function handleAnchorClick(event: MouseEvent) {
  if (
    event.defaultPrevented
    || event.button !== 0
    || event.metaKey
    || event.ctrlKey
    || event.shiftKey
    || event.altKey
    || !(event.target instanceof Element)
  ) return

  const anchor = event.target.closest<HTMLAnchorElement>('a[href^="#"]')
  const href = anchor?.getAttribute('href')
  if (!href || href.length === 1) return

  const target = document.getElementById(href.slice(1))
  if (!target) return

  event.preventDefault()
  target.scrollIntoView({ behavior: reducedMotion ? 'auto' : 'smooth', block: 'start' })
  window.history.replaceState(window.history.state, '', href)
}

async function waitForFonts() {
  if (!document.fonts?.ready) return

  await Promise.race([
    document.fonts.ready,
    new Promise<void>((resolve) => window.setTimeout(resolve, 450)),
  ])
}

function setupRevealObserver() {
  const elements = Array.from(document.querySelectorAll<HTMLElement>('[data-reveal]'))

  if (reducedMotion || typeof IntersectionObserver !== 'function') {
    elements.forEach((element) => element.classList.add('is-visible'))
    return
  }

  revealObserver = new IntersectionObserver((entries, observer) => {
    entries.forEach((entry) => {
      if (!entry.isIntersecting) return
      entry.target.classList.add('is-visible')
      observer.unobserve(entry.target)
    })
  }, { threshold: .12 })

  elements.forEach((element) => revealObserver?.observe(element))
}

onMounted(async () => {
  await nextTick()
  await waitForFonts()
  if (unmounted) return
  documentReady = true
  setupRevealObserver()
  reportReadiness()
})

onBeforeUnmount(() => {
  unmounted = true
  revealObserver?.disconnect()
})
</script>

<style scoped>
.aiwelink-home {
  --home-canvas: #f7f7f7;
  --home-canvas-layer: linear-gradient(145deg, #fff 0%, #f7f7f7 52%, #ededed 100%);
  --home-text: #101010;
  --home-muted: #5f5f5f;
  --home-faint: #858585;
  --home-soft: rgba(0, 0, 0, .045);
  --home-hover: rgba(0, 0, 0, .08);
  --home-primary: #111111;
  --home-primary-rgb: 17, 17, 17;
  --home-primary-contrast: #fff;
  --home-accent: #242424;

  position: relative;
  min-height: 100vh;
  overflow-x: clip;
  color: var(--home-text);
  background: var(--home-canvas-layer);
  font-family: "Aptos", "Segoe UI Variable", "Microsoft YaHei", system-ui, sans-serif;
  letter-spacing: 0;
  transition: color 240ms ease, background 240ms ease;
}

.aiwelink-home.is-dark {
  --home-canvas: #050608;
  --home-canvas-layer: #050608;
  --home-text: #f6f3eb;
  --home-muted: #a5aab3;
  --home-faint: #6f7681;
  --home-soft: rgba(255, 255, 255, .055);
  --home-hover: rgba(255, 255, 255, .09);
  --home-primary: #ffc648;
  --home-primary-rgb: 255, 198, 72;
  --home-primary-contrast: #181005;
  --home-accent: #ef3f72;
}

.homepage-content {
  position: relative;
  z-index: 1;
  min-height: 100vh;
}

.compose-layer {
  --compose-delay: 0ms;
}

.navigation-layer { --compose-delay: 40ms; }
.network-layer { --compose-delay: 160ms; }
.hero-layer { --compose-delay: 300ms; }
.use-cases-layer { --compose-delay: 520ms; }
.pricing-layer { --compose-delay: 680ms; }
.models-layer { --compose-delay: 840ms; }
.cta-layer { --compose-delay: 980ms; }

[data-intro-stage="preparing"] .compose-layer,
[data-intro-stage="composing"] .compose-layer {
  opacity: 0;
  filter: blur(8px);
  transform: translateY(18px);
}

[data-intro-stage="revealing"] .compose-layer {
  animation: compose-layer-in 760ms cubic-bezier(.16, 1, .3, 1) both;
  animation-delay: var(--compose-delay);
}

[data-intro-stage="ready"] .compose-layer {
  opacity: 1;
  filter: none;
  transform: none;
}

:deep([data-reveal]) {
  opacity: 0;
  transform: translateY(22px);
  transition: opacity 560ms ease, transform 560ms ease;
}

:deep([data-reveal].is-visible) {
  opacity: 1;
  transform: translateY(0);
}

:deep([id]) { scroll-margin-top: 72px; }

.intro-exit-leave-active { transition: opacity 460ms ease; }
.intro-exit-leave-to { opacity: 0; }

@keyframes compose-layer-in {
  from { opacity: 0; transform: translateY(18px); filter: blur(8px); }
  to { opacity: 1; transform: translateY(0); filter: blur(0); }
}

@media (prefers-reduced-motion: reduce) {
  .aiwelink-home { transition: none; }
  .compose-layer,
  :deep([data-reveal]) {
    opacity: 1 !important;
    filter: none !important;
    transform: none !important;
    animation: none !important;
    transition: none !important;
  }

  .intro-exit-leave-active { transition-duration: 120ms; }
}
</style>
