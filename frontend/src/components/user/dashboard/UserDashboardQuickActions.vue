<template>
  <section class="dashboard-section dashboard-embedded-section">
    <header class="dashboard-section-header">
      <div>
        <h2 class="text-sm font-semibold text-theme-text">{{ t('dashboard.quickActions') }}</h2>
        <p class="mt-0.5 text-[11px] text-theme-muted">{{ t('dashboard.generateNewKey') }}</p>
      </div>
    </header>
    <div>
      <button type="button" @click="router.push('/keys')" class="dashboard-command group">
        <Icon name="key" size="sm" class="dashboard-command-icon" />
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-theme-text">{{ t('dashboard.createApiKey') }}</p>
          <p class="mt-0.5 truncate text-[11px] text-theme-muted">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="dashboard-command-arrow" />
      </button>

      <button type="button" @click="router.push('/usage')" class="dashboard-command group">
        <Icon name="chart" size="sm" class="dashboard-command-icon" />
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-theme-text">{{ t('dashboard.viewUsage') }}</p>
          <p class="mt-0.5 truncate text-[11px] text-theme-muted">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="dashboard-command-arrow" />
      </button>

      <button v-if="canUseBatchImage" type="button" @click="router.push('/batch-image')" class="dashboard-command group">
        <Icon name="sparkles" size="sm" class="dashboard-command-icon" />
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-theme-text">{{ t('dashboard.batchImageAgent') }}</p>
          <p class="mt-0.5 truncate text-[11px] text-theme-muted">{{ t('dashboard.batchImageAgentDesc') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="dashboard-command-arrow" />
      </button>

      <button type="button" @click="router.push('/redeem')" class="dashboard-command group">
        <Icon name="gift" size="sm" class="dashboard-command-icon" />
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-theme-text">{{ t('dashboard.redeemCode') }}</p>
          <p class="mt-0.5 truncate text-[11px] text-theme-muted">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="dashboard-command-arrow" />
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
const router = useRouter()
const { t } = useI18n()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

onMounted(() => {
  void refreshBatchImageAccess()
})
</script>
