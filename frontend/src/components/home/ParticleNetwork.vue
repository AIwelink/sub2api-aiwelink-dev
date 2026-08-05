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
  burstVx: number
  burstVy: number
  radius: number
  opacity: number
  rose: boolean
}

const LINK_DISTANCE = 150
const POINTER_DISTANCE = 150
const LIGHT_AMBER = '198, 121, 20'
const DARK_GOLD = '255, 198, 72'
const WARM_ROSE = '239, 63, 114'
const CLUSTER_CHECK_INTERVAL = 90
const CLUSTER_RADIUS_RATIO = .16
const CLUSTER_PARTICLE_RATIO = .4
const CLUSTER_IMPULSE = 8
const BURST_DECAY = .975

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
let clusterCheckCooldown = 0

function particleCount() {
  if (width <= 760) return 45

  const referenceDensity = Math.round((width * height * 120) / 800_000)
  return Math.min(180, Math.max(72, referenceDensity))
}

function createParticle(index: number): Particle {
  const angle = Math.random() * Math.PI * 2
  const speed = .16 + Math.random() * .24
  return {
    x: Math.random() * width,
    y: Math.random() * height,
    vx: Math.cos(angle) * speed,
    vy: Math.sin(angle) * speed,
    burstVx: 0,
    burstVy: 0,
    radius: .2 + Math.random() * 2.3,
    opacity: .18 + Math.random() * .24,
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
  const speed = Math.hypot(particle.vx, particle.vy)
  if (speed > .62) {
    const scale = .62 / speed
    particle.vx *= scale
    particle.vy *= scale
  }

  particle.x += particle.vx + particle.burstVx
  particle.y += particle.vy + particle.burstVy
  particle.burstVx *= BURST_DECAY
  particle.burstVy *= BURST_DECAY

  if (Math.abs(particle.burstVx) < .005) particle.burstVx = 0
  if (Math.abs(particle.burstVy) < .005) particle.burstVy = 0

  if (particle.x <= particle.radius) {
    particle.x = particle.radius
    particle.vx = Math.abs(particle.vx)
    particle.burstVx = Math.abs(particle.burstVx)
  } else if (particle.x >= width - particle.radius) {
    particle.x = width - particle.radius
    particle.vx = -Math.abs(particle.vx)
    particle.burstVx = -Math.abs(particle.burstVx)
  }

  if (particle.y <= particle.radius) {
    particle.y = particle.radius
    particle.vy = Math.abs(particle.vy)
    particle.burstVy = Math.abs(particle.burstVy)
  } else if (particle.y >= height - particle.radius) {
    particle.y = height - particle.radius
    particle.vy = -Math.abs(particle.vy)
    particle.burstVy = -Math.abs(particle.burstVy)
  }
}

function applyPairForces() {
  for (let firstIndex = 0; firstIndex < particles.length; firstIndex += 1) {
    const first = particles[firstIndex]
    for (let secondIndex = firstIndex + 1; secondIndex < particles.length; secondIndex += 1) {
      const second = particles[secondIndex]
      const deltaX = first.x - second.x
      const deltaY = first.y - second.y
      const distance = Math.hypot(deltaX, deltaY)
      if (distance === 0 || distance > LINK_DISTANCE) continue

      const attractionX = deltaX / 600_000
      const attractionY = deltaY / 1_200_000
      first.vx -= attractionX
      first.vy -= attractionY
      second.vx += attractionX
      second.vy += attractionY
    }
  }
}

function disperseDenseCluster() {
  if (clusterCheckCooldown > 0) {
    clusterCheckCooldown -= 1
    return
  }
  clusterCheckCooldown = CLUSTER_CHECK_INTERVAL

  const clusterRadius = Math.min(width, height) * CLUSTER_RADIUS_RATIO
  let clusteredParticles: Particle[] = []

  particles.forEach((seed) => {
    const nearbyParticles = particles.filter((particle) => (
      Math.hypot(particle.x - seed.x, particle.y - seed.y) <= clusterRadius
    ))
    if (nearbyParticles.length > clusteredParticles.length) {
      clusteredParticles = nearbyParticles
    }
  })

  if (clusteredParticles.length < particles.length * CLUSTER_PARTICLE_RATIO) return

  const clusterCenter = clusteredParticles.reduce((center, particle) => ({
    x: center.x + particle.x / clusteredParticles.length,
    y: center.y + particle.y / clusteredParticles.length,
  }), { x: 0, y: 0 })

  clusteredParticles.forEach((particle, index) => {
    let deltaX = particle.x - clusterCenter.x
    let deltaY = particle.y - clusterCenter.y
    let distance = Math.hypot(deltaX, deltaY)

    if (distance < .01) {
      const angle = (index / clusteredParticles.length) * Math.PI * 2
      deltaX = Math.cos(angle)
      deltaY = Math.sin(angle)
      distance = 1
    }

    particle.burstVx = (deltaX / distance) * CLUSTER_IMPULSE
    particle.burstVy = (deltaY / distance) * CLUSTER_IMPULSE
  })
}

function connectionColor(isDark: boolean, rose: boolean) {
  if (isDark) return DARK_GOLD
  return rose ? WARM_ROSE : LIGHT_AMBER
}

function drawConnections(isDark: boolean) {
  if (!context) return

  for (let firstIndex = 0; firstIndex < particles.length; firstIndex += 1) {
    const first = particles[firstIndex]
    for (let secondIndex = firstIndex + 1; secondIndex < particles.length; secondIndex += 1) {
      const second = particles[secondIndex]
      const deltaX = first.x - second.x
      const deltaY = first.y - second.y
      const distance = Math.hypot(deltaX, deltaY)
      if (distance > LINK_DISTANCE) continue

      const baseAlpha = (1 - distance / LINK_DISTANCE) * (isDark ? .24 : .18)
      const color = connectionColor(isDark, first.rose || second.rose)

      context.beginPath()
      context.moveTo(first.x, first.y)
      context.lineTo(second.x, second.y)
      context.lineWidth = .9
      context.strokeStyle = `rgba(${color}, ${baseAlpha})`
      context.stroke()
    }
  }
}

function drawPointerConnections(isDark: boolean) {
  if (!context || !props.interactive || pointer.x < 0 || pointer.y < 0) return

  particles.forEach((particle) => {
    const distance = Math.hypot(particle.x - pointer.x, particle.y - pointer.y)
    if (distance > POINTER_DISTANCE) return

    const alpha = (1 - distance / POINTER_DISTANCE) * .4
    context!.beginPath()
    context!.moveTo(particle.x, particle.y)
    context!.lineTo(pointer.x, pointer.y)
    context!.lineWidth = 1
    context!.strokeStyle = `rgba(${connectionColor(isDark, particle.rose)}, ${alpha})`
    context!.stroke()
  })
}

function drawParticles(isDark: boolean) {
  if (!context) return

  particles.forEach((particle) => {
    const pointOpacity = Math.min(.64, particle.opacity + (isDark ? .16 : 0))
    context!.beginPath()
    context!.arc(particle.x, particle.y, particle.radius, 0, Math.PI * 2)
    context!.fillStyle = particle.rose
      ? `rgba(${WARM_ROSE}, ${Math.min(.72, pointOpacity + .08)})`
      : `rgba(${isDark ? DARK_GOLD : LIGHT_AMBER}, ${pointOpacity})`
    context!.fill()
  })
}

function drawFrame(update: boolean) {
  if (!context) return
  const isDark = document.documentElement.classList.contains('dark')

  context.clearRect(0, 0, width, height)
  if (update) {
    disperseDenseCluster()
    applyPairForces()
    particles.forEach(updateParticle)
  }
  drawConnections(isDark)
  drawPointerConnections(isDark)
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
