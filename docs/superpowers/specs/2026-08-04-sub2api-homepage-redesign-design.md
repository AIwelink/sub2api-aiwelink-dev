# Sub2API AIWeLink Public Homepage Redesign

**Date:** 2026-08-04

## Goal

Redesign the default Sub2API public homepage as a focused AIWeLink product experience for developers and small teams. The page must communicate one unified API, concrete usage scenarios, clear pricing, broad model access, and a direct route into the product.

The memorable interaction is a short AI-generated startup sequence: the page opens on a black brand screen, shows an AI thinking state while the homepage initializes in the background, and then appears as if an editor is generating its components.

## Scope And Existing Mode Priority

The redesign applies only to the default public `HomeView` experience at `/home`.

Existing mode priority remains unchanged:

1. Admin-provided custom HTML or URL content renders first when configured.
2. Compact home mode renders when enabled and no custom content exists.
3. The redesigned AIWeLink homepage renders only when neither override is active.

The change does not alter authenticated dashboards, admin pages, authentication routes, backend behavior, public settings contracts, routing, or theme persistence.

## Approved Direction

- Medium-length homepage: hero, use cases, pricing, model coverage, and final CTA.
- Full-viewport animated particle-link background inspired by the interaction principles in `Jaxon1216/homepage`.
- Black/gold night mode and restrained light mode, with saturated gold and rose accents.
- No card wall, large section containers, decorative bottom frames, or model carousel.
- Overall scale is restrained; the hero title is the only oversized type.
- AIWeLink brand text uses a fixed gold and does not change on theme toggle or hover.
- The startup sequence plays once per full document load or refresh. SPA navigation and in-page anchors do not replay it.

## Visual System

### Core colors

- Night canvas: `#050608`
- Night text: `#F6F3EB`
- Light canvas: `#F4F6F8`
- Light text: `#171A1F`
- Fixed AIWeLink gold: `#FFC648`
- Rose emphasis: `#EF3F72`
- Night muted text: `#A5AAB3`
- Light muted text: `#626B77`

The intro is always rendered on the night canvas, independent of the saved page theme. The resolved theme becomes visible during the homepage composition phase.

### Typography and spacing

- Use a strong display face for the hero and a compact monospaced face for indices, loading status, and generated component lines.
- Do not scale font size continuously with viewport width. Use fixed responsive breakpoints.
- Desktop content width is approximately `1040px`; mobile horizontal padding is `16px`.
- Desktop hero title is approximately `62px`; mobile is approximately `36px`.
- Sections use open space and typographic grouping instead of framed containers.

## Information Architecture

### Navigation

The fixed navigation contains:

- `AIWELINK` wordmark in fixed gold.
- Anchor links for `使用场景`, `价格`, and `模型覆盖`.
- Theme toggle.
- Registration or authenticated dashboard action using the existing authentication state and routes.

The navigation remains compact and does not introduce a separate marketing header.

### Hero

- Eyebrow: `ONE ENDPOINT · EVERY MODEL`.
- Primary title: `AIWELINK API`.
- `AIWELINK` remains fixed gold; `API` uses the rose accent.
- Supporting copy: `一个 API，连接所有主流模型。为开发者和小团队提供稳定、统一、透明的模型接入体验。`
- Primary action: `点击开始`.
- Secondary action: `了解接入方式`.

The title stays centered and is the strongest visual signal. Supporting copy is smaller and explains the offer instead of competing with it.

### 01 / Use Cases

Display three unframed scenarios:

- Coding: Codex and Claude Code.
- 科研与深度学习.
- Agent 开发接入.

Use large indices, short headings, and one supporting line. Do not place each scenario in a bordered card.

### 02 / Pricing

The primary pricing expression is `¥1 = $10`.

Supporting statements:

- 倍率低至 `¥0.1–0.2 / $1`.
- 不同模型按对应倍率扣减.
- 实际消耗以账单为准.

The equality is the visual anchor. The section remains typographic and unframed rather than becoming a pricing-card grid.

### 03 / Models

Heading: `多种模型，一个统一入口。`

Show only three static names:

- GPT
- Claude
- Gemini

Do not use a rotating carousel and do not describe the set as `三种模型`.

### Final CTA

Use a short, unframed final call to action that leads unauthenticated users to the existing registration or login flow and authenticated users to their dashboard. It repeats the gold primary action treatment without adding another decorative panel.

## Particle-Link Background

The background is a fixed full-viewport canvas behind all homepage sections.

- Approximately 120 particles on desktop and 45 on mobile.
- Nearby particles connect with low-opacity lines.
- Pointer proximity increases local line strength without blocking links or buttons.
- Gold is dominant; rose appears sparingly at selected nodes.
- The canvas uses `requestAnimationFrame`, caps device pixel ratio at `2`, and resizes without shifting content.
- Animation pauses while the document is hidden and resumes when visible.
- All listeners and animation frames are cleaned up when the homepage unmounts.

Under `prefers-reduced-motion: reduce`, render a static low-density network or a still first frame instead of continuous movement.

## Startup And Composition Sequence

### Playback rule

The sequence plays once per browser document lifetime:

- A full open or hard refresh resets and plays it.
- In-page anchor navigation does not replay it.
- Navigating away and returning through the SPA does not replay it.

Use module-lifetime state rather than `localStorage` or `sessionStorage`; storage would incorrectly suppress the animation after a refresh.

### State model

The sequence has three explicit states:

```text
preparing -> composing -> ready
```

