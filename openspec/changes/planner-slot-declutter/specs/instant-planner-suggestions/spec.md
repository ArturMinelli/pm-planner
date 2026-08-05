## Purpose

Describe the slot presentation the planner actually uses and keeps: a two-up Entrada/Saída grid per journey, with balance information owned exclusively by the Resumo panel.

## MODIFIED Requirements

### Requirement: Timeline presentation of suggestions

The planner SHALL present each journey as a card containing its entry and exit slots side by side in a two-column grid that collapses to one column when the available width is insufficient. A slot SHALL show its label, its time value, and the calculator control; the calculated slot SHALL additionally show a `calculada` badge. Balance badges and alternative exit times SHALL NOT appear on individual slot rows. An inline ordering warning SHALL appear only when times conflict.

The previously specified vertical timeline presentation is withdrawn: it was never implemented, and the two-up grid keeps a journey's entry and exit visually paired, which reads better for a day built from journey spans.

#### Scenario: Clean slot row

- **WHEN** a slot has a suggestion and is not registered
- **THEN** the slot shows the label and suggested time as the primary visual element without extra chrome

#### Scenario: Entry and exit are paired

- **WHEN** a journey card is rendered at the default window width
- **THEN** its entry and exit slots appear side by side within the same card

#### Scenario: Balance lives in summary

- **WHEN** a balance adjustment applies today
- **THEN** the alternative exit time and the bank explanation appear only in the Resumo panel, and no balance badge or alternative-time chip is rendered in any slot

#### Scenario: Grid collapses when space is insufficient

- **WHEN** a journey card is too narrow to fit both slots side by side
- **THEN** the slots stack into a single column, each full width, with no horizontal overflow

#### Scenario: Ordering warning renders as a distinct chip

- **WHEN** a journey's times conflict and the inline ordering warning appears
- **THEN** it renders with its warning background and border, using tokens that are actually defined, rather than falling back to unstyled text

## ADDED Requirements

### Requirement: Resumo panel explains the balance trade-off

The Resumo panel SHALL own the explanation of the balance trade-off. Its "Banco de Horas" and "Horário alternativo" rows SHALL each expose an explanation covering which direction the balance is being used, the resulting estimated balance, the multiplier when credit is being generated, and any applied daily cap. The explanation SHALL be reachable by keyboard and SHALL be associated with its row for assistive technology.

#### Scenario: Explanation available on the alternative exit row

- **WHEN** a balance adjustment applies today and the user activates the explanation on the "Horário alternativo" row
- **THEN** the explanation states how much bank time the alternative exit uses or generates and the resulting estimated balance

#### Scenario: Explanation reachable by keyboard

- **WHEN** the user tabs to the "Banco de Horas" row explanation control and activates it
- **THEN** the explanation opens, and pressing Escape closes it

#### Scenario: Alternative exit is attributed to its slot

- **WHEN** the Resumo panel shows an alternative exit time
- **THEN** it identifies which slot the alternative applies to, since the time is no longer displayed next to that slot

#### Scenario: Balance unavailable

- **WHEN** the balance could not be fetched for today
- **THEN** the Resumo panel communicates that the balance is unavailable and that the normal exit time remains valid, and no slot field shows an unavailable-balance badge
