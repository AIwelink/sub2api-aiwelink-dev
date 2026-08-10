# AIWeLink Dual Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply the approved C light theme and B dark theme to every sub2api frontend route while preserving existing theme behavior and business semantics.

**Architecture:** Introduce one CSS custom-property theme boundary consumed by Tailwind semantic colors. Keep the existing `html.dark` state flow, migrate primary foregrounds to an `on-primary` token, and expose the same tokens to Chart.js through one reactive palette composable.

**Tech Stack:** Vue 3, TypeScript, Tailwind CSS 3, Chart.js, Vitest, Vite

---

## File Map

- Create `frontend/src/styles/theme.css`: light and dark color token definitions.
- Create `frontend/src/styles/__tests__/themeTokens.spec.ts`: token and Tailwind contract tests.
- Modify `frontend/tailwind.config.js`: map primary, accent, ink, surface, glow, and gradients to CSS variables.
- Modify `frontend/src/main.ts`: load theme tokens before application styles.
- Modify `frontend/src/style.css`: consume surface tokens and `on-primary` in shared components.
- Create `frontend/src/__tests__/primaryForegrounds.spec.ts`: prevent white-on-gold primary controls.
- Create `frontend/src/composables/useThemePalette.ts`: resolve chart colors and observe `html.dark` changes.
- Create `frontend/src/composables/__tests__/useThemePalette.spec.ts`: verify token resolution and theme observation.
- Modify chart and dashboard files that currently hard-code the old teal brand color.
- Modify public, authentication, onboarding, usage, settings, and provider defaults that hard-code teal or white-on-primary combinations.
- Create `frontend/src/__tests__/legacyBrandColors.spec.ts`: reject the former teal brand literals while permitting semantic teal/cyan classes.

### Task 1: Establish the semantic theme boundary

**Files:**
- Create: `frontend/src/styles/theme.css`
- Create: `frontend/src/styles/__tests__/themeTokens.spec.ts`
- Modify: `frontend/tailwind.config.js`
- Modify: `frontend/src/main.ts`

- [ ] **Step 1: Write the failing token contract test**

Create a Vitest test that reads `theme.css` and `tailwind.config.js` as text and asserts the approved core values and variable-backed Tailwind mapping:

```ts
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const themeCss = readFileSync(fileURLToPath(new URL('../theme.css', import.meta.url)), 'utf8')
const tailwindConfig = readFileSync(fileURLToPath(new URL('../../../tailwind.config.js', import.meta.url)), 'utf8')

function block(selector: string) {
  const escaped = selector.replace('.', '\\.')
  return themeCss.match(new RegExp(`${escaped}\\s*\\{([\\s\\S]*?)\\}`))?.[1] ?? ''
}

describe('AIWeLink theme tokens', () => {
  it('defines the approved C light palette', () => {
    expect(block(':root')).toContain('--color-primary-500: 186 54 80')
    expect(block(':root')).toContain('--color-on-primary: 255 255 255')
    expect(block(':root')).toContain('--color-canvas: 245 247 248')
  })

  it('defines the approved B dark palette', () => {
    expect(block('.dark')).toContain('--color-primary-500: 233 190 115')
    expect(block('.dark')).toContain('--color-on-primary: 8 19 28')
    expect(block('.dark')).toContain('--color-canvas: 7 18 26')
  })

  it('maps Tailwind brand colors to CSS variables', () => {
    expect(tailwindConfig).toContain('var(--color-primary-500')
    expect(tailwindConfig).toContain("'on-primary'")
    expect(tailwindConfig).not.toContain("500: '#14b8a6'")
  })
})
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run from `frontend`:

```bash
npm run test:run -- src/styles/__tests__/themeTokens.spec.ts
```

Expected: FAIL because `theme.css` does not exist.

- [ ] **Step 3: Define complete light, dark, accent, ink, and surface tokens**

Create `theme.css` with the exact ramps from the approved design. Store RGB channels so Tailwind opacity modifiers continue to work:

```css
:root {
  --color-primary-50: 255 241 243;
  --color-primary-100: 255 228 232;
  --color-primary-200: 253 203 212;
  --color-primary-300: 249 168 184;
  --color-primary-400: 232 108 132;
  --color-primary-500: 186 54 80;
  --color-primary-600: 165 45 69;
  --color-primary-700: 135 38 59;
  --color-primary-800: 112 35 55;
  --color-primary-900: 96 33 51;
  --color-primary-950: 53 14 25;

  --color-accent-50: 255 249 235;
  --color-accent-100: 255 240 199;
  --color-accent-200: 255 225 138;
  --color-accent-300: 247 213 143;
  --color-accent-400: 240 202 135;
  --color-accent-500: 233 190 115;
  --color-accent-600: 213 163 82;
  --color-accent-700: 185 129 49;
  --color-accent-800: 147 97 35;
  --color-accent-900: 101 63 32;
  --color-accent-950: 58 32 14;

  --color-on-primary: 255 255 255;
  --color-canvas: 245 247 248;
  --color-surface: 255 255 255;
  --color-surface-muted: 240 243 244;
  --color-theme-border: 228 232 234;
  --color-theme-text: 37 51 59;
  --color-theme-muted: 101 114 122;
}

