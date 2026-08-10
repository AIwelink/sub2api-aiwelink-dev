import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { i18n } from '@/i18n'
import enMessages from '@/i18n/locales/en'
import VersionBadge from '@/components/common/VersionBadge.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: { version?: string }) =>
        key === 'version.basedOnSub2API'
          ? `Based on Sub2API v${params?.version ?? ''}`
          : key,
    }),
  }
})

const appStore = {
  versionLoading: false,
  currentVersion: '0.1.170-2.4',
  upstreamVersion: '0.1.170',
  latestVersion: '0.1.170-2.4',
  hasUpdate: false,
  buildType: 'release',
  releaseInfo: null,
  fetchVersion: vi.fn(),
  clearVersionCache: vi.fn(),
}

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ isAdmin: true }),
  useAppStore: () => appStore,
}))

vi.mock('@/api/admin/system', () => ({
  performUpdate: vi.fn(),
  restartService: vi.fn(),
  getRollbackVersions: vi.fn(),
  rollback: vi.fn(),
}))

describe('VersionBadge', () => {
  beforeEach(() => {
    appStore.fetchVersion.mockReset()
    i18n.global.setLocaleMessage('en', enMessages)
    i18n.global.locale.value = 'en'
  })

  it('shows the AIWeLink version and upstream Sub2API baseline', async () => {
    const wrapper = mount(VersionBadge, {
      global: {
        plugins: [createPinia(), i18n],
        stubs: { Icon: true },
      },
    })

    expect(wrapper.get('button').text()).toContain('AIWeLink v0.1.170-2.4')
    await wrapper.get('button').trigger('click')
    expect(wrapper.text()).toContain('Based on Sub2API v0.1.170')
  })
})
