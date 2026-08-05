## 1. Planner column ratio

- [x] 1.1 Update `.grid-2` desktop rule in `index.css` to `minmax(0, 1.5fr) minmax(14rem, 1fr)` (~60/40 journey-primary)
- [x] 1.2 Verify planner page at ~1000px: journey column wider than summary, no large dead space beside cards
- [x] 1.3 Verify single-column stack still works at ≤800px (unchanged from planner-ui-polish)

## 2. Settings two-column layout

- [x] 2.1 Remove `narrow` prop from `SettingsPage` so page uses full 1080px max-width
- [x] 2.2 Add `.settings-grid` wrapper around the three settings forms in `SettingsPage.tsx`
- [x] 2.3 Add `.settings-grid` CSS: two equal columns at ≥801px, single column at ≤800px
- [x] 2.4 Layout row 1: Account + Planner defaults side by side; row 2: Reminders spans full width (`grid-column: 1 / -1`)
- [x] 2.5 Verify settings page at ~1000px fills main panel without narrow floating column

## 3. Page header alignment

- [x] 3.1 Visually verify planner date picker aligns with summary column right edge at desktop width
- [x] 3.2 Adjust `page-header-row` spacing only if misalignment remains after grid ratio change

## 4. Tests

- [x] 4.1 Add or update `PlannerPage.test.tsx` to assert `.grid-2` structure is present
- [x] 4.2 Add settings page test asserting `.settings-grid` wrapper and two-column class at wide viewport (or structure test via component render)
- [x] 4.3 Run frontend test suite and fix any regressions

## 5. Validation

- [x] 5.1 Manual resize check: planner and settings at 820px, 1000px, 1200px — no dead space, smooth reflow
- [x] 5.2 Run `openspec validate --change page-column-layout`

### 5.1 CSS breakpoint verification (logical)

Breakpoint: `@media (min-width: 801px)` enables two-column rules; `≤800px` stacks single column (unchanged from planner-ui-polish).

| Viewport | Planner (`.grid-2`) | Settings (`.settings-grid`) |
|----------|---------------------|-----------------------------|
| **820px** | Two columns: `1.5fr / 1fr` with summary min 14rem (~60/40 journey-primary). Page fills `min(100%, 1080px)` — no narrow cap. | Two columns: Account + Planner side by side; Reminders spans `grid-column: 1 / -1`. |
| **1000px** | Same 60/40 ratio; ~1000px content width uses full main panel without dead bands beside cards. | Same 2-col + full-width reminders row; cards fill column width via `minmax(0, 1fr)`. |
| **1200px** | No additional breakpoint; grid scales proportionally within 1080px max page width. | Same; no narrow `Page` wrapper — full 1080px max-width. |

Header alignment: both pages share `.page` (`width: min(100%, 1080px)`); planner date picker in `page-header-actions` aligns with body grid boundary.
