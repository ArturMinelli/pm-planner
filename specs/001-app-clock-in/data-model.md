# Data Model

- **Location**: `latitude`, `longitude` (float), `accuracy` (m), `address` (service-format string), `source` ∈ {detected, default}.
- **Punch**: `id`, `date` (YYYY-MM-DD), `time` (HH:MM:SS, as reported by the service), `latitude`, `longitude`, `address`, `disabled`.
- **ClockInAttempt** (UI state): `idle → locating → sending → success | failed | uncertain`. `uncertain → (verify day) → success | idle`. `failed` retries return to `locating`.
- **Rules**: a punch is never sent without a Location (default guarantees this); a new punch within 60 s of the last requires explicit confirmation; displayed time is the service's, never the local clock.
