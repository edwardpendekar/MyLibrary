package mailer

import (
	"context"

	"github.com/rs/zerolog"
)

// consoleMailer logs the reset link instead of sending an email. It exists so
// the forgot-password flow works out of the box in local/dev environments
// where no SMTP relay has been configured.
type consoleMailer struct {
	log zerolog.Logger
}

func NewConsoleMailer(log zerolog.Logger) Mailer {
	return &consoleMailer{log: log}
}

func (m *consoleMailer) SendPasswordReset(_ context.Context, toEmail, _ string, resetURL string) error {
	m.log.Warn().
		Str("to", toEmail).
		Str("reset_url", resetURL).
		Msg("SMTP not configured — logging password reset link instead of emailing it")
	return nil
}
