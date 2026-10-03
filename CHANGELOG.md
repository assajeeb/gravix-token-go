# Changelog

All notable changes to `github.com/assajeeb/gravix-token-go` (module path
`github.com/gravitycompile/gravix-token-go` before 2026-09-28). This project
follows [semantic versioning](https://semver.org/); releases are git tags `vX.Y.Z`.

## v0.1.2 (2026-10-03)

- Error messages no longer quote caller values. Validation errors from `ToJWT`
  name the field and the rule (for example `app ID must not contain "_"`, `TTL must
  be in (0, 24h0m0s]`) but never include the App ID, certificate, room, identity,
  TTL or any other value you passed, so they are safe to log or send to crash
  reporting.
- Module path changed from `github.com/gravitycompile/gravix-token-go` to
  `github.com/assajeeb/gravix-token-go` (2026-09-28). Update imports and `go get` lines.
- License changed from Apache-2.0 to MIT (Copyright (c) 2024-2026 Gravity
  Compile). All code is Gravity Compile's.

## v0.1.1

- `CanUpdateOwnMetadata` is granted by default (`CanUpdateOwnMetadata(false)` turns
  it off). Parity with the old Gravix auth service; the Flutter SDK publishes its
  camera facing as a participant attribute.

## v0.1.0

- First release. Builder API (`New(appID, certificate)`, `Room`, `Identity`, `Name`,
  `Metadata`, `Attributes`, `TTL`, grant setters, `ToJWT`), app-prefixed rooms
  (`RoomName` / `DisplayName`), input validation. Only dependency:
  `github.com/golang-jwt/jwt/v5`.
- `conformance/` module: tokens verify with the Gravix server's own verifier.
