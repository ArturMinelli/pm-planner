## 1. CSS layout

- [x] 1.1 Remove `.settings-grid` two-column rules from `@media (min-width: 801px)` in `index.css` (delete `grid-template-columns: repeat(2, …)` and `:last-child` span rule)
- [x] 1.2 Verify `.settings-grid` keeps single-column `display: grid; gap: 1rem` at all widths

## 2. Tests

- [x] 2.1 Update `SettingsPage.test.tsx` to describe single-column stacked layout (wrapper still holds three cards in order)
- [x] 2.2 Run frontend tests and confirm pass

## 3. Visual verification

- [x] 3.1 Confirm settings cards stack vertically at wide desktop width (~1000px+)
- [x] 3.2 Confirm planner 60/40 layout unchanged
