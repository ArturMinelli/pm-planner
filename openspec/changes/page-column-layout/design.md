## Context

See `proposal.md` for motivation. The desktop frontend uses a flex sidebar + main layout. The planner page wraps journey list and summary in `.grid-2`, currently `1fr / 0.82fr` with an 18rem minimum on the summary column at ≥801px. Settings uses `<Page narrow>` capped at 700px, forcing a single narrow column on wide screens. Grilling confirmed: fix column proportions app-wide (not overflow), journey-primary 60/40 on planner, settings two-column on wide screens, success = no dead space.

## Goals / Non-Goals

**Goals:**

- Journey list ~60%, summary ~40% on planner at desktop widths.
- Settings cards in a two-column grid at desktop widths; single column ≤800px.
- Remove large unused horizontal space inside the main panel at ~900px+.
- Align page headers with the body grid below.

**Non-Goals:**

- Changing suggestion algorithm, backend calls, or form behavior.
- New sidebar patterns or visual rebrand.
- Revisiting overflow/stacking rules from `planner-ui-polish` (keep those intact).
- Three-column settings or denser form grids inside cards.

## Decisions

### 1. Planner grid ratio: `1.5fr 1fr` (~60/40)

**Choice:** Replace `.grid-2` desktop rule `minmax(0, 1fr) minmax(18rem, 0.82fr)` with `minmax(0, 1.5fr) minmax(14rem, 1fr)`. The 1.5:1 ratio gives journey-primary weight; lowering summary min from 18rem to 14rem prevents the summary from stealing space at medium-wide widths.

**Alternatives:** CSS `grid-template-columns: 60% 40%` — rejected; fr units handle gap and min-width constraints better.

### 2. Widen settings page container

**Choice:** Remove `narrow` from `SettingsPage` (or use full `Page` width at 1080px max). Wrap the three settings cards in a new `.settings-grid` class: `display: grid; gap: 1rem` with `grid-template-columns: repeat(2, minmax(0, 1fr))` at ≥801px. Third card (reminders) spans full width or sits in column 1 row 2 — prefer natural flow: account + planner side by side, reminders full width below (2-col for first two only) OR all three in 2-col with reminders spanning both columns if tall.

**Layout choice:** Account | Planner defaults on row 1; Reminders spans full width on row 2 (reminders form is long). This avoids cramming three uneven cards.

**Alternatives:** Keep narrow page — rejected per grilling (dead space on wide screens).

### 3. Shared breakpoint at 801px

**Choice:** Reuse existing 801px breakpoint for both planner and settings grids so behavior stays consistent with `planner-ui-polish`.

**Alternatives:** New 900px breakpoint — rejected; would conflict with existing cascade.

### 4. Page header alignment via shared max-width

**Choice:** Both planner and settings use the same `Page` max-width (1080px). Planner date picker stays in `page-header-actions`; no structural change needed if body grid and header share the same `.page` container (already true). Verify visually that date picker right edge aligns with summary column right edge after ratio change.

**Alternatives:** CSS subgrid on header — rejected as overkill for two pages.

### 5. CSS-only where possible; minimal markup

**Choice:** Add `.settings-grid` wrapper in `SettingsPage.tsx` only. Planner needs no markup change — ratio is pure CSS on `.grid-2`.

**Alternatives:** Inline grid styles per page — rejected for consistency.

## Risks / Trade-offs

- **[Summary panel feels cramped at ~820px]** Lower min-width may squeeze stat rows → Summary already stacks stats; verify at 820–900px; bump min to 15rem if needed.
- **[Reminders card full-width feels unbalanced]** → Acceptable; reminders content is long; alternative is span-2 which is equivalent.
- **[Settings forms wider inputs]** Wider cards mean longer input lines → Improves use of space per grilling goal; fields already stretch.

## Migration Plan

CSS + one markup wrapper. No data migration. Rollback by reverting CSS and SettingsPage wrapper.

## Open Questions

None — grilling resolved column ratio, settings layout, scope, and success criteria.
