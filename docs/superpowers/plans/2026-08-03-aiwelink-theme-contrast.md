# AIWeLink Theme Contrast Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the approved higher-contrast black-gold dark theme and saturated rose light theme across the Sub2API frontend, with restrained primary-button sheen and selected-state glow.

**Architecture:** Keep `src/styles/theme.css` as the runtime source of truth and keep Tailwind fallbacks synchronized for pre-CSS rendering and generated utilities. Express shared motion and glow in `src/style.css`, retain component-local active-tab rules where scoped CSS would override the shared rule, and migrate only branded public-page surfaces that currently bypass semantic depth tokens.

**Tech Stack:** Vue 3, TypeScript, Tailwind CSS, Vitest, CSS custom properties, Vite

---

### Task 1: Lock and implement the approved palette

**Files:**
- Modify: `frontend/src/styles/__tests__/themeTokens.spec.ts`
- Modify: `frontend/src/composables/__tests__/useThemePalette.spec.ts`
- Modify: `frontend/src/styles/theme.css`
- Modify: `frontend/tailwind.config.js`
- Modify: `frontend/src/composables/useThemePalette.ts`

- [ ] **Step 1: Write the failing exact-token and contrast tests**

Replace the old palette expectations in `themeTokens.spec.ts` with exact assertions for the approved semantic colors and add real WCAG contrast calculation:

```ts
function contrast(foreground: [number, number, number], background: [number, number, number]): number {
  const luminance = ([r, g, b]: [number, number, number]) => {
    const channels = [r, g, b].map((channel) => {
      const value = channel / 255
      return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
    })
    return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722
  }
  const a = luminance(foreground)
  const b = luminance(background)
  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05)
}

expect(light).toContain('--color-primary-500: 210 31 75')
expect(light).toContain('--color-accent-500: 244 189 56')
expect(light).toContain('--color-surface-muted: 237 241 243')
expect(light).toContain('--color-theme-border: 217 224 228')
expect(light).toContain('--color-theme-text: 32 42 49')
expect(light).toContain('--color-theme-muted: 99 113 122')

expect(dark).toContain('--color-primary-500: 255 194 71')
expect(dark).toContain('--color-accent-500: 226 49 92')
expect(dark).toContain('--color-on-primary: 24 16 5')
expect(dark).toContain('--color-canvas: 3 5 7')
expect(dark).toContain('--color-surface: 13 17 22')
expect(dark).toContain('--color-surface-muted: 23 29 36')
expect(dark).toContain('--color-theme-border: 48 57 69')
expect(dark).toContain('--color-theme-text: 250 247 240')
expect(dark).toContain('--color-theme-muted: 170 178 188')

expect(contrast([210, 31, 75], [255, 255, 255])).toBeGreaterThanOrEqual(4.5)
expect(contrast([255, 194, 71], [24, 16, 5])).toBeGreaterThanOrEqual(4.5)
expect(contrast([99, 113, 122], [245, 247, 248])).toBeGreaterThanOrEqual(4.5)
expect(contrast([170, 178, 188], [13, 17, 22])).toBeGreaterThanOrEqual(4.5)
```

Add a composable fallback assertion inside a Vue effect scope:

```ts
const scope = effectScope()
const palette = scope.run(() => useThemePalette())
expect(palette?.value.primary).toBe('#D21F4B')
expect(palette?.value.accent).toBe('#F4BD38')
expect(palette?.value.grid).toBe('#D9E0E4')
expect(palette?.value.text).toBe('#63717A')
scope.stop()
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/styles/__tests__/themeTokens.spec.ts src/composables/__tests__/useThemePalette.spec.ts
```

Expected: FAIL because the source still contains `186 54 80`, `233 190 115`, the blue-black canvas, and the old chart fallbacks.

- [ ] **Step 3: Implement the approved ramps and semantic tokens**

Update `theme.css` so light mode uses rose primary `210 31 75`, gold accent `244 189 56`, and the approved light semantics. Update `.dark` so the primary ramp is gold centered on `255 194 71`, the accent ramp is rose centered on `226 49 92`, the ink 700/800/900/950 levels are `48 57 69`, `23 29 36`, `13 17 22`, and `3 5 7`, and all approved dark semantic values are exact.

Synchronize every light-mode fallback in `tailwind.config.js`, including:

```js
500: token('primary-500', '210 31 75')
500: token('accent-500', '244 189 56')
700: token('ink-700', '48 57 69')
800: token('ink-800', '23 29 36')
900: token('ink-900', '13 17 22')
950: token('ink-950', '3 5 7')
canvas: token('canvas', '245 247 248')
'surface-muted': token('surface-muted', '237 241 243')
'theme-border': token('theme-border', '217 224 228')
'theme-text': token('theme-text', '32 42 49')
'theme-muted': token('theme-muted', '99 113 122')
```

