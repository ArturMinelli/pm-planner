## Purpose

Enables the planner to open ready-to-use on any selected date by auto-fetching work-day data and falling back to Settings defaults when the API is unavailable.

## ADDED Requirements

### Requirement: Auto-load on mount and date change

The planner SHALL automatically fetch the full work-day payload (time cards, shift schedule, employee balance) when the page opens and whenever the user changes the selected date. The manual **Carregar** control SHALL NOT be present.

#### Scenario: Page opens with today's date

- **WHEN** the user opens the planner page in the desktop app
- **THEN** the system fetches work-day data for today's date without user action
- **AND** journey fields and summary remain disabled until the fetch completes or fails

#### Scenario: User changes the selected date

- **WHEN** the user picks a different date in the header date picker
- **THEN** the system immediately fetches work-day data for the new date
- **AND** all planner state from the previous date is fully replaced when the fetch succeeds

### Requirement: Loading state uses skeleton placeholders

While a fetch is in progress, the planner SHALL show skeleton placeholders inside the journey list and summary cards and SHALL keep time fields disabled.

#### Scenario: Fetch in progress

- **WHEN** work-day data is being fetched
- **THEN** skeleton placeholders are visible in the journey and summary areas
- **AND** journey time inputs are not editable

### Requirement: Successful fetch populates planner from API

On a successful fetch, the planner SHALL display journeys, base target, balance adjustment, solved exit, originals line, and summary exactly as produced by the existing full load pipeline.

#### Scenario: API returns work-day data

- **WHEN** the fetch succeeds for the selected date
- **THEN** journeys reflect assigned stamps and registered punches from the API
- **AND** the base target comes from the shift schedule (or the existing default when shift time is absent)
- **AND** employee balance is applied when available
- **AND** journey fields become editable

### Requirement: Fetch failure falls back to Settings defaults

When the work-day fetch fails, the planner SHALL populate from planner Settings defaults and SHALL remain usable. A non-blocking warning SHALL inform the user that API data could not be loaded.

#### Scenario: API unreachable on page open

- **WHEN** the fetch fails for any reason (network, auth, empty response, server error)
- **THEN** the system builds two journeys from Settings anchor times (Entrada 1, Saída 1, Entrada 2, Saída 2)
- **AND** all slots are marked unregistered
- **AND** the solved exit is the last journey exit
- **AND** the base target is 8 hours 30 minutes
- **AND** balance-dependent alternative exit is omitted
- **AND** the originals line reads `(nenhum)`
- **AND** a warning banner explains the fallback
- **AND** journey fields become editable

#### Scenario: Settings anchors are partially configured

- **WHEN** the fetch fails and the user has saved custom anchor times in Settings
- **THEN** the fallback journeys use the saved anchor times (with invalid slots falling back per existing anchor resolution rules)

### Requirement: Date picker lives in the page header

The date picker SHALL be rendered in the planner page header. The standalone load card SHALL NOT exist.

#### Scenario: Planner page layout

- **WHEN** the user views the planner page
- **THEN** a date input is visible in the page header next to the title
- **AND** no separate load card or **Carregar** button is shown

### Requirement: CLI falls back to Settings defaults on fetch failure

The `pm plan` command SHALL attempt the same full work-day fetch as today. When that fetch fails, it SHALL render a plan from Settings defaults instead of exiting with an error.

#### Scenario: CLI with API failure

- **WHEN** the user runs `pm plan` and the work-day fetch fails
- **THEN** the CLI prints a defaults-based plan using Settings anchor times and an 8h30 base target
- **AND** the CLI indicates that API data was unavailable
