# AIWeLink Homepage Approved Adjustments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the approved login-first homepage behavior, Chinese copy, and light-theme particle visibility without changing registration or upstream backend behavior.

**Architecture:** Keep the existing homepage component boundaries and computed destinations. Add regression coverage at the component and upstream-sync contract levels, then restore the smallest production diffs from the previously merged commits.

**Tech Stack:** Vue 3, TypeScript, Vue Test Utils, Vitest, Bash, pnpm

---

### Task 1: Establish The Focused Baseline

**Files:**
- Inspect: `frontend/package.json`
- Test: `frontend/src/components/home/__tests__/HomepageContent.spec.ts`
- Test: `frontend/src/components/home/__tests__/ParticleNetwork.spec.ts`
- Test: `deploy/tests/upstream-sync-contract-test.sh`

- [ ] **Step 1: Install the locked frontend dependencies**

Run from `frontend`:

```powershell
corepack pnpm install --frozen-lockfile
```

Expected: exit 0 without changing `pnpm-lock.yaml`.

- [ ] **Step 2: Run the existing focused component tests**

```powershell
corepack pnpm exec vitest run src/components/home/__tests__/HomepageContent.spec.ts src/components/home/__tests__/ParticleNetwork.spec.ts
```

Expected: PASS on the unmodified baseline.

- [ ] **Step 3: Run the existing sync contract**

```bash
bash deploy/tests/upstream-sync-contract-test.sh
```

Expected: `upstream sync contract passed: AIWeLink 0.1.177-1 (Sub2API 0.1.177)`.

### Task 2: Add Login-First And Copy Regression Tests

**Files:**
- Modify: `frontend/src/components/home/__tests__/HomepageContent.spec.ts`
- Modify: `deploy/tests/upstream-sync-contract-test.sh`

- [ ] **Step 1: Replace the guest registration expectation with all three login targets**

Mount `HomepageNavigation` with the existing stubs and assert the guest command text and route. Keep both authenticated destinations covered:

```ts
it('uses login for guest entry points and the supplied dashboard for authenticated users', () => {
  const guestContent = mountContent()
  const memberContent = mountContent(true, '/admin/dashboard')
  const guestNavigation = mountNavigation(false)
  const memberNavigation = mountNavigation(true, '/admin/dashboard')

  expect(guestNavigation.get('.navigation-command').text()).toBe('登录')
  expect(guestNavigation.get('.navigation-command').findComponent(RouterLinkStub).props('to')).toBe('/login')
  expect(guestContent.get('[data-testid="hero-primary"]').findComponent(RouterLinkStub).props('to')).toBe('/login')
  expect(guestContent.get('.final-command').findComponent(RouterLinkStub).props('to')).toBe('/login')
  expect(memberNavigation.get('.navigation-command').findComponent(RouterLinkStub).props('to')).toBe('/admin/dashboard')
  expect(memberContent.get('[data-testid="hero-primary"]').findComponent(RouterLinkStub).props('to')).toBe('/admin/dashboard')
  expect(memberContent.get('.final-command').findComponent(RouterLinkStub).props('to')).toBe('/admin/dashboard')
})
```

- [ ] **Step 2: Add approved copy and width assertions**

Extend the rendered-content test with:

```ts
expect(text).toContain('一个 API，连接所有主流模型。为开发者、团队、企业提供稳定、统一、透明的模型接入。')
expect(text).toContain('从 Coding 到 科研 与 Agent 接入')
expect(text).toContain('让模型成为工作流的一部分。')
expect(text).toContain('现在，开始你的创造')
expect(readFileSync(resolve('src/components/home/HomepageHero.vue'), 'utf8')).toContain('max-width: 700px')
```

- [ ] **Step 3: Add the upstream-sync login contract**

Define the three component paths and, for each component, require `: '/login'` and reject `: '/register'`. Require `HomepageNavigation.vue` to use `t('home.login')`.

- [ ] **Step 4: Verify RED**

Run:

```powershell
corepack pnpm exec vitest run src/components/home/__tests__/HomepageContent.spec.ts
```

Expected: FAIL because the current routes, copy, and width still use the old values.

Run:

```bash
bash deploy/tests/upstream-sync-contract-test.sh
```

Expected: FAIL because `HomepageNavigation.vue` still targets `/register`.

### Task 3: Add Particle Visibility Regression Tests

**Files:**
- Modify: `frontend/src/components/home/__tests__/ParticleNetwork.spec.ts`

- [ ] **Step 1: Add a fill-alpha helper**

