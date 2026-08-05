## MODIFIED Requirements

### Requirement: Settings page uses two-column card grid at wide widths

The settings page SHALL arrange its settings cards (account, planner defaults, reminders) in a single vertical column at all viewport widths. Cards SHALL appear in order: account, then planner defaults, then reminders. Each card SHALL span the full width of the settings content area and fill the available horizontal space without splitting into side-by-side columns.

#### Scenario: Single-column settings at desktop width

- **WHEN** the user views the settings page at a viewport wider than ~800px
- **THEN** all settings cards stack vertically in one column
- **AND** no two cards appear side by side

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
- **THEN** the stacked settings cards span the full main content width without a narrow column floating in unused space