.dark {
  --color-primary-50: 255 249 235;
  --color-primary-100: 255 240 199;
  --color-primary-200: 255 225 138;
  --color-primary-300: 247 213 143;
  --color-primary-400: 240 202 135;
  --color-primary-500: 233 190 115;
  --color-primary-600: 213 163 82;
  --color-primary-700: 185 129 49;
  --color-primary-800: 147 97 35;
  --color-primary-900: 101 63 32;
  --color-primary-950: 58 32 14;

  --color-accent-50: 255 241 243;
  --color-accent-100: 255 228 232;
  --color-accent-200: 253 203 212;
  --color-accent-300: 249 168 184;
  --color-accent-400: 232 108 132;
  --color-accent-500: 186 54 80;
  --color-accent-600: 165 45 69;
  --color-accent-700: 135 38 59;
  --color-accent-800: 112 35 55;
  --color-accent-900: 96 33 51;
  --color-accent-950: 53 14 25;

  --color-on-primary: 8 19 28;
  --color-canvas: 7 18 26;
  --color-surface: 11 24 35;
  --color-surface-muted: 13 29 39;
  --color-theme-border: 38 55 64;
  --color-theme-text: 244 237 226;
  --color-theme-muted: 154 172 181;
}

:root {
  --color-ink-50: 248 250 250;
  --color-ink-100: 237 241 241;
  --color-ink-200: 220 228 230;
  --color-ink-300: 189 201 205;
  --color-ink-400: 154 172 181;
  --color-ink-500: 113 134 144;
  --color-ink-600: 75 98 109;
  --color-ink-700: 38 55 64;
  --color-ink-800: 13 29 39;
  --color-ink-900: 11 24 35;
  --color-ink-950: 7 18 26;
}
```

- [ ] **Step 4: Map Tailwind to the variables and import tokens before app styles**

Use one helper in `tailwind.config.js`:

```js
const token = (name, fallback) =>
  `rgb(var(--color-${name}, ${fallback}) / <alpha-value>)`
