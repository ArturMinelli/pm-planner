## Why

At the app's default 980x720 window, the calculated `Saída` field paints outside its journey card: the accent border and the alternative-time chip both cross the card boundary. The field is asked to hold a label, three badges (`registrada`, `calculada`, `Banco -12:25`), a time value, the calculator button, and an alternative-time chip inside roughly 170px of inner width.

The reflow rules that would rescue it are keyed to **viewport** width (`@media (max-width: 800px)`), but the overflow happens when a **column** is narrow inside a still-wide window, so those rules never fire. The single `@container (max-width: 11rem)` rule sits just above the actual field width and therefore triggers unreliably.

Two of the three crowding badges are already redundant: `Banco -12:25` and the `19:43` chip duplicate the Resumo panel's "Banco de Horas" and "Horário alternativo" rows. `instant-suggestions-ux-redesign` already requires that those never appear on slot rows, and its tasks are marked complete — so this is drift against an existing requirement, not new design.

## What Changes

- Remove the balance badge (`Banco -12:25`), its click-tooltip, and the alternative-time chip (`19:43`) from slot fields, bringing the code in line with the existing requirement that these live only in the Resumo panel.
- Move the balance trade-off explanation to tooltips on the Resumo panel's "Banco de Horas" and "Horário alternativo" rows, so no information is lost.
- **BREAKING (spec)**: replace the `registrada` badge with a distinct punch-derived field treatment plus an explanation. This supersedes the `dynamic-journeys` requirement that a registered slot display a `registrada` badge. Registered slots stay editable, as they are today.
- Keep the `calculada` badge as the only badge inside a slot field.
- Change the calculated-field highlight so it consumes no extra horizontal width (inset outline or box-shadow rather than border plus padding).
- Replace viewport-keyed reflow rules for slot internals with **container-based** rules, so a narrow column inside a wide window reflows correctly. Stacking becomes a genuine fallback near the 720px window minimum rather than the primary mechanism.
- Harden cards against overflow (`min-width: 0`, overflow containment) so a future regression cannot paint outside a card.
- **BREAKING (spec)**: amend the unimplemented "vertical timeline" presentation requirement to describe the two-up Entrada/Saída grid the code actually uses and that we are keeping.
- Give the planner's journey column a non-zero minimum width. It is currently `minmax(0, 1.5fr)`, so the existing rule against squeezing columns "below readable minimums" is unenforced for the primary column — the upstream reason slot fields collapse to ~174px.
- Center the capped page content, so a window wider than ~1300px no longer leaves the whole leftover band on one side (about 600px at 1920px), and delete the now-callerless `.page.narrow` rule.
- Fix `.planner-warning`, which references undefined `--warn` / `--warn-bg` tokens instead of `--warning` / `--warning-bg`. The invalid `var()` values drop both the background and the entire `border` shorthand, so the time-conflict warning currently renders as unstyled text rather than a chip.
- Correct the stale 600px lower bound in the responsiveness requirement to the 720px window minimum actually enforced by `desktop/main.go`, and record that the absence of a scrollbar is not evidence of compliance, since `.main { overflow-x: hidden }` masks overflow rather than preventing it.
- Explicitly out of scope: sidebar width (218px), default window size (980x720), and the 60/40 planner column ratio all stay as they are. The fix must work at both 980px and the 720px minimum.

## Capabilities

### New Capabilities

None. This change corrects and clarifies existing capabilities.

### Modified Capabilities

- `desktop-ui-layout`: slot-field reflow requirements move from viewport-based to container-based sizing; badge-wrapping requirement is replaced because the badges themselves are removed; the responsiveness floor is corrected from 600px to the enforced 720px minimum; the planner journey column gains an enforced width floor; capped page content must be centered; adds a requirement that content never paints outside a card.
- `instant-planner-suggestions`: the "vertical timeline" presentation requirement is amended to the two-up Entrada/Saída grid; the existing "balance lives in summary" requirement is restated with the Resumo tooltip as its explicit home.
- `journey-planning`: registered slots are indicated by read-only styling and a tooltip instead of a `registrada` badge; the Resumo panel becomes the sole owner of alternative exit time and bank explanation.

## Impact

Code:

- `desktop/frontend/src/features/planner/PlannerSlotField.tsx` — remove balance badge, tooltip, and alternative chip; add read-only treatment for registered slots.
- `desktop/frontend/src/features/planner/PlannerSummaryPanel.tsx` — host the balance explanation tooltip.
- `desktop/frontend/src/index.css` — `.slot-field`, `.clockout-label-row`, `.clockout-values`, `.slot-solved`, `.journey-group*`, container queries, and the `@media (max-width: 800px)` block; delete now-dead `.balance-badge`, `.alternative-clockout`, `.registered-badge`, and `.clockout-tooltip*` rules if unused elsewhere.
- `desktop/frontend/src/features/planner/PlannerJourneyList.tsx` / `PlannerJourneyGroup.tsx` — prop threading for values no longer rendered in slots.

Not affected: the Go planner core and its balance calculations. `BalanceAdjustment` data is still computed and consumed, only its presentation moves.

Archive ordering constraint: `page-column-layout` adds a two-column settings grid requirement that `settings-single-column-layout` later reverses to single-column at all widths. The content is already corrected in the later change, so no further spec edit is needed here — but `page-column-layout` MUST be synced or archived before `settings-single-column-layout`, or the reversal is lost and a reverted layout is reintroduced.

Risk: the alternative exit time becomes visible in one place instead of two, so the Resumo rows must clearly attribute it to the calculated slot.
