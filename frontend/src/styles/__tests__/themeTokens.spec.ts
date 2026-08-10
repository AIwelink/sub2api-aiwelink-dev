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

function contrast(
  foreground: [number, number, number],
  background: [number, number, number]
): number {
  const luminance = ([r, g, b]: [number, number, number]) => {
    const channels = [r, g, b].map((channel) => {
      const value = channel / 255
      return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
    })

    return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722
  }

  const foregroundLuminance = luminance(foreground)
  const backgroundLuminance = luminance(background)

  return (
    (Math.max(foregroundLuminance, backgroundLuminance) + 0.05) /
    (Math.min(foregroundLuminance, backgroundLuminance) + 0.05)
  )
}

describe('AIWeLink theme tokens', () => {
  it('defines the approved vivid rose light palette', () => {
    const light = block(readThemeCss(), ':root')

    expect(light).toContain('--color-primary-500: 210 31 75')
    expect(light).toContain('--color-accent-500: 244 189 56')
    expect(light).toContain('--color-on-primary: 255 255 255')
    expect(light).toContain('--color-canvas: 245 247 248')
    expect(light).toContain('--color-surface: 255 255 255')
    expect(light).toContain('--color-surface-muted: 237 241 243')
    expect(light).toContain('--color-theme-border: 217 224 228')
    expect(light).toContain('--color-theme-text: 32 42 49')
    expect(light).toContain('--color-theme-muted: 99 113 122')
  })

  it('defines the approved neutral black and vivid gold dark palette', () => {
    const dark = block(readThemeCss(), '.dark')

    expect(dark).toContain('--color-primary-500: 255 194 71')
    expect(dark).toContain('--color-accent-500: 226 49 92')
    expect(dark).toContain('--color-on-primary: 24 16 5')
    expect(dark).toContain('--color-canvas: 3 5 7')
    expect(dark).toContain('--color-surface: 13 17 22')
    expect(dark).toContain('--color-surface-muted: 23 29 36')
    expect(dark).toContain('--color-theme-border: 48 57 69')
    expect(dark).toContain('--color-theme-text: 250 247 240')
    expect(dark).toContain('--color-theme-muted: 170 178 188')
  })

  it('keeps primary and muted semantic pairs at WCAG AA contrast', () => {
    expect(contrast([210, 31, 75], [255, 255, 255])).toBeGreaterThanOrEqual(4.5)
    expect(contrast([255, 194, 71], [24, 16, 5])).toBeGreaterThanOrEqual(4.5)
    expect(contrast([99, 113, 122], [245, 247, 248])).toBeGreaterThanOrEqual(4.5)
    expect(contrast([170, 178, 188], [13, 17, 22])).toBeGreaterThanOrEqual(4.5)
  })

  it('maps Tailwind brand colors to CSS variables', () => {
    const tailwindConfig = readFileSync(tailwindPath, 'utf8')

    expect(tailwindConfig).toContain('const token = (name, fallback) =>')
    expect(tailwindConfig).toContain("500: token('primary-500', '210 31 75')")
    expect(tailwindConfig).toContain('var(--color-primary-500')
    expect(tailwindConfig).toContain("'on-primary'")
    expect(tailwindConfig).toContain("canvas: token('canvas', '245 247 248')")
    expect(tailwindConfig).toContain("'surface-muted': token('surface-muted', '237 241 243')")
    expect(tailwindConfig).toContain("'theme-border': token('theme-border', '217 224 228')")
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
