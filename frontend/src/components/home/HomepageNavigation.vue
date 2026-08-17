<template>
  <header class="home-navigation">
    <nav class="navigation-inner" :aria-label="t('home.redesign.hero.eyebrow')">
      <a href="#top" data-testid="home-wordmark" class="aiwelink-wordmark">AIwelink</a>

      <div class="navigation-actions">
        <div class="navigation-links">
          <a href="#use-cases">{{ t('home.redesign.nav.useCases') }}</a>
          <a href="#pricing">{{ t('home.redesign.nav.pricing') }}</a>
          <a href="#models">{{ t('home.redesign.nav.models') }}</a>
        </div>

        <LocaleSwitcher class="locale-switcher" />

        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="navigation-icon"
          :title="t('home.viewDocs')"
        >
          <Icon name="book" size="sm" />
        </a>

        <button
          data-testid="home-theme-toggle"
          type="button"
          class="navigation-icon"
          :title="dark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="$emit('toggle-theme')"
        >
          <Icon :name="dark ? 'sun' : 'moon'" size="sm" />
        </button>

        <router-link :to="destination" class="navigation-command">
          {{ authenticated ? t('home.dashboard') : t('home.login') }}
        </router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  authenticated: boolean
  dashboardPath: string
  dark: boolean
  docUrl: string
}>()

defineEmits<{
  'toggle-theme': []
}>()

const { t } = useI18n()
const destination = computed(() => props.authenticated ? props.dashboardPath : '/login')
</script>

<style scoped>
.home-navigation {
  position: fixed;
  inset: 0 0 auto;
  z-index: 40;
  height: 64px;
  color: var(--home-text);
  background: color-mix(in srgb, var(--home-canvas) 72%, transparent);
  backdrop-filter: blur(16px);
}

.navigation-inner {
  display: flex;
  width: min(1040px, calc(100% - 48px));
  height: 64px;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin: 0 auto;
}

.aiwelink-wordmark {
  flex: 0 0 auto;
  color: var(--home-primary);
  font-family: "Bahnschrift", "Aptos Display", "Microsoft YaHei", sans-serif;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 0;
  text-decoration: none;
  transition: none;
}

.navigation-actions,
.navigation-links {
  display: flex;
  align-items: center;
}

.navigation-actions { gap: 12px; }
.navigation-links { gap: 28px; margin-right: 4px; }

.navigation-links a {
  color: var(--home-muted);
  font-size: 13px;
  text-decoration: none;
  transition: color 180ms ease;
}

.navigation-links a:hover { color: var(--home-text); }

.navigation-icon {
  display: inline-grid;
  width: 38px;
  height: 38px;
  flex: 0 0 38px;
  place-items: center;
  border: 0;
  border-radius: 5px;
  color: var(--home-text);
  background: transparent;
  cursor: pointer;
  transition: background 180ms ease, color 180ms ease;
}

.navigation-icon:hover { background: var(--home-hover); }

.navigation-command {
  display: inline-flex;
  min-width: 74px;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  padding: 0 20px;
  border-radius: 5px;
  color: var(--home-primary-contrast);
  background: var(--home-primary);
  box-shadow: 0 8px 26px rgba(var(--home-primary-rgb), .26);
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.navigation-command:hover {
  transform: translateY(-1px);
  box-shadow: 0 11px 32px rgba(var(--home-primary-rgb), .34);
}

@media (max-width: 760px) {
  .navigation-inner { width: min(100% - 32px, 660px); }
  .navigation-links, .locale-switcher { display: none; }
  .navigation-actions { gap: 7px; }
  .navigation-command { min-width: 62px; min-height: 40px; padding: 0 14px; }
}

@media (prefers-reduced-motion: reduce) {
  .navigation-command { transition: none; }
  .navigation-command:hover { transform: none; }
}
</style>
