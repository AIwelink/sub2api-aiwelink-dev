import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import DingTalkOAuthSection from '@/components/auth/DingTalkOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>
}))

const locationState = vi.hoisted(() => ({
  current: { href: 'http://localhost/register' } as { href: string }
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('OAuth login sections', () => {
  beforeEach(() => {
    routeState.query = { redirect: '/billing?plan=pro', aff: 'AFF123' }
    locationState.current = { href: 'http://localhost/register' }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState.current
    })
    window.sessionStorage.clear()
  })

  it.each([
    ['linuxdo', LinuxDoOAuthSection],
    ['dingtalk', DingTalkOAuthSection],
    ['oidc', OidcOAuthSection]
  ] as const)('passes affiliate data to the %s start endpoint without storing it', async (provider, component) => {
    const wrapper = mount(component, { props: { affCode: 'AFF456' } })

    await wrapper.get('button').trigger('click')

    expect(locationState.current.href).toBe(
      `/api/v1/auth/oauth/${provider}/start?redirect=%2Fbilling%3Fplan%3Dpro&aff_code=AFF456`
    )
    expect(window.sessionStorage.getItem('oauth_aff_code')).toBeNull()
  })
})
