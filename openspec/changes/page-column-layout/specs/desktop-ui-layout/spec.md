## MODIFIED Requirements

### Requirement: Planner grid reflows at medium breakpoints

The planner page two-column layout (journey list + summary) SHALL use a journey-primary column ratio of approximately 60% journey list and 40% summary at desktop widths wider than the medium breakpoint. The layout SHALL switch to a single column when horizontal space is insufficient (~800px). The summary panel SHALL appear below the journey list without squeezing columns below readable minimums. Both columns SHALL fill their allocated width without large empty gaps beside cards.

#### Scenario: Journey-primary columns at wide width

- **WHEN** the viewport is wider than the medium breakpoint (~800px)
- **THEN** the journey list column is visibly wider than the summary column (approximately 60/40)
- **AND** neither column shows large unused horizontal space beside its card content

#### Scenario: Two columns at wide width

- **WHEN** the viewport is wider than the medium breakpoint
- **THEN** journey list and summary display side by side

#### Scenario: Single column at medium width

- **WHEN** the viewport is at or below the medium breakpoint (~800px)
- **THEN** journey list and summary stack vertically with full-width cards

## ADDED Requirements

### Requirement: Settings page uses two-column card grid at wide widths

The settings page SHALL arrange its settings cards (account, planner defaults, reminders) in a responsive two-column grid when the viewport is wider than the medium breakpoint (~800px). Cards SHALL stack to a single column at or below that breakpoint. Each card SHALL span one grid cell and fill the available column width without leaving large empty gutters.

#### Scenario: Two-column settings at desktop width

- **WHEN** the user views the settings page at a viewport wider than ~800px
- **THEN** settings cards display in a two-column grid
- **AND** cards fill their columns without large dead space between columns

#### Scenario: Single-column settings at medium width

- **WHEN** the viewport is at or below ~800px on the settings page
- **THEN** all settings cards stack vertically at full width

### Requirement: Page content fills available column width

Page layouts (planner, settings, shared page chrome) SHALL use the available main content width so columns distribute space evenly. Content areas SHALL NOT leave large unused horizontal margins or gutters inside the main panel at desktop widths from ~900px upward.

#### Scenario: Planner fills main panel

- **WHEN** the planner page is viewed at ~1000px viewport width
- **THEN** the two-column grid spans the main content area with balanced column widths and no large empty band beside the grid

#### Scenario: Settings fills main panel

- **WHEN** the settings page is viewed at ~1000px viewport width
- **THEN** the settings card grid spans the main content area without a single narrow column floating in unused space

### Requirement: Page headers align to content grid

Page headers (title, description, and actions such as the date picker) SHALL align with the same horizontal content boundaries as the page body grid below, so header actions do not sit in dead zones misaligned from the columns beneath.

#### Scenario: Planner header aligns with grid

- **WHEN** the planner page is viewed at desktop width
- **THEN** the date picker action area aligns horizontally with the right edge of the summary column or the page content grid boundary

#### Scenario: Settings header aligns with cards

- **WHEN** the settings page is viewed at desktop width
- **THEN** the page header text aligns with the left edge of the settings card grid below
