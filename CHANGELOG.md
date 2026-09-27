# Changelog

All notable changes to `github.com/gravitycompile/gravix-token-go`. This project
follows [semantic versioning](https://semver.org/); releases are git tags `vX.Y.Z`.

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