Update `useThemePalette.ts` fallbacks to `#D21F4B`, `#F4BD38`, `#D9E0E4`, and `#63717A`.

- [ ] **Step 4: Run focused theme tests and verify GREEN**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/styles/__tests__/themeTokens.spec.ts src/__tests__/primaryForegrounds.spec.ts src/composables/__tests__/useThemePalette.spec.ts
```

Expected: 3 files PASS with no failed assertions.

- [ ] **Step 5: Commit palette work**

```powershell
git add frontend/src/styles/theme.css frontend/tailwind.config.js frontend/src/styles/__tests__/themeTokens.spec.ts frontend/src/composables/useThemePalette.ts frontend/src/composables/__tests__/useThemePalette.spec.ts
git commit -m "feat: strengthen AIWeLink theme contrast"
```

### Task 2: Add primary-only sheen and selected-state glow

**Files:**
- Create: `frontend/src/styles/__tests__/themeEffects.spec.ts`
- Modify: `frontend/src/style.css`
- Modify: `frontend/src/views/admin/SettingsView.vue`
- Modify: `frontend/src/views/admin/ChannelsView.vue`

- [ ] **Step 1: Write the failing effect contract**

Create `themeEffects.spec.ts` that reads the three style sources and asserts the concrete contract:

```ts
expect(shared).toContain('.btn-primary::before')
expect(shared).toContain('@keyframes primary-button-sheen')
expect(shared).toContain('.btn-primary:hover:not(:disabled)')
expect(shared).toContain('.btn-primary:disabled::before')
expect(shared).toContain('@media (prefers-reduced-motion: reduce)')
expect(shared).toContain('animation: none')
expect(shared).toContain('transform: none')
expect(shared).not.toMatch(/\.btn-(?:secondary|danger)::before/)
expect(shared).toContain('.sidebar-link-active')
expect(shared).toContain('.tab-active')
expect(settings).toContain('rgb(var(--color-primary-500) / 0.14)')
expect(channels).toContain('rgb(var(--color-primary-500) / 0.14)')
```

- [ ] **Step 2: Run the effect test and verify RED**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/styles/__tests__/themeEffects.spec.ts
```

Expected: FAIL because there is no sheen keyframe, disabled sheen rule, reduced-motion override, or localized tab glow.

- [ ] **Step 3: Implement the shared button and active-state effects**

In `style.css`, make `.btn-primary` a clipped, isolated positioning context with a static brand shadow. Add a low-opacity diagonal `::before` sheen, a `primary-button-sheen` keyframe, a one-pixel hover lift, and explicit disabled and reduced-motion behavior:

```css
.btn-primary {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  box-shadow:
    0 8px 22px rgb(var(--color-primary-500) / 0.22),
    0 1px 0 rgb(255 255 255 / 0.24) inset;
}

.btn-primary::before {
  position: absolute;
  top: -65%;
  bottom: -65%;
  left: -55%;
  width: 34%;
  pointer-events: none;
  content: "";
  background: linear-gradient(90deg, transparent, rgb(255 255 255 / 0.3), transparent);
  transform: skewX(-18deg);
  animation: primary-button-sheen 3.8s ease-in-out infinite;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 12px 30px rgb(var(--color-primary-500) / 0.34);
}

.btn-primary:disabled { box-shadow: none; }
.btn-primary:disabled::before { display: none; }

@keyframes primary-button-sheen {
  0%, 58% { left: -55%; opacity: 0; }
  66% { opacity: 0.7; }
  84%, 100% { left: 125%; opacity: 0; }
}
```

Give `.sidebar-link-active` and `.tab-active` an inset one-pixel highlight plus a restrained brand shadow. Add equivalent `0.14` primary-alpha shadows to scoped `.settings-tab-active` and `.channel-tab-active` rules so their local styles do not cancel the shared effect.

Extend the existing reduced-motion media query with:

```css
.btn-primary::before { animation: none; opacity: 0; }
.btn-primary:hover:not(:disabled) { transform: none; }
```

