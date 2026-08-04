<template>
  <section class="final-cta" data-reveal>
    <div class="final-inner">
      <p>{{ t('home.redesign.cta.eyebrow') }}</p>
      <h2>{{ t('home.redesign.cta.title') }}</h2>
      <span>{{ t('home.redesign.cta.description') }}</span>
      <router-link :to="destination" class="final-command">
        {{ authenticated ? t('home.goToDashboard') : t('home.redesign.cta.register') }}
      </router-link>
    </div>
  </section>
  <footer class="home-footer">
    <div><span>© {{ year }} AIWELINK</span><span>GPT · Claude · Gemini</span></div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  authenticated: boolean
  dashboardPath: string
}>()

const { t } = useI18n()
const destination = computed(() => props.authenticated ? props.dashboardPath : '/register')
const year = new Date().getFullYear()
</script>

<style scoped>
.final-cta {
  position: relative;
  display: grid;
  min-height: 54svh;
  place-items: center;
  padding: 92px 24px;
  text-align: center;
}

.final-inner { width: min(760px, 100%); }
.final-inner > p { margin: 0 0 20px; color: #ef3f72; font-family: "Cascadia Code", monospace; font-size: 11px; font-weight: 700; }
.final-inner h2 { margin: 0; color: var(--home-text); font-size: 46px; line-height: 1.15; }
.final-inner > span { display: block; margin-top: 18px; color: var(--home-muted); font-size: 15px; }
.final-command { display: inline-flex; min-height: 46px; align-items: center; justify-content: center; margin-top: 30px; padding: 0 24px; border-radius: 5px; color: #181005; background: #ffc648; box-shadow: 0 10px 30px rgba(255, 198, 72, .28); font-size: 14px; font-weight: 700; text-decoration: none; transition: transform 180ms ease, box-shadow 180ms ease; }
.final-command:hover { transform: translateY(-1px); box-shadow: 0 14px 36px rgba(255, 198, 72, .36); }
.home-footer { position: relative; padding: 28px 24px 38px; color: var(--home-faint); font-size: 12px; }
.home-footer > div { display: flex; width: min(1040px, 100%); justify-content: space-between; gap: 20px; margin: 0 auto; }

@media (max-width: 760px) {
  .final-cta { min-height: 48svh; padding: 72px 16px; }
  .final-inner h2 { font-size: 32px; }
  .home-footer > div { flex-direction: column; gap: 8px; }
}

@media (prefers-reduced-motion: reduce) {
  .final-command { transition: none; }
  .final-command:hover { transform: none; }
}
</style>
