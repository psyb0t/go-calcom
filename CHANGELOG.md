# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.0, 2026-10-17

- Added `calcom`, a REST client generated with oapi-codegen from the Cal.com repository's public OpenAPI document, titled Cal.diy API v2, spec version 1.0.0, OpenAPI 3.0.0. It covers all 82 paths and 121 operations in that document: bookings, event types, availability, schedules, calendars, teams, webhooks and profile.
- Number fields are `float64` instead of oapi-codegen's default `float32`.
- The malformed self-referential reassignment response data is an open JSON object, preserving fields from the missing upstream response schemas.
- `make generate` downloads the spec from https://raw.githubusercontent.com/calcom/cal.com/main/docs/api-reference/v2/openapi.json, gives duplicate calendar event operation IDs distinct names, and regenerates the client. The spec is not committed.
- Added local HTTP tests, the README example, development tooling and the `go-calcom` agent skill.