```ts
function maxFillAlpha(color: string) {
  const alphas = fillStyles
    .filter((style) => style.startsWith(`rgba(${color},`))
    .map((style) => Number(style.match(/,\s*([\d.]+)\)$/)?.[1]))
  return Math.max(...alphas)
}
```

- [ ] **Step 2: Require the exact 30 percent light-theme increase**

Assert `0.18 * 1.3` for the strongest base link and amber fill, and `(0.18 + 0.08) * 1.3` for the rose fill. Keep the existing color-family assertions.

- [ ] **Step 3: Lock dark and pointer behavior**

Pair each pointer `lineTo` call with its `moveTo` coordinates and assert the exact formula `(1 - distance / 150) * .4 * 1.3` in light mode. In dark mode, assert the same formula without the multiplier and keep the existing connection and fill alpha values unchanged.

- [ ] **Step 4: Verify RED**

```powershell
corepack pnpm exec vitest run src/components/home/__tests__/ParticleNetwork.spec.ts
```

Expected: FAIL because the current light-theme alpha is not multiplied by 1.3.

After GREEN, temporarily change the multiplier to `1.2` and rerun this file. The regular-link and coordinate-derived pointer assertions must fail; restore `1.3` and rerun to PASS before committing.

### Task 4: Restore The Minimal Production Behavior

**Files:**
- Modify: `frontend/src/components/home/HomepageNavigation.vue`
- Modify: `frontend/src/components/home/HomepageHero.vue`
- Modify: `frontend/src/components/home/HomepageFinalCta.vue`
- Modify: `frontend/src/components/home/ParticleNetwork.vue`
- Modify: `frontend/src/i18n/locales/zh/landing.ts`

- [ ] **Step 1: Route all guest entry points to login**

Use this destination in all three components:

```ts
const destination = computed(() => props.authenticated ? props.dashboardPath : '/login')
```

In navigation, render `authenticated ? t('home.dashboard') : t('home.login')`.

- [ ] **Step 2: Restore the approved copy and width**

Set `.hero-description` to `max-width: 700px` and restore the four strings listed in the design document.

- [ ] **Step 3: Restore light-theme particle visibility**

Add:

```ts
const LIGHT_VISIBILITY_MULTIPLIER = 1.3

function themeAlpha(alpha: number, isDark: boolean, maximum = 1) {
  return Math.min(maximum, isDark ? alpha : alpha * LIGHT_VISIBILITY_MULTIPLIER)
}
```

Apply `themeAlpha` to regular connections, pointer connections, amber particles, and rose particles while preserving `.64` and `.72` particle caps.

- [ ] **Step 4: Verify GREEN**

```powershell
corepack pnpm exec vitest run src/components/home/__tests__/HomepageContent.spec.ts src/components/home/__tests__/ParticleNetwork.spec.ts
```

Expected: both test files PASS.

```bash
bash deploy/tests/upstream-sync-contract-test.sh
```

Expected: sync contract PASS.

### Task 5: Validate And Publish

**Files:**
- Verify: all files changed by Tasks 2-4

- [ ] **Step 1: Run focused static checks**

```powershell
corepack pnpm run typecheck
corepack pnpm exec eslint src/components/home/HomepageNavigation.vue src/components/home/HomepageHero.vue src/components/home/HomepageFinalCta.vue src/components/home/ParticleNetwork.vue src/components/home/__tests__/HomepageContent.spec.ts src/components/home/__tests__/ParticleNetwork.spec.ts src/i18n/locales/zh/landing.ts
git diff --check
```

Expected: all commands exit 0.

- [ ] **Step 2: Run browser QA**

Start the Vite server and verify the flow `/home -> navigation, Hero, final CTA -> /login` at desktop and mobile widths. Confirm meaningful content, no framework overlay, no relevant console errors, and screenshot evidence.

- [ ] **Step 3: Request independent review**

Review the complete diff against the design. Fix every Critical or Important finding and rerun affected checks.

- [ ] **Step 4: Commit and publish**

```powershell
git add frontend/src/components/home/HomepageNavigation.vue frontend/src/components/home/HomepageHero.vue frontend/src/components/home/HomepageFinalCta.vue frontend/src/components/home/ParticleNetwork.vue frontend/src/components/home/__tests__/HomepageContent.spec.ts frontend/src/components/home/__tests__/ParticleNetwork.spec.ts frontend/src/i18n/locales/zh/landing.ts deploy/tests/upstream-sync-contract-test.sh
git add -f docs/superpowers/specs/2026-08-17-homepage-approved-adjustments-design.md docs/superpowers/plans/2026-08-17-homepage-approved-adjustments.md
git commit -m "fix(home): restore login-first homepage adjustments"
git push -u origin codex/restore-homepage-approved-adjustments
```

Create a ready PR with base `aiwelink-dev`. Do not add reviewers, request approval, enable auto-merge, or contact `Attillo123`.

- [ ] **Step 5: Confirm required GitHub checks**

Wait for `frontend-tests`, `frontend-checks`, and `ci-gate`. If any fails, inspect its logs, fix the root cause on the same branch, rerun focused local verification, and push the follow-up commit.
