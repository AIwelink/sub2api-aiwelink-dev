import { defineComponent } from 'vue'
import { mount, RouterLinkStub } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it, vi } from 'vitest'

import HomepageFinalCta from '../HomepageFinalCta.vue'
import HomepageHero from '../HomepageHero.vue'
import HomepageModels from '../HomepageModels.vue'
import HomepageNavigation from '../HomepageNavigation.vue'
import HomepagePricing from '../HomepagePricing.vue'
import HomepageUseCases from '../HomepageUseCases.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  const messages = (await import('@/i18n/locales/zh/landing')).default as Record<string, unknown>

  function translate(key: string) {
    const value = key.split('.').reduce<unknown>((current, segment) => {
      if (!current || typeof current !== 'object') return undefined
      return (current as Record<string, unknown>)[segment]
    }, messages)
    return typeof value === 'string' ? value : key
  }

  return { ...actual, useI18n: () => ({ t: translate }) }
})

const ContentHarness = defineComponent({
  components: {
    HomepageHero,
    HomepageUseCases,
    HomepagePricing,
    HomepageModels,
    HomepageFinalCta,
  },
  props: {
    authenticated: Boolean,
    dashboardPath: { type: String, default: '/dashboard' },
  },
  template: `
    <main>
      <HomepageHero :authenticated="authenticated" :dashboard-path="dashboardPath" />
      <HomepageUseCases />
      <HomepagePricing />
      <HomepageModels />
      <HomepageFinalCta :authenticated="authenticated" :dashboard-path="dashboardPath" />
    </main>
  `,
})

function mountContent(authenticated = false, dashboardPath = '/dashboard') {
  return mount(ContentHarness, {
    props: { authenticated, dashboardPath },
    global: {
      stubs: { RouterLink: RouterLinkStub, Icon: true },
    },
  })
}

describe('AIWeLink homepage content', () => {
  it('keeps warm and blue accents out of content components', () => {
    const componentSources = [
      'HomepageFinalCta.vue',
      'HomepageHero.vue',
      'HomepageModels.vue',
      'HomepageNavigation.vue',
      'HomepagePricing.vue',
      'HomepageUseCases.vue',
    ].map((file) => readFileSync(resolve('src/components/home', file), 'utf8'))

    componentSources.forEach((source) => {
      expect(source).not.toContain('#ef3f72')
      expect(source).not.toContain('#2563eb')
      expect(source).not.toContain('37, 99, 235')
    })
  })

  it('renders the approved unframed content without a model count or carousel', () => {
    const wrapper = mountContent()
    const text = wrapper.text()

    expect(wrapper.get('h1').text().replace(/\s+/g, ' ').trim()).toBe('AIwelink API')
    expect(text).toContain('Codex')
    expect(text).toContain('Claude Code')
    expect(text).toContain('科研与深度学习')
    expect(text).toContain('Agent 开发接入')
    expect(text).toContain('¥1 = $10')
    expect(text).toContain('¥0.1–0.2 / $1')
    expect(text).toContain('GPT')
    expect(text).toContain('Claude')
    expect(text).toContain('Gemini')
    expect(text).not.toContain('三种模型')
    expect(wrapper.find('[data-testid="model-carousel"]').exists()).toBe(false)
    expect(wrapper.find('.card').exists()).toBe(false)
  })

  it('uses registration for guests and the supplied dashboard for authenticated users', () => {
    const guestLink = mountContent().get('[data-testid="hero-primary"]')
    const memberLink = mountContent(true, '/admin/dashboard').get('[data-testid="hero-primary"]')

    expect(guestLink.findComponent(RouterLinkStub).props('to')).toBe('/register')
    expect(memberLink.findComponent(RouterLinkStub).props('to')).toBe('/admin/dashboard')
  })

  it('uses a directional animated track for the hero scroll cue', () => {
    const wrapper = mountContent()
    const cue = wrapper.get('.scroll-cue')

    expect(cue.get('.scroll-track').exists()).toBe(true)
    expect(cue.get('.scroll-runner').exists()).toBe(true)
    expect(cue.get('.scroll-chevron').exists()).toBe(true)
  })

  it('keeps the wordmark fixed and emits the existing theme action', async () => {
    const wrapper = mount(HomepageNavigation, {
      props: {
        authenticated: false,
        dashboardPath: '/dashboard',
        dark: true,
        docUrl: '',
      },
      global: {
        stubs: { RouterLink: RouterLinkStub, LocaleSwitcher: true, Icon: true },
      },
    })

    expect(wrapper.get('[data-testid="home-wordmark"]').text()).toBe('AIwelink')
    await wrapper.get('[data-testid="home-theme-toggle"]').trigger('click')
    expect(wrapper.emitted('toggle-theme')).toHaveLength(1)
  })
})