- [ ] **Step 4: Run effect and primary foreground tests and verify GREEN**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/styles/__tests__/themeEffects.spec.ts src/__tests__/primaryForegrounds.spec.ts
```

Expected: 2 files PASS; no effect selector targets secondary or destructive buttons.

- [ ] **Step 5: Commit interaction effects**

```powershell
git add frontend/src/style.css frontend/src/styles/__tests__/themeEffects.spec.ts frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/ChannelsView.vue
git commit -m "feat: add restrained brand glow effects"
```

### Task 3: Let public and authentication surfaces use semantic depth

**Files:**
- Create: `frontend/src/styles/__tests__/publicThemeSurfaces.spec.ts`
- Modify: `frontend/src/views/HomeView.vue`
- Verify: `frontend/src/components/layout/AuthLayout.vue`

- [ ] **Step 1: Write the failing public-surface contract**

Create `publicThemeSurfaces.spec.ts`:

```ts
const home = readFileSync(resolve(process.cwd(), 'src/views/HomeView.vue'), 'utf8')
const auth = readFileSync(resolve(process.cwd(), 'src/components/layout/AuthLayout.vue'), 'utf8')

expect(home).toContain('bg-canvas text-theme-text')
expect(home).toContain('from-canvas via-surface-muted to-canvas')
expect(home).toContain('border-theme-border')
expect(home).toContain('bg-surface/80')
expect(home).not.toMatch(/dark:(?:from|via|to)-dark-(?:900|950)/)
expect(auth).toContain('from-canvas via-surface-muted to-canvas')
```

- [ ] **Step 2: Run the public-surface test and verify RED**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/styles/__tests__/publicThemeSurfaces.spec.ts
```

Expected: FAIL because `HomeView.vue` still uses gray/ink page backgrounds and fixed dark gradient stops.

- [ ] **Step 3: Migrate branded homepage surfaces**

Update compact and full homepage wrappers, borders, muted text, translucent badges/cards, and footer to `canvas`, `surface`, `surface-muted`, `theme-border`, `theme-text`, and `theme-muted` utilities. Use `btn btn-primary` for the two principal homepage router links so they receive the shared foreground, shadow, sheen, disabled, focus, and motion behavior. Keep provider identity colors, layout, routing, copy, and decorations unchanged.

Representative replacements:

```vue
class="flex min-h-screen flex-col bg-canvas text-theme-text"
class="border-b border-theme-border px-4 py-4 sm:px-6"
class="relative flex min-h-screen flex-col overflow-hidden bg-gradient-to-br from-canvas via-surface-muted to-canvas"
class="inline-flex items-center gap-2.5 rounded-full border border-theme-border/70 bg-surface/80 ..."
class="group rounded-2xl border border-theme-border/70 bg-surface/70 ..."
class="btn btn-primary mt-8 min-h-10"
```

- [ ] **Step 4: Run all theme contract tests and verify GREEN**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/styles/__tests__/themeTokens.spec.ts src/styles/__tests__/themeEffects.spec.ts src/styles/__tests__/publicThemeSurfaces.spec.ts src/__tests__/primaryForegrounds.spec.ts src/composables/__tests__/useThemePalette.spec.ts
```

Expected: 5 files PASS.

- [ ] **Step 5: Commit public surface migration**

```powershell
git add frontend/src/views/HomeView.vue frontend/src/styles/__tests__/publicThemeSurfaces.spec.ts
git commit -m "feat: apply semantic depth to public surfaces"
```

### Task 4: Verify production quality and browser rendering

**Files:**
- Verify: `frontend/src/styles/theme.css`
- Verify: `frontend/src/style.css`
- Verify: `frontend/src/views/HomeView.vue`
- Verify: `frontend/src/components/layout/AuthLayout.vue`

- [ ] **Step 1: Run static quality gates**

Run:

```powershell
cd frontend
npm.cmd run lint:check
npm.cmd run typecheck
npm.cmd run build
```

Expected: all commands exit 0. Record any historical failure separately instead of modifying unrelated application behavior.

- [ ] **Step 2: Run the frontend test suite**

Run:

```powershell
cd frontend
npm.cmd run test:run
```

Expected: theme-related tests pass. If the known historical `GroupsView` mock failures remain, confirm they are unchanged and report them explicitly.

- [ ] **Step 3: Start the frontend and inspect real pages**

Run a Vite development server on an available localhost port. In desktop and mobile viewports, inspect light and dark home, login/auth, authenticated sidebar/tabs, a dense table, and a modal/dialog. Verify:

```text
- dark canvas is visibly deeper than surfaces and elevated surfaces
- gold/rose controls remain readable and non-neon
- only primary buttons show moving sheen
- active navigation/tabs have localized static glow without layout shift
- reduced-motion disables sheen and hover translation
- no text overflow, overlap, blank content, or console errors
```

- [ ] **Step 4: Review the final diff and commit any verification-only corrections**

Run:

```powershell
git diff --check
git status --short
git diff --stat HEAD~3..HEAD
```

Expected: no whitespace errors; only theme implementation files plus the forced-added plan/spec documents are part of this work. Existing `.gitignore` and deployment files remain unstaged.
