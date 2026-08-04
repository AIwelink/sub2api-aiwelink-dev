# AIWeLink Dual Theme Design

Date: 2026-08-01

## 1. Goal

Replace sub2api's current teal brand theme with a dual AIWeLink theme derived from the existing homepage palette. The light theme uses the approved C direction and the dark theme uses the approved B direction.

The theme applies to the entire frontend:

- authentication, setup, legal, not-found, and other public routes;
- user-facing pages and payment flows;
- admin pages, operational tools, tables, dialogs, and onboarding;
- charts, data visualizations, loading states, focus states, and responsive layouts.

The change is visual only. Existing theme selection, system preference detection, local persistence, routing, business behavior, copy, and backend contracts remain unchanged.

## 2. Approved Direction

### Light theme: C, rose refresh

- Canvas: cool gray-white.
- Surfaces: clean white.
- Primary action: AIWeLink rose.
- Secondary brand emphasis: muted gold.
- Text: ink blue.
- Sidebar: white, with a rose active state.

This mode remains restrained and operational. It should feel related to the homepage without turning the application into an illustrated or decorative interface.

### Dark theme: B, ink night

- Canvas: deep ink blue.
- Surfaces: layered ink blue.
- Primary action: AIWeLink gold.
- Secondary brand emphasis: rose.
- Text: moon white.
- Sidebar: the darkest ink surface, with a gold active indicator.

This mode carries the strongest homepage identity. It must retain the density, scanning speed, and contrast expected from a control panel.

## 3. Source Palette

The theme is derived from colors already present in the AIWeLink homepage:

| Role | Color | Source use |
| --- | --- | --- |
| Deep ink | `#08131C` | Homepage shell and scene background |
| Ink surface | `#0B1823` | Homepage hero surface |
| Moon white | `#F7F1E7` | Homepage foreground text |
| Gold | `#E9BE73` | Kicker, energy, and focus accents |
| Rose | `#BA3650` | Homepage entry action |
| Mist blue | `#9AACB5` | Supporting cool neutral |

The application does not reuse the homepage illustration, motion, serif display typography, or decorative effects. Only the brand color relationships are translated.

## 4. Semantic Token Architecture

Use semantic CSS custom properties as the single theme boundary. Tailwind color entries reference these variables with alpha support, so existing classes such as `bg-primary-500`, `text-primary-600`, and `dark:bg-dark-900` continue to work.

The light theme defines the default `:root` values. The existing `.dark` class overrides the same variables. No additional theme store or runtime theme service is introduced.

### Core light tokens

| Token | Value |
| --- | --- |
| `canvas` | `#F5F7F8` |
| `surface` | `#FFFFFF` |
| `surface-muted` | `#F0F3F4` |
| `border` | `#E4E8EA` |
| `text` | `#25333B` |
| `text-muted` | `#65727A` |
| `primary` | `#BA3650` |
| `primary-hover` | `#A52D45` |
| `on-primary` | `#FFFFFF` |
| `accent` | `#C7A25E` |

### Core dark tokens

| Token | Value |
| --- | --- |
| `canvas` | `#07121A` |
| `surface` | `#0B1823` |
| `surface-raised` | `#0D1D27` |
| `border` | `#263740` |
| `text` | `#F4EDE2` |
| `text-muted` | `#9AACB5` |
| `primary` | `#E9BE73` |
| `primary-hover` | `#D5A352` |
| `on-primary` | `#08131C` |
| `accent` | `#BA3650` |

The full `50` through `950` ramps are fixed as follows:

| Scale | Light primary: rose | Dark primary: gold | Ink neutral |
| --- | --- | --- | --- |
| `50` | `#FFF1F3` | `#FFF9EB` | `#F8FAFA` |
| `100` | `#FFE4E8` | `#FFF0C7` | `#EDF1F1` |
| `200` | `#FDCBD4` | `#FFE18A` | `#DCE4E6` |
| `300` | `#F9A8B8` | `#F7D58F` | `#BDC9CD` |
| `400` | `#E86C84` | `#F0CA87` | `#9AACB5` |
| `500` | `#BA3650` | `#E9BE73` | `#718690` |
| `600` | `#A52D45` | `#D5A352` | `#4B626D` |
| `700` | `#87263B` | `#B98131` | `#263740` |
| `800` | `#702337` | `#936123` | `#0D1D27` |
| `900` | `#602133` | `#653F20` | `#0B1823` |
| `950` | `#350E19` | `#3A200E` | `#07121A` |

Component code consumes semantic scale names and does not choose between rose and gold directly.

### Contrast requirements

The approved core pairs meet WCAG AA for normal text:

- white on light rose `#BA3650`: 5.63:1;
- deep ink `#08131C` on dark gold `#E9BE73`: 10.78:1;
- moon white `#F4EDE2` on dark canvas `#07121A`: 16.26:1;
- ink text `#25333B` on light canvas `#F5F7F8`: 12.10:1;
- mist blue `#9AACB5` on dark canvas `#07121A`: 8.05:1.

Primary foreground color must therefore be semantic. Existing white-on-primary combinations are migrated to an `on-primary` utility or component rule so dark gold controls never retain white text.

## 5. Component Rules

### Navigation and layout

