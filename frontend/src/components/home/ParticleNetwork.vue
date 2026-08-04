<template>
  <canvas ref="canvas" class="particle-network" aria-hidden="true" />
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(defineProps<{
  interactive?: boolean
}>(), {
  interactive: true,
})

const emit = defineEmits<{
  ready: []
}>()

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  radius: number
  rose: boolean
}

const canvas = ref<HTMLCanvasElement | null>(null)
const pointer = { x: -1000, y: -1000 }

let context: CanvasRenderingContext2D | null = null
let particles: Particle[] = []
let width = 0
let height = 0
let frameId: number | null = null
let resizeObserver: ResizeObserver | null = null
let themeObserver: MutationObserver | null = null
let mediaQuery: MediaQueryList | null = null
let reducedMotion = false
let destroyed = false

function particleCount() {
  return width <= 760 ? 45 : 120
}

function createParticle(index: number): Particle {
  const angle = Math.random() * Math.PI * 2
  const speed = .09 + Math.random() * .16
  return {
    x: Math.random() * width,
    y: Math.random() * height,
    vx: Math.cos(angle) * speed,
    vy: Math.sin(angle) * speed,
    radius: .8 + Math.random() * 1.2,
    rose: index % 11 === 0,
  }
}

function resizeParticles(previousWidth: number, previousHeight: number) {
  if (particles.length === 0 || previousWidth <= 0 || previousHeight <= 0) {
    particles = Array.from({ length: particleCount() }, (_, index) => createParticle(index))
    return
  }

  const scaleX = width / previousWidth
  const scaleY = height / previousHeight
  particles.forEach((particle) => {
    particle.x *= scaleX
    particle.y *= scaleY
  })

  const targetCount = particleCount()
  if (particles.length > targetCount) {
    particles = particles.slice(0, targetCount)
    return
  }

  while (particles.length < targetCount) {
    particles.push(createParticle(particles.length))
  }
}

function resizeCanvas() {
  if (!canvas.value || !context) return

  const previousWidth = width
  const previousHeight = height
  const bounds = canvas.value.getBoundingClientRect()
  width = Math.max(1, bounds.width || window.innerWidth)
  height = Math.max(1, bounds.height || window.innerHeight)
  const pixelRatio = Math.min(window.devicePixelRatio || 1, 2)

  canvas.value.width = Math.round(width * pixelRatio)
  canvas.value.height = Math.round(height * pixelRatio)
  context.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0)
  resizeParticles(previousWidth, previousHeight)
  drawFrame(false)
}

function updateParticle(particle: Particle) {
  particle.x += particle.vx
  particle.y += particle.vy

  if (particle.x < -8) particle.x = width + 8
  if (particle.x > width + 8) particle.x = -8
  if (particle.y < -8) particle.y = height + 8
  if (particle.y > height + 8) particle.y = -8
}

function drawConnections(isDark: boolean) {
  if (!context) return

  const maxDistance = width <= 760 ? 112 : 132
  const pointerRadius = 185

  for (let firstIndex = 0; firstIndex < particles.length; firstIndex += 1) {
    const first = particles[firstIndex]
    for (let secondIndex = firstIndex + 1; secondIndex < particles.length; secondIndex += 1) {
      const second = particles[secondIndex]
      const deltaX = first.x - second.x
      const deltaY = first.y - second.y
      const distance = Math.hypot(deltaX, deltaY)
      if (distance > maxDistance) continue

      const midpointX = (first.x + second.x) / 2
      const midpointY = (first.y + second.y) / 2
      const pointerDistance = Math.hypot(midpointX - pointer.x, midpointY - pointer.y)
      const pointerBoost = props.interactive && pointerDistance < pointerRadius
        ? (1 - pointerDistance / pointerRadius) * .12
        : 0
      const baseAlpha = (1 - distance / maxDistance) * (isDark ? .22 : .145)

      context.beginPath()
      context.moveTo(first.x, first.y)
      context.lineTo(second.x, second.y)
      context.lineWidth = pointerBoost > 0 ? .85 : .55
      context.strokeStyle = `rgba(255, 198, 72, ${baseAlpha + pointerBoost})`
      context.stroke()
    }
  }
}

function drawParticles(isDark: boolean) {
  if (!context) return

  particles.forEach((particle) => {
    context!.beginPath()
    context!.arc(particle.x, particle.y, particle.radius, 0, Math.PI * 2)
    context!.fillStyle = particle.rose
      ? `rgba(239, 63, 114, ${isDark ? .82 : .66})`
      : `rgba(255, 198, 72, ${isDark ? .72 : .58})`
    context!.fill()
  })
}

function drawFrame(update: boolean) {
  if (!context) return
  const isDark = document.documentElement.classList.contains('dark')

  context.clearRect(0, 0, width, height)
  if (update) particles.forEach(updateParticle)
  drawConnections(isDark)
  drawParticles(isDark)
}

function animate() {
  if (destroyed || reducedMotion || document.hidden) {
    frameId = null
    return
  }

  drawFrame(true)
  frameId = requestAnimationFrame(animate)
}

function startAnimation() {
  if (destroyed || reducedMotion || document.hidden || frameId !== null || !context) return
  frameId = requestAnimationFrame(animate)
}

function stopAnimation() {
  if (frameId === null) return
  cancelAnimationFrame(frameId)
  frameId = null
}

function handlePointerMove(event: PointerEvent) {
  if (!props.interactive) return
  pointer.x = event.clientX
  pointer.y = event.clientY
}

function handlePointerLeave() {
  pointer.x = -1000
  pointer.y = -1000
}

function handleVisibilityChange() {
  if (document.hidden) stopAnimation()
  else startAnimation()
}

function handleMotionChange(event: MediaQueryListEvent) {
  reducedMotion = event.matches
  if (reducedMotion) {
    stopAnimation()
    drawFrame(false)
  } else {
    startAnimation()
  }
}

onMounted(() => {
  const element = canvas.value
  context = element?.getContext('2d') ?? null

  if (!element || !context) {
    emit('ready')
    return
  }

  mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  reducedMotion = mediaQuery.matches
  mediaQuery.addEventListener?.('change', handleMotionChange)

  window.addEventListener('pointermove', handlePointerMove)
  window.addEventListener('pointerleave', handlePointerLeave)
  document.addEventListener('visibilitychange', handleVisibilityChange)

  if (typeof ResizeObserver === 'function') {
    resizeObserver = new ResizeObserver(resizeCanvas)
    resizeObserver.observe(element)
  } else {
    window.addEventListener('resize', resizeCanvas)
  }

  themeObserver = new MutationObserver(() => drawFrame(false))
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

  resizeCanvas()
  startAnimation()
  emit('ready')
})

onBeforeUnmount(() => {
  destroyed = true
  stopAnimation()
  resizeObserver?.disconnect()
  themeObserver?.disconnect()
  mediaQuery?.removeEventListener?.('change', handleMotionChange)
  window.removeEventListener('resize', resizeCanvas)
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerleave', handlePointerLeave)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.particle-network {
  position: fixed;
  inset: 0;
  z-index: 0;
  display: block;
  width: 100%;
  height: 100%;
  pointer-events: none;
}
</style>
