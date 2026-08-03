## Purpose

Deliver immediate, journey-count-adaptive clock suggestions in the desktop planner so users always see coherent entry and exit times without waiting on network round-trips.

## ADDED Requirements

### Requirement: Entry suggestions update synchronously on every edit

The desktop planner SHALL compute and display suggested entry times for every journey immediately when the user edits any slot, adds or removes a journey, or toggles the solved exit. Registered entry slots SHALL keep their punch time and SHALL NOT be overwritten by anchor suggestions.

#### Scenario: Edit entry shows anchor immediately

- **WHEN** the user changes Entrada 2 on a 2-journey day with unregistered slots
- **THEN** all other unregistered entry slots refresh their displayed suggestions within the same interaction frame without a perceptible delay

#### Scenario: Registered entry is preserved

- **WHEN** Entrada 1 is marked registrada at 08:05
- **THEN** Entrada 1 continues to show 08:05 and is not replaced by the morning anchor

### Requirement: Exit suggestions follow solved-slot selection immediately

The desktop planner SHALL calculate exactly one exit from the daily target when any exit is unsolved. Changing which exit is solved SHALL recalculate that exit's suggestion immediately. Registered exits SHALL never be chosen as the solved slot.

#### Scenario: Toggle solved exit recalculates instantly

- **WHEN** the user activates calculation on Saída 1 on a day with target 6h00
- **THEN** Saída 1 becomes the calculated exit and its time updates immediately to fill the remaining target

#### Scenario: Registered exit cannot be solved

- **WHEN** Saída 2 is registrada
- **THEN** the calculator control for Saída 2 is disabled and Saída 2 is never auto-selected as the solved exit

### Requirement: Adding a journey seeds suggestions immediately

When the user adds a journey, the system SHALL create entry suggestion at previous exit plus break duration and set the new exit as the solved slot with a calculated or anchor-derived time without waiting for a debounced backend call.

#### Scenario: Third journey on add

- **WHEN** the user clicks "Adicionar jornada" after Saída 2 at 15:00 with 1h break
- **THEN** Entrada 3 shows 16:00 and Saída 3 is the active calculated exit with a coherent suggested time

### Requirement: Timeline presentation of suggestions

The planner SHALL present slots in a vertical timeline: label, prominent suggested time, optional registrada pill, and inline ordering warning only when times conflict. Balance badges and alternative exit times SHALL NOT appear on individual slot rows.

#### Scenario: Clean slot row

- **WHEN** a slot has a suggestion and is not registrada
- **THEN** the row shows the label and suggested time as the primary visual element without extra chrome on that row

#### Scenario: Balance lives in summary

- **WHEN** a balance adjustment applies today
- **THEN** the alternative exit and bank explanation appear only in the Resumo panel

### Requirement: Solved exit is visually distinct

The active calculated exit SHALL use a dedicated visual treatment (highlighted suggestion chip or filled field) and plain-language affordance such as "calculada" instead of a generic disabled input.

#### Scenario: Solved exit highlight

- **WHEN** Saída 2 is the solved exit
- **THEN** Saída 2 is visually marked as calculada and its time is read-only
