<template>
  <section class="dashboard-section">
    <header class="dashboard-section-header">
      <div>
        <h2 class="text-sm font-semibold text-theme-text">{{ t('dashboard.title') }}</h2>
        <p class="mt-0.5 text-[11px] text-theme-muted">{{ t('dashboard.welcomeMessage') }}</p>
      </div>
    </header>

    <div class="dashboard-metric-grid">
      <article v-if="!isSimple" class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="dollar" size="sm" />{{ t('dashboard.balance') }}</div>
        <p class="dashboard-metric-value dashboard-value-gold">${{ formatBalance(balance) }}</p>
        <p class="dashboard-metric-detail">{{ t('common.available') }}</p>
      </article>

      <article class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="key" size="sm" />{{ t('dashboard.apiKeys') }}</div>
        <p class="dashboard-metric-value">{{ stats?.total_api_keys || 0 }}</p>
        <p class="dashboard-metric-detail text-emerald-600 dark:text-emerald-400">{{ stats?.active_api_keys || 0 }} {{ t('common.active') }}</p>
      </article>

      <article class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="chart" size="sm" />{{ t('dashboard.todayRequests') }}</div>
        <p class="dashboard-metric-value">{{ formatNumber(stats?.today_requests || 0) }}</p>
        <p class="dashboard-metric-detail">{{ t('common.total') }}: {{ formatNumber(stats?.total_requests || 0) }}</p>
      </article>

      <article class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="dollar" size="sm" />{{ t('dashboard.todayCost') }}</div>
        <p class="dashboard-metric-value dashboard-value-pink" :title="t('dashboard.actual')">${{ formatCost(stats?.today_actual_cost || 0) }}</p>
        <p class="dashboard-metric-detail" :title="t('dashboard.standard')">{{ t('common.total') }} ${{ formatCost(stats?.total_actual_cost || 0) }} / ${{ formatCost(stats?.total_cost || 0) }}</p>
      </article>

      <article class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="cube" size="sm" />{{ t('dashboard.todayTokens') }}</div>
        <p class="dashboard-metric-value dashboard-value-gold">{{ formatTokens(stats?.today_tokens || 0) }}</p>
        <p class="dashboard-metric-detail">{{ t('dashboard.input') }} {{ formatTokens(stats?.today_input_tokens || 0) }} / {{ t('dashboard.output') }} {{ formatTokens(stats?.today_output_tokens || 0) }}</p>
      </article>

      <article class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="database" size="sm" />{{ t('dashboard.totalTokens') }}</div>
        <p class="dashboard-metric-value">{{ formatTokens(stats?.total_tokens || 0) }}</p>
        <p class="dashboard-metric-detail">{{ t('dashboard.input') }} {{ formatTokens(stats?.total_input_tokens || 0) }} / {{ t('dashboard.output') }} {{ formatTokens(stats?.total_output_tokens || 0) }}</p>
      </article>

      <article class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="bolt" size="sm" />{{ t('dashboard.performance') }}</div>
        <p class="dashboard-metric-value">{{ formatTokens(stats?.rpm || 0) }} <span class="text-[10px] font-normal text-theme-muted">RPM</span></p>
        <p class="dashboard-metric-detail">{{ formatTokens(stats?.tpm || 0) }} TPM</p>
      </article>

      <article class="dashboard-metric">
        <div class="dashboard-metric-label"><Icon name="clock" size="sm" />{{ t('dashboard.avgResponse') }}</div>
        <p class="dashboard-metric-value dashboard-value-pink">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
        <p class="dashboard-metric-detail">{{ t('dashboard.averageTime') }}</p>
      </article>
    </div>
  </section>

  <section v-if="!isSimple && platformCards.length > 0" class="dashboard-section">
    <header class="dashboard-section-header">
      <div>
        <h2 class="text-sm font-semibold text-theme-text">{{ t('dashboard.platformBreakdown') }}</h2>
        <p class="mt-0.5 text-[11px] text-theme-muted">{{ t('dashboard.platformCount', { count: sortedPlatforms.length }) }}</p>
      </div>
    </header>

    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4">
      <article
        v-for="item in platformCards"
        :key="item.platform"
        class="border-b border-[rgb(var(--workbench-border))] p-4 last:border-b-0 md:border-r md:[&:nth-child(even)]:border-r-0 xl:border-b-0 xl:[&:nth-child(even)]:border-r xl:last:border-r-0"
      >
        <div class="flex items-center justify-between gap-3">
          <span class="truncate text-xs font-semibold text-theme-text">
            {{ item.isOther ? t('dashboard.platformOther') : platformLabel(item.platform) }}
          </span>
          <span class="font-mono text-xs text-[rgb(var(--workbench-accent-pink))]" :title="t('dashboard.actual')">${{ formatCost(item.total_actual_cost) }}</span>
        </div>
        <dl class="mt-3 grid grid-cols-3 gap-3 text-[11px]">
          <div><dt class="text-theme-muted">{{ t('dashboard.todayCost') }}</dt><dd class="mt-1 font-mono text-theme-text">${{ formatCost(item.today_actual_cost) }}</dd></div>
          <div><dt class="text-theme-muted">{{ t('dashboard.requests') }}</dt><dd class="mt-1 font-mono text-theme-text">{{ item.total_requests > 0 ? formatNumber(item.total_requests) : '-' }}</dd></div>
          <div><dt class="text-theme-muted">{{ t('dashboard.tokens') }}</dt><dd class="mt-1 font-mono text-theme-text">{{ item.total_tokens > 0 ? formatTokens(item.total_tokens) : '-' }}</dd></div>
        </dl>

        <div v-if="hasAnyLimit(item.quota) && !item.isOther" class="mt-4 space-y-2 border-t border-[rgb(var(--workbench-border))] pt-3">
          <p class="font-mono text-[10px] uppercase text-theme-muted">{{ t('dashboard.platformQuota.title') }}</p>
          <template v-for="w in (['daily', 'weekly', 'monthly'] as const)" :key="w">
            <div v-if="quotaVal(item.quota, `${w}_limit_usd`) != null" class="space-y-1">
              <template v-if="(quotaVal(item.quota, `${w}_limit_usd`) as number) === 0">
                <div class="flex items-center justify-between text-[11px]"><span class="text-theme-muted">{{ t(`dashboard.platformQuota.${w}`) }}</span><span class="font-mono text-red-500">{{ t('dashboard.platformQuota.disabled') }}</span></div>
                <div class="h-0.5 w-full bg-red-500" />
              </template>
              <template v-else>
                <div class="flex items-center justify-between gap-2 text-[11px]"><span class="text-theme-muted">{{ t(`dashboard.platformQuota.${w}`) }}</span><span class="truncate font-mono text-theme-text">${{ formatUsd((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0) }} / ${{ formatUsd(quotaVal(item.quota, `${w}_limit_usd`) as number) }}</span></div>
                <div class="h-0.5 w-full bg-[rgb(var(--workbench-border))]"><div class="h-full transition-all" :class="quotaBarClass(calcPercent((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${w}_limit_usd`) as number))" :style="{ width: calcPercent((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${w}_limit_usd`) as number) + '%' }" /></div>
                <p v-if="quotaVal(item.quota, `${w}_window_resets_at`)" class="text-[10px] text-theme-muted">{{ t('dashboard.platformQuota.resetsAt', { time: formatResetTime(quotaVal(item.quota, `${w}_window_resets_at`) as string) }) }}</p>
              </template>
            </div>
          </template>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import type { PlatformQuotaItem } from '@/types'

interface FusedPlatformCard {
  platform: string
  total_actual_cost: number
  today_actual_cost: number
  total_requests: number
  total_tokens: number
  isOther?: boolean
  quota?: PlatformQuotaItem
}

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
  platformQuotas?: PlatformQuotaItem[] | null
}>()
const { t } = useI18n()

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity'
}

