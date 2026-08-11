# AIwelink Clean Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the authenticated shell and user dashboard card-heavy UI with the compact, divider-led AIwelink workbench defined in the approved concept while preserving every existing route, permission, API request, prop, emit, and feature flag.

**Architecture:** Add workbench-specific semantic classes and tokens without changing the legacy business component system globally. `AppLayout`, `AppSidebar`, and `AppHeader` define the shared shell; the existing dashboard child components retain data ownership and Chart.js behavior while their templates become continuous work surfaces.

**Tech Stack:** Vue 3, TypeScript, Pinia, Tailwind CSS, Chart.js/vue-chartjs, Vitest, Vue Test Utils, Playwright fallback for visual QA.

---

### Task 1: Lock the visual contract with failing tests

**Files:**
- Create: `frontend/src/components/layout/__tests__/workbenchVisualContract.spec.ts`
- Modify: `frontend/src/components/__tests__/Dashboard.spec.ts`

- [ ] **Step 1: Add shell source-contract tests**

Read the four shell/theme source files and assert the new semantic hooks exist while legacy mesh, glass, 256px/72px shell dimensions, glow logo treatment, and shadowed active navigation are absent. The assertions must include:

```ts
expect(layoutSource).toContain('workbench-shell')
expect(layoutSource).not.toContain('bg-mesh-gradient')
expect(headerSource).toContain('workbench-header')
expect(headerSource).not.toContain('class="glass')
expect(sidebarSource).toContain("sidebarCollapsed ? 'w-16' : 'w-56'")
expect(sidebarSource).not.toContain('shadow-glow')
expect(themeSource).toContain('--workbench-canvas: 247 247 247')
expect(themeSource).toContain('--workbench-accent-gold: 211 148 24')
```

- [ ] **Step 2: Add dashboard component-contract tests**

Read the dashboard view and four dashboard component sources. Assert that the view contains `workbench-dashboard`, that each child uses a `dashboard-section` or `dashboard-section-grid` surface, and that the dashboard sources no longer contain the shared `card`, `stat-card`, `rounded-2xl`, or `shadow-card` classes.

- [ ] **Step 3: Run the new tests and confirm RED**

Run:

```powershell
npx --yes pnpm@9.15.9 exec vitest run src/components/layout/__tests__/workbenchVisualContract.spec.ts src/components/__tests__/Dashboard.spec.ts
```

Expected: FAIL because the semantic workbench hooks are not yet present and legacy card/glass/mesh hooks remain.

- [ ] **Step 4: Commit the failing contract tests**

```powershell
git add frontend/src/components/layout/__tests__/workbenchVisualContract.spec.ts frontend/src/components/__tests__/Dashboard.spec.ts
git commit -m "test: define clean workbench visual contract"
```

### Task 2: Build the compact shared workbench shell

**Files:**
- Modify: `frontend/src/styles/theme.css`
- Modify: `frontend/src/style.css`
- Modify: `frontend/src/components/layout/AppLayout.vue`
- Modify: `frontend/src/components/layout/AppSidebar.vue`
- Modify: `frontend/src/components/layout/AppHeader.vue`
- Test: `frontend/src/components/layout/__tests__/workbenchVisualContract.spec.ts`
- Test: `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`
- Test: `frontend/src/styles/__tests__/themeTokens.spec.ts`
- Test: `frontend/src/styles/__tests__/themeEffects.spec.ts`

- [ ] **Step 1: Add scoped workbench tokens**

Define light and dark variables on `:root` and `.dark`:

```css
--workbench-canvas: 247 247 247;
--workbench-surface: 255 255 255;
--workbench-surface-hover: 239 239 239;
--workbench-text: 16 16 16;
--workbench-muted: 95 95 95;
--workbench-border: 222 222 222;
--workbench-border-strong: 200 200 200;
--workbench-accent-gold: 211 148 24;
--workbench-accent-pink: 210 31 75;
```

Dark values are `5 6 8`, `13 17 22`, `23 29 36`, `246 243 235`, `165 170 179`, `48 57 69`, `255 194 71`, and `226 49 92` respectively.

- [ ] **Step 2: Add semantic shell classes**

Implement `workbench-shell`, `workbench-main`, `workbench-content`, `workbench-header`, `workbench-header-action`, `workbench-balance`, `workbench-user-trigger`, `dashboard-section`, `dashboard-section-header`, and keyboard focus states. Keep radii at 0-6px except avatars and actual floating menus; add reduced-motion handling.

- [ ] **Step 3: Flatten `AppLayout`**

Remove the mesh decoration, apply `workbench-shell`, change desktop offsets to `lg:ml-16` and `lg:ml-56`, and let `workbench-content` own responsive padding and the 1440px maximum content width.

- [ ] **Step 4: Compact `AppSidebar`**

Change expanded/collapsed widths to 224px/64px, logo/header height to 56px, navigation icon size to 18px, and navigation rows to 36px. Remove logo glow/rounded base and active-item shadow; keep a 2px warm accent rail, subtle hover fill, existing mobile drawer, scroll restoration, menus, feature gates, theme action, and collapse action.

- [ ] **Step 5: Flatten `AppHeader`**

Change height to 56px and use an opaque divided surface. Preserve documentation, model plaza, language, announcement, balance, user menu, onboarding, and logout behavior. Present balance as compact text and the user as an avatar/name row without gradient or pill framing.

