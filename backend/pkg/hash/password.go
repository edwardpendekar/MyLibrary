// Package hash wraps password hashing so the rest of the codebase never imports
// bcrypt directly (makes it a one-place change if the algorithm is ever upgraded).
package hash

import "golang.org/x/crypto/bcrypt"

const cost = 12

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func VerifyPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
