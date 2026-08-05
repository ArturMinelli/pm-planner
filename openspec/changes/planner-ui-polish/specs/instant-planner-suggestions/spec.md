## MODIFIED Requirements

### Requirement: Timeline presentation of suggestions

The planner SHALL present slots in a vertical timeline: label, prominent suggested time, optional registrada pill, and inline ordering warning only when times conflict. Balance badges and alternative exit times SHALL NOT appear on individual slot rows. At medium and narrow widths, the label row and value row SHALL stack vertically so badges and controls wrap instead of overflowing.

#### Scenario: Clean slot row

- **WHEN** a slot has a suggestion and is not registrada
- **THEN** the row shows the label and suggested time as the primary visual element without extra chrome on that row

#### Scenario: Balance lives in summary

- **WHEN** a balance adjustment applies today
- **THEN** the alternative exit and bank explanation appear only in the Resumo panel

#### Scenario: Slot row reflows at medium width

- **WHEN** the viewport is at or below the medium breakpoint and a slot row has label, badges, and value controls
- **THEN** badges wrap within the label area and value controls stack without horizontal overflow
