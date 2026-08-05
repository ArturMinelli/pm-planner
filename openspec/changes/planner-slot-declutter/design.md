## Context

The desktop frontend is React 19 in a Wails webview, styled entirely by one global stylesheet, `desktop/frontend/src/index.css`. The window defaults to 980x720 with a 720x540 minimum (`desktop/main.go`). The planner page uses a fixed 218px sidebar, a page capped at 1080px, and a `minmax(0, 1.5fr) / minmax(14rem, 1fr)` split above 801px.

At the 980px default that leaves each slot field roughly 170px of inner width. `box-sizing: border-box` is global, so the `.slot-solved` padding is not the culprit. The actual cause is that `.clockout-values` is a nowrap flex row whose children are all shrink-blocked:

- `.alternative-clockout` and `.balance-badge` set `flex-shrink: 0` (index.css:940-949)
- `.calculator-btn` sets `flex-shrink: 0` with a fixed `2rem` (index.css:718-722)
- `.calculada-time` has no `min-width: 0` and carries `1.15rem` tabular text plus padding (index.css:649-660)

Their combined minimum exceeds the field width, and because nothing may shrink and the row cannot wrap, the overflow is painted outside the field and outside `.journey-group`, which has no overflow containment (index.css:602-609).

The rules written to prevent this are keyed to the viewport. `.clockout-values { flex-direction: column }` lives inside `@media (max-width: 800px)` (index.css:1174), so it never applies at a 980px window even though the field is only 170px wide. The one container query, `@container (max-width: 11rem)` (index.css:924), targets 176px — within a few pixels of the real field width, so it fires unpredictably.

This is drift, not new design: `instant-suggestions-ux-redesign` already requires that balance badges and alternative exit times never appear on slot rows, and its tasks are all marked complete.

## Goals / Non-Goals

**Goals:**

- Nothing paints outside a card at any window width from 720px to full desktop.
- Slot reflow keys on the element's own available width, not the viewport's.
- No information is lost: the balance trade-off explanation moves to the Resumo panel, which already shows both numbers.
- Reduce, then constrain — cut redundant content first so the remaining layout has slack rather than depending on precise thresholds.

**Non-Goals:**

- Sidebar width, default window size, and the 60/40 column split are unchanged.
- No vertical timeline redesign; the two-up Entrada/Saída grid stays and the spec is amended to match.
- No change to the Go planner core or to how `BalanceAdjustment` is computed.
- No CSS framework, no design token overhaul, no automated visual regression harness.

## Decisions

### Cut content before adding CSS

Removing the balance badge and the alternative-time chip takes the value row down to a time value plus a `2rem` button, which fits ~170px with room to spare, and the label row down to `Saída 2` plus one badge. Most of the overflow disappears without any new sizing rule.

The alternative was to keep everything and make it reflow, which was the previous attempt (`planner-ui-polish`) and it failed. Adding rules to fit six elements into 170px produces threshold-sensitive CSS that breaks on the next content change. Both removed items already appear in the Resumo panel, so this costs no information.

### Reflow on container size, never on viewport size

`.slot-field` already declares `container-type: inline-size`, so container queries are available; the mistake was the threshold and the fact that the important rule (`.clockout-values` stacking) was left in a viewport media query. Slot-internal rules move to `@container`, and the viewport `@media (max-width: 800px)` block keeps only genuinely viewport-scoped concerns such as the `.grid-2` collapse and the sidebar.

Thresholds must be derived by measuring the rendered field in the running app at 720, 980, and 1280px, not guessed. The `11rem` value failed precisely because it was set a few pixels from the real width. Where a threshold can be avoided in favor of intrinsic sizing (`flex-wrap: wrap`, allowing `flex-shrink`, `min-width: 0`), prefer that: a wrap point computed by the browser cannot drift out of calibration.

### Allow shrinking; make the row wrappable

`flex-shrink: 0` is removed from the elements that remain in a space-constrained row, and `.clockout-values` becomes wrappable. `min-width: 0` moves onto `.calculada-time` so a long value cannot force the row wider than its track.

### Highlight the calculated field without consuming width

`.slot-solved` currently adds a 1px border plus `0.55rem 0.65rem` padding. Since the highlight can appear or disappear as the user toggles the calculator, its cost must be zero: use an inset `outline` (which does not participate in layout) or `box-shadow: inset`, keeping the background tint. This makes toggling the calculator layout-neutral, so a field that fits cannot stop fitting by becoming calculated.

### Overflow containment as a structural backstop

Correct sizing is the fix; containment is insurance. `.journey-group` and `.card` get overflow containment so that a future addition without matching sizing rules degrades into clipping rather than drawing over the card border. `overflow: clip` is preferred over `hidden` because it does not create a scroll container, so it will not interfere with the `.clockout-tooltip` popovers that remain elsewhere on the page.

### Balance explanation moves to Resumo, anchored inward

The Resumo column can be as narrow as `14rem`, while `.clockout-tooltip` is up to `19rem` wide and anchored `left: 0`. Reused as-is it would overflow the window's right edge. The popover is therefore anchored inward from the right edge of the Resumo panel and clamped to the viewport, reusing the existing `min(19rem, calc(100vw - 3rem))` width idiom that already guards this.

The existing markup pattern is kept: a `<button>` toggling a sibling `[role="tooltip"]` with `aria-describedby`, Escape to close, close on blur. This preserves keyboard access and the assistive-technology association that the current in-slot badge already had.