```

Map `primary`, `accent`, `dark`, `canvas`, `surface`, `surface-muted`, `theme-border`, `theme-text`, `theme-muted`, and `on-primary`. Replace teal literals in glow, gradient, mesh, and keyframes with `rgb(var(--color-primary-500) / alpha)`.

In `main.ts`, import in this order:

```ts
import './styles/theme.css'
import './style.css'
```

- [ ] **Step 5: Run the token test and build CSS**

Run:

```bash
npm run test:run -- src/styles/__tests__/themeTokens.spec.ts
npm run build
```

Expected: PASS and a successful Vite production build.

- [ ] **Step 6: Commit the token boundary**

```bash
git add frontend/src/styles/theme.css frontend/src/styles/__tests__/themeTokens.spec.ts frontend/tailwind.config.js frontend/src/main.ts
git commit -m "feat(frontend): add AIWeLink theme tokens"
```

### Task 2: Make primary foreground contrast semantic

**Files:**
- Create: `frontend/src/__tests__/primaryForegrounds.spec.ts`
- Modify: `frontend/src/style.css`
- Modify: `frontend/src/views/KeyUsageView.vue`
- Modify: `frontend/src/views/HomeView.vue`
- Modify: `frontend/src/views/NotFoundView.vue`
- Modify: `frontend/src/views/setup/SetupWizardView.vue`
- Modify: `frontend/src/views/user/RedeemView.vue`
- Modify: `frontend/src/components/auth/LoginAgreementPrompt.vue`
- Modify: `frontend/src/views/public/LegalDocumentView.vue`
- Modify: `frontend/src/views/admin/RiskControlView.vue`
- Modify: `frontend/src/views/admin/SettingsView.vue`
- Modify: `frontend/src/views/admin/orders/AdminPaymentDashboardView.vue`
- Modify: `frontend/src/components/account/AccountTestModal.vue`
- Modify: `frontend/src/components/account/AccountStatsModal.vue`
- Modify: `frontend/src/components/account/EditAccountModal.vue`
- Modify: `frontend/src/components/account/BulkEditAccountModal.vue`
- Modify: `frontend/src/components/account/CreateAccountModal.vue`
- Modify: `frontend/src/components/account/ModelWhitelistSelector.vue`
- Modify: `frontend/src/components/account/HeaderOverrideJsonTools.vue`
- Modify: `frontend/src/components/admin/account/AccountTestModal.vue`
- Modify: `frontend/src/components/admin/account/AccountStatsModal.vue`
- Modify: `frontend/src/components/admin/account/ScheduledTestsPanel.vue`
- Modify: `frontend/src/components/admin/user/UserAllowedGroupsModal.vue`
- Modify: `frontend/src/components/common/DateRangePicker.vue`
- Modify: `frontend/src/components/common/VersionBadge.vue`
- Modify: `frontend/src/components/layout/AppHeader.vue`
- Modify: `frontend/src/components/modelPlaza/PlazaFilterBar.vue`
- Modify: `frontend/src/components/modelPlaza/PlazaNavBar.vue`
- Modify: `frontend/src/components/payment/PaymentProviderDialog.vue`
- Modify: `frontend/src/components/payment/ProviderCard.vue`
- Modify: `frontend/src/components/user/profile/ProfileAvatarCard.vue`
- Modify: `frontend/src/components/user/profile/ProfileInfoCard.vue`
- Modify: `frontend/src/utils/platformColors.ts`

- [ ] **Step 1: Write the failing foreground audit**

Recursively scan Vue class attributes and TypeScript class strings. Report any class group that combines `bg-primary-*` or `from-primary-*` with `text-white`:

```ts
const primaryBackground = /(?:bg|from)-primary-(?:400|500|600|700)/
const whiteForeground = /text-white/

