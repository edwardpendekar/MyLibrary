// Package mailer sends transactional emails (currently just password resets).
// When SMTP isn't configured (local dev), it falls back to logging the link
// instead of failing the request — see NewFromConfig.
package mailer

import "context"

type Mailer interface {
	SendPasswordReset(ctx context.Context, toEmail, toName, resetURL string) error
}
