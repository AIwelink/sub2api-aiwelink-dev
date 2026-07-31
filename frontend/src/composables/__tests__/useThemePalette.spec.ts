import { beforeEach, describe, expect, it, vi } from 'vitest'
import { observeThemeChanges, readThemeColor } from '../useThemePalette'

describe('theme palette', () => {
  beforeEach(() => {
    document.documentElement.className = ''
    document.documentElement.removeAttribute('style')
  })

  it('resolves RGB channel tokens for Chart.js', () => {
    document.documentElement.style.setProperty('--color-primary-500', '186 54 80')

    expect(readThemeColor('--color-primary-500', '#000000')).toBe('rgb(186, 54, 80)')
    expect(readThemeColor('--color-missing', '#123456')).toBe('#123456')
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
