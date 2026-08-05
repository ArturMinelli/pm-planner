## 1. Baseline measurement

- [x] 1.1 Run the planner in dev mode with a two-journey day where Saída 2 is calculated and a balance adjustment applies, reproducing the overflow
- [x] 1.2 Measure the rendered inline size of `.slot-field` at 720px, 980px, and 1280px window widths and record the three values in `design.md` under Open Questions
- [x] 1.3 Capture "before" screenshots at all three widths for comparison

## 2. Move balance presentation to the Resumo panel

- [x] 2.1 Add an explanation control to the Resumo panel's "Banco de Horas" and "Horário alternativo" rows in `PlannerSummaryPanel.tsx`, reusing the `<button>` + `[role="tooltip"]` + `aria-describedby` pattern from `PlannerSlotField.tsx`
- [x] 2.2 Move the `tooltipCopy` helper out of `PlannerSlotField.tsx` so both the summary panel and any future consumer can use it
- [x] 2.3 Make the Resumo "Horário alternativo" row name which slot the alternative applies to
- [x] 2.4 Handle the balance-unavailable case in the Resumo panel, stating that the normal exit time remains valid
- [x] 2.5 Anchor the Resumo popover inward and clamp it to the viewport so it cannot overflow the window edge from a 14rem column; verify at 720px

## 3. Declutter the slot field

- [x] 3.1 Remove the balance badge, its tooltip markup, and the `tooltipOpen` state from `PlannerSlotField.tsx`
- [x] 3.2 Remove the alternative-time chip (`.alternative-clockout`) from `PlannerSlotField.tsx`
- [x] 3.3 Replace the `registrada` badge with a punch-derived field treatment via a modifier class, keeping the input editable — do not set the `readonly` attribute
- [x] 3.4 Add the explanation that registered punches cannot be calculated, associated with the disabled calculator control
- [x] 3.5 Drop the now-unused `balance`, `balanceError`, and `alternativeTime` props from `PlannerSlotField` and stop threading them through `PlannerJourneyGroup.tsx` and `PlannerJourneyList.tsx`
- [x] 3.6 Confirm `PlannerSummaryPanel` receives the balance data it now needs, adding props in `PlannerPage.tsx` if required

## 4. Rework slot-field sizing

- [x] 4.1 Remove `flex-shrink: 0` from elements remaining in `.clockout-values` and add `flex-wrap: wrap` so the row wraps intrinsically instead of overflowing
- [x] 4.2 Add `min-width: 0` to `.calculada-time` so a long value cannot force its track wider
- [x] 4.3 Convert `.slot-solved` to an inset outline or inset box-shadow so the highlight consumes no layout width, keeping the background tint
- [x] 4.4 Move slot-internal reflow rules out of `@media (max-width: 800px)` into `@container` queries on `.slot-field`, using the widths measured in 1.2
- [x] 4.5 Recalibrate or remove the `@container (max-width: 11rem)` rule, which resolves to 176px against a ~174px field — under 2px of headroom
- [x] 4.6 Raise the `9.5rem` track floor in `.journey-group-fields`, or move its single-column override into a container query, so a track can no longer be narrower than the value row's min-content width
- [x] 4.7 Leave only viewport-scoped concerns in `@media (max-width: 800px)`: the `.grid-2` collapse, settings grids, and the sidebar

## 5. Overflow backstop

- [x] 5.1 Add `overflow: clip` (not `hidden`, to avoid creating a scroll container) to `.journey-group` and `.card` — the audit confirmed neither sets `overflow`, and `container-type: inline-size` does not provide paint containment
- [x] 5.2 Verify the remaining `.clockout-tooltip` popovers in the Resumo panel are not clipped by the containment added

## 6. Page and column width corrections

- [x] 6.1 Give the `.grid-2` journey track a non-zero minimum in place of `minmax(0, 1.5fr)`, sized to hold two slot fields usably, keeping the ~60/40 ratio
- [x] 6.2 Confirm the grid collapses to one column when the journey minimum can no longer be met, rather than shrinking past it
- [x] 6.3 Center the capped `.page` content within the main panel and verify at 1920px that leftover space is split evenly
- [x] 6.4 Delete the callerless `.page.narrow` rule after grepping to confirm no component still applies it
- [x] 6.5 Fix `.planner-warning` to use `--warning` / `--warning-bg` instead of the undefined `--warn` / `--warn-bg`, and verify the chip's background and border both render
- [x] 6.6 Grep the stylesheet for any other `var(--...)` referencing a token absent from `:root`

## 7. Dead code removal

- [x] 7.1 Delete the `.balance-badge`, `.alternative-clockout`, and `.registered-badge` rules and their tone variants
- [x] 7.2 Delete their overrides in the `@media (max-width: 800px)` block
- [x] 7.3 Grep the frontend for each removed class name to confirm no remaining references

## 8. Verification

- [x] 8.1 Verify at 720px, 980px, and 1280px that no element paints outside a card — check element boxes directly, since `.main { overflow-x: hidden }` hides overflow and makes a missing scrollbar meaningless as a pass signal
- [x] 8.2 Verify the calculated-slot highlight toggles without shifting layout
- [x] 8.3 Verify a registered slot reads as punch-derived, still accepts edits, and has a disabled calculator control with an explanation
- [x] 8.4 Verify the balance explanation is reachable by keyboard in the Resumo panel and closes on Escape
- [x] 8.5 Verify with a three-journey day, and with a day where every exit is registered so no slot is calculated
- [x] 8.6 Verify the ordering-conflict warning renders as a styled chip
- [x] 8.7 Verify page centering at 1920px in addition to the three reference widths
- [x] 8.8 Run the frontend typecheck, tests, and build, and capture "after" screenshots at all widths

## 9. Spec reconciliation

- [x] 9.1 Confirm the archive ordering constraint is respected: `page-column-layout` must be synced before `settings-single-column-layout`, or the settings single-column reversal is lost
- [x] 9.2 Record which audit findings for `planner-ui-polish` and `page-column-layout` this change addresses versus defers
- [x] 9.3 Confirm no requirement in this change's delta specs is left contradicted by the code, then archive the change
