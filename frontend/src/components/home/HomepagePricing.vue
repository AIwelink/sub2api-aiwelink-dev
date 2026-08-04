<template>
  <section id="pricing" class="home-section" data-reveal>
    <div class="home-inner">
      <div class="section-heading">
        <p class="section-index">{{ t('home.redesign.pricing.index') }}</p>
        <div>
          <h2>{{ t('home.redesign.pricing.title') }}</h2>
          <p class="section-lead">{{ t('home.redesign.pricing.lead') }}</p>
        </div>
      </div>

      <div class="pricing-layout">
        <div class="pricing-primary">
          <strong>{{ t('home.redesign.pricing.equation') }}</strong>
          <p>{{ t('home.redesign.pricing.equationLabel') }}</p>
        </div>
        <div class="pricing-facts">
          <article>
            <span>LOW</span>
            <div>
              <p>{{ t('home.redesign.pricing.multiplierLabel') }}</p>
              <strong>{{ t('home.redesign.pricing.multiplier') }}</strong>
            </div>
          </article>
          <article>
            <span>BILL</span>
            <div>
              <p>{{ t('home.redesign.pricing.billingLabel') }}</p>
              <small>{{ t('home.redesign.pricing.billingDescription') }}</small>
            </div>
          </article>
        </div>
      </div>
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
.section-index { margin: 7px 0 0; color: #ef3f72; font-family: "Cascadia Code", monospace; font-size: 11px; font-weight: 700; }
h2 { max-width: 680px; margin: 0; color: var(--home-text); font-size: 37px; line-height: 1.18; font-weight: 760; }
.section-lead { max-width: 560px; margin: 16px 0 0; color: var(--home-muted); font-size: 15px; line-height: 1.75; }
.pricing-layout { display: grid; grid-template-columns: 1.05fr .95fr; align-items: start; gap: 86px; }
.pricing-primary strong { display: block; color: #ffc648; font-family: "Bahnschrift", sans-serif; font-size: 62px; line-height: 1; white-space: nowrap; }
.pricing-primary p { margin: 18px 0 0; color: var(--home-muted); font-size: 15px; }
.pricing-facts { display: grid; gap: 34px; }
.pricing-facts article { display: grid; grid-template-columns: 68px 1fr; align-items: start; gap: 0; }
.pricing-facts article > span { padding-top: 5px; color: #ef3f72; font-family: "Cascadia Code", monospace; font-size: 10px; font-weight: 700; }
.pricing-facts p { margin: 0; color: var(--home-muted); font-size: 13px; }
.pricing-facts strong { display: block; margin-top: 7px; color: var(--home-text); font-size: 24px; line-height: 1.2; }
.pricing-facts small { display: block; margin-top: 7px; color: var(--home-text); font-size: 15px; }

@media (max-width: 760px) {
  .home-section { padding: 72px 0; }
  .home-inner { width: min(100% - 32px, 660px); }
  .section-heading { grid-template-columns: 1fr; gap: 11px; margin-bottom: 40px; }
  .section-index { margin-top: 0; }
  h2 { font-size: 28px; }
  .section-lead { font-size: 14px; }
  .pricing-layout { grid-template-columns: 1fr; gap: 46px; }
  .pricing-primary strong { font-size: 44px; }
}
</style>
