<template>
  <section class="dashboard-section dashboard-embedded-section">
    <header class="dashboard-section-header">
      <div>
        <h2 class="text-sm font-semibold text-theme-text">{{ t('dashboard.recentUsage') }}</h2>
        <p class="mt-0.5 text-[11px] text-theme-muted">{{ t('dashboard.last7Days') }}</p>
      </div>
      <router-link to="/usage" class="inline-flex items-center gap-1 text-[11px] font-medium text-theme-muted transition-colors hover:text-theme-text">
        {{ t('dashboard.viewAllUsage') }}
        <Icon name="arrowRight" size="sm" />
      </router-link>
    </header>
    <div>
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner size="md" />
      </div>
      <div v-else-if="data.length === 0" class="py-6">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <div v-else>
        <div class="hidden grid-cols-[minmax(0,1fr)_100px_128px_124px] gap-4 border-b border-[rgb(var(--workbench-border))] px-4 py-2 font-mono text-[10px] text-theme-muted md:grid md:px-5">
          <span>{{ t('dashboard.model') }}</span>
          <span class="text-right">{{ t('dashboard.tokens') }}</span>
          <span class="text-right">{{ t('dashboard.actual') }}</span>
          <span class="text-right">{{ t('dashboard.time') }}</span>
        </div>
        <div
          v-for="log in data"
          :key="log.id"
          class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-4 border-b border-[rgb(var(--workbench-border))] px-4 py-3 transition-colors last:border-b-0 hover:bg-[rgb(var(--workbench-surface-hover))] md:grid-cols-[minmax(0,1fr)_100px_128px_124px] md:px-5"
        >
          <div class="min-w-0">
            <p class="truncate text-xs font-medium text-theme-text" :title="log.model">{{ log.model }}</p>
            <p class="mt-0.5 truncate font-mono text-[10px] text-theme-muted">{{ log.inbound_endpoint || log.request_type || 'API' }}</p>
          </div>
          <p class="hidden text-right font-mono text-[11px] text-theme-muted md:block">{{ (log.input_tokens + log.output_tokens).toLocaleString() }}</p>
          <p class="text-right font-mono text-[11px]">
            <span class="text-emerald-600 dark:text-emerald-400" :title="t('dashboard.actual')">${{ formatCost(log.actual_cost) }}</span>
            <span class="text-theme-muted" :title="t('dashboard.standard')"> / ${{ formatCost(log.total_cost) }}</span>
          </p>
          <div class="hidden text-right md:block">
            <p class="font-mono text-[11px] text-theme-text">{{ formatTime(log.created_at) }}</p>
            <p class="mt-0.5 font-mono text-[10px] text-theme-muted">{{ formatDate(log.created_at) }}</p>
          </div>
          <p class="col-span-2 font-mono text-[10px] text-theme-muted md:hidden">{{ formatDateTime(log.created_at) }} · {{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
}>()
const { t } = useI18n()
const formatCost = (c: number) => c.toFixed(4)
const formatTime = (value: string) => new Date(value).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', hour12: false })
const formatDate = (value: string) => new Date(value).toLocaleDateString(undefined, { month: '2-digit', day: '2-digit' })
</script>
