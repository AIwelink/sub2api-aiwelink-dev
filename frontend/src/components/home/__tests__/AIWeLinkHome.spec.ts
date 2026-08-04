import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { defineComponent, nextTick, onMounted } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { resetHomepageIntroForTests } from '@/composables/useHomepageIntro'
import AIWeLinkHome from '../AIWeLinkHome.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const ParticleStub = defineComponent({
  emits: ['ready'],
  setup(_, { emit }) {
    onMounted(() => emit('ready'))
    return () => null
  },
})

const NavigationStub = defineComponent({
  emits: ['toggle-theme'],
  template: `
    <div>
      <button data-testid="stub-theme" @click="$emit('toggle-theme')">theme</button>
      <a href="#pricing" data-testid="stub-pricing-link">pricing</a>
    </div>
  `,
})

const PricingStub = defineComponent({
  template: '<section id="pricing">pricing target</section>',
})

function mountHome(attachTo?: HTMLElement) {
  return mount(AIWeLinkHome, {
    attachTo,
    props: {
      authenticated: false,
      dashboardPath: '/dashboard',
      dark: true,
      docUrl: '',
    },
    global: {
      stubs: {
        ParticleNetwork: ParticleStub,
        HomepageIntro: { props: ['stage'], template: '<div data-testid="stub-intro" :data-stage="stage" />' },
        HomepageNavigation: NavigationStub,
        HomepageHero: true,
        HomepageUseCases: true,
        HomepagePricing: PricingStub,
        HomepageModels: true,
        HomepageFinalCta: true,
      },
    },
  })
}

describe('AIWeLinkHome', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    resetHomepageIntroForTests()
    vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: false } as MediaQueryList)
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('marks the homepage root when the dark theme is active', () => {
    const wrapper = mountHome()

    expect(wrapper.classes()).toContain('is-dark')
  })

  it('moves from thinking to editor composition, component reveal, and then the ready homepage', async () => {
    const wrapper = mountHome()
    await nextTick()
    await flushPromises()

    expect(wrapper.attributes('data-intro-stage')).toBe('preparing')
    vi.advanceTimersByTime(1200)
    await nextTick()
    expect(wrapper.attributes('data-intro-stage')).toBe('composing')

    vi.advanceTimersByTime(1000)
    await nextTick()
    expect(wrapper.attributes('data-intro-stage')).toBe('revealing')

    vi.advanceTimersByTime(1799)
    await nextTick()
    expect(wrapper.attributes('data-intro-stage')).toBe('revealing')

    vi.advanceTimersByTime(1)
    await nextTick()
    expect(wrapper.attributes('data-intro-stage')).toBe('ready')

    vi.advanceTimersByTime(500)
    await nextTick()
    await flushPromises()
    expect(wrapper.find('[data-testid="stub-intro"]').exists()).toBe(false)
  })

  it('uses a slower blurred entrance for every composing layer during reveal', () => {
    const source = readFileSync(resolve('src/components/home/AIWeLinkHome.vue'), 'utf8')

    expect(source).toContain('animation: compose-layer-in 760ms')
    expect(source).toMatch(
      /@keyframes compose-layer-in\s*{\s*from\s*{\s*opacity:\s*0;\s*transform:\s*translateY\(18px\);\s*filter:\s*blur\(8px\);\s*}/,
    )
  })

  it('does not replay after the homepage remounts in the same document', async () => {
    const first = mountHome()
    first.unmount()
    const second = mountHome()
    await nextTick()

    expect(second.attributes('data-intro-stage')).toBe('ready')
  })

  it('forwards the existing theme toggle action', async () => {
    const wrapper = mountHome()
    await wrapper.get('[data-testid="stub-theme"]').trigger('click')

    expect(wrapper.emitted('toggle-theme')).toHaveLength(1)
  })

  it('scrolls same-page navigation anchors without remounting the homepage', async () => {
    const scrollIntoView = vi.fn()
    const originalHistoryState = window.history.state
    const originalUrl = window.location.href
    const routerState = { position: 7, current: '/home' }
    window.history.replaceState(routerState, '', originalUrl)
    const replaceState = vi.spyOn(window.history, 'replaceState')
    replaceState.mockClear()
    const originalScrollIntoView = Object.getOwnPropertyDescriptor(Element.prototype, 'scrollIntoView')
    const host = document.createElement('div')
    document.body.appendChild(host)
    Object.defineProperty(Element.prototype, 'scrollIntoView', {
      configurable: true,
      value: scrollIntoView,
    })

    let wrapper: ReturnType<typeof mountHome> | undefined
    try {
      wrapper = mountHome(host)
      await wrapper.get('[data-testid="stub-pricing-link"]').trigger('click')

      expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth', block: 'start' })
      expect(replaceState).toHaveBeenCalledWith(routerState, '', '#pricing')
      expect(wrapper.attributes('data-intro-stage')).toBe('preparing')
    } finally {
      wrapper?.unmount()
      host.remove()
      window.history.replaceState(originalHistoryState, '', originalUrl)
      if (originalScrollIntoView) {
        Object.defineProperty(Element.prototype, 'scrollIntoView', originalScrollIntoView)
      } else {
        delete (Element.prototype as Element & { scrollIntoView?: unknown }).scrollIntoView
      }
    }
  })

  it('does not create a reveal observer after unmount while fonts are loading', async () => {
    let resolveFonts!: (value: FontFaceSet) => void
    const fontsReady = new Promise<FontFaceSet>((resolve) => {
      resolveFonts = resolve
    })
    const originalFonts = Object.getOwnPropertyDescriptor(document, 'fonts')
    Object.defineProperty(document, 'fonts', {
      configurable: true,
      value: { ready: fontsReady },
    })

    const observer = vi.fn(() => ({
      observe: vi.fn(),
      unobserve: vi.fn(),
      disconnect: vi.fn(),
    }))
    vi.stubGlobal('IntersectionObserver', observer)

    try {
      const wrapper = mountHome()
      await nextTick()
      wrapper.unmount()
      resolveFonts({} as FontFaceSet)
      await flushPromises()

      expect(observer).not.toHaveBeenCalled()
    } finally {
      if (originalFonts) {
        Object.defineProperty(document, 'fonts', originalFonts)
      } else {
        delete (document as Document & { fonts?: FontFaceSet }).fonts
      }
    }
  })
})