### Registered slots keep a locked look but stay editable

`dynamic-journeys` requires that registered punches remain editable, and that stays true. The field gets muted, punch-derived styling plus an explanation, but accepts input. The existing global `input[readonly]` rule (index.css:407-410) supplies the muted look but must not be applied via an actual `readonly` attribute, since that would remove editability; a modifier class carries the appearance instead.

### Dead code removal

`.balance-badge`, `.alternative-clockout`, and `.registered-badge` become unused and are deleted along with their `@media` overrides at index.css:1169-1186, rather than left as orphans in a 1200-line global stylesheet. `.clockout-tooltip*` rules are retained because the Resumo panel now uses them.

## Risks / Trade-offs

- **The alternative exit time is now shown in one place instead of two, and is farther from the slot it applies to.** → The Resumo row explicitly names the slot it belongs to, and it sits in the panel the user already reads for the day's totals.
- **Container query thresholds could be miscalibrated again.** → Prefer intrinsic wrapping over thresholds wherever possible; where a threshold is required, derive it from a measurement in the running app rather than from arithmetic, and verify at all three reference widths.
- **`overflow: clip` could hide a future layout bug instead of making it visible.** → It is a backstop, not the mechanism; the three-width verification is what catches regressions, and clipping inside a card is a better failure mode than painting across it.
- **Removing three of four badges could feel like a loss of information density.** → Two were verbatim duplicates of Resumo rows, and the third is replaced by a field treatment that conveys the same state.
- **This is the second attempt at the same bug.** → The distinguishing move is fixing the mechanism (container-relative sizing) rather than adding more viewport breakpoints, plus removing the content that made the field impossible to fit.
- **Slot state is threaded through `PlannerJourneyList` and `PlannerJourneyGroup` as props.** → Removing balance rendering from slots leaves unused props to clean up; the balance data must be rerouted to `PlannerSummaryPanel`, which already receives `loaded` and `summary`.

## Migration Plan

Frontend-only and stateless: no data migration, no backend change. Rollback is a revert of the frontend commit. The Go core continues to compute `BalanceAdjustment` unchanged throughout, so presentation can be reverted independently of any calculation.

## Width budget at the 980px default

Derived by tracing the box chain; to be confirmed by measurement in the running app (task 1.2):

| Box | Width |
| --- | --- |
| `main` (980 − 218 sidebar − 1px border) | 761px |
| `.page` content box (− 48px padding) | 713px |
| `.grid-2` journey track / summary track | 418px / 279px |
| `.journey-group` inner | 359px |
| `.journey-group-fields` → 2 tracks | ~174px each |
| calculated slot content box (− ~23px `.slot-solved`) | ~151px |
| minimum width of the value row | ~187px |

The row overshoots its box by roughly 35px, which is the overflow in the screenshot. Two consequences follow. First, the `9.5rem` (152px) track floor in `.journey-group-fields` admits a track narrower than the value row's min-content width, so the floor is itself part of the bug. Second, `@container (max-width: 11rem)` resolves to 176px against a ~174px field — under 2px of headroom. That query is the only thing currently preventing badge-row overflow at the default size, and any change to sidebar width, page padding, grid gap, or the track floor flips it off.

## Open Questions

- Resolved by measurement (task 1.2) in a layout fixture that mirrors the planner chrome + two-journey overflow case. Rendered `.slot-field` (calculated Saída 2) inline sizes:

  | Window width | `.slot-field` inline size |
  | --- | --- |
  | 720px | 624.81px |
  | 980px | 172.61px |
  | 1280px | 262.61px |

  At 980px the alternative-time chip paints ~32px past the slot and ~17px past the journey group (overflow reproduced). Screenshots: `artifacts/before-{720,980,1280}.png`.
- Container-query thresholds derived from those widths: stack/reflow below ~12rem (192px) so the 980px default reflows while 1280px (~263px) stays on one line; label-action wrapping below ~14rem (224px). Journey fields collapse to one column below `24rem` group width (named container `journey`).
- Resolved by the drift audit: containment is applied to both `.card` and `.journey-group`, since neither sets `overflow` today and `container-type: inline-size` provides layout, style, and inline-size containment but *not* paint containment. Only `.main { overflow-x: hidden }` currently stops the spill at the panel edge. Resumo cards opt out via `:has(.summary-tooltip-wrap)` so explanation popovers are not clipped.

## Archive ordering (task 9.1)

Neither `page-column-layout` nor `settings-single-column-layout` has been synced/archived yet. When they are, sync or archive `page-column-layout` **before** `settings-single-column-layout`, or the settings single-column reversal is lost.

## Audit findings addressed vs deferred (task 9.2)

| Prior change | Addressed here | Deferred |
| --- | --- | --- |
| `planner-ui-polish` | Slot overflow at 980px; viewport-keyed reflow that never fired; badge crowding; calculated highlight layout shift | Sidebar width, default window size |
| `page-column-layout` | Journey column `minmax(0, …)` floor; capped page left-alignment; `.page.narrow` dead rule | Settings grid (owned by `settings-single-column-layout`) |
| Other | `.planner-warning` undefined `--warn` tokens; balance/alternative chrome drift vs `instant-suggestions-ux-redesign` | Go planner core / balance computation |
