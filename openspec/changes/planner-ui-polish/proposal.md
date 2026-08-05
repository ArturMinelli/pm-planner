## Why

The desktop app layout breaks when the window is resized to medium widths (~800px), especially where the planner two-column grid and slot-field label rows compete for horizontal space. Labels, badges, and action buttons in journey slot fields overflow and clip instead of reflowing, making the UI feel unpolished even though suggestion logic works. A focused layout pass is needed now so the app stays readable at any reasonable window size without horizontal scroll.

## What Changes

- **Responsive breakpoints**: Tune grid and field layouts so the planner two-column view reflows cleanly around ~800px; slot fields stack label, badges, and values vertically when space is tight.
- **Slot-field overflow fix**: Reflow `clockout-label-row` and `clockout-values` so badges (`registrada`, `calculada`, balance) wrap or stack instead of overflowing the field width.
- **App-wide layout polish**: Audit sidebar, page headers, settings forms, summary panel, and reminder chips for overflow and cramped spacing; apply consistent `min-width: 0`, flex-wrap, and stacking rules.
- **No horizontal scroll**: Ensure no clipped text or horizontal scrollbar at any reasonable desktop window size (sidebar keeps current horizontal nav at narrow widths).
- **No behavior changes**: Suggestion calculation, backend calls, and data flow remain unchanged — this is layout and presentation only.

## Capabilities

### New Capabilities

- `desktop-ui-layout`: Responsive layout rules and overflow-safe presentation for the desktop frontend, including planner slot fields, summary panel, settings, and shared chrome.

### Modified Capabilities

- `instant-planner-suggestions`: Slot-field presentation requirements updated to specify vertical stacking of labels/badges/values and overflow-safe badge rows at medium widths.

## Impact

- **Desktop frontend CSS**: `index.css` (breakpoints, slot-field, grid-2, stat rows, settings grids).
- **Planner components**: `PlannerSlotField.tsx`, `PlannerJourneyGroup.tsx`, `PlannerSummaryPanel.tsx` (markup tweaks if needed for stacking).
- **Shared UI**: `ui.tsx`, `PlannerDefaultsForm.tsx`, sidebar/nav layout.
- **Tests**: Component/layout tests asserting no overflow classes or responsive structure at key breakpoints.
- **No backend or API changes.**
