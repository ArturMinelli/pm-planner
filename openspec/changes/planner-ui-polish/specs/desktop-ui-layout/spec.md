## Purpose

Ensure the desktop frontend layout reflows cleanly at all reasonable window sizes so labels, badges, and controls never overflow or cause horizontal scroll.

## ADDED Requirements

### Requirement: No horizontal scroll at reasonable window sizes

The desktop app SHALL NOT show horizontal scrollbars or clip text at any viewport width from 600px through full desktop width. Content SHALL reflow using stacking, wrapping, or grid column changes instead of overflowing.

#### Scenario: Resize planner window

- **WHEN** the user resizes the desktop window to 800px wide on the planner page
- **THEN** all visible text and controls remain fully readable without horizontal scroll

#### Scenario: Narrow window still usable

- **WHEN** the viewport is 600px wide
- **THEN** the app remains usable with stacked layouts and no clipped labels

### Requirement: Planner grid reflows at medium breakpoints

The planner page two-column layout (journey list + summary) SHALL switch to a single column when horizontal space is insufficient (~800px). The summary panel SHALL appear below the journey list without squeezing columns below readable minimums.

#### Scenario: Two columns at wide width

- **WHEN** the viewport is wider than the medium breakpoint
- **THEN** journey list and summary display side by side

#### Scenario: Single column at medium width

- **WHEN** the viewport is at or below the medium breakpoint (~800px)
- **THEN** journey list and summary stack vertically with full-width cards

### Requirement: Slot fields stack label and badge rows

Journey slot fields SHALL stack their label row and value row vertically when horizontal space is tight. Badges (`registrada`, `calculada`, balance) SHALL wrap to a new line within the label area instead of overflowing past the field boundary.

#### Scenario: Badges wrap in label row

- **WHEN** a slot field label row contains a label plus multiple badges at medium width
- **THEN** badges wrap below or beside the label without clipping or pushing content outside the field

#### Scenario: Values stack below label row

- **WHEN** a slot field has a calculated time, calculator button, and alternative time at narrow width
- **THEN** value controls stack vertically with full-width touch targets

### Requirement: Sidebar keeps horizontal nav at narrow widths

At narrow viewport widths, the sidebar SHALL continue using the existing horizontal scrollable nav pattern. The sidebar SHALL NOT introduce a hamburger menu or icon-only collapse in this change.

#### Scenario: Narrow sidebar nav

- **WHEN** the viewport is below the mobile breakpoint
- **THEN** navigation links display in a horizontal scrollable row as today

### Requirement: App-wide overflow-safe text

Page headers, settings form labels, stat rows, reminder chips, and toast messages SHALL use overflow-safe CSS (min-width constraints, wrapping, or truncation with accessible full text) so long Portuguese labels never clip.

#### Scenario: Long settings label

- **WHEN** a settings field label exceeds the available column width
- **THEN** the label wraps to multiple lines instead of overflowing

#### Scenario: Summary stat row

- **WHEN** a summary stat label and value share a row at medium width
- **THEN** both remain visible without overlap or horizontal scroll