expect(offenders, offenders.join('\n')).toEqual([])
```

Also assert that `.btn-primary` applies `text-on-primary`.

- [ ] **Step 2: Run the audit and verify it lists current offenders**

Run:

```bash
npm run test:run -- src/__tests__/primaryForegrounds.spec.ts
```

Expected: FAIL with existing primary controls such as `KeyUsageView.vue` and `AppHeader.vue`.

- [ ] **Step 3: Add and apply the semantic foreground**

Change shared primary styling to:

```css
.btn-primary {
  @apply bg-gradient-to-r from-primary-500 to-primary-600;
  @apply text-on-primary shadow-md shadow-primary-500/25;
}
```

Replace direct primary-background `text-white` combinations with `text-on-primary`. Preserve `text-white` for red, green, blue, provider, and other non-primary controls.

Update the default platform button class to:

```ts
const BUTTON_DEFAULT = 'bg-primary-500 text-on-primary hover:bg-primary-600 dark:hover:bg-primary-400'
```

- [ ] **Step 4: Apply canvas and surface tokens to shared layout primitives**

Use `bg-canvas`, `text-theme-text`, `bg-surface`, `bg-surface-muted`, and `border-theme-border` in the body, application shell, sidebar, header glass layer, auth shell, cards, inputs, dropdowns, and dialogs. Retain status-specific and provider-specific colors.

- [ ] **Step 5: Run foreground, login, and sidebar tests**

Run:

```bash
npm run test:run -- src/__tests__/primaryForegrounds.spec.ts src/components/__tests__/LoginForm.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts
```

Expected: PASS.

- [ ] **Step 6: Commit contrast and shared surfaces**

```bash
git add frontend/src/__tests__/primaryForegrounds.spec.ts frontend/src/style.css frontend/src
git commit -m "feat(frontend): apply semantic theme foregrounds"
```

### Task 3: Make charts consume live theme colors

**Files:**
- Create: `frontend/src/composables/useThemePalette.ts`
- Create: `frontend/src/composables/__tests__/useThemePalette.spec.ts`
- Modify: `frontend/src/components/charts/ModelDistributionChart.vue`
- Modify: `frontend/src/components/charts/GroupDistributionChart.vue`
- Modify: `frontend/src/components/charts/EndpointDistributionChart.vue`
- Modify: `frontend/src/components/charts/TokenUsageTrend.vue`
- Modify: `frontend/src/views/admin/DashboardView.vue`
- Modify: `frontend/src/views/admin/ops/components/OpsSwitchRateTrendChart.vue`
- Modify: `frontend/src/views/KeyUsageView.vue`
- Modify: `frontend/src/components/charts/__tests__/ModelDistributionChart.spec.ts`
- Modify: `frontend/src/components/charts/__tests__/GroupDistributionChart.spec.ts`
- Modify: `frontend/src/views/admin/__tests__/DashboardView.spec.ts`

- [ ] **Step 1: Write failing resolver and observer tests**

Test a pure resolver and the class observer separately:

```ts
document.documentElement.style.setProperty('--color-primary-500', '186 54 80')
expect(readThemeColor('--color-primary-500', '#000000')).toBe('rgb(186, 54, 80)')

const onChange = vi.fn()
const stop = observeThemeChanges(onChange)
document.documentElement.classList.toggle('dark')
await vi.waitFor(() => expect(onChange).toHaveBeenCalled())
stop()
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run:

```bash
npm run test:run -- src/composables/__tests__/useThemePalette.spec.ts
```

Expected: FAIL because the composable does not exist.

- [ ] **Step 3: Implement the resolver and composable**

Export `readThemeColor`, `observeThemeChanges`, and `useThemePalette`. The composable increments a local revision when the root `class` changes and exposes a computed palette containing:

```ts
interface ThemePalette {
  primary: string
  primaryAlpha: string
  accent: string
  grid: string
  text: string
  tooltipSurface: string
  chartSeries: string[]
}
```

The first chart series is `primary`; later series remain categorical and distinguishable.

- [ ] **Step 4: Replace hard-coded chart brand colors**

Each listed chart imports `useThemePalette()` and makes its data/options computed values depend on the reactive palette. Replace `#14b8a6`, stale gray tooltip colors, and non-reactive `documentElement.classList.contains('dark')` computed values where they control theme styling.

Update chart expectations so the first brand series is the resolved primary token and theme changes produce a different primary series without remounting.

- [ ] **Step 5: Run chart tests**

Run:

```bash
npm run test:run -- src/composables/__tests__/useThemePalette.spec.ts src/components/charts/__tests__ src/views/admin/__tests__/DashboardView.spec.ts
```

Expected: PASS.

- [ ] **Step 6: Commit live chart theming**

