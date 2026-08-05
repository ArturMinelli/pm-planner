## Purpose

Ensure planner layout reflows based on the space an element actually has, not on viewport width, so content never paints outside its container at any supported window size.

## MODIFIED Requirements

### Requirement: No horizontal scroll at reasonable window sizes

The desktop app SHALL NOT show horizontal scrollbars or clip text at any window width from the application's 720px minimum through full desktop width. Content SHALL reflow using stacking, wrapping, or grid column changes instead of overflowing.

The lower bound is corrected from 600px to 720px: `desktop/main.go` sets `MinWidth: 720`, so no window narrower than that is reachable and a 600px guarantee is untestable. Additionally, the absence of a scrollbar SHALL NOT be treated as evidence of compliance, because `.main` sets `overflow-x: hidden` and therefore masks overflow rather than preventing it; compliance requires that no element extend beyond its container's box.

#### Scenario: Resize planner window

- **WHEN** the user resizes the desktop window to 800px wide on the planner page
- **THEN** all visible text and controls remain fully readable without horizontal scroll

#### Scenario: Minimum window width still usable

- **WHEN** the window is at its 720px minimum width
- **THEN** the app remains usable with stacked layouts and no clipped labels

#### Scenario: Hidden overflow does not count as passing

- **WHEN** an element's content would extend past its container
- **THEN** the layout is non-compliant even though `.main { overflow-x: hidden }` suppresses the scrollbar

### Requirement: Planner grid reflows at medium breakpoints

The planner page two-column layout (journey list + summary) SHALL use a journey-primary column ratio of approximately 60% journey list and 40% summary at desktop widths wider than the medium breakpoint. The layout SHALL switch to a single column when horizontal space is insufficient (~800px). Both columns SHALL fill their allocated width without large empty gaps beside cards.

The journey column SHALL declare a non-zero minimum width sufficient to hold a journey card's two slot fields at a usable width. The current `minmax(0, 1.5fr)` track leaves the primary column with no floor, so the existing clause forbidding columns "squeezed below readable minimums" is unenforced exactly where it matters most: at the 980px default the journey track resolves to 418px, which squeezes each slot field to roughly 174px.

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

#### Scenario: Journey column has an enforced floor

- **WHEN** available width would reduce the journey column below its declared minimum
- **THEN** the grid collapses to a single column rather than shrinking the journey column past that minimum

### Requirement: Page content fills available column width

Page layouts (planner, settings, shared page chrome) SHALL use the available main content width so columns distribute space evenly. Content areas SHALL NOT leave large unused horizontal margins or gutters inside the main panel at desktop widths from ~900px upward. Where page width is capped, the capped content SHALL be centered within the main panel so the leftover space is distributed evenly rather than accumulating on one side.

#### Scenario: Planner fills main panel

- **WHEN** the planner page is viewed at ~1000px viewport width
- **THEN** the two-column grid spans the main content area with balanced column widths and no large empty band beside the grid

#### Scenario: Settings fills main panel

- **WHEN** the settings page is viewed at ~1000px viewport width
- **THEN** the stacked settings cards span the full main content width without a narrow column floating in unused space

#### Scenario: Capped page is centered on wide windows

- **WHEN** the window is wide enough that the page hits its width cap, for example 1920px
- **THEN** the capped content is centered in the main panel rather than left-aligned with all leftover space collected on the right

### Requirement: Slot fields stack label and badge rows

Journey slot fields SHALL reflow based on their **own inline size**, not the viewport width, because a slot field can be narrow inside a wide window. A slot field SHALL contain at most one badge (`calculada`); the balance badge and alternative exit time SHALL NOT be rendered inside a slot field. When a slot field's inline size cannot fit its label and remaining content on one line, the label row and value row SHALL stack vertically. No slot-field descendant SHALL render outside the slot field's boundary at any window size from the 720px application minimum through full desktop width.

#### Scenario: Narrow column inside a wide window

- **WHEN** the window is at its 980px default width, making a slot field roughly 170px wide
- **THEN** the slot field's label, `calculada` badge, time value, and calculator button all render inside the slot field boundary, and the slot field renders inside its journey card boundary

#### Scenario: Reflow is driven by field width not viewport width

- **WHEN** a slot field is too narrow for its contents while the viewport is wider than 800px
- **THEN** the field reflows (stacks or wraps) anyway, because the rule is keyed to the field's own inline size

#### Scenario: Values stack at the application minimum width

- **WHEN** the window is at its 720px minimum width
- **THEN** the time value and calculator button stack vertically with full-width targets and no horizontal overflow

#### Scenario: Calculated highlight consumes no extra width

- **WHEN** a slot field becomes the calculated slot and gains its highlight treatment
- **THEN** the field's outer width is unchanged, so gaining the highlight cannot itself cause overflow

## ADDED Requirements

### Requirement: Card content never paints outside its card

Every card and journey group SHALL constrain its descendants so no child renders outside the card's border, regardless of content length or available width. Layout containers SHALL permit their children to shrink (no shrink-blocked flex children in space-constrained rows) and SHALL apply overflow containment as a structural backstop.

#### Scenario: Overlong content stays inside the card

- **WHEN** a journey card's contents are wider than the card's inner width for any reason
- **THEN** the content is clipped, wrapped, or shrunk inside the card border rather than drawn over or past it

#### Scenario: Backstop survives future content additions

- **WHEN** a new badge or control is later added to a slot field without accompanying sizing rules
- **THEN** it cannot paint outside the journey card, because the card enforces overflow containment independently

### Requirement: Planner layout verified at three reference widths

Planner layout changes SHALL be verified in the running application at 720px (application minimum), 980px (default), and 1280px (wide) window widths. At each width there SHALL be no horizontal scrollbar and no element painting outside its card.

#### Scenario: Verification at reference widths

- **WHEN** the planner page is rendered at 720px, 980px, and 1280px window width with a two-journey day where the second exit is calculated and a balance adjustment applies
- **THEN** each width shows no horizontal scroll and no element outside a card boundary
