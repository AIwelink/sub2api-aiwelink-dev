import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const home = readFileSync(resolve(process.cwd(), 'src/views/HomeView.vue'), 'utf8')
const auth = readFileSync(resolve(process.cwd(), 'src/components/layout/AuthLayout.vue'), 'utf8')

describe('public theme surfaces', () => {
  it('uses semantic depth tokens for both homepage variants', () => {
    expect(home).toContain('bg-canvas text-theme-text')
    expect(home).toContain('from-canvas via-surface-muted to-canvas')
    expect(home).toContain('border-theme-border')
    expect(home).toContain('bg-surface/80')
    expect(home).not.toMatch(/dark:(?:from|via|to)-dark-(?:900|950)/)
  })

  it('keeps authentication on the same semantic background system', () => {
    expect(auth).toContain('from-canvas via-surface-muted to-canvas')
  })
})
