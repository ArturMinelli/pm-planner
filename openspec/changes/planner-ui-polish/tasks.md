## 1. Breakpoint consolidation

- [x] 1.1 Audit existing `@media` rules in `index.css` (760px, 900px) and align to ~800px single-column `grid-2` breakpoint
- [x] 1.2 Ensure `.page`, `.main`, and `.grid-2` children have `min-width: 0` to prevent flex/grid overflow

## 2. Slot field overflow fix

- [x] 2.1 Update `.clockout-label-row` and `.exit-label-actions` to `flex-wrap` with consistent gap; force badge row below label at medium width if needed (`flex-basis: 100%`)
- [x] 2.2 Align `.clockout-values` column stack breakpoint with the new medium breakpoint; ensure calculator button and alternative time are full-width when stacked
- [x] 2.3 Verify balance badge tooltip positioning when badges wrap; adjust `.clockout-tooltip` offset if clipped
- [x] 2.4 Add or update `PlannerSlotField` component test asserting label/badge row structure supports wrapping

## 3. Planner page layout

- [x] 3.1 Confirm `grid-2` stacks journey list above summary at ~800px without horizontal scroll
- [x] 3.2 Fix `.journey-group-fields` grid `minmax` values if fields still overflow at medium width
- [x] 3.3 Verify `PlannerSummaryPanel` stat rows wrap labels without clipping at medium width

## 4. App-wide overflow audit

- [x] 4.1 Fix settings page long labels (`PlannerDefaultsForm`, check grids) with wrap-safe field styles
- [x] 4.2 Fix page header date picker row wrapping in `.page-header-row` at medium width
- [x] 4.3 Fix reminder lead builder grid and chip overflow at narrow widths
- [x] 4.4 Confirm sidebar horizontal nav still works and does not cause page-level horizontal scroll

## 5. Verification

- [x] 5.1 Manually resize desktop window at 600px, 800px, and 1000px on planner and settings pages — confirm no horizontal scroll or clipped text
- [x] 5.2 Run frontend test suite (`npm test` in `desktop/frontend`)
