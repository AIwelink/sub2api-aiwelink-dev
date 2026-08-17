import { computed, onScopeDispose, ref, type ComputedRef } from 'vue'

export interface ThemePalette {
  primary: string
  primaryAlpha: string
  accent: string
  grid: string
  text: string
  tooltipSurface: string
  chartSeries: string[]
}

const RGB_CHANNELS = /^([\d.]+)[\s,]+([\d.]+)[\s,]+([\d.]+)$/

export function readThemeColor(customProperty: string, fallback: string): string {
  if (typeof document === 'undefined' || typeof getComputedStyle === 'undefined') {
    return fallback
  }

  const value = getComputedStyle(document.documentElement).getPropertyValue(customProperty).trim()
  if (!value) return fallback

  const channels = value.match(RGB_CHANNELS)
  return channels ? `rgb(${channels[1]}, ${channels[2]}, ${channels[3]})` : value
}

export function observeThemeChanges(onChange: () => void): () => void {
  if (typeof document === 'undefined' || typeof MutationObserver === 'undefined') {
    return () => undefined
  }

  const observer = new MutationObserver(onChange)
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

  return () => observer.disconnect()
}

function withAlpha(color: string, alpha: number): string {
  const channels = color.match(/rgb\(\s*([\d.]+)\s*,\s*([\d.]+)\s*,\s*([\d.]+)\s*\)/)
  return channels ? `rgba(${channels[1]}, ${channels[2]}, ${channels[3]}, ${alpha})` : color
}

export function useThemePalette(): ComputedRef<ThemePalette> {
  const revision = ref(0)
  const stop = observeThemeChanges(() => {
    revision.value += 1
  })

  onScopeDispose(stop)

  return computed(() => {
    void revision.value

    const primary = readThemeColor('--color-primary-500', '#D21F4B')
    const accent = readThemeColor('--color-accent-500', '#F4BD38')

    return {
      primary,
      primaryAlpha: withAlpha(primary, 0.125),
      accent,
      grid: readThemeColor('--color-theme-border', '#D9E0E4'),
      text: readThemeColor('--color-theme-muted', '#63717A'),
      tooltipSurface: readThemeColor('--color-surface', '#FFFFFF'),
      chartSeries: [
        primary,
        accent,
        '#9AACB5',
        '#8B5CF6',
        '#10B981',
        '#F59E0B',
        '#EF4444',
        '#3B82F6',
        '#EC4899',
        '#F97316',
        '#06B6D4',
        '#A855F7'
      ]
    }
  })
}
