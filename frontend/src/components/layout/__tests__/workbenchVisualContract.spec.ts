import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const source = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')

const layoutSource = source('src/components/layout/AppLayout.vue')
const sidebarSource = source('src/components/layout/AppSidebar.vue')
const headerSource = source('src/components/layout/AppHeader.vue')
const sharedStyles = source('src/style.css')
const themeSource = source('src/styles/theme.css')

const dashboardSources = [
  source('src/views/user/DashboardView.vue'),
  source('src/components/user/dashboard/UserDashboardStats.vue'),
  source('src/components/user/dashboard/UserDashboardCharts.vue'),
  source('src/components/charts/TokenUsageTrend.vue'),
  source('src/components/user/dashboard/UserDashboardRecentUsage.vue'),
  source('src/components/user/dashboard/UserDashboardQuickActions.vue')
]

describe('authenticated workbench visual contract', () => {
  it('uses the compact divider-led shell', () => {
    expect(layoutSource).toContain('workbench-shell')
    expect(layoutSource).toContain("sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-56'")
    expect(layoutSource).not.toContain('bg-mesh-gradient')

    expect(headerSource).toContain('workbench-header')
    expect(headerSource).toContain('class="flex h-full items-center')
    expect(headerSource).not.toContain('class="glass')
    expect(headerSource).not.toContain('h-16')
    expect(sharedStyles).toMatch(/\.workbench-header\s*\{[^}]*height: 56px/)

    expect(sidebarSource).toContain("sidebarCollapsed ? 'w-16' : 'w-56'")
    expect(sidebarSource).not.toContain('shadow-glow')
    expect(sharedStyles).toContain('.sidebar-link-active::before')
  })

  it('defines accessible light and dark workbench colors', () => {
    expect(themeSource).toContain('--workbench-canvas: 247 247 247')
    expect(themeSource).toContain('--workbench-surface: 255 255 255')
    expect(themeSource).toContain('--workbench-text: 16 16 16')
    expect(themeSource).toContain('--workbench-accent-gold: 211 148 24')
    expect(themeSource).toContain('--workbench-accent-pink: 210 31 75')
    expect(themeSource).toContain('--workbench-canvas: 5 6 8')
    expect(themeSource).toContain('--workbench-accent-gold: 255 194 71')
  })

  it('renders the dashboard as continuous sections instead of cards', () => {
    const combined = dashboardSources.join('\n')

    expect(dashboardSources[0]).toContain('workbench-dashboard')
    dashboardSources.slice(1).forEach((component) => {
      expect(component).toMatch(/dashboard-section(?:-grid)?/)
    })
    expect(combined).not.toMatch(/class="[^"]*\bcard\b/)
    expect(combined).not.toContain('stat-card')
    expect(combined).not.toContain('rounded-2xl')
    expect(combined).not.toContain('shadow-card')
    expect(dashboardSources[4]).toContain("t('usage.time')")
    expect(dashboardSources[4]).not.toContain("t('dashboard.time')")
  })
})
