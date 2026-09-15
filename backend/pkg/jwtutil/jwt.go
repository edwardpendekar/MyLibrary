// Package jwtutil issues and verifies short-lived JWT access tokens.
// Refresh tokens are deliberately NOT JWTs (see pkg/hash.HashToken) — they are
// opaque random strings whose hash is checked against the refresh_tokens table,
// which lets us revoke them server-side; a JWT cannot be revoked before it expires.
package jwtutil

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Issuer struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewIssuer(secret string, ttl time.Duration, issuer string) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl, issuer: issuer}
}

func (i *Issuer) GenerateAccessToken(userID int64, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(i.ttl)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (i *Issuer) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return i.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GenerateOpaqueToken returns a URL-safe random string used as a refresh token value.
// Only its SHA-256 hash (pkg/hash.HashToken) is ever persisted.
func GenerateOpaqueToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
