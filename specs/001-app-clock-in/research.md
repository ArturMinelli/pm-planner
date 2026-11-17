# Research

## R1. What does a punch request look like?
**Decision**: `POST {base}/api/time_cards/register`, JSON body `{image:null, employee:<session employee>, time_card:{latitude, longitude, address, original_latitude, original_longitude, original_address, location_edited:false, accuracy, accuracy_method:null, image:null, info:null}, _path:"/registrar-ponto", _appVersion, _device:{manufacturer,model,uuid,version}}`, with the existing auth headers plus `api-version: 2` and a `uuid` header. Response is **202** with `time_card`, `message_id`, `timestamp`.
**Rationale**: Taken directly from the HAR (values inspected structurally; PII not copied).
**Alternatives**: none; this is the only observed write endpoint.

## R2. Where does the `employee` object come from?
**Decision**: `GET /api/session` (read-only) returns `session.employee`, byte-identical to the one the web app posts. Fetch it per punch (or cache for the session) and embed it.
**Open risk**: the server may accept a minimal `{id}`. This cannot be tested against production, so we replay the full object, which is what the real client does.

## R3. How do we get the address string in the service's format?
**Decision**: `GET /api/geo_locations?lat&lng&employee_id` returns `address.formatted_address` (a plus-code style string). Use its result for `address`/`original_address`. It is read-only.
**Fallback**: if it fails, send the human-readable address text.

## R4. Automatic location
**Decision**: webview `navigator.geolocation.getCurrentPosition` with a 3 s timeout and no permission UI; on any error/denial/timeout use `DEFAULT_LOCATION`. Accuracy comes from the position (HAR shows ~1100 m).
**Rationale**: no third-party call and no IP leak. **Alternative rejected**: IP-based lookup (sends the user's IP to a third party, coarse, extra dependency).
**Risk**: Wails/WebKitGTK may not support geolocation, in which case the default is used every time. Verify on the real desktop build.

## R5. Default location coordinates
**Decision**: `DEFAULT_LOCATION` = Av. Presidente Kennedy, 861, Campinas, São José/SC, 88102-401, with latitude/longitude to be geocoded once at implementation time. Then run them through R3 for the service-format address and hard-code the results as constants.
**Action for the user**: confirm the coordinates (or provide ones from a previous punch at that office). We will not guess them.

## R6. Duplicate protection
**Decision**: frontend disables the button while in flight; the backend holds a mutex plus a 60 s "last punch" guard and requires `confirmNear:true` to override (FR-009). After a timeout or network error, call `last_cached` (read-only), and if a new punch exists, report success instead of retrying (FR-008).

## R7. Never touching production in tests
**Decision**: three layers. (1) `cmd/pm-fake-pontomais` serves `/api/session`, `/api/geo_locations`, `/api/time_cards/register` and `/api/time_cards/current/last_cached` with scenario switches (ok, 401, 500, slow, offline). (2) `pm-dev --fake` sets `Client.BaseURL` to it and injects static `AuthHeaders` (skips sign-in, which has a separate hard-coded URL). (3) Playwright `browser_network_requests` must show no host other than localhost, and a route block on `*.pontomais.com.br`. The UI shows a visible "TEST MODE" banner when fake (FR-012).
