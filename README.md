# gravix-token-go

Mint Gravix Cloud join tokens **in your own Go backend**. The Gravix server verifies them; no Gravix service is called at join time.

## Install

```sh
go get github.com/gravitycompile/gravix-token-go@latest
```

## Usage

```go
import gravixtoken "github.com/gravitycompile/gravix-token-go"

jwt, err := gravixtoken.New(os.Getenv("GRAVIX_APP_ID"), os.Getenv("GRAVIX_APP_CERTIFICATE")).
	Room("standup").          // server-side room "<AppID>_standup"
	Identity(user.ID).
	Name(user.Name).
	ToJWT()
```

Builder methods: `Room`, `Identity` (required), `Name`, `Metadata`, `Attributes`, `TTL` (default 6h, max 24h), `CanPublish`, `CanSubscribe`, `CanPublishData`, `CanPublishSources`, `Hidden`, `RoomAdmin`. Helpers: `RoomName`, `DisplayName`. Only dependency: `github.com/golang-jwt/jwt/v5`. Go 1.21+.

`conformance/` (a separate module) verifies these tokens — and tokens from the TypeScript library (`GRAVIX_FOREIGN_TOKENS`) — with the Gravix server's own verifier.

## Development

```sh
go test ./...                      # the library (go 1.21+)
(cd conformance && go test ./...)  # server-verifier conformance (go 1.26+)
```

The conformance module needs Go 1.26 because its dependency `github.com/livekit/protocol`
(the server's verifier) declares `go 1.26`. It is a separate module with a `replace ../`,
so the library's users never download it or inherit that requirement.

A TypeScript version with the same options is `@gravitycompile/gravix-token` on npm.

## License

Apache-2.0. See [LICENSE](./LICENSE).
