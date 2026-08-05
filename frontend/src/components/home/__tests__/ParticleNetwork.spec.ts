import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import ParticleNetwork from '../ParticleNetwork.vue'

const strokeStyles: string[] = []
const fillStyles: string[] = []

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
  get fillStyle() {
    return fillStyles.at(-1) ?? ''
  },
  set fillStyle(value: string) {
    fillStyles.push(value)
  },
  lineWidth: 1,
  globalAlpha: 1,
}

function maxConnectionAlpha() {
  const alphas = strokeStyles.map((style) => Number(style.match(/,\s*([\d.]+)\)$/)?.[1]))
  return Math.max(...alphas)
}

function mockClusteredParticles(clusteredCount: number) {
  let randomCall = 0
  vi.spyOn(Math, 'random').mockImplementation(() => {
    const particleIndex = Math.floor(randomCall / 6)
    const fieldIndex = randomCall % 6
    const distributedIndex = particleIndex - clusteredCount
    randomCall += 1

    if (particleIndex < clusteredCount) return [.5, .5, .2, .2, .5, .5][fieldIndex]
    return [
      .5,
      .5,
      .55 + (distributedIndex % 8) * .05,
      .1 + Math.floor(distributedIndex / 8) * .16,
      .5,
      .5,
    ][fieldIndex]
  })
}

describe('ParticleNetwork', () => {
  beforeEach(() => {
    strokeStyles.length = 0
    fillStyles.length = 0
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

  it('uses warm gold and sparse rose only in the light-theme canvas', async () => {
    vi.spyOn(Math, 'random').mockReturnValue(0)
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()

    expect(maxConnectionAlpha()).toBeGreaterThanOrEqual(.14)
    expect(strokeStyles.some((style) => style.startsWith('rgba(198, 121, 20,'))).toBe(true)
    expect(strokeStyles.some((style) => style.startsWith('rgba(239, 63, 114,'))).toBe(true)
    expect(strokeStyles.some((style) => style.startsWith('rgba(59, 130, 246,'))).toBe(false)
    expect(fillStyles.some((style) => style.startsWith('rgba(198, 121, 20,'))).toBe(true)
    expect(fillStyles.some((style) => style.startsWith('rgba(239, 63, 114,'))).toBe(true)
    wrapper.unmount()
  })

  it('keeps dark-theme connection lines visibly present without pointer boost', async () => {
    document.documentElement.classList.add('dark')
    vi.spyOn(Math, 'random').mockReturnValue(0)
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()

    expect(maxConnectionAlpha()).toBeGreaterThanOrEqual(.22)
    expect(strokeStyles.some((style) => style.startsWith('rgba(255, 198, 72,'))).toBe(true)
    wrapper.unmount()
  })

  it('draws grab lines from nearby particles directly to the pointer', async () => {
    vi.spyOn(Math, 'random').mockReturnValue(0)
    const wrapper = mount(ParticleNetwork)
    await nextTick()
    const animationFrame = vi.mocked(requestAnimationFrame).mock.calls[0]?.[0]

    context.lineTo.mockClear()
    strokeStyles.length = 0
    window.dispatchEvent(new MouseEvent('pointermove', { clientX: 50, clientY: 50 }))
    animationFrame?.(0)

    expect(context.lineTo).toHaveBeenCalledWith(50, 50)
    expect(maxConnectionAlpha()).toBeGreaterThan(.2)
    expect(maxConnectionAlpha()).toBeLessThanOrEqual(.42)
    wrapper.unmount()
  })

  it('scales desktop density by canvas area and uses reference-sized points', async () => {
    vi.spyOn(Math, 'random').mockReturnValue(.9)
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()

    const radii = context.arc.mock.calls.map(([, , radius]) => Number(radius))
    expect(radii).toHaveLength(72)
    expect(Math.max(...radii)).toBeGreaterThan(2)
    expect(Math.max(...radii)).toBeLessThanOrEqual(2.5)
    wrapper.unmount()
  })

  it('fully bursts a centrally collapsed particle cluster back outward', async () => {
    vi.spyOn(Math, 'random').mockReturnValue(.5)
    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()
    const animationFrame = vi.mocked(requestAnimationFrame).mock.calls[0]?.[0]

    try {
      context.arc.mockClear()
      animationFrame?.(0)
      const firstFrameCenters = context.arc.mock.calls.map(([x, y]) => ({ x: Number(x), y: Number(y) }))
      const firstFrameXSpread = Math.max(...firstFrameCenters.map(({ x }) => x)) - Math.min(...firstFrameCenters.map(({ x }) => x))
      const firstFrameYSpread = Math.max(...firstFrameCenters.map(({ y }) => y)) - Math.min(...firstFrameCenters.map(({ y }) => y))

      expect(Math.max(firstFrameXSpread, firstFrameYSpread)).toBeGreaterThan(15)

      for (let frame = 0; frame < 40; frame += 1) {
        context.arc.mockClear()
        animationFrame?.(frame)
      }

      const centers = context.arc.mock.calls.map(([x, y]) => ({ x: Number(x), y: Number(y) }))
      const xSpread = Math.max(...centers.map(({ x }) => x)) - Math.min(...centers.map(({ x }) => x))
      const ySpread = Math.max(...centers.map(({ y }) => y)) - Math.min(...centers.map(({ y }) => y))

      expect(Math.max(xSpread, ySpread)).toBeGreaterThan(120)
    } finally {
      wrapper.unmount()
    }
  })

  it('bursts a dense particle cluster away from the viewport center', async () => {
    const clusteredCount = 30
    mockClusteredParticles(clusteredCount)

    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()
    const animationFrame = vi.mocked(requestAnimationFrame).mock.calls[0]?.[0]

    try {
      context.arc.mockClear()
      animationFrame?.(0)

      const clusteredCenters = context.arc.mock.calls.slice(0, clusteredCount)
        .map(([x, y]) => ({ x: Number(x), y: Number(y) }))
      const xSpread = Math.max(...clusteredCenters.map(({ x }) => x)) - Math.min(...clusteredCenters.map(({ x }) => x))
      const ySpread = Math.max(...clusteredCenters.map(({ y }) => y)) - Math.min(...clusteredCenters.map(({ y }) => y))

      expect(Math.max(xSpread, ySpread)).toBeGreaterThan(15)
    } finally {
      wrapper.unmount()
    }
  })

  it('does not burst when only one third of particles are clustered', async () => {
    const clusteredCount = 24
    mockClusteredParticles(clusteredCount)

    const wrapper = mount(ParticleNetwork, { props: { interactive: false } })
    await nextTick()
    const animationFrame = vi.mocked(requestAnimationFrame).mock.calls[0]?.[0]

    try {
      context.arc.mockClear()
      animationFrame?.(0)

      const clusteredCenters = context.arc.mock.calls.slice(0, clusteredCount)
        .map(([x, y]) => ({ x: Number(x), y: Number(y) }))
      const xSpread = Math.max(...clusteredCenters.map(({ x }) => x)) - Math.min(...clusteredCenters.map(({ x }) => x))
      const ySpread = Math.max(...clusteredCenters.map(({ y }) => y)) - Math.min(...clusteredCenters.map(({ y }) => y))

      expect(Math.max(xSpread, ySpread)).toBeLessThan(1)
    } finally {
      wrapper.unmount()
    }
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
