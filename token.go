// Package gravixtoken mints Gravix Cloud join tokens in your own backend.
//
// Your backend holds the App ID and App Certificate from the Gravix console and
// signs a token for each user locally, in microseconds. Your app passes the token
// to the Gravix SDK, and the Gravix server verifies it. No Gravix service is
// called at join time, so a join never waits on (or fails because of) a token
// server.
//
//	jwt, err := gravixtoken.New(appID, appCertificate).
//		Room("standup").Identity(userID).Name(displayName).
//		ToJWT()
//
// Never ship the App Certificate in a mobile or web app: mint on your server.
package gravixtoken

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// Separator between the App ID and your room name. Room names are global on
	// the Gravix servers, so every room of an app is "<AppID>_<room>"; this
	// library adds the prefix, and the server refuses a room outside it.
	Separator = "_"
	// DefaultTTL is how long a token stays valid for joining.
	DefaultTTL = 6 * time.Hour
	// MaxTTL bounds a token's life.
	MaxTTL = 24 * time.Hour
)

// Builder collects one token's claims. Methods chain; ToJWT signs.
type Builder struct {
	appID, cert                string
	room, identity, name, meta string
	attrs                      map[string]string
	ttl                        time.Duration
	pub, sub, data             bool
	hidden, admin              bool
	canPublishSources          []string
	now                        func() time.Time
}

// New starts a token for the app. Publish, subscribe and data are allowed by default.
func New(appID, appCertificate string) *Builder {
	return &Builder{appID: appID, cert: appCertificate, ttl: DefaultTTL, pub: true, sub: true, data: true, now: time.Now}
}

// Room sets the room to join; the "<AppID>_" prefix is added if missing.
func (b *Builder) Room(name string) *Builder { b.room = RoomName(b.appID, name); return b }

// Identity is the user's unique id within the room (required).
func (b *Builder) Identity(id string) *Builder { b.identity = id; return b }

// Name is the display name other participants see.
func (b *Builder) Name(n string) *Builder { b.name = n; return b }

// Metadata is an opaque string visible to other participants.
func (b *Builder) Metadata(m string) *Builder { b.meta = m; return b }

// Attributes are key/value pairs visible to other participants. The key
// "gravix.app" is reserved: the server overwrites it with the verified App ID.
func (b *Builder) Attributes(a map[string]string) *Builder { b.attrs = a; return b }

// TTL sets how long the token can be used to join (default 6h, max 24h).
func (b *Builder) TTL(d time.Duration) *Builder { b.ttl = d; return b }

// CanPublish (default true); false makes a listener.
func (b *Builder) CanPublish(v bool) *Builder { b.pub = v; return b }

// CanSubscribe (default true).
func (b *Builder) CanSubscribe(v bool) *Builder { b.sub = v; return b }

// CanPublishData (default true): data messages and chat.
func (b *Builder) CanPublishData(v bool) *Builder { b.data = v; return b }

// CanPublishSources limits publishing to these sources ("camera", "microphone",
// "screen_share", "screen_share_audio"). Empty = any.
func (b *Builder) CanPublishSources(s ...string) *Builder { b.canPublishSources = s; return b }

// Hidden: the participant is invisible to others (recorders, bots).
func (b *Builder) Hidden(v bool) *Builder { b.hidden = v; return b }

// RoomAdmin: may mute/remove others in THIS room only.
func (b *Builder) RoomAdmin(v bool) *Builder { b.admin = v; return b }

type videoGrant struct {
	RoomJoin          bool     `json:"roomJoin"`
	Room              string   `json:"room"`
	RoomAdmin         bool     `json:"roomAdmin,omitempty"`
	CanPublish        *bool    `json:"canPublish"`
	CanSubscribe      *bool    `json:"canSubscribe"`
	CanPublishData    *bool    `json:"canPublishData"`
	CanPublishSources []string `json:"canPublishSources,omitempty"`
	Hidden            bool     `json:"hidden,omitempty"`
}

type claims struct {
	jwt.RegisteredClaims
	Name       string            `json:"name,omitempty"`
	Metadata   string            `json:"metadata,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Video      *videoGrant       `json:"video"`
}

// ToJWT validates the claims and signs them (HS256 with the App Certificate).
func (b *Builder) ToJWT() (string, error) {
	switch {
	case b.appID == "":
		return "", errors.New("gravixtoken: app ID is empty")
	case strings.Contains(b.appID, Separator):
		return "", fmt.Errorf("gravixtoken: app ID %q contains %q", b.appID, Separator)
	case len(b.cert) < 32:
		return "", errors.New("gravixtoken: app certificate is shorter than 32 characters")
	case b.room == "":
		return "", errors.New("gravixtoken: room is required")
	case b.identity == "":
		return "", errors.New("gravixtoken: identity is required")
	case b.ttl <= 0 || b.ttl > MaxTTL:
		return "", fmt.Errorf("gravixtoken: TTL %v must be in (0, %v]", b.ttl, MaxTTL)
	}
	now := b.now()
	pub, sub, data := b.pub, b.sub, b.data
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    b.appID,
			Subject:   b.identity,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(b.ttl)),
		},
		Name:       b.name,
		Metadata:   b.meta,
		Attributes: b.attrs,
		Video: &videoGrant{
			RoomJoin: true, Room: b.room, RoomAdmin: b.admin,
			CanPublish: &pub, CanSubscribe: &sub, CanPublishData: &data,
			CanPublishSources: b.canPublishSources, Hidden: b.hidden,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(b.cert))
}

// RoomName is the server-side name of an app's room ("<AppID>_<room>").
func RoomName(appID, room string) string {
	if strings.HasPrefix(room, appID+Separator) {
		return room
	}
	return appID + Separator + room
}

// DisplayName strips the app prefix for showing a room name to users.
func DisplayName(appID, room string) string {
	return strings.TrimPrefix(room, appID+Separator)
}
