## Purpose

Indicate registered punches through field affordance rather than a badge, and make the Resumo panel the sole owner of alternative exit time.

## MODIFIED Requirements

### Requirement: Registered punches are pre-filled and never solved

Slots matched to a real PontoMais punch SHALL be pre-filled with the punch time and SHALL remain editable, so the user can plan a hypothetical against a punch already made. Registration SHALL be communicated by a distinct field treatment rather than a `registrada` badge: the field SHALL read as originating from a real punch and SHALL carry an explanation to that effect. A registered exit SHALL NOT be selected as the solved (calculated) slot, and its calculator control SHALL be disabled.

The `registrada` badge is removed because a slot field does not have room for a badge row at the application's default window width; the field's own treatment communicates the same state without competing for horizontal space.

#### Scenario: Registered exit is shown not recalculated

- **WHEN** all four slots of a 2-journey day have registered punches including Saída 2 at 14:41
- **THEN** Saída 2 displays 14:41 and is NOT replaced by a target-derived time such as 15:41

#### Scenario: Registered slot reads as punch-derived

- **WHEN** a slot time was assigned from a real punch
- **THEN** that slot's field carries a distinct treatment identifying it as punch-derived, with no `registrada` badge

#### Scenario: Registered slot stays editable

- **WHEN** the user types a different time into a registered slot
- **THEN** the edit is accepted and the plan recalculates, because registered slots remain editable for what-if planning

#### Scenario: Registered slot explains why it cannot be calculated

- **WHEN** the user inspects or focuses a registered slot's disabled calculator control
- **THEN** an explanation states that registered punches cannot be calculated

### Requirement: Summary shows per-journey spans and totals

The summary panel SHALL display the daily target, one row per journey span, total worked time, and overtime (hora extra). When a balance adjustment applies today, the summary panel SHALL be the only place an alternative solved exit time is shown, and it SHALL identify which slot that alternative applies to.

#### Scenario: Three journeys show three span rows

- **WHEN** the day has 3 journeys with valid times
- **THEN** the summary lists "1ª Jornada", "2ª Jornada", "3ª Jornada" with their respective durations plus Total and Hora Extra

#### Scenario: Alternative exit shown only in summary

- **WHEN** a balance adjustment applies today and Saída 2 is the solved slot
- **THEN** the alternative exit time appears in the summary panel attributed to Saída 2, and does not appear beside the Saída 2 field
