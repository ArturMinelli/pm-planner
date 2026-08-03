## MODIFIED Requirements

### Requirement: Exactly one exit is solved at a time

The system SHALL calculate exactly one exit slot from the daily target at any time. Entry slots SHALL never be calculated from the target, but unregistered entry slots SHALL display immediate anchor-derived suggestions. The user SHALL select which exit is solved via an inline calculator control on each Saída field. The active solved exit SHALL be read-only, visually highlighted as calculada, and update its suggestion synchronously on any edit.

#### Scenario: User moves solved slot to Saída 1

- **WHEN** the user activates the calculator toggle on Saída 1
- **THEN** Saída 1 becomes read-only, calculada, and recalculated immediately, and the previously solved exit becomes a normal editable field

#### Scenario: Solved exit fills remaining target time

- **WHEN** journey 1 spans 3h42 and the daily target is 6h00 with Saída 2 as the solved exit and Entrada 2 at 13:23
- **THEN** Saída 2 is calculated as 13:23 plus the remaining 2h18 (15:41) immediately, unless Saída 2 is registered

### Requirement: Single source of truth for plan math

All journey span, solved-exit, and summary calculations SHALL be performed in the shared Go planner core. The desktop app SHALL call a backend recalculate binding for summary totals, balance, and alternative exit. Slot suggestion times for display MAY be computed locally using the same rules as the Go core so they update without network latency.

#### Scenario: Desktop recalculates summary via backend

- **WHEN** the user edits any journey time in the desktop planner
- **THEN** the app calls the Go recalculate endpoint and updates journey spans, total, overtime, and balance from the response

#### Scenario: Slot suggestions do not wait for backend

- **WHEN** the user edits any journey time in the desktop planner
- **THEN** displayed slot suggestions update immediately before or without waiting for the summary response
