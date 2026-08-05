## 1. Core algorithm — solve any slot

- [x] 1.1 Generalize Go `SolveSlot` to compute either entry or exit for a journey index from daily target
- [x] 1.2 Replace `solvedExit` with `solvedSlot` (`journeyIndex` + `isEntry`) in Go payload, app layer, and Wails bindings
- [x] 1.3 Update `DefaultSolvedSlot` to pick last unregistered slot (entry or exit)
- [x] 1.4 Add Go tests for solve-entry and solve-exit across 2–4 journeys

## 2. Local suggestion mirror (TypeScript)

- [x] 2.1 Add `plannerSuggestions.ts` mirroring anchor resolution and `SolveSlot`
- [x] 2.2 Add TS tests aligned with Go fixtures for instant suggestions

## 3. Planner page state

- [x] 3.1 Remove 400 ms debounce; split raw journeys vs display journeys with instant local solve
- [x] 3.2 Fire `recalculateDay` immediately for summary/balance with stale-request guard
- [x] 3.3 Allow editing registrada slots; toggle solved on any unregistered Entrada or Saída

## 4. UI redesign — journey cards

- [x] 4.1 Simplify `PlannerJourneyGroup` — clean card layout, calculada highlight on active slot
- [x] 4.2 Add calculator toggle on both Entrada and Saída rows
- [x] 4.3 Show balance badge + alternative time on calculated slot AND in Resumo panel
- [x] 4.4 Update `index.css` for cleaner cards, calculada chip, reduced chrome

## 5. Summary panel

- [x] 5.1 Move alternative exit / bank explanation into Resumo with clear labels
- [x] 5.2 Keep journey spans, total, overtime; show inline ordering warnings

## 6. Tests and build

- [x] 6.1 Update component tests for journey cards, instant toggle, editable registrada
- [x] 6.2 Run frontend and Go test suites
- [x] 6.3 Build and install desktop app via `scripts/setup.sh`