```bash
git add frontend/src/composables/useThemePalette.ts frontend/src/composables/__tests__/useThemePalette.spec.ts frontend/src/components/charts frontend/src/views/admin/DashboardView.vue frontend/src/views/admin/ops/components/OpsSwitchRateTrendChart.vue frontend/src/views/KeyUsageView.vue
git commit -m "feat(frontend): theme charts with live palette"
```

### Task 4: Remove remaining legacy brand teal literals

**Files:**
- Create: `frontend/src/__tests__/legacyBrandColors.spec.ts`
- Modify: `frontend/src/styles/onboarding.css`
- Modify: `frontend/src/views/HomeView.vue`
- Modify: `frontend/src/views/KeyUsageView.vue`
- Modify: `frontend/src/views/admin/SettingsView.vue`
- Modify: `frontend/src/utils/platformColors.ts`

- [ ] **Step 1: Write the failing legacy-color audit**

Scan frontend source and Tailwind config for the exact former brand literals:

```ts
const legacyBrand = /#14b8a6|#0d9488|#0f766e|rgba?\(20,\s*184,\s*166/gi
expect(matches, matches.join('\n')).toEqual([])
```

Do not reject `teal-*` or `cyan-*` class names globally because some represent model families, providers, or data categories.

- [ ] **Step 2: Run the audit and verify current hard-coded files fail**

Run:

```bash
npm run test:run -- src/__tests__/legacyBrandColors.spec.ts
```

Expected: FAIL with onboarding, HomeView, KeyUsageView, charts, settings, and platform default matches.

- [ ] **Step 3: Replace visual brand literals with tokens**

Use CSS variable syntax in CSS and Vue style blocks:

```css
color: rgb(var(--color-primary-500));
box-shadow: 0 0 0 3px rgb(var(--color-primary-500) / 0.2);
```

Use the theme palette resolver for TypeScript and Chart.js values. Preserve categorical cyan/teal classes and the composite provider color `#06b6d4`.

- [ ] **Step 4: Run the audit and full unit suite**

Run:

```bash
npm run test:run -- src/__tests__/legacyBrandColors.spec.ts
npm run test:run
```

Expected: PASS.

- [ ] **Step 5: Commit the migration audit**

```bash
git add frontend/src/__tests__/legacyBrandColors.spec.ts frontend/src
git commit -m "refactor(frontend): remove legacy teal branding"
```

### Task 5: Verify the complete frontend

**Files:**
- Modify only files required by failures found during verification.

- [ ] **Step 1: Run static verification**

Run from `frontend`:

```bash
npm run lint:check
npm run typecheck
npm run test:run
npm run build
```

Expected: every command exits successfully.

- [ ] **Step 2: Start the Vite development server**

Run:

```bash
npm run dev -- --host 127.0.0.1
```

Use an available port and keep the process running for visual verification.

- [ ] **Step 3: Check the selected route matrix in both modes**

At desktop `1440x900` and mobile `390x844`, inspect public/login, setup if reachable, user dashboard, admin dashboard, a dense table, a form dialog, payment, legal, and not-found states. Confirm:

- C light mode uses white/gray surfaces and rose primary controls.
- B dark mode uses ink surfaces and gold primary controls.
- Gold controls use ink foregrounds.
- Charts change colors after toggling without a reload.
- No content overlaps, disappears, or becomes unreadable.

- [ ] **Step 4: Inspect console and rendered assets**

Confirm there are no Vue warnings, CSS parsing errors, missing assets, blank chart canvases, or horizontal overflow introduced by the theme.

- [ ] **Step 5: Re-run focused checks after any visual fixes**

Run:

```bash
npm run lint:check
npm run typecheck
npm run test:run -- src/styles/__tests__/themeTokens.spec.ts src/__tests__/primaryForegrounds.spec.ts src/__tests__/legacyBrandColors.spec.ts src/composables/__tests__/useThemePalette.spec.ts
npm run build
```

Expected: PASS.

- [ ] **Step 6: Commit final verification fixes**

```bash
git add frontend
git commit -m "fix(frontend): polish AIWeLink dual theme"
```
