# AIWeLink Theme Contrast and Saturation Design

**Date:** 2026-08-03

## Goal

Increase visual separation in dark mode and strengthen the rose and gold brand colors in both themes without turning the operational interface into a decorative or high-noise experience.

## Approved Direction

The design uses a neutral deep-black canvas rather than the previous blue-black canvas. Dark surfaces step upward through cool neutral grays so cards, inputs, sidebars, and dialogs remain visibly separate from the page background. Brand colors use a vivid, non-neon rose in light mode and gold in dark mode.

## Color System

### Light mode

- Canvas: `#F5F7F8`
- Surface: `#FFFFFF`
- Muted surface: `#EDF1F3`
- Border: `#D9E0E4`
- Primary rose: `#D21F4B`
- Accent gold: `#F4BD38`
- Primary foreground: `#FFFFFF`
- Text: `#202A31`
- Muted text: `#63717A`

The primary rose against white text has a contrast ratio of approximately 5.2:1.

### Dark mode

- Canvas: `#030507`
- Surface: `#0D1116`
- Muted/elevated surface: `#171D24`
- Border: `#303945`
- Primary gold: `#FFC247`
- Accent rose: `#E2315C`
- Primary foreground: `#181005`
- Text: `#FAF7F0`
- Muted text: `#AAB2BC`

The primary gold against the dark primary foreground has a contrast ratio of approximately 11.7:1. The canvas, base surface, elevated surface, and border form four visibly distinct depth levels.

## Token Mapping

The existing CSS variable boundary remains the source of truth. The primary and accent ramps are rebuilt around the approved 500 values. Semantic canvas, surface, muted surface, border, text, and muted text tokens receive the approved values.

Hard-coded dark background utilities in branded public and authentication surfaces are replaced with semantic theme utilities where they prevent the new depth system from taking effect. Provider identity colors and categorical chart colors remain unchanged.

## Glow and Motion

Primary command buttons receive a restrained brand-colored shadow at rest, a stronger shadow and small upward translation on hover, and a low-opacity diagonal sheen. The sheen never changes layout dimensions.

Current navigation and tab selections receive a thin accent border, inner highlight, and localized static glow. They do not receive a continuously moving sheen.

Secondary, destructive, and disabled controls do not receive brand glow. Under `prefers-reduced-motion: reduce`, moving sheen and translation are disabled while static contrast and shadows remain.

## Scope

- Shared theme tokens and Tailwind fallbacks
- Shared button, surface, input, table, dialog, and navigation primitives
- Public home and authentication surfaces that still use fixed dark colors
- Current sidebar and tab selection states
- Theme-aware charts through the existing palette composable

The change does not alter layout, routing, persistence, business logic, provider identity colors, or categorical status colors.

## Verification

- Contract tests for exact approved tokens
- Contrast tests for primary button foregrounds and semantic text pairs
- CSS contract tests for primary-only sheen, selected-state glow, and reduced-motion handling
- Existing theme, chart, lint, typecheck, and production build checks
- Browser screenshots for public, authentication, and authenticated layouts in light and dark modes at desktop and mobile widths
