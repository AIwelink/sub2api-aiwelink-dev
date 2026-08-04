<template>
  <div class="homepage-intro" :data-stage="stage">
    <div v-if="stage === 'preparing'" class="intro-center" role="status" aria-live="polite">
      <h1 data-testid="intro-brand" class="intro-brand" aria-label="AIWELINK API">
        <span class="intro-brand-word">AIWELINK</span>
        {{ ' ' }}
        <span class="intro-brand-api">API</span>
      </h1>

      <div class="thinking-row">
        <span data-testid="intro-spinner" class="thinking-spinner" aria-hidden="true" />
        <span>{{ t('home.redesign.intro.thinking') }}</span>
      </div>

      <div
        data-testid="intro-progress"
        class="thinking-progress"
        aria-hidden="true"
      >
        <span />
      </div>
    </div>

    <div
      v-else-if="stage === 'composing'"
      data-testid="intro-editor"
      class="intro-editor"
      aria-hidden="true"
    >
      <div class="editor-heading">
        <span class="editor-status" />
        <span>AIWELINK / homepage.vue</span>
      </div>
      <div class="editor-lines">
        <code
          v-for="(line, index) in componentLines"
          :key="line"
          class="editor-line"
          :style="{ '--line-index': index }"
        >
          <span class="line-number">{{ String(index + 1).padStart(2, '0') }}</span>
          <span class="line-bracket">&lt;</span><span class="line-name">{{ line }}</span><span class="line-bracket"> /&gt;</span>
        </code>
        <span class="editor-caret" />
      </div>
    </div>

    <div
      v-else
      data-testid="intro-fade"
      class="intro-fade"
      aria-hidden="true"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import type { HomepageIntroStage } from '@/composables/useHomepageIntro'

defineProps<{
  stage: Exclude<HomepageIntroStage, 'ready'>
}>()

const { t } = useI18n()

const componentLines = ['Navigation', 'Hero', 'UseCases', 'Pricing', 'Models']
</script>

<style scoped>
.homepage-intro {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: grid;
  place-items: center;
  overflow: hidden;
  color: #f6f3eb;
  background: #000;
  pointer-events: none;
}

.homepage-intro[data-stage="revealing"] {
  animation: intro-fade-out 520ms cubic-bezier(.22, 1, .36, 1) both;
}

.intro-fade {
  position: absolute;
  inset: 0;
}

.intro-center {
  display: grid;
  width: min(520px, calc(100vw - 40px));
  justify-items: center;
  text-align: center;
}

.intro-brand {
  display: flex;
  min-height: 68px;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin: 0;
  font-family: "Bahnschrift", "Aptos Display", "Microsoft YaHei", sans-serif;
  font-size: 58px;
  font-weight: 800;
  line-height: 1;
  letter-spacing: 0;
  animation: brand-arrive 620ms cubic-bezier(.2, .75, .2, 1) both;
}

.intro-brand-word {
  color: #ffc648;
  text-shadow: 0 0 28px rgba(255, 198, 72, .24);
}

.intro-brand-api {
  color: #ef3f72;
  text-shadow: 0 0 24px rgba(239, 63, 114, .2);
}

.thinking-row {
  display: flex;
  min-height: 24px;
  align-items: center;
  gap: 10px;
  margin-top: 32px;
  color: #a5aab3;
  font-family: "Cascadia Code", "SFMono-Regular", Consolas, monospace;
  font-size: 13px;
  line-height: 1.5;
  letter-spacing: 0;
  animation: status-arrive 380ms 420ms ease both;
}

.thinking-spinner {
  width: 14px;
  height: 14px;
  flex: 0 0 14px;
  border: 1.5px solid rgba(255, 198, 72, .24);
  border-top-color: #ffc648;
  border-right-color: #ef3f72;
  border-radius: 50%;
  animation: spinner-turn 760ms linear infinite;
}

.thinking-progress {
  position: relative;
  width: 180px;
  height: 1px;
  margin-top: 18px;
  overflow: hidden;
  background: rgba(255, 255, 255, .1);
  animation: status-arrive 380ms 520ms ease both;
}

.thinking-progress span {
  position: absolute;
  inset: 0 auto 0 0;
  width: 42%;
  background: #ffc648;
  box-shadow: 46px 0 0 #ef3f72;
  animation: progress-travel 900ms ease-in-out infinite;
}

.intro-editor {
  width: min(560px, calc(100vw - 40px));
  color: #d9dde3;
  font-family: "Cascadia Code", "SFMono-Regular", Consolas, monospace;
  font-size: 14px;
  line-height: 1.75;
  letter-spacing: 0;
}

.editor-heading {
  display: flex;
  min-height: 24px;
  align-items: center;
  gap: 9px;
  margin-bottom: 18px;
  color: #727985;
  font-size: 11px;
  text-transform: uppercase;
}

.editor-status {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ffc648;
  box-shadow: 0 0 12px rgba(255, 198, 72, .64);
}

.editor-lines {
  position: relative;
  display: grid;
  min-height: 154px;
  align-content: start;
}

.editor-line {
  display: grid;
  grid-template-columns: 34px auto 1fr auto;
  width: max-content;
  max-width: 100%;
  opacity: 0;
  transform: translateY(7px);
  animation: line-write 220ms ease forwards;
  animation-delay: calc(var(--line-index) * 130ms);
  white-space: nowrap;
}

.line-number {
  color: #454b55;
  user-select: none;
}

.line-bracket { color: #ef3f72; }
.line-name { color: #ffc648; }

.editor-caret {
  width: 8px;
  height: 17px;
  margin-top: 4px;
  margin-left: 34px;
  background: #ffc648;
  animation: caret-blink 620ms steps(1, end) infinite;
}

@keyframes brand-arrive {
  from { opacity: 0; transform: translateY(12px); filter: blur(8px); }
  to { opacity: 1; transform: translateY(0); filter: blur(0); }
}

@keyframes status-arrive {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes spinner-turn { to { transform: rotate(360deg); } }

@keyframes progress-travel {
  0% { transform: translateX(-160%); }
  55%, 100% { transform: translateX(340%); }
}

@keyframes line-write {
  to { opacity: 1; transform: translateY(0); }
}

@keyframes intro-fade-out {
  from { opacity: 1; }
  to { opacity: 0; }
}

@keyframes caret-blink {
  50% { opacity: 0; }
}

@media (max-width: 640px) {
  .intro-brand {
    min-height: 48px;
    gap: 10px;
    font-size: 38px;
  }

  .thinking-row { margin-top: 26px; }
  .intro-editor { font-size: 12px; }
  .editor-line { grid-template-columns: 28px auto 1fr auto; }
  .editor-caret { margin-left: 28px; }
}

@media (prefers-reduced-motion: reduce) {
  .intro-brand,
  .thinking-row,
  .thinking-progress,
  .editor-line {
    opacity: 1;
    transform: none;
    animation: none;
  }

  .thinking-spinner,
  .thinking-progress span,
  .editor-caret {
    animation: none;
  }

  .homepage-intro[data-stage="revealing"] {
    animation-duration: 120ms;
  }
}
</style>
