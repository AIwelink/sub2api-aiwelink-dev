import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import ParticleNetwork from '../ParticleNetwork.vue'

const strokeStyles: string[] = []

const context = {
  clearRect: vi.fn(),
  beginPath: vi.fn(),
  moveTo: vi.fn(),
  lineTo: vi.fn(),
  stroke: vi.fn(),
  arc: vi.fn(),
  fill: vi.fn(),
  setTransform: vi.fn(),
  get strokeStyle() {
    return strokeStyles.at(-1) ?? ''
  },
  set strokeStyle(value: string) {
    strokeStyles.push(value)
  },
  fillStyle: '',
  lineWidth: 1,
  globalAlpha: 1,
}

function maxConnectionAlpha() {
  const alphas = strokeStyles.map((style) => Number(style.match(/,\s*([\d.]+)\)$/)?.[1]))
  return Math.max(...alphas)
}

describe('ParticleNetwork', () => {
  beforeEach(() => {
    strokeStyles.length = 0
    document.documentElement.classList.remove('dark')
    Object.values(context).forEach((value) => {
      if (typeof value === 'function' && 'mockClear' in value) value.mockClear()
    })

    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(
      context as unknown as CanvasRenderingContext2D,
    )
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      width: 800,
      height: 600,
      top: 0,
      right: 800,
      bottom: 600,
      left: 0,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    })
    vi.stubGlobal('requestAnimationFrame', vi.fn(() => 42))
    vi.stubGlobal('cancelAnimationFrame', vi.fn())
    vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: false } as MediaQueryList)
  })

  afterEach(() => {
    document.documentElement.classList.remove('dark')
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('keeps light-theme connection lines visibly present without pointer boost', async () => {
    vi.spyOn(Math, 'random').mockReturnValue(0)
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()

    expect(maxConnectionAlpha()).toBeGreaterThanOrEqual(.14)
    wrapper.unmount()
  })

  it('keeps dark-theme connection lines visibly present without pointer boost', async () => {
    document.documentElement.classList.add('dark')
    vi.spyOn(Math, 'random').mockReturnValue(0)
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()

    expect(maxConnectionAlpha()).toBeGreaterThanOrEqual(.22)
    wrapper.unmount()
  })

  it('keeps the pointer connection boost restrained', async () => {
    vi.spyOn(Math, 'random').mockReturnValue(0)
    const wrapper = mount(ParticleNetwork)
    await nextTick()
    const baseAlpha = maxConnectionAlpha()
    const animationFrame = vi.mocked(requestAnimationFrame).mock.calls[0]?.[0]

    window.dispatchEvent(new MouseEvent('pointermove', { clientX: 0, clientY: 0 }))
    strokeStyles.length = 0
    animationFrame?.(0)

    expect(maxConnectionAlpha() - baseAlpha).toBeLessThanOrEqual(.12)
    wrapper.unmount()
  })

  it('sizes and paints the canvas before reporting readiness', async () => {
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()

    const canvas = wrapper.get('canvas').element as HTMLCanvasElement
    expect(canvas.width).toBe(800)
    expect(canvas.height).toBe(600)
    expect(context.fill).toHaveBeenCalled()
    expect(wrapper.emitted('ready')).toHaveLength(1)
    expect(requestAnimationFrame).toHaveBeenCalled()
  })

  it('preserves particle positions when a small canvas resize keeps the same density', async () => {
    let resizeCallback!: ResizeObserverCallback
    let canvasWidth = 800
    class CapturingResizeObserver {
      constructor(callback: ResizeObserverCallback) {
        resizeCallback = callback
      }

      observe = vi.fn()
      disconnect = vi.fn()
      unobserve = vi.fn()
    }

    vi.stubGlobal('ResizeObserver', CapturingResizeObserver)
    vi.mocked(HTMLElement.prototype.getBoundingClientRect).mockImplementation(() => ({
      width: canvasWidth,
      height: 600,
      top: 0,
      right: canvasWidth,
      bottom: 600,
      left: 0,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    }))
    const random = vi.spyOn(Math, 'random').mockReturnValue(.25)
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()
    random.mockClear()

    canvasWidth = 792
    resizeCallback([], {} as ResizeObserver)

    expect(random).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('renders one still frame when reduced motion is enabled', async () => {
    vi.mocked(window.matchMedia).mockReturnValue({ matches: true } as MediaQueryList)
    mount(ParticleNetwork)
    await nextTick()

    expect(context.fill).toHaveBeenCalled()
    expect(requestAnimationFrame).not.toHaveBeenCalled()
  })

  it('still reports readiness when Canvas 2D is unavailable', async () => {
    vi.mocked(HTMLCanvasElement.prototype.getContext).mockReturnValue(null)
    const wrapper = mount(ParticleNetwork)
    await nextTick()

    expect(wrapper.emitted('ready')).toHaveLength(1)
    expect(requestAnimationFrame).not.toHaveBeenCalled()
  })

  it('cancels animation and removes listeners on unmount', async () => {
    const addListener = vi.spyOn(window, 'addEventListener')
    const removeListener = vi.spyOn(window, 'removeEventListener')
    const wrapper = mount(ParticleNetwork)
    await nextTick()

    const pointerHandler = addListener.mock.calls.find(([event]) => event === 'pointermove')?.[1]
    expect(pointerHandler).toBeTypeOf('function')

    wrapper.unmount()

    expect(cancelAnimationFrame).toHaveBeenCalledWith(42)
    expect(removeListener).toHaveBeenCalledWith('pointermove', pointerHandler)
  })
})
