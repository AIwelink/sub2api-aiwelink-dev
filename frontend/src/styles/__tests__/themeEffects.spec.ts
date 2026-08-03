import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const shared = readFileSync(resolve(process.cwd(), 'src/style.css'), 'utf8')
const settings = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')
const channels = readFileSync(resolve(process.cwd(), 'src/views/admin/ChannelsView.vue'), 'utf8')

describe('AIWeLink theme effects', () => {
  it('adds a restrained sheen to primary buttons only', () => {
    expect(shared).toContain('.btn-primary::before')
    expect(shared).toContain('@keyframes primary-button-sheen')
    expect(shared).toContain('.btn-primary:hover:not(:disabled)')
    expect(shared).not.toMatch(/\.btn-(?:secondary|danger)::before/)
  })

  it('removes primary motion and glow from disabled buttons', () => {
    expect(shared).toContain('.btn-primary:disabled')
    expect(shared).toContain('.btn-primary:disabled::before')
    expect(shared).toContain('display: none')
  })

  it('honors reduced-motion preferences', () => {
    expect(shared).toContain('@media (prefers-reduced-motion: reduce)')
    expect(shared).toMatch(/\.btn-primary::before\s*\{\s*animation: none/)
    expect(shared).toMatch(/\.btn-primary:hover:not\(:disabled\)\s*\{\s*transform: none/)
  })

  it('uses static localized glow on every selected navigation family', () => {
    expect(shared).toContain('.sidebar-link-active')
    expect(shared).toContain('.tab-active')
    expect(shared).toContain('rgb(var(--color-primary-500) / 0.14)')
    expect(settings).toContain('rgb(var(--color-primary-500) / 0.14)')
    expect(channels).toContain('rgb(var(--color-primary-500) / 0.14)')
  })
})
