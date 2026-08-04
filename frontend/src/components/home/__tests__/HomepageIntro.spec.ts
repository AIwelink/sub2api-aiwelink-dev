import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import HomepageIntro from '../HomepageIntro.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key === 'home.redesign.intro.thinking' ? 'AI 思考中' : key,
  }),
}))

describe('HomepageIntro', () => {
  it('shows the fixed brand and accessible thinking status while preparing', () => {
    const wrapper = mount(HomepageIntro, { props: { stage: 'preparing' } })

    expect(wrapper.get('[data-testid="intro-brand"]').attributes('aria-label')).toBe('AIWELINK API')
    expect(wrapper.get('[role="status"]').text()).toContain('AI 思考中')
    expect(wrapper.get('[data-testid="intro-spinner"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="intro-progress"]').attributes('aria-hidden')).toBe('true')
  })

  it('shows concise decorative component output while composing', () => {
    const wrapper = mount(HomepageIntro, { props: { stage: 'composing' } })

    const editor = wrapper.get('[data-testid="intro-editor"]')
    expect(editor.attributes('aria-hidden')).toBe('true')
    expect(editor.text()).toContain('<Navigation />')
    expect(editor.text()).toContain('<Hero />')
    expect(editor.text()).toContain('<UseCases />')
    expect(editor.text()).toContain('<Pricing />')
    expect(editor.text()).toContain('<Models />')
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
  })

  it('removes the editor and fades the black intro without a directional curtain', () => {
    const wrapper = mount(HomepageIntro, { props: { stage: 'revealing' } })

    expect(wrapper.get('[data-testid="intro-fade"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="intro-curtain"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="intro-editor"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('<Navigation />')
  })
})
