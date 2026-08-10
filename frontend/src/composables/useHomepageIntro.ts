import { getCurrentScope, onScopeDispose, ref, type Ref } from 'vue'

export type HomepageIntroStage = 'preparing' | 'composing' | 'revealing' | 'ready'

interface HomepageIntroOptions {
  reducedMotion: boolean
}

interface HomepageIntroController {
  stage: Ref<HomepageIntroStage>
  markContentReady: () => void
  skip: () => void
  dispose: () => void
}

const PREPARING_DURATION_MS = 1200
const REDUCED_MOTION_DURATION_MS = 180
const COMPOSING_DURATION_MS = 1000
const REVEALING_DURATION_MS = 1800
const DEFENSIVE_TIMEOUT_MS = 5500

let introPlayedForDocument = false

export function useHomepageIntro(options: HomepageIntroOptions): HomepageIntroController {
  const shouldPlay = !introPlayedForDocument
  introPlayedForDocument = true

  const stage = ref<HomepageIntroStage>(shouldPlay ? 'preparing' : 'ready')
  const timers = new Set<ReturnType<typeof setTimeout>>()
  const root = typeof document === 'undefined' ? null : document.documentElement
  const previousOverflow = root?.style.overflow ?? ''
  const previousScrollbarGutter = root?.style.scrollbarGutter ?? ''

  let contentReady = false
  let minimumElapsed = false
  let disposed = false

  function schedule(callback: () => void, delay: number) {
    const timer = setTimeout(() => {
      timers.delete(timer)
      callback()
    }, delay)
    timers.add(timer)
  }

  function clearTimers() {
    timers.forEach((timer) => clearTimeout(timer))
    timers.clear()
  }

  function restorePageLock() {
    if (!root) return
    root.style.overflow = previousOverflow
    root.style.scrollbarGutter = previousScrollbarGutter
  }

  function finish() {
    if (disposed || stage.value === 'ready') return
    clearTimers()
    stage.value = 'ready'
    restorePageLock()
  }

  function beginReveal() {
    if (disposed || stage.value !== 'composing') return
    stage.value = 'revealing'
    schedule(finish, REVEALING_DURATION_MS)
  }

  function advanceWhenReady() {
    if (disposed || stage.value !== 'preparing' || !contentReady || !minimumElapsed) return

    if (options.reducedMotion) {
      finish()
      return
    }

    stage.value = 'composing'
    schedule(beginReveal, COMPOSING_DURATION_MS)
  }

  function markContentReady() {
    contentReady = true
    advanceWhenReady()
  }

  function dispose() {
    if (disposed) return
    disposed = true
    clearTimers()
    restorePageLock()
  }

  if (shouldPlay) {
    if (root) {
      root.style.scrollbarGutter = 'stable'
      root.style.overflow = 'hidden'
    }

    schedule(() => {
      minimumElapsed = true
      advanceWhenReady()
    }, options.reducedMotion ? REDUCED_MOTION_DURATION_MS : PREPARING_DURATION_MS)

    schedule(finish, DEFENSIVE_TIMEOUT_MS)
  }

  if (getCurrentScope()) onScopeDispose(dispose)

  return {
    stage,
    markContentReady,
    skip: finish,
    dispose,
  }
}

export function resetHomepageIntroForTests() {
  introPlayedForDocument = false
}
