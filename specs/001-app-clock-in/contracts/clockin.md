# Contracts

## Internal: `POST /api/clockin` (dev server) / `ClockIn(req)` (Wails)
Request: `{ "location": {latitude, longitude, accuracy, source}, "confirmNear": false }`
Response 200: `{ "status": "registered", "punch": {id,date,time,address}, "locationSource": "detected|default" }`
Errors (all JSON `{status, message, registered:boolean}`): `401 session_expired`, `409 too_close` (needs `confirmNear`), `502 service_rejected`, `504 uncertain` (`registered:null`), `503 offline`.
`registered` always states definitively whether a punch exists (FR-007).

## External (PontoMais, replayed from HAR)
- `GET /api/session` → `session.employee`
- `GET /api/geo_locations?lat=&lng=&employee_id=` → `address.formatted_address`
- `POST /api/time_cards/register` → 202 + `time_card`
- `GET /api/time_cards/current/last_cached` → `time_cards[]` (id, date, time, address, latitude, longitude, disabled)
Production calls are forbidden in tests; see research R7.