The state machine owns the overlay, page visibility, scroll lock, and transition cleanup. It does not control business data or authentication.

### Phase 1: Preparing

Target duration: `0–1.4s`, with a minimum visible duration of approximately `1.2s`.

- Start on a pure black viewport.
- Center only `AIWELINK API`.
- Animate the gold and rose glyphs with a restrained character reveal or highlight pass.
- Show `AI 思考中` below the title.
- Add an indeterminate circular spinner and a thin flowing progress line.
- Do not show a percentage because the work is not a measurable download.
- Initialize the canvas, theme-resolved homepage, observers, and local assets behind the overlay.

Readiness means the default homepage DOM has mounted, Vue has completed the next render tick, the canvas has a non-zero size, and fonts have either resolved or reached a short bounded wait. The intro does not depend on an external AI or API request.

If readiness takes longer than the minimum duration, remain in the thinking state until the local page is ready. Never reveal a blank or partially sized homepage.

### Phase 2: Composing

Target duration: approximately `0.9–1.1s`.

The thinking UI transitions into a frameless editor-style output. It has a monospaced cursor and syntax-colored lines but no terminal window chrome or large card border.

Lines appear in order:

```text
<Navigation />
<Hero />
<UseCases />
<Pricing />
<Models />
```

Each generated line triggers the corresponding homepage layer to resolve behind the overlay. The visible first viewport builds from navigation to hero to particle network; below-fold sections enter their ready state without forcing the page to scroll.

The effect uses short opacity, mask, and vertical-position transitions. It must not simulate typing every character of large source files or make the visitor wait through verbose logs.

### Phase 3: Ready

- Fade the editor output and black overlay away.
- Release scroll lock.
- Start normal pointer response on the particle network.
- Preserve the resolved page theme and all normal navigation behavior.
- Focus remains where the browser placed it; the animation does not steal focus.

### Failure and reduced-motion behavior

- A defensive timeout removes the overlay if animation coordination fails after the homepage DOM is usable.
- If the canvas cannot initialize, reveal the homepage with a static background rather than blocking entry.
- With reduced motion enabled, show the brand briefly, skip the spinner rotation and generated-line sequence, then use a short opacity transition into the complete page.

## Component Boundaries

Implementation should separate the responsibilities even if `HomeView` remains the route-level owner:

- `HomepageIntro`: state presentation, status semantics, and composition-line animation.
- `ParticleNetwork`: canvas sizing, particles, pointer response, visibility pause, and cleanup.
- `HomepageNavigation`: anchors, theme control, and existing auth-aware destination.
- `HomepageHero`: brand hierarchy, supporting copy, and actions.
- `HomepageUseCases`, `HomepagePricing`, `HomepageModels`, and `HomepageFinalCta`: static content sections.
- A small intro-sequence composable or state controller coordinates timing and readiness without embedding canvas internals.

These components consume existing app-store settings and authentication state through explicit props or narrow computed values. No new global store is required.

## Accessibility And Responsive Behavior

- The thinking message uses `role="status"` with a concise accessible label.
- Decorative generated code and the canvas are hidden from assistive technology.
- The overlay has no interactive controls and never traps focus.
- Keyboard navigation begins normally after the overlay exits.
- Gold, rose, body text, and controls meet WCAG AA against their active backgrounds.
- Mobile keeps the title, status, and spinner centered with stable dimensions and no horizontal overflow.
- Generated lines fit the smallest supported viewport through wrapping or reduced fixed type sizes.
- Theme controls, buttons, and links retain visible focus states.

## Testing And Verification

Automated coverage should verify:

1. Custom-content and compact-home precedence remains unchanged.
2. The intro state order is `preparing -> composing -> ready`.
3. The minimum preparing duration and readiness gate both apply.
4. A second `HomeView` mount in the same document skips the intro.
5. A fresh document load starts the intro again.
6. Reduced-motion mode skips continuous and generated-line motion.
7. Failure cleanup cannot leave scroll locked or the homepage hidden.
8. Canvas listeners and animation frames are removed on unmount.
9. Auth-aware CTA destinations and the existing theme preference remain unchanged.

Browser verification must cover dark and light themes at desktop and mobile widths. Capture the preparing, composing, and ready states. Confirm that the canvas is nonblank, buttons remain clickable, text does not overlap, and the intro does not replay during anchor navigation or an SPA route return.

## Acceptance Criteria

- Every full open or hard refresh begins with a black `AIWELINK API` brand screen.
- `AI 思考中`, a circular indeterminate spinner, and a flowing progress line appear during background initialization.
- Approximately `1–2s` after start, the page begins composing in a concise AI editor-output style.
- The complete homepage is interactive after the overlay exits and is never exposed in an incomplete layout.
- Internal navigation does not replay the sequence.
- AIWeLink brand text remains fixed gold in both themes.
- The final homepage includes the approved hero, use cases, pricing, GPT/Claude/Gemini model coverage, and CTA.
- The page contains no model carousel, card wall, large section frames, or bottom container frame.
- Reduced-motion, slow initialization, and canvas failure all produce usable fallbacks.
- Existing custom home, compact home, authentication, routing, and public settings behavior remains intact.

## Non-Goals

- Calling an AI service to generate the homepage at runtime.
- Showing a real download percentage or blocking on nonessential network requests.
- Rebuilding authenticated product pages or authentication forms.
- Changing pricing calculations, billing contracts, model availability APIs, or backend configuration.
- Adding more homepage content before the current structure and interaction are approved in production.
