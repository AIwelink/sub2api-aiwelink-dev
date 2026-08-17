import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const home = readFileSync(resolve(process.cwd(), 'src/views/HomeView.vue'), 'utf8')
const aiwelinkHome = readFileSync(resolve(process.cwd(), 'src/components/home/AIWeLinkHome.vue'), 'utf8')
const auth = readFileSync(resolve(process.cwd(), 'src/components/layout/AuthLayout.vue'), 'utf8')

describe('public theme surfaces', () => {
  it('keeps compact and branded home variants on their intended theme boundaries', () => {
    expect(home).toContain('bg-canvas text-theme-text')
    expect(home).toContain('border-theme-border')
    expect(home).toContain('hover:bg-surface-muted')
    expect(home).not.toMatch(/dark:(?:from|via|to)-dark-(?:900|950)/)

    expect(aiwelinkHome).toContain('class="aiwelink-home"')
    expect(aiwelinkHome).toContain(":class=\"{ 'is-dark': dark }\"")
    expect(aiwelinkHome).toContain('--home-canvas-layer: linear-gradient(')
    expect(aiwelinkHome).toMatch(/\.aiwelink-home\.is-dark\s*{[^}]*--home-canvas-layer:\s*#050608/s)
  })

  it('keeps authentication on the same semantic background system', () => {
    expect(auth).toContain('from-canvas via-surface-muted to-canvas')
  })
})
