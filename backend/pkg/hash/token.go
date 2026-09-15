package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken is used for opaque bearer tokens (refresh tokens) where we only ever
// need to compare equality, never recover the original value, so a fast SHA-256
// digest is sufficient and avoids bcrypt's cost on every refresh call.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
