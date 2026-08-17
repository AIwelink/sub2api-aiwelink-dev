import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const home = readFileSync(resolve(process.cwd(), 'src/views/HomeView.vue'), 'utf8')
const aiwelinkHome = readFileSync(resolve(process.cwd(), 'src/components/home/AIWeLinkHome.vue'), 'utf8')
const auth = readFileSync(resolve(process.cwd(), 'src/components/layout/AuthLayout.vue'), 'utf8')

describe('public theme surfaces', () => {
  it('keeps the official compact home intact and themes the branded default home', () => {
    expect(home).toContain('bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white')
    expect(home).toContain('border-gray-200 px-4 py-4 sm:px-6 dark:border-dark-800')
    expect(home).toContain('text-gray-500 hover:bg-gray-100 dark:text-dark-400')
    expect(home).not.toContain('bg-canvas text-theme-text')

    expect(aiwelinkHome).toContain('class="aiwelink-home"')
    expect(aiwelinkHome).toContain(":class=\"{ 'is-dark': dark }\"")
    expect(aiwelinkHome).toContain('--home-canvas-layer: linear-gradient(')
    expect(aiwelinkHome).toMatch(/\.aiwelink-home\.is-dark\s*{[^}]*--home-canvas-layer:\s*#050608/s)
  })

  it('keeps authentication on the same semantic background system', () => {
    expect(auth).toContain('from-canvas via-surface-muted to-canvas')
  })
})
