<template>
  <section id="models" class="home-section" data-reveal>
    <div class="home-inner">
      <div class="section-heading">
        <p class="section-index">{{ t('home.redesign.models.index') }}</p>
        <div>
          <h2>{{ t('home.redesign.models.title') }}</h2>
          <p class="section-lead">{{ t('home.redesign.models.lead') }}</p>
        </div>
      </div>

      <ul class="model-list" aria-label="GPT, Claude, Gemini">
        <li>GPT</li>
        <li>Claude</li>
        <li>Gemini</li>
      </ul>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<style scoped>
.home-section { position: relative; padding: 96px 0; }
.home-inner { width: min(1040px, calc(100% - 48px)); margin: 0 auto; }
.section-heading { display: grid; grid-template-columns: 150px 1fr; gap: 38px; margin-bottom: 54px; }
.section-index { margin: 7px 0 0; color: var(--home-primary); font-family: "Cascadia Code", monospace; font-size: 11px; font-weight: 700; }
h2 { max-width: 680px; margin: 0; color: var(--home-text); font-size: 37px; line-height: 1.18; font-weight: 760; }
.section-lead { max-width: 560px; margin: 16px 0 0; color: var(--home-muted); font-size: 15px; line-height: 1.75; }
.model-list { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 48px; margin: 0; padding: 8px 0 0; list-style: none; }
.model-list li { color: var(--home-text); font-family: "Bahnschrift", "Aptos Display", sans-serif; font-size: 30px; font-weight: 760; opacity: .9; }
.model-list li:nth-child(2) { color: var(--home-accent); }
.model-list li:nth-child(3) { color: var(--home-primary); }

@media (max-width: 760px) {
  .home-section { padding: 72px 0; }
  .home-inner { width: min(100% - 32px, 660px); }
  .section-heading { grid-template-columns: 1fr; gap: 11px; margin-bottom: 40px; }
  .section-index { margin-top: 0; }
  h2 { font-size: 28px; }
  .section-lead { font-size: 14px; }
  .model-list { grid-template-columns: 1fr; gap: 18px; }
  .model-list li { font-size: 23px; }
}
</style>
