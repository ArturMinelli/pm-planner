## Context

See `proposal.md` for motivation. The desktop frontend uses a flex sidebar + main layout, a `grid-2` two-column planner at ≥900px, and slot fields with inline label/badge rows (`clockout-label-row`) and horizontal value rows (`clockout-values`). At ~800px the two-column grid and inline badge rows compete for space, causing label overflow in journey slot fields. Grilling confirmed: layout breaks (not input lag), overflow is in slot fields, fix via vertical stacking, scope is whole app, sidebar stays horizontal at narrow widths, success = no horizontal scroll.

## Goals / Non-Goals

**Goals:**

- Eliminate horizontal scroll and clipped text from 600px upward across all desktop pages.
- Reflow planner `grid-2` at ~800px (earlier than current 900px-only column rule).
- Stack slot-field label/badge rows and value rows at medium widths.
- Audit and fix overflow in settings, summary stats, page headers, and reminder chips.

**Non-Goals:**

- Changing suggestion algorithm, backend calls, or debounce behavior.
- New sidebar patterns (hamburger, icon collapse).
- Mobile-first redesign or touch gesture changes.
- Visual rebrand or color/typography overhaul.

## Decisions

### 1. Lower the two-column breakpoint to ~800px

**Choice:** Add a `@media (max-width: 800px)` rule (or adjust existing 760px/900px rules) so `grid-2` becomes single-column at the width where users report breakage.

**Alternatives:** Keep 900px breakpoint and only fix slot fields — rejected because the whole column feels cramped at 800px.

### 2. Slot fields: flex-wrap + column stack via CSS, minimal markup changes

**Choice:** Update `.clockout-label-row` to `flex-wrap: wrap` with `gap`, and `.exit-label-actions` to wrap badges. At medium width, set `.clockout-values` to `flex-direction: column` (already done at 760px — align breakpoints). Ensure `.field` and `.slot-field` have `min-width: 0`.

**Alternatives:** Shorter badge copy ("calc.") — rejected per user preference for stacking over abbreviation.

### 3. Badge row becomes a second line under the label

**Choice:** On wrap, label stays on first line; badges flow to second line within the label row container. Use `flex-basis: 100%` on `.exit-label-actions` at medium widths if needed to force badge row below label.

**Alternatives:** Move badges below the time value — rejected; badges are metadata about the slot, not the value.

### 4. App-wide overflow pass in `index.css` only where possible

**Choice:** Centralize responsive fixes in CSS. Only touch component markup where semantic structure blocks reflow (e.g., if a flex child lacks `min-width: 0`).

**Alternatives:** Per-component inline styles — rejected for consistency.

### 5. Component tests for structure, not pixel-perfect layout

**Choice:** Add tests asserting responsive class structure or CSS module hooks where practical; manual resize verification for visual polish.

## Risks / Trade-offs

- **[Breakpoint mismatch]** Existing 760px and 900px rules may conflict → Consolidate into a clear cascade: 800px single-column grid, 760px slot-field stack (merge or align).
- **[Taller slot fields]** Stacking increases vertical height → Acceptable trade-off for readability; journey list scrolls within main panel.
- **[Balance badge tooltip position]** Tooltip anchor may shift when badges wrap → Verify tooltip still appears above badge on wrap; adjust `clockout-tooltip` left offset if needed.

## Migration Plan

CSS-only change with optional minor markup tweaks. No data migration. Deploy with frontend build; rollback by reverting CSS/component diff.
