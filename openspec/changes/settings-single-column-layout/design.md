## Context

See `proposal.md` for motivation. Settings currently uses `.settings-grid` with `grid-template-columns: repeat(2, 1fr)` at ≥801px (`page-column-layout`), placing account and planner defaults side by side and spanning reminders full width. Grilling confirmed: always single column at all widths, full main content width (same as planner `Page`), keep Account → Planner → Reminders order, planner 60/40 unchanged.

## Goals / Non-Goals

**Goals:**

- Settings cards stack vertically at every viewport width.
- Cards use full main content width (1080px `Page` max).
- Preserve card order: Account → Planner → Reminders.
- Remove two-column settings CSS without affecting planner `.grid-2`.

**Non-Goals:**

- Changing planner column ratio or journey/summary layout.
- Reordering settings cards or merging forms.
- Changing form fields, validation, or save behavior.
- New visual styling beyond layout structure.

## Decisions

### 1. CSS-only change on `.settings-grid`

**Choice:** Keep the `.settings-grid` wrapper in `SettingsPage.tsx` (useful for tests and gap spacing) but remove the `@media (min-width: 801px)` rules that set two columns and the `:last-child` span. Default single-column grid (`display: grid; gap: 1rem`) applies at all widths.

**Alternatives:** Replace grid with flex column — rejected; grid gap already works and tests reference `.settings-grid`.

### 2. Full-width via existing `Page` container

**Choice:** Settings already uses `<Page>` without `narrow` (from `page-column-layout`). No markup change needed for width — only remove two-column CSS.

**Alternatives:** Reintroduce `narrow` — rejected per grilling (full width desired).

### 3. Planner CSS untouched

**Choice:** Leave `.grid-2` 60/40 rules in the same `@media (min-width: 801px)` block; only delete `.settings-grid` rules from that block.

**Alternatives:** Separate media queries — rejected; unnecessary churn.

### 4. Test updates

**Choice:** Update `SettingsPage.test.tsx` description/assertions to reflect single-column layout (wrapper still has 3 children stacked, not side-by-side columns).

## Risks / Trade-offs

- **[Longer scroll on wide screens]** Single column means more vertical scroll → Acceptable per user preference; content is easier to scan.
- **[Wider input fields]** Full-width cards stretch inputs → Improves readability for long values (login, anchor times).

## Migration Plan

CSS deletion + test label update. No data migration. Rollback by restoring two-column `.settings-grid` rules.

## Open Questions

None — grilling resolved layout, width, order, and scope.
