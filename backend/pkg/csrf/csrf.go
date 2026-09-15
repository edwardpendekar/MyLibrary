// Package csrf implements the double-submit cookie pattern: a random token is
// set in a JS-readable cookie at login, and the frontend must echo it back in
// a request header on every state-changing call. A cross-site attacker can
// trick a browser into sending cookies automatically but cannot read the
// cookie's value to also set the header, so the two values won't match.
//
// This is defense-in-depth on top of SameSite=Strict auth cookies (which
// already block the cookie from being sent cross-site at all in modern
// browsers); together they cover both the modern and legacy-browser cases.
package csrf

import (
	"crypto/rand"
	"encoding/base64"
)

const (
	CookieName = "csrf_token"
	HeaderName = "X-CSRF-Token"
)

func GenerateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