- Light sidebar and header use white surfaces and neutral borders.
- Light active navigation uses a pale rose background with rose text.
- Dark sidebar uses the deepest ink surface.
- Dark active navigation uses a restrained gold tint or a narrow gold indicator, not a large bright fill.
- Main content remains unframed; existing individual cards, dialogs, and repeated items retain their current structural roles.

### Actions and controls

- Light primary buttons use rose with white text.
- Dark primary buttons use gold with ink text.
- Focus rings follow the mode's primary color and remain clearly visible on both canvas and raised surfaces.
- Inputs, switches, selected tabs, progress bars, links, and selection states consume the same primary tokens.
- Destructive, success, warning, and informational actions keep their semantic red, green, amber, and blue identities.

### Cards, tables, and dialogs

- Light surfaces stay white against the gray-white canvas.
- Dark surfaces use two ink levels to distinguish canvas, normal surface, and raised overlays without excessive borders.
- Hover and selected states use low-opacity semantic brand colors.
- Dense tables preserve current spacing and hierarchy; theme work does not alter row height, column behavior, or responsive overflow.

### Authentication and public pages

- Authentication, setup, legal, and not-found routes use the same dual token system.
- Existing background decorations are recolored from teal/cyan to the appropriate rose/gold/ink relationships.
- No homepage imagery, ornamental scene assets, or new marketing composition is added to application routes.

### Charts and visual data

- The first series uses the current mode's primary brand color.
- Subsequent series use rose, gold, mist blue, violet, green, and other distinguishable categorical colors.
- Categorical platform colors remain stable when color conveys identity rather than generic branding.
- Chart colors are resolved from CSS variables at render time so theme toggles update the visualization without a reload.

## 6. Migration Boundary

The current `primary-*` theme appears across approximately 147 frontend source files. The migration therefore changes the central color implementation first and only makes targeted component edits where semantics or contrast require them.

Required hard-coded migrations include:

- teal primary ramps, glows, gradients, mesh values, and animation shadows in `tailwind.config.js`;
- primary component styles and semantic foreground utilities in `src/style.css`;
- teal onboarding highlights and buttons in `src/styles/onboarding.css`;
- hard-coded teal chart series and primary chart defaults;
- hard-coded teal borders, shadows, and decorative values in `HomeView.vue`, `KeyUsageView.vue`, and admin surfaces;
- direct white-on-primary controls that need `on-primary` behavior.

Teal and cyan values that represent a model family, platform identity, data category, or status are not globally replaced. Each hard-coded occurrence is classified before modification.

## 7. Theme Behavior and Data Flow

The existing startup flow remains authoritative:

```text
page load
-> read localStorage theme preference
-> otherwise inspect prefers-color-scheme
-> toggle the existing html.dark class before Vue mounts
-> CSS variables resolve the correct light or dark palette
-> Tailwind utilities and charts consume those variables
```

The existing sidebar theme control continues to toggle the same `.dark` class and save the same preference. There is no migration of user settings and no new storage key.

If a CSS variable is unavailable, every Tailwind color reference includes a static fallback matching the light theme. This prevents an unstyled or transparent primary control during partial loading.

## 8. Accessibility and Responsive Behavior

- All text and interactive foreground/background pairs meet WCAG AA contrast.
- Focus-visible states remain present on every interactive control.
- Color is not the sole indicator for success, failure, selection, or status.
- Theme changes do not resize controls, charts, cards, or navigation.
- Desktop and mobile retain the same semantic hierarchy and theme identity.
- Reduced-motion behavior remains unchanged because the theme introduces no new motion.

## 9. Testing and Verification

Implementation verification covers:

1. Automated tests for light and dark token presence, theme-specific `on-primary`, and unchanged startup/persistence behavior.
2. Existing frontend unit and integration tests.
3. Type checking and production build.
4. A hard-coded brand-color audit that permits only documented categorical teal/cyan uses.
5. Visual checks in both themes at desktop and mobile widths for login, setup, user dashboard, admin dashboard, tables, forms, dialogs, payment, public/legal, and not-found routes.
6. Chart rendering checks before and after a live theme toggle.
7. Contrast checks for primary controls, navigation selection, links, inputs, focus rings, badges, and disabled states.

## 10. Acceptance Criteria

- Every frontend route uses C for light mode and B for dark mode.
- Light primary actions are rose with white foregrounds.
- Dark primary actions are gold with ink foregrounds.
- The dark canvas and surfaces use AIWeLink ink blues; primary text uses moon white.
- Existing theme toggle, system preference, and persistence behavior are unchanged.
- Business status colors and categorical identity colors retain their meaning.
- No generic teal remains as a brand primary color.
- No white-on-gold primary control remains.
- Layout, copy, routing, API behavior, and backend code are unchanged.
- Automated checks pass and representative desktop/mobile screenshots show no unreadable, overlapping, or blank content.

## 11. Non-goals

- Recreating the AIWeLink homepage scene inside sub2api.
- Changing typography, layout density, navigation structure, or component architecture beyond the theme boundary.
- Adding new theme choices or an end-user color picker.
- Changing backend behavior, API contracts, database state, or deployment configuration.
- Rebranding semantic status colors or third-party provider colors.
