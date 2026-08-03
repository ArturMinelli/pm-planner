## Purpose

Enable users to plan work days with a variable number of journeys (entry + exit pairs), choose which exit is calculated from the daily target, and never re-suggest clock-out times that were already registered in PontoMais.

## ADDED Requirements

### Requirement: Default journey count

The system SHALL present at least 2 journeys when loading a day with no punches. When loading a day with existing punches, the system SHALL open with `max(2, ceil(punch_count / 2))` journeys.

#### Scenario: Empty day loads two journeys

- **WHEN** the user loads a work day with zero registered punches
- **THEN** the planner displays exactly 2 journeys, each with an entry and an exit slot

#### Scenario: Six punches open three journeys

- **WHEN** the user loads a work day with 6 registered punches
- **THEN** the planner displays 3 journeys pre-filled from those punches

### Requirement: User can add and remove journeys per day

The system SHALL allow the user to add a journey via an "Adicionar jornada" control and remove any journey after the first via a remove control on that journey group. Removing a journey that contains a registered punch SHALL be disabled.

#### Scenario: Add third journey

- **WHEN** the user clicks "Adicionar jornada" on a 2-journey day
- **THEN** a third journey appears with its entry seeded at the previous exit plus the configured break duration and its exit set as the new solved slot

#### Scenario: Cannot remove journey with registered punch

- **WHEN** a journey contains at least one slot marked as registered
- **THEN** the remove control for that journey is disabled with an explanatory tooltip

### Requirement: Registered punches are pre-filled and never solved

Slots matched to a real PontoMais punch SHALL be marked `registrada`, pre-filled with the punch time, and remain editable. A registered exit SHALL NOT be selected as the solved (calculated) slot.

#### Scenario: Registered exit is shown not recalculated

- **WHEN** all four slots of a 2-journey day have registered punches including Saída 2 at 14:41
- **THEN** Saída 2 displays 14:41 and is NOT replaced by a target-derived time such as 15:41

#### Scenario: Registered badge on matched slots

- **WHEN** a slot time was assigned from a real punch
- **THEN** that slot displays a `registrada` badge alongside the time input

### Requirement: Exactly one exit is solved at a time

The system SHALL calculate exactly one exit slot from the daily target at any time. Entry slots SHALL never be calculated. The user SHALL select which exit is solved via an inline calculator toggle on each Saída field. The active solved exit SHALL be read-only and visually highlighted.

#### Scenario: User moves solved slot to Saída 1

- **WHEN** the user activates the calculator toggle on Saída 1
- **THEN** Saída 1 becomes read-only and calculated, and the previously solved exit becomes a normal editable field

#### Scenario: Solved exit fills remaining target time

- **WHEN** journey 1 spans 3h42 and the daily target is 6h00 with Saída 2 as the solved exit and Entrada 2 at 13:23
- **THEN** Saída 2 is calculated as 13:23 plus the remaining 2h18 (15:41), unless Saída 2 is registered

### Requirement: Solved exit auto-shifts away from registered punches

When the default solved exit (last exit without a punch) would be a registered slot, the system SHALL auto-shift the solved slot to the last exit that has no registered punch. When every exit has a registered punch, no exit SHALL be solved.

#### Scenario: All exits punched shows real total

- **WHEN** every exit slot has a registered punch
- **THEN** no exit is calculated, and the summary shows the sum of actual journey spans and any overtime

#### Scenario: Auto-shift skips registered last exit

- **WHEN** Saída 2 is registered and Saída 3 is empty on a 3-journey day
- **THEN** Saída 3 becomes the solved exit, not Saída 2

### Requirement: Stamp assignment preserves anchor proximity

The system SHALL map chronological punches to planner slots by minimizing total distance to anchor times while preserving punch order. When no punch maps to the first entry anchor, the first entry SHALL default to the morning anchor time, not the chronologically first punch.

#### Scenario: Lone midday punch maps to Saída 1

- **WHEN** the only registered punch is 12:28 and anchors are 08:00 / 12:00 / 13:30 / 18:00
- **THEN** Entrada 1 defaults to 08:00 and Saída 1 is assigned 12:28

### Requirement: Derived anchors for journeys beyond two

For journey index 3 and above, entry anchors SHALL be derived as the previous journey's exit plus the break duration. Exit anchors SHALL be derived from the entry plus a proportional share of remaining target time. The four user-configured anchor times in Settings SHALL apply only to journeys 1 and 2.

#### Scenario: Third journey entry derived from break

- **WHEN** journey 2 exit is 15:00 and the break duration is 1 hour
- **THEN** journey 3 entry anchor defaults to 16:00

### Requirement: Ordering validation without cascade on solve

When a calculated exit time is after the next journey's entry, the system SHALL display an inline validation warning on the affected pair. The system SHALL NOT auto-cascade other slot times as a result of solving. Manual time edits via arrow keys MAY continue to enforce minimum-gap ordering via forward cascade.

#### Scenario: Solved Saída 1 after Entrada 2 shows warning

- **WHEN** Saída 1 is solved and its calculated time is after Entrada 2
- **THEN** an inline warning is shown on that journey pair and other times remain unchanged

### Requirement: Summary shows per-journey spans and totals

The summary panel SHALL display the daily target, one row per journey span, total worked time, and overtime (hora extra). When a balance adjustment applies today, an alternative solved exit time SHALL be shown for the active solved slot.

#### Scenario: Three journeys show three span rows

- **WHEN** the day has 3 journeys with valid times
- **THEN** the summary lists "1ª Jornada", "2ª Jornada", "3ª Jornada" with their respective durations plus Total and Hora Extra

### Requirement: Single source of truth for plan math

All journey span, solved-exit, and summary calculations SHALL be performed in the shared Go planner core. The desktop app SHALL call a backend recalculate binding on edit rather than duplicating math locally.

#### Scenario: Desktop recalculates via backend

- **WHEN** the user edits any journey time in the desktop planner
- **THEN** the app calls the Go recalculate endpoint and updates the summary from the response

### Requirement: CLI and reminders support dynamic journeys

The CLI `pm plan` command and the reminder scheduler SHALL use the same dynamic journey model. Reminder slot keys SHALL be generated as `in{n}` and `out{n}` with labels `Entrada N` and `Saída N`.

#### Scenario: Reminder for third journey entry

- **WHEN** a day plan has 3 journeys and reminders are enabled
- **THEN** the scheduler creates reminders for `in3` labeled "Entrada 3" at the planned time
