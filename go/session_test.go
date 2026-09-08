package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// gatewaySessionJSON is a real payload taken from Taronja Gateway's own
// json.Marshal(db.Session) (verified against a real db.Session value in
// that repository, not hand-guessed) — see the gateway's own
// AGENTS.md/db/schema.go for the source of truth this mirrors. Kept as one
// literal here rather than built field by field so the test is exercising
// exactly what actually comes over the wire, key names included.
const gatewaySessionJSON = `{"ID":42,"CreatedAt":"2019-01-01T00:00:00Z","UpdatedAt":"2019-01-02T00:00:00Z","DeletedAt":"2020-01-02T03:04:05Z","token":"tok-123","userId":"user-abc","username":"jdoe","email":"jdoe@example.com","isAuthenticated":true,"isAdmin":true,"validUntil":"2030-01-01T00:00:00Z","provider":"google","closedOn":"2021-06-07T08:09:10Z","lastActivity":"2022-03-04T05:06:07Z","sessionName":"my-session","createdFrom":"login","ipAddress":"1.2.3.4","userAgent":"curl/8.0","referrer":"https://example.com","browserFamily":"Chrome","browserVersion":"120","osFamily":"Linux","osVersion":"6.1","deviceFamily":"Other","deviceBrand":"","deviceModel":"","geoLocation":"Somewhere","latitude":1.5,"longitude":-2.5,"city":"Testville","zipCode":"12345","country":"Testland","countryCode":"TL","region":"Test Region","continent":"Testinent","fingerprint":"abcdef1234","fingerprintType":"ja4_tls"}`

func TestParseSessionHeader(t *testing.T) {
	t.Run("decodes a real gateway-shaped payload field for field", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderUserData, gatewaySessionJSON)

		session, err := ParseSessionHeader(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if session.ID != 42 {
			t.Errorf("ID = %v, want 42 (gorm.Model's untagged field must decode under its literal Go name)", session.ID)
		}
		if !session.CreatedAt.Equal(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)) {
			t.Errorf("CreatedAt = %v", session.CreatedAt)
		}
		if session.DeletedAt == nil || !session.DeletedAt.Equal(time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)) {
			t.Errorf("DeletedAt = %v, want a non-nil 2020-01-02T03:04:05Z", session.DeletedAt)
		}
		if session.Token != "tok-123" {
			t.Errorf("Token = %q", session.Token)
		}
		if session.UserID != "user-abc" {
			t.Errorf("UserID = %q", session.UserID)
		}
		if !session.IsAdmin || !session.IsAuthenticated {
			t.Errorf("IsAdmin/IsAuthenticated did not decode as true")
		}
		if session.ClosedOn == nil || !session.ClosedOn.Equal(time.Date(2021, 6, 7, 8, 9, 10, 0, time.UTC)) {
			t.Errorf("ClosedOn = %v", session.ClosedOn)
		}
		if session.IPAddress != "1.2.3.4" || session.BrowserFamily != "Chrome" || session.Country != "Testland" {
			t.Errorf("embedded ClientInfo fields did not decode correctly: %+v", session)
		}
		// These two specifically replaced an older, differently-named
		// field ("JA4Fingerprint") a hand-copied Session struct might
		// still expect — the whole reason this shared package exists is
		// to keep exactly this kind of rename from silently drifting.
		if session.Fingerprint != "abcdef1234" {
			t.Errorf("Fingerprint = %q", session.Fingerprint)
		}
		if session.FingerprintType != "ja4_tls" {
			t.Errorf("FingerprintType = %q", session.FingerprintType)
		}
	})

	t.Run("a nullable DeletedAt/ClosedOn decodes as nil, not a zero time", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderUserData, `{"token":"t","DeletedAt":null,"closedOn":null}`)

		session, err := ParseSessionHeader(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if session.DeletedAt != nil {
			t.Errorf("DeletedAt = %v, want nil", session.DeletedAt)
		}
		if session.ClosedOn != nil {
			t.Errorf("ClosedOn = %v, want nil", session.ClosedOn)
		}
	})

	t.Run("returns an error when the header is missing", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := ParseSessionHeader(r)
		if err == nil {
			t.Fatal("expected an error for a missing header, got nil")
		}
	})

	t.Run("returns an error for malformed JSON", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderUserData, `{not valid json`)
		_, err := ParseSessionHeader(r)
		if err == nil {
			t.Fatal("expected an error for malformed JSON, got nil")
		}
	})
}
