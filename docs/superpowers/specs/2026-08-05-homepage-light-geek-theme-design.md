# Homepage Light Geek Theme

**Date:** 2026-08-05

## Goal

Refine the AIwelink public homepage light mode into a restrained geek-style interface. The page should read primarily as black, white, and gray, while the animated particle network remains the only warm-colored visual layer.

## Approved Direction

- Apply this change only to the redesigned public homepage light mode.
- Keep the existing dark black-gold theme unchanged.
- Use near-white, black, and neutral gray as the light-mode content palette.
- Allow subtle grayscale gradients for depth; do not use colored gradients in content surfaces.
- Reserve warm gold and sparse rose exclusively for canvas particles and connection lines.
- Remove blue, gold, and rose emphasis from visible light-mode text, buttons, navigation, section indices, pricing, model names, focus outlines, and scroll indicators.

## Light Palette

- Canvas: a subtle grayscale gradient from `#ffffff` through `#f7f7f7` to `#ededed`.
- Primary text: `#101010`.
- Muted text: neutral medium gray with accessible contrast.
- Faint text: a lighter neutral gray for secondary metadata only.
- Primary surfaces and buttons: near-black with white text.
- Soft and hover surfaces: translucent black over the light canvas.
- Shadows: neutral black alpha only, without colored glow.

The gradient must remain quiet enough that the page still reads as a clean white interface. It may create spatial depth but must not resemble a decorative color wash.

## Content Treatment

- `AIwelink API` is entirely monochrome in light mode.
- The navigation wordmark, hero eyebrow, section indices, pricing equation, fact labels, model names, and final CTA eyebrow use grayscale theme tokens.
- Primary actions use a black-to-dark-gray treatment with white text.
- Secondary actions use a pale neutral surface.
- Existing layout, copy, typography scale, intro sequence, and section reveal timing remain unchanged.
- Content remains unframed; do not introduce card walls or large bordered containers.

## Particle Network

- Replace the light-mode blue particle palette with the warm palette previously associated with `API`.
- Use warm gold as the dominant light-mode particle and line color.
- Use rose sparingly on selected particles and their local connections so the field does not become a single flat color.
- Preserve current density, attraction, cursor connection, cluster detection, 40% burst threshold, impulse `8`, resize behavior, and reduced-motion behavior.
- Maintain enough contrast for the animation to remain visible over the grayscale canvas without competing with content.

## Accessibility And Motion

- Light-mode body text and controls must retain WCAG AA contrast.
- Keyboard focus remains visible using monochrome outlines or shadows.
- Theme switching must not flash the previous blue palette.
- Reduced-motion behavior remains a static network frame.

## Verification

- Component tests assert monochrome light-mode tokens and unchanged dark-mode gold tokens.
- Particle tests assert warm light-mode connections and retain dark-mode coverage.
- Homepage tests verify light-mode hard-coded rose and blue accents are absent from content components.
- Run homepage tests, type checking, lint, and production build.
- Inspect desktop and mobile light mode to confirm the grayscale gradient is subtle, text remains legible, the canvas is visible, and no warm color appears outside the canvas.

## Non-Goals

- Redesigning homepage layout or copy.
- Changing the startup intro, dark mode, authenticated pages, or global application theme tokens.
- Altering particle physics or explosion behavior.
