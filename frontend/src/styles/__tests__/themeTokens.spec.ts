import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const themePath = resolve(process.cwd(), 'src/styles/theme.css')
const tailwindPath = resolve(process.cwd(), 'tailwind.config.js')

function readThemeCss(): string {
  expect(existsSync(themePath), 'theme.css should define the shared theme boundary').toBe(true)
  return readFileSync(themePath, 'utf8')
}

function block(css: string, selector: string): string {
  const escaped = selector.replace('.', '\\.')
  return css.match(new RegExp(`${escaped}\\s*\\{([\\s\\S]*?)\\}`))?.[1] ?? ''
}

describe('AIWeLink theme tokens', () => {
  it('defines the approved C light palette', () => {
    const light = block(readThemeCss(), ':root')

    expect(light).toContain('--color-primary-500: 186 54 80')
    expect(light).toContain('--color-on-primary: 255 255 255')
    expect(light).toContain('--color-canvas: 245 247 248')
    expect(light).toContain('--color-theme-text: 37 51 59')
  })

  it('defines the approved B dark palette', () => {
    const dark = block(readThemeCss(), '.dark')

    expect(dark).toContain('--color-primary-500: 233 190 115')
    expect(dark).toContain('--color-on-primary: 8 19 28')
    expect(dark).toContain('--color-canvas: 7 18 26')
    expect(dark).toContain('--color-theme-text: 244 237 226')
  })

  it('maps Tailwind brand colors to CSS variables', () => {
    const tailwindConfig = readFileSync(tailwindPath, 'utf8')

    expect(tailwindConfig).toContain('const token = (name, fallback) =>')
    expect(tailwindConfig).toContain("500: token('primary-500', '186 54 80')")
    expect(tailwindConfig).toContain('var(--color-primary-500')
    expect(tailwindConfig).toContain("'on-primary'")
    expect(tailwindConfig).toContain("canvas: token('canvas', '245 247 248')")
    expect(tailwindConfig).not.toContain("500: '#14b8a6'")
  })

  it('provides a static fallback for every theme variable reference', () => {
    const tailwindConfig = readFileSync(tailwindPath, 'utf8')
    const variableReferences = tailwindConfig.match(/var\(--color-[^)]+\)/g) ?? []
    const missingFallbacks = variableReferences.filter((reference) => !reference.includes(','))

    expect(variableReferences.length).toBeGreaterThan(0)
    expect(missingFallbacks, missingFallbacks.join('\n')).toEqual([])
  })
})
