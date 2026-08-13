<template>
  <section class="dashboard-section">
    <header class="dashboard-section-header flex-wrap">
      <div>
        <h2 class="text-sm font-semibold text-theme-text">{{ t('dashboard.modelDistribution') }}</h2>
        <p class="mt-0.5 text-[11px] text-theme-muted">{{ t('dashboard.timeRange') }}</p>
      </div>
      <div class="flex min-w-0 flex-1 flex-wrap items-center justify-end gap-2 sm:gap-3">
        <div class="flex items-center gap-2">
          <DateRangePicker :start-date="startDate" :end-date="endDate" @update:startDate="$emit('update:startDate', $event)" @update:endDate="$emit('update:endDate', $event)" @change="$emit('dateRangeChange', $event)" />
        </div>
        <button type="button" @click="$emit('refresh')" :disabled="loading" class="dashboard-tool-button">
          {{ t('common.refresh') }}
        </button>
        <div class="flex items-center gap-2">
          <span class="hidden text-[11px] font-medium text-theme-muted sm:inline">{{ t('dashboard.granularity') }}</span>
          <div class="w-24">
            <Select :model-value="granularity" :options="[{value:'day', label:t('dashboard.day')}, {value:'hour', label:t('dashboard.hour')}]" @update:model-value="$emit('update:granularity', $event)" @change="$emit('granularityChange')" />
          </div>
        </div>
      </div>
    </header>

    <div class="dashboard-chart-grid grid grid-cols-1 lg:grid-cols-[minmax(0,1.4fr)_minmax(360px,1fr)]">
      <TokenUsageTrend :trend-data="trend" :loading="loading" />

      <div class="relative min-w-0 overflow-hidden">
        <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center bg-white/80 dark:bg-[#0d1116]/80">
          <LoadingSpinner size="md" />
        </div>
        <div class="flex min-h-[260px] flex-col items-center gap-4 p-4 sm:flex-row sm:gap-5 md:p-5 lg:flex-col xl:flex-row">
          <div class="h-36 w-36 shrink-0 sm:h-40 sm:w-40">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div v-else class="flex h-full items-center justify-center text-xs text-theme-muted">{{ t('dashboard.noDataAvailable') }}</div>
          </div>
          <div class="max-h-52 w-full min-w-0 flex-1 overflow-auto">
            <table class="w-full text-[11px]">
              <thead>
                <tr class="font-mono text-[10px] text-theme-muted">
                  <th class="pb-2 text-left font-normal">{{ t('dashboard.model') }}</th>
                  <th class="pb-2 text-right font-normal">{{ t('dashboard.requests') }}</th>
                  <th class="pb-2 text-right font-normal">{{ t('dashboard.tokens') }}</th>
                  <th class="pb-2 text-right font-normal">{{ t('dashboard.actual') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(model, index) in models" :key="model.model" class="border-t border-[rgb(var(--workbench-border))] hover:bg-[rgb(var(--workbench-surface-hover))]">
                  <td class="max-w-[120px] truncate py-2 font-medium text-theme-text" :title="model.model"><span class="mr-2 inline-block h-1.5 w-1.5" :style="{ backgroundColor: modelColor(index) }"></span>{{ model.model }}</td>
                  <td class="py-2 text-right font-mono text-theme-muted">{{ formatNumber(model.requests) }}</td>
                  <td class="py-2 text-right font-mono text-theme-muted">{{ formatTokens(model.total_tokens) }}</td>
                  <td class="py-2 text-right font-mono text-emerald-600 dark:text-emerald-400" :title="`${t('dashboard.standard')}: $${formatCost(model.cost)}`">${{ formatCost(model.actual_cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import { useThemePalette } from '@/composables/useThemePalette'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler } from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ loading: boolean, startDate: string, endDate: string, granularity: string, trend: TrendDataPoint[], models: ModelStat[] }>()
defineEmits(['update:startDate', 'update:endDate', 'update:granularity', 'dateRangeChange', 'granularityChange', 'refresh'])
const { t } = useI18n()
const themePalette = useThemePalette()

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: props.models.map((_, index) => modelColor(index))
  }]
})

const modelColor = (index: number) => [
  themePalette.value.accent,
  themePalette.value.primary,
  ...themePalette.value.chartSeries.slice(2)
][index % themePalette.value.chartSeries.length]

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens`
      }
    }
  }
}
</script>
