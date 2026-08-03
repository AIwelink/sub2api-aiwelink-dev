import { beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { observeThemeChanges, readThemeColor, useThemePalette } from '../useThemePalette'

describe('theme palette', () => {
  beforeEach(() => {
    document.documentElement.className = ''
    document.documentElement.removeAttribute('style')
  })

  it('resolves RGB channel tokens for Chart.js', () => {
    document.documentElement.style.setProperty('--color-primary-500', '210 31 75')

    expect(readThemeColor('--color-primary-500', '#000000')).toBe('rgb(210, 31, 75)')
    expect(readThemeColor('--color-missing', '#123456')).toBe('#123456')
  })

  it('uses the approved light palette when CSS tokens are unavailable', () => {
    const scope = effectScope()
    const palette = scope.run(() => useThemePalette())

    expect(palette?.value.primary).toBe('#D21F4B')
    expect(palette?.value.accent).toBe('#F4BD38')
    expect(palette?.value.grid).toBe('#D9E0E4')
    expect(palette?.value.text).toBe('#63717A')

    scope.stop()
  })

  it('observes root theme class changes and can be stopped', async () => {
    const onChange = vi.fn()
    const stop = observeThemeChanges(onChange)

    document.documentElement.classList.toggle('dark')
    await vi.waitFor(() => expect(onChange).toHaveBeenCalledTimes(1))

    stop()
    document.documentElement.classList.toggle('dark')
    await Promise.resolve()
    expect(onChange).toHaveBeenCalledTimes(1)
  })
})
