// This file is hand-written, not generated — unlike apiclient.gen.go in
// this same package, which oapi-codegen regenerates from Taronja
// Gateway's OpenAPI spec on every release. Session never appears as an
// HTTP request/response body of the gateway's own management API (it's
// never part of that spec), so it wouldn't be generated here even if the
// gateway repository's own generation step ran again — it only ever shows
// up inside the X-User-Data header value the gateway attaches to requests
// it proxies to *other* services. See client.yml's rsync step in the
// gateway repository: this file (and this one specifically, by name) is
// excluded from the generated-output sync precisely so a release doesn't
// delete it.
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Header and cookie names Taronja Gateway uses to forward session data to
// backend routes it proxies to. See the gateway's own README.md
// ("Headers Sent to Backend Routes" / "X-User-Data JSON Structure") for
// the full contract these values are part of.
const (
	// HeaderUserId carries just the authenticated user's ID — cheaper to
	// read than decoding HeaderUserData's full JSON when that's all a
	// route needs.
	HeaderUserId = "X-User-Id"
	// HeaderUserData carries the full Session, JSON-encoded — see
	// ParseSessionHeader.
	HeaderUserData = "X-User-Data"
	// CookieSession is the gateway's own session cookie name, forwarded
	// unchanged on proxied requests when the original client
	// authenticated via a session cookie rather than a bearer token.
	CookieSession = "tg_session_token"
)

// Session mirrors db.Session from Taronja Gateway's own repository, field
// for field and JSON tag for JSON tag — this is exactly the shape decoded
// from a request's HeaderUserData header. It's a hand-maintained copy, not
// a generated one: Session is never a request/response body anywhere in
// the gateway's OpenAPI spec (it only ever appears inside a header
// value), so apiclient.gen.go's generator has no way to produce it. This
// is the single copy every consuming service should import instead of
// hand-copying its own — see this file's own top-of-file comment for how
// it survives being regenerated out of existence.
//
// Keeping this in sync with db.Session when the gateway's own schema
// changes is still a manual step (there's no compiler or CI check tying
// the two together across repositories) — but it's now exactly one place
// to update, not one per consumer.
type Session struct {
	// Embedded via gorm.Model on the gateway side, which has no JSON tags
	// of its own, so these three (plus DeletedAt below) serialize under
	// their literal Go field names rather than a camelCase tag.
	ID        uint      `json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	// DeletedAt is gorm.DeletedAt on the gateway side, a nullable
	// timestamp with its own MarshalJSON: it encodes as a plain JSON
	// timestamp string when set, or JSON null when not — a plain
	// *time.Time here decodes either shape correctly.
	DeletedAt *time.Time `json:"DeletedAt"`

	Token           string     `json:"token"`
	UserID          string     `json:"userId"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	IsAuthenticated bool       `json:"isAuthenticated"`
	IsAdmin         bool       `json:"isAdmin"`
	ValidUntil      time.Time  `json:"validUntil"`
	Provider        string     `json:"provider"`
	ClosedOn        *time.Time `json:"closedOn"`
	LastActivity    time.Time  `json:"lastActivity"`
	SessionName     string     `json:"sessionName"`
	CreatedFrom     string     `json:"createdFrom"`

	// Embedded via db.ClientInfo on the gateway side.
	IPAddress      string  `json:"ipAddress"`
	UserAgent      string  `json:"userAgent"`
	Referrer       string  `json:"referrer"`
	BrowserFamily  string  `json:"browserFamily"`
	BrowserVersion string  `json:"browserVersion"`
	OSFamily       string  `json:"osFamily"`
	OSVersion      string  `json:"osVersion"`
	DeviceFamily   string  `json:"deviceFamily"`
	DeviceBrand    string  `json:"deviceBrand"`
	DeviceModel    string  `json:"deviceModel"`
	GeoLocation    string  `json:"geoLocation"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	City           string  `json:"city"`
	ZipCode        string  `json:"zipCode"`
	Country        string  `json:"country"`
	CountryCode    string  `json:"countryCode"`
	Region         string  `json:"region"`
	Continent      string  `json:"continent"`
	// Fingerprint is the client's single fingerprint value for this
	// session — TLS-level JA4, a reduced-entropy "stable" fingerprint, or
	// JA4H, in that priority order (see FingerprintType for which one).
	// Empty if none was available. Note this replaces what used to be a
	// JA4-only field on the gateway side; a consumer still keying off a
	// field literally named "JA4Fingerprint" is reading a name that no
	// longer exists in the gateway's own schema.
	Fingerprint string `json:"fingerprint"`
	// FingerprintType names which algorithm produced Fingerprint:
	// "ja4_tls", "stable", or "ja4h". Empty exactly when Fingerprint is.
	FingerprintType string `json:"fingerprintType"`
}

// ParseSessionHeader decodes the Session Taronja Gateway attached to r via
// HeaderUserData. Returns an error if the header is missing/empty or
// isn't valid JSON — a request that never went through the gateway, or
// one on a route where authentication isn't required, simply has no such
// header, so a caller should treat that as "no session available" rather
// than necessarily a hard failure, depending on whether the route itself
// requires one.
func ParseSessionHeader(r *http.Request) (Session, error) {
	raw := r.Header.Get(HeaderUserData)
	if raw == "" {
		return Session{}, fmt.Errorf("%s header not present", HeaderUserData)
	}
	var session Session
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return Session{}, fmt.Errorf("invalid %s header: %w", HeaderUserData, err)
	}
	return session, nil
}
