import { nextTick } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import {
  resetHomepageIntroForTests,
  useHomepageIntro,
} from '../useHomepageIntro'

describe('useHomepageIntro', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    resetHomepageIntroForTests()
    document.documentElement.style.overflow = ''
    document.documentElement.style.scrollbarGutter = ''
  })

  afterEach(() => {
    vi.useRealTimers()
    document.documentElement.style.overflow = ''
    document.documentElement.style.scrollbarGutter = ''
  })

  it('waits for both the minimum duration and content readiness', async () => {
    const intro = useHomepageIntro({ reducedMotion: false })

    intro.markContentReady()
    vi.advanceTimersByTime(1199)
    await nextTick()
    expect(intro.stage.value).toBe('preparing')

    vi.advanceTimersByTime(1)
    await nextTick()
    expect(intro.stage.value).toBe('composing')

    vi.advanceTimersByTime(999)
    await nextTick()
    expect(intro.stage.value).toBe('composing')

    vi.advanceTimersByTime(1)
    await nextTick()
    expect(intro.stage.value).toBe('revealing')

    vi.advanceTimersByTime(1799)
    await nextTick()
    expect(intro.stage.value).toBe('revealing')

    vi.advanceTimersByTime(1)
    await nextTick()
    expect(intro.stage.value).toBe('ready')
  })

  it('stays in the thinking state when the page is not ready', async () => {
    const intro = useHomepageIntro({ reducedMotion: false })

    vi.advanceTimersByTime(1500)
    await nextTick()
    expect(intro.stage.value).toBe('preparing')

    intro.markContentReady()
    await nextTick()
    expect(intro.stage.value).toBe('composing')
  })

  it('skips playback on a second mount in the same document', () => {
    const first = useHomepageIntro({ reducedMotion: false })
    const second = useHomepageIntro({ reducedMotion: false })

    expect(first.stage.value).toBe('preparing')
    expect(second.stage.value).toBe('ready')
  })

  it('uses a short static intro when reduced motion is requested', async () => {
    const intro = useHomepageIntro({ reducedMotion: true })

    intro.markContentReady()
    vi.advanceTimersByTime(179)
    await nextTick()
    expect(intro.stage.value).toBe('preparing')

    vi.advanceTimersByTime(1)
    await nextTick()
    expect(intro.stage.value).toBe('ready')
  })

  it('releases the page after the defensive timeout', async () => {
    const intro = useHomepageIntro({ reducedMotion: false })

    vi.advanceTimersByTime(5499)
    await nextTick()
    expect(intro.stage.value).toBe('preparing')

    vi.advanceTimersByTime(1)
    await nextTick()
    expect(intro.stage.value).toBe('ready')
  })

  it('restores the previous overflow style on completion and disposal', async () => {
    document.documentElement.style.overflow = 'clip'
    document.documentElement.style.scrollbarGutter = 'auto'
    const intro = useHomepageIntro({ reducedMotion: false })

    expect(document.documentElement.style.overflow).toBe('hidden')
    expect(document.documentElement.style.scrollbarGutter).toBe('stable')
    intro.skip()
    await nextTick()
    expect(document.documentElement.style.overflow).toBe('clip')
    expect(document.documentElement.style.scrollbarGutter).toBe('auto')

    resetHomepageIntroForTests()
    const second = useHomepageIntro({ reducedMotion: false })
    expect(document.documentElement.style.overflow).toBe('hidden')
    expect(document.documentElement.style.scrollbarGutter).toBe('stable')
    second.dispose()
    expect(document.documentElement.style.overflow).toBe('clip')
    expect(document.documentElement.style.scrollbarGutter).toBe('auto')
  })
})
