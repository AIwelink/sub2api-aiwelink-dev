<template>
  <section id="top" class="homepage-hero">
    <div class="hero-inner">
      <p class="hero-eyebrow">{{ t('home.redesign.hero.eyebrow') }}</p>
      <h1>
        <span class="hero-brand">AIwelink</span>
        {{ ' ' }}
        <span class="hero-api">API</span>
      </h1>
      <p class="hero-description">{{ t('home.redesign.hero.description') }}</p>

      <div class="hero-actions">
        <router-link data-testid="hero-primary" :to="destination" class="hero-primary">
          {{ authenticated ? t('home.goToDashboard') : t('home.redesign.hero.primary') }}
        </router-link>
        <a href="#use-cases" class="hero-secondary">{{ t('home.redesign.hero.secondary') }}</a>
      </div>

      <a href="#use-cases" class="scroll-cue" aria-label="Scroll to use cases">
        <span class="scroll-label">SCROLL</span>
        <span class="scroll-track" aria-hidden="true">
          <span class="scroll-runner" />
        </span>
        <span class="scroll-chevron" aria-hidden="true" />
      </a>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  authenticated: boolean
  dashboardPath: string
}>()

const { t } = useI18n()
const destination = computed(() => props.authenticated ? props.dashboardPath : '/login')
</script>

<style scoped>
.homepage-hero {
  position: relative;
  display: grid;
  min-height: 92svh;
  place-items: center;
  padding: 108px 24px 48px;
  text-align: center;
}

.hero-inner { width: min(860px, 100%); }

.hero-eyebrow {
  margin: 0 0 24px;
  color: var(--home-primary);
  font-family: "Cascadia Code", "SFMono-Regular", Consolas, monospace;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.5;
  letter-spacing: 0;
}

h1 {
  margin: 0;
  color: var(--home-text);
  font-family: "Bahnschrift", "Aptos Display", "Microsoft YaHei", sans-serif;
  font-size: 62px;
  font-weight: 800;
  line-height: 1.02;
  letter-spacing: 0;
}

.hero-brand { color: var(--home-primary); transition: none; }
.hero-api { color: var(--home-accent); transition: none; }

.hero-description {
  max-width: 700px;
  margin: 24px auto 0;
  color: var(--home-muted);
  font-size: 16px;
  line-height: 1.8;
}

.hero-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 30px;
}

.hero-primary,
.hero-secondary {
  display: inline-flex;
  min-height: 46px;
  align-items: center;
  justify-content: center;
  padding: 0 22px;
  border-radius: 5px;
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
  transition: transform 180ms ease, box-shadow 180ms ease, background 180ms ease;
}

.hero-primary {
  position: relative;
  overflow: hidden;
  color: var(--home-primary-contrast);
  background: var(--home-primary);
  box-shadow: 0 10px 30px rgba(var(--home-primary-rgb), .28);
}

.hero-primary::after {
  position: absolute;
  inset: -20% auto -20% -45%;
  width: 26%;
  content: "";
  background: rgba(255, 255, 255, .42);
  transform: skewX(-22deg);
  animation: command-shine 3.4s ease-in-out infinite;
}

.hero-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 38px rgba(var(--home-primary-rgb), .38);
}

.hero-secondary {
  color: var(--home-text);
  background: var(--home-soft);
}

.hero-secondary:hover { background: var(--home-hover); }

.scroll-cue {
  display: grid;
  width: 48px;
  justify-items: center;
  gap: 7px;
  margin: 58px auto 0;
  color: var(--home-faint);
  font-family: "Cascadia Code", monospace;
  font-size: 9px;
  text-decoration: none;
  transition: color 180ms ease;
}

.scroll-track {
  position: relative;
  width: 14px;
  height: 30px;
  overflow: hidden;
}

.scroll-track::before {
  position: absolute;
  top: 3px;
  bottom: 3px;
  left: 50%;
  width: 3px;
  content: "";
  background: radial-gradient(circle, rgba(var(--home-primary-rgb), .44) 1px, transparent 1.5px) center top / 3px 8px repeat-y;
  transform: translateX(-50%);
  transition: opacity 180ms ease;
}

.scroll-runner {
  position: absolute;
  top: -5px;
  left: 50%;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--home-primary);
  box-shadow: 0 0 9px rgba(var(--home-primary-rgb), .7);
  transform: translateX(-50%);
  animation: scroll-runner-travel 1.7s cubic-bezier(.45, 0, .2, 1) infinite;
}

.scroll-chevron {
  width: 7px;
  height: 7px;
  margin-top: -4px;
  border-right: 1px solid var(--home-primary);
  border-bottom: 1px solid var(--home-primary);
  opacity: .72;
  transform: rotate(45deg);
  transition: opacity 180ms ease, transform 180ms ease;
}

.scroll-cue:hover,
.scroll-cue:focus-visible {
  color: var(--home-text);
}

.scroll-cue:hover .scroll-track::before,
.scroll-cue:focus-visible .scroll-track::before {
  opacity: 1;
}

.scroll-cue:hover .scroll-chevron,
.scroll-cue:focus-visible .scroll-chevron {
  opacity: 1;
  transform: translateY(2px) rotate(45deg);
}

.scroll-cue:focus-visible {
  outline: 1px solid var(--home-primary);
  outline-offset: 6px;
}

@keyframes command-shine {
  0%, 62% { transform: translateX(0) skewX(-22deg); opacity: 0; }
  72% { opacity: .8; }
  100% { transform: translateX(650%) skewX(-22deg); opacity: 0; }
}

@keyframes scroll-runner-travel {
  0% { transform: translate(-50%, 0); opacity: 0; }
  14% { opacity: 1; }
  72% { opacity: 1; }
  100% { transform: translate(-50%, 36px); opacity: 0; }
}

@media (max-width: 760px) {
  .homepage-hero { min-height: 88svh; padding: 96px 16px 38px; }
  .hero-eyebrow { margin-bottom: 20px; font-size: 10px; }
  h1 { font-size: 36px; }
  .hero-description { margin-top: 20px; font-size: 14px; line-height: 1.75; }
  .hero-actions { width: min(280px, 100%); flex-direction: column; margin-right: auto; margin-left: auto; }
  .hero-primary, .hero-secondary { width: 100%; }
  .scroll-cue { margin-top: 40px; }
}

@media (prefers-reduced-motion: reduce) {
  .hero-primary, .hero-secondary { transition: none; }
  .hero-primary::after { display: none; }
  .hero-primary:hover { transform: none; }
  .scroll-runner { top: 12px; animation: none; }
  .scroll-cue,
  .scroll-track::before,
  .scroll-chevron { transition: none; }
}
</style>
