<template>
  <section id="use-cases" class="home-section" data-reveal>
    <div class="home-inner">
      <div class="section-heading">
        <p class="section-index">{{ t('home.redesign.useCases.index') }}</p>
        <div>
          <h2>{{ t('home.redesign.useCases.title') }}</h2>
          <p class="section-lead">{{ t('home.redesign.useCases.lead') }}</p>
        </div>
      </div>

      <div class="use-case-grid">
        <article v-for="item in items" :key="item.number" class="use-case-item">
          <b>{{ item.number }}</b>
          <h3>{{ t(item.title) }}</h3>
          <p>{{ t(item.description) }}</p>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const items = [
  { number: '01', title: 'home.redesign.useCases.coding.title', description: 'home.redesign.useCases.coding.description' },
  { number: '02', title: 'home.redesign.useCases.research.title', description: 'home.redesign.useCases.research.description' },
  { number: '03', title: 'home.redesign.useCases.agent.title', description: 'home.redesign.useCases.agent.description' },
]
</script>

<style scoped>
.home-section { position: relative; padding: 96px 0; }
.home-inner { width: min(1040px, calc(100% - 48px)); margin: 0 auto; }
.section-heading { display: grid; grid-template-columns: 150px 1fr; gap: 38px; margin-bottom: 54px; }
.section-index { margin: 7px 0 0; color: var(--home-primary); font-family: "Cascadia Code", monospace; font-size: 11px; font-weight: 700; }
h2 { max-width: 680px; margin: 0; color: var(--home-text); font-size: 37px; line-height: 1.18; font-weight: 760; }
.section-lead { max-width: 560px; margin: 16px 0 0; color: var(--home-muted); font-size: 15px; line-height: 1.75; }
.use-case-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 52px; }
.use-case-item { min-width: 0; }
.use-case-item b { display: block; margin-bottom: 22px; color: var(--home-primary); font-family: "Cascadia Code", monospace; font-size: 25px; line-height: 1; }
.use-case-item h3 { margin: 0; color: var(--home-text); font-size: 19px; line-height: 1.3; }
.use-case-item p { margin: 10px 0 0; color: var(--home-muted); font-size: 14px; line-height: 1.65; }

@media (max-width: 760px) {
  .home-section { padding: 72px 0; }
  .home-inner { width: min(100% - 32px, 660px); }
  .section-heading { grid-template-columns: 1fr; gap: 11px; margin-bottom: 40px; }
  .section-index { margin-top: 0; }
  h2 { font-size: 28px; }
  .section-lead { font-size: 14px; }
  .use-case-grid { grid-template-columns: 1fr; gap: 38px; }
}
</style>