- [ ] **Step 6: Run shell tests and confirm GREEN**

Run:

```powershell
npx --yes pnpm@9.15.9 exec vitest run src/components/layout/__tests__/workbenchVisualContract.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts src/styles/__tests__/themeTokens.spec.ts src/styles/__tests__/themeEffects.spec.ts
```

Expected: PASS with no unhandled errors.

- [ ] **Step 7: Commit the shared shell**

```powershell
git add frontend/src/styles/theme.css frontend/src/style.css frontend/src/components/layout/AppLayout.vue frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/AppHeader.vue
git commit -m "feat: redesign authenticated workbench shell"
```

### Task 3: Recompose the user dashboard as continuous work surfaces

**Files:**
- Modify: `frontend/src/views/user/DashboardView.vue`
- Modify: `frontend/src/components/user/dashboard/UserDashboardStats.vue`
- Modify: `frontend/src/components/user/dashboard/UserDashboardCharts.vue`
- Modify: `frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue`
- Modify: `frontend/src/components/user/dashboard/UserDashboardQuickActions.vue`
- Test: `frontend/src/components/__tests__/Dashboard.spec.ts`
- Test: `frontend/src/components/layout/__tests__/workbenchVisualContract.spec.ts`

- [ ] **Step 1: Create the dashboard page grid**

Use `workbench-dashboard` as the page root. Keep the same load orchestration and child props/events, place stats and charts as full-width sections, then use a 2/3 + 1/3 divider-led grid for recent usage and quick actions that stacks on narrow screens.

- [ ] **Step 2: Convert statistics to a metric band**

Render the eight current values in a responsive 4-column band with shared vertical/horizontal dividers. Keep platform quota aggregation and status semantics; replace colored icon tiles with 16-18px line icons and reserve gold/pink for main values and status marks.

- [ ] **Step 3: Open the chart workspace**

Keep date range, granularity, refresh emits, datasets, tooltips, responsive options, model aggregation, and empty/loading behavior. Remove nested card wrappers; give the trend chart the wider track and the model distribution/table the narrower track, separated by 1px rules.

- [ ] **Step 4: Flatten recent usage and quick actions**

Present recent usage as divided rows retaining model, endpoint, token, fee, status, and time information. Present quick actions as compact icon/title/description/arrow command rows and preserve the existing route destinations.

- [ ] **Step 5: Run dashboard tests and confirm GREEN**

Run:

```powershell
npx --yes pnpm@9.15.9 exec vitest run src/components/__tests__/Dashboard.spec.ts src/components/layout/__tests__/workbenchVisualContract.spec.ts
```

Expected: PASS with current data behavior and the new visual contract intact.

- [ ] **Step 6: Commit the dashboard redesign**

```powershell
git add frontend/src/views/user/DashboardView.vue frontend/src/components/user/dashboard/UserDashboardStats.vue frontend/src/components/user/dashboard/UserDashboardCharts.vue frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue frontend/src/components/user/dashboard/UserDashboardQuickActions.vue
git commit -m "feat: redesign user dashboard workspace"
```

### Task 4: Verify behavior, rendering, and delivery

**Files:**
- Modify only when verification finds an in-scope defect in the files listed above.

- [ ] **Step 1: Run focused regression tests**

```powershell
npx --yes pnpm@9.15.9 exec vitest run src/components/__tests__/Dashboard.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts src/components/layout/__tests__/workbenchVisualContract.spec.ts src/styles/__tests__/themeTokens.spec.ts src/styles/__tests__/themeEffects.spec.ts
```

Expected: all focused tests PASS without unhandled errors.

- [ ] **Step 2: Run static and production checks**

```powershell
npx --yes pnpm@9.15.9 run typecheck
npx --yes pnpm@9.15.9 run lint:check
npx --yes pnpm@9.15.9 run build
```

Expected: all commands exit 0. Any unrelated baseline failure must be recorded with its exact source and must not be hidden by changing unrelated code.

- [ ] **Step 3: Start the frontend and run browser QA**

Start Vite on a free localhost port. Because the in-app browser control skill's required Node REPL tool is unavailable in this environment, use the bundled Playwright runtime to capture 1440x900 light, 1440x900 dark, and 390x844 mobile screenshots. Check console errors, overflow, sidebar collapse/drawer, theme switch, user menu, dashboard refresh, loading, and empty states.

- [ ] **Step 4: Compare production screenshots with the approved concept**

Inspect all screenshots and verify compact 224px/64px navigation, 56px header, continuous surfaces, readable gold in light mode, high-contrast dark mode, no nested cards, stable chart dimensions, and no text overlap.

- [ ] **Step 5: Push and open a draft PR**

```powershell
git status --short
git push -u origin feature/workbench-clean-ui
$prBodyPath = Join-Path $env:TEMP 'aiwelink-workbench-clean-ui-pr.md'
gh pr create --draft --base aiwelink-dev --head feature/workbench-clean-ui --title "feat: redesign authenticated workbench" --body-file $prBodyPath
```

Before the command, create `$prBodyPath` with a Summary section covering the compact shell and continuous dashboard surfaces, a Scope section stating that APIs and deployment are unchanged, a Testing section containing exact command results, a Known Baseline Issue section naming the unrelated `GroupsView` mock omission if still present, a Screenshots section, and a Rollback section that instructs reverting the feature commits.
