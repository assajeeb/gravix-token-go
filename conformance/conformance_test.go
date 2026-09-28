// Conformance: tokens from this library verify with the Gravix server's own
// verifier (livekit/protocol/auth, the code in pkg/service/auth.go), carrying the
// grants the server reads. A separate module, so the library itself stays light.
package conformance

import (
	"os"
	"strings"
	"testing"
	"time"

	gravixtoken "github.com/assajeeb/gravix-token-go"
	"github.com/livekit/protocol/auth"
)

const cert = "cert-0123456789abcdef0123456789abcdef"

func verify(t *testing.T, jwt, secret string) *auth.ClaimGrants {
	t.Helper()
	v, err := auth.ParseAPIToken(jwt)
	if err != nil {
		t.Fatal(err)
	}
	_, g, err := v.Verify(secret)
	if err != nil {
		t.Fatal(err)
	}
	if v.APIKey() != "APPconf" {
		t.Fatalf("iss %q", v.APIKey())
	}
	return g
}

func TestServerVerifierAcceptsAndReadsTheGrants(t *testing.T) {
	jwt, err := gravixtoken.New("APPconf", cert).Room("standup").Identity("alice").Name("Alice").
		Metadata("m").Attributes(map[string]string{"seat": "1"}).CanPublish(false).TTL(time.Hour).ToJWT()
	if err != nil {
		t.Fatal(err)
	}
	g := verify(t, jwt, cert)
	if g.Identity != "alice" || g.Name != "Alice" || g.Metadata != "m" || g.Attributes["seat"] != "1" {
		t.Fatalf("grants %+v", g)
	}
	if !g.Video.RoomJoin || g.Video.Room != "APPconf_standup" || g.Video.GetCanPublish() || !g.Video.GetCanSubscribe() || !g.Video.GetCanPublishData() || !g.Video.GetCanUpdateOwnMetadata() {
		t.Fatalf("video %+v", g.Video)
	}
	if !strings.HasPrefix(g.Video.Room, "APPconf_") {
		t.Fatal("the server's app scoping would refuse this room")
	}
}

func TestWrongCertificateFails(t *testing.T) {
	jwt, _ := gravixtoken.New("APPconf", cert).Room("r").Identity("a").ToJWT()
	v, _ := auth.ParseAPIToken(jwt)
	if _, _, err := v.Verify("another-cert-0123456789abcdef0123456"); err == nil {
		t.Fatal("verified with the wrong certificate")
	}
}

// Tokens minted by other Gravix token libraries (e.g. the TypeScript one) are
// dropped as files in GRAVIX_FOREIGN_TOKENS (one "<jwt>" per line, certificate
// = cert above, app APPconf) and must verify here too.
func TestForeignLibraryTokens(t *testing.T) {
	path := os.Getenv("GRAVIX_FOREIGN_TOKENS")
	if path == "" {
		t.Skip("GRAVIX_FOREIGN_TOKENS not set")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		g := verify(t, strings.TrimSpace(line), cert)
		if !strings.HasPrefix(g.Video.Room, "APPconf_") || g.Identity == "" || !g.Video.RoomJoin {
			t.Fatalf("foreign token grants %+v", g)
		}
		n++
	}
	t.Logf("%d foreign tokens verified", n)
}