const platformLabel = (p: string) => PLATFORM_LABELS[p] ?? p

const sortedPlatforms = computed(() => {
  const list = props.stats?.by_platform ?? []
  return [...list].sort((a, b) => b.total_actual_cost - a.total_actual_cost)
})

// 处理"各平台之和 < 总值"的差值：后端按平台聚合时过滤了无法归属平台的行
// （group 与 account 都缺 platform）。这里把差值作为"其他"卡片显式展示，
// 避免 Row 1 总值与 Row 3 平台拆分加总对不上、用户困惑。
const OTHER_THRESHOLD = 0.0001
const platformCards = computed<FusedPlatformCard[]>(() => {
  // 建立 by_platform Map
  const byPlat = new Map<string, (typeof sortedPlatforms.value)[number]>()
  for (const item of props.stats?.by_platform ?? []) byPlat.set(item.platform, item)

  // 建立 quota Map
  const byQuota = new Map<string, PlatformQuotaItem>()
  for (const q of props.platformQuotas ?? []) byQuota.set(q.platform, q)

  // union 平台集合。后端 by_platform / quota 接口均不会返回 platform='__other__'，
  // 无需显式排除；__other__ 由下方差值补差逻辑单独追加。
  const platforms = new Set<string>([...byPlat.keys(), ...byQuota.keys()])

  const PLATFORM_ORDER = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
  const cards: FusedPlatformCard[] = []

  for (const p of platforms) {
    const stat = byPlat.get(p)
    cards.push({
      platform: p,
      total_actual_cost: stat?.total_actual_cost ?? 0,
      today_actual_cost: stat?.today_actual_cost ?? 0,
      total_requests: stat?.total_requests ?? 0,
      total_tokens: stat?.total_tokens ?? 0,
      quota: byQuota.get(p),
    })
  }

  // 排序：按 PLATFORM_ORDER，未知平台按名称排序
  cards.sort((a, b) => {
    const ai = PLATFORM_ORDER.indexOf(a.platform)
    const bi = PLATFORM_ORDER.indexOf(b.platform)
    if (ai === -1 && bi === -1) return a.platform.localeCompare(b.platform)
    if (ai === -1) return 1
    if (bi === -1) return -1
    return ai - bi
  })

  // __other__ 补差逻辑：只对 by_platform 有 usage 数据的总和计算
  const total = props.stats?.total_actual_cost ?? 0
  const today = props.stats?.today_actual_cost ?? 0
  const sumTotal = cards.reduce((s, c) => s + c.total_actual_cost, 0)
  const sumToday = cards.reduce((s, c) => s + c.today_actual_cost, 0)
  const diffTotal = Math.max(0, total - sumTotal)
  const diffToday = Math.max(0, today - sumToday)

  if (diffTotal > OTHER_THRESHOLD || diffToday > OTHER_THRESHOLD) {
    cards.push({
      platform: '__other__',
      total_actual_cost: diffTotal,
      today_actual_cost: diffToday,
      total_requests: 0,
      total_tokens: 0,
      isOther: true,
    })
  }

  return cards
})

