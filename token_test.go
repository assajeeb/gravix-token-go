package gravixtoken

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func decode(t *testing.T, jwt string) map[string]any {
	t.Helper()
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWT: %q", jwt)
	}
	b, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestTokenCarriesTheAppPrefixedRoomAndDefaults(t *testing.T) {
	jwt, err := New("APPabc", "secret-0123456789abcdef0123456789ab").
		Room("standup").Identity("alice").Name("Alice").TTL(2 * time.Hour).ToJWT()
	if err != nil {
		t.Fatal(err)
	}
	c := decode(t, jwt)
	if c["iss"] != "APPabc" || c["sub"] != "alice" || c["name"] != "Alice" {
		t.Fatalf("claims %v", c)
	}
	v := c["video"].(map[string]any)
	if v["room"] != "APPabc_standup" || v["roomJoin"] != true || v["canPublish"] != true || v["canSubscribe"] != true || v["canPublishData"] != true || v["canUpdateOwnMetadata"] != true {
		t.Fatalf("video grant %v", v)
	}
	if exp, nbf := c["exp"].(float64), c["nbf"].(float64); exp-nbf != 7200 {
		t.Fatalf("ttl %v", exp-nbf)
	}
}

func TestAnAlreadyPrefixedRoomIsNotPrefixedTwice(t *testing.T) {
	jwt, _ := New("APPabc", "s-0123456789abcdef0123456789abcdef").Room("APPabc_standup").Identity("a").ToJWT()
	if got := decode(t, jwt)["video"].(map[string]any)["room"]; got != "APPabc_standup" {
		t.Fatalf("room %v", got)
	}
}

func TestListenerOnlyAndAttributes(t *testing.T) {
	jwt, _ := New("APPabc", "s-0123456789abcdef0123456789abcdef").Room("r").Identity("b").
		CanPublish(false).Attributes(map[string]string{"seat": "3"}).Metadata(`{"x":1}`).ToJWT()
	c := decode(t, jwt)
	v := c["video"].(map[string]any)
	if v["canPublish"] != false || v["canSubscribe"] != true {
		t.Fatalf("grant %v", v)
	}
	if c["attributes"].(map[string]any)["seat"] != "3" || c["metadata"] != `{"x":1}` {
		t.Fatalf("claims %v", c)
	}
}

func TestRefusesWhatTheServerWouldRefuse(t *testing.T) {
	cases := []struct {
		name string
		b    *Builder
	}{
		{"no app id", New("", "s-0123456789abcdef0123456789abcdef").Room("r").Identity("a")},
		{"app id with _", New("APP_x", "s-0123456789abcdef0123456789abcdef").Room("r").Identity("a")},
		{"short certificate", New("APPabc", "short").Room("r").Identity("a")},
		{"no room", New("APPabc", "s-0123456789abcdef0123456789abcdef").Identity("a")},
		{"no identity", New("APPabc", "s-0123456789abcdef0123456789abcdef").Room("r")},
		{"ttl over 24h", New("APPabc", "s-0123456789abcdef0123456789abcdef").Room("r").Identity("a").TTL(25 * time.Hour)},
	}
	for _, c := range cases {
		if _, err := c.b.ToJWT(); err == nil {
			t.Errorf("%s: no error", c.name)
		}
	}
}

func TestRoomNameHelpers(t *testing.T) {
	if RoomName("APPa", "x") != "APPa_x" || DisplayName("APPa", "APPa_x") != "x" || DisplayName("APPa", "other") != "other" {
		t.Fatal("helpers")
	}
}

// The Gravix SDKs publish their own attributes (Flutter: cameraFacing); the old
// auth service granted this by default, so the libraries do too, and it can be off.
func TestCanUpdateOwnMetadataDefaultsOnAndCanBeOff(t *testing.T) {
	jwt, _ := New("APPabc", "s-0123456789abcdef0123456789abcdef").Room("r").Identity("a").CanUpdateOwnMetadata(false).ToJWT()
	if got := decode(t, jwt)["video"].(map[string]any)["canUpdateOwnMetadata"]; got != false {
		t.Fatalf("canUpdateOwnMetadata %v", got)
	}
}

// Errors end up in app logs and crash reports: they name the field, never the value.
func TestErrorsNameTheFieldButNeverQuoteCallerValues(t *testing.T) {
	const app, room, id, badCert, cert = "LEAKAPP9f3c", "leakroom77", "leakuser55", "leakcert", "s-0123456789abcdef0123456789abcdef"
	cases := []struct {
		field string
		b     *Builder
	}{
		{"app ID", New("", cert).Room(room).Identity(id)},
		{"app ID", New(app+"_x", cert).Room(room).Identity(id)},
		{"app certificate", New(app, badCert).Room(room).Identity(id)},
		{"room", New(app, cert).Identity(id)},
		{"identity", New(app, cert).Room(room)},
		{"TTL", New(app, cert).Room(room).Identity(id).TTL(25*time.Hour + 17*time.Second)},
	}
	for _, c := range cases {
		_, err := c.b.ToJWT()
		if err == nil {
			t.Fatalf("%s: no error", c.field)
		}
		msg := err.Error()
		if !strings.Contains(msg, c.field) {
			t.Errorf("error %q does not name %q", msg, c.field)
		}
		for _, v := range []string{app, room, id, badCert, cert, "25h0m17s"} {
			if strings.Contains(msg, v) {
				t.Errorf("error %q quotes a caller value", msg)
			}
		}
	}
}
