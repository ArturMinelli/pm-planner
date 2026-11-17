# Feature Specification: In-App Clock-In

**Feature Branch**: `001-app-clock-in`

**Created**: 2026-11-17

**Status**: Draft

**Input**: User description: "Build a clock-in functionality for this app, where the user can register their clock-ins through the app itself, following the behavior observed in the provided PontoMais HAR capture. It must work smoothly, and its UI/UX must be validated in a browser without ever hitting the production time-tracking service during testing."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Register a clock-in with one action (Priority: P1)

A signed-in user opens the planner and sees today's punches along with a prominent "Clock in" action. They press it, confirm if prompted, and the app records the punch with the time-tracking service. The new punch appears in today's list immediately, with the time it was registered, and the planner recalculates the remaining schedule.

**Why this priority**: This is the entire purpose of the feature: replacing the trip to the official PontoMais app/website with a single action inside the planner.

**Independent Test**: With a signed-in user and a simulated time-tracking service, press "Clock in" and verify a new punch with the current time appears in today's list and a success message is shown.

**Acceptance Scenarios**:

1. **Given** a signed-in user viewing today, **When** they press "Clock in", **Then** a punch is submitted once, a success confirmation is shown, and the punch appears in today's list with its registered time.
2. **Given** a punch was just registered, **When** the planner refreshes, **Then** the day's worked time, balance and remaining slots reflect the new punch.
3. **Given** the user presses "Clock in", **When** the request is in progress, **Then** the action is disabled and shows a progress state so it cannot be triggered twice.

---

### User Story 2 - Clear feedback on success, failure and offline (Priority: P2)

When registration fails (no connection, expired session, service rejection, or the service is slow), the user is told plainly what happened and whether the punch was recorded, and can retry safely without creating duplicate punches.

**Why this priority**: A clock-in has legal/payroll weight; a user who is unsure whether a punch was saved will double-punch or not punch at all.

**Independent Test**: Simulate each failure type and verify the message, that no punch is shown as recorded, and that retry works and does not duplicate.

**Acceptance Scenarios**:

1. **Given** the service is unreachable, **When** the user clocks in, **Then** an error says the punch was NOT registered and a retry action is offered.
2. **Given** the session expired, **When** the user clocks in, **Then** the user is guided to sign in again and the punch is not lost silently.
3. **Given** the request outcome is uncertain (timeout), **When** the app recovers, **Then** it checks the day's punches before allowing a retry and warns if a punch was in fact registered.

---

### User Story 3 - Location detected automatically, with a default (Priority: P3)

The time-tracking service requires a location with each punch. The app detects the user's location automatically, with no interaction. If detection is unavailable, denied or fails, it silently uses a built-in default location: Av. Presidente Kennedy, 861 - 1º Andar - Campinas, São José - SC, 88102-401. The location used is shown at punch time.

**Why this priority**: Required for punches to be accepted, but it must never add friction to Stories 1–2.

**Independent Test**: Simulate a detected location and verify the punch carries it; simulate detection failure and verify the punch carries the default address in the format the service expects.

**Acceptance Scenarios**:

1. **Given** the location can be detected, **When** the user clocks in, **Then** the punch uses the detected location with no prompt.
2. **Given** detection is unavailable, denied or times out, **When** the user clocks in, **Then** the punch uses the default location with no prompt and the UI shows that the default was used.
3. **Given** any location source, **When** a punch is sent, **Then** location is structured exactly as the service expects (coordinates, address, accuracy).

---

### Edge Cases

- Rapid double-click or double key press on "Clock in": only one punch is submitted.
- A punch registered very close to a previous one (e.g., within a minute): the user is warned and must confirm.
- Clocking in from another device/app between planner loads: the day's list refreshes before/after punching so the view stays accurate.
- Local clock differs from the service's time: the registered time shown is the one the service reports.
- Day rolls over to the next date while the screen is open.
- Location permission denied or detection slow: the punch is not delayed noticeably and falls back to the default.
- Window offline at launch or loses connectivity mid-request.
- Testing mode: during automated or manual UI validation, no request may reach the production service.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to register a clock-in from within the planner with a single prominent action.
- **FR-002**: The system MUST submit exactly one punch per user action, and prevent repeat submissions while one is in progress.
- **FR-003**: The system MUST automatically detect the user's location for each punch without user interaction, and MUST fall back to the built-in default location (Av. Presidente Kennedy, 861 - 1º Andar - Campinas, São José - SC, 88102-401) when detection is unavailable, denied or fails.
- **FR-004**: The location sent MUST be in the structure the time-tracking service expects, and the UI MUST show which location (detected or default) is being used.
- **FR-005**: The system MUST show the newly registered punch (with the time confirmed by the service) in today's list immediately after success.
- **FR-006**: The system MUST refresh the day's totals, balance and remaining schedule after a successful punch.
- **FR-007**: The system MUST state clearly whether a punch was registered or not for every outcome (success, rejection, offline, expired session, timeout).
- **FR-008**: The system MUST verify against the day's recorded punches before permitting a retry after an uncertain outcome, to avoid duplicates.
- **FR-009**: The system MUST ask for confirmation when a new punch is very close in time to the previous one.
- **FR-010**: The system MUST require authentication using the user's existing sign-in and guide them to re-authenticate when it has expired.
- **FR-011**: The clock-in action MUST be reachable and usable by keyboard and expose accessible labels and status announcements.
- **FR-012**: The system MUST support a safe test mode in which clock-in requests are intercepted/simulated and never reach the production time-tracking service, and test mode MUST be unmistakably indicated in the UI.
- **FR-013**: The UI and UX of the feature MUST be verified in a real browser against simulated responses covering success, failure, offline, expired-session and duplicate-click cases.

### Key Entities

- **Punch**: A single clock-in record: date, time (as confirmed by the service), location, and whether it is disabled/invalid.
- **Work Location**: Coordinates, address and accuracy used for a punch, from automatic detection or the built-in default.
- **Day Punch List**: The ordered punches for the current day that drive worked time, balance and remaining schedule.
- **Clock-In Attempt**: A transient state (idle, in progress, succeeded, failed, uncertain) shown to the user.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can register a clock-in in 2 clicks or fewer from opening the planner.
- **SC-002**: After pressing "Clock in", the user sees a definitive result (registered or not) within 5 seconds in normal conditions.
- **SC-003**: 100% of simulated rapid double-presses result in exactly one punch.
- **SC-004**: In 100% of failure scenarios tested, the message states whether the punch was registered, and retry never produces a duplicate.
- **SC-005**: Zero requests reach the production time-tracking service during all testing.
- **SC-006**: The registered punch appears in today's list and updated totals within 2 seconds of success.

## Assumptions

- Users are already signed in with their existing PontoMais credentials; this feature reuses that sign-in.
- The punch request format follows the behavior observed in the provided capture (punch with location, device info and an optional photo that is not used here); photo capture and justifications/edited punches are out of scope.
- Location is auto-detected; a desktop app may lack a reliable source, hence the default fallback. The default address must be converted to coordinates and the service's address format during planning (not guessed), and detected coordinates must be reverse-geocoded to an address where required.
- Only registering a punch is in scope; editing or deleting past punches is out of scope.
- Delivery will be orchestrated with lower-cost sub-agents under an orchestrator who reviews and validates the output; this is a process choice, not a product requirement.
- Test mode (simulated service) is available for validation so the production service is never touched.