// Quota helpers

type QuotaWindow = 'daily' | 'weekly' | 'monthly'
type QuotaField = `${QuotaWindow}_limit_usd` | `${QuotaWindow}_usage_usd` | `${QuotaWindow}_window_resets_at`

function quotaVal(q: PlatformQuotaItem | undefined, key: QuotaField): PlatformQuotaItem[QuotaField] {
  return q?.[key]
}

function hasAnyLimit(q: PlatformQuotaItem | undefined): boolean {
  if (!q) return false
  return q.daily_limit_usd != null || q.weekly_limit_usd != null || q.monthly_limit_usd != null
}

function calcPercent(usage: number, limit: number): number {
  if (!limit || limit <= 0) return 0
  return Math.min(100, Math.max(0, Math.round((usage / limit) * 100)))
}

function quotaBarClass(p: number): string {
  if (p >= 95) return 'bg-red-500'
  if (p >= 75) return 'bg-amber-500'
  return 'bg-green-500'
}

// 与 formatBalance 一致使用 Intl.NumberFormat 做半偶舍入，避免 toFixed 在不同 JS 引擎
// 下偶发截断而非四舍五入（与后端展示精度不一致）。
const usdFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})
function formatUsd(n: number): string {
  if (!Number.isFinite(n)) return '0.00'
  return usdFormatter.format(n)
}

function formatResetTime(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>
