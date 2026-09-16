package mailer

import (
	"context"
	"fmt"
	"net/smtp"
)

type smtpMailer struct {
	host, port, username, password, from string
}

func NewSMTPMailer(host string, port int, username, password, from string) Mailer {
	return &smtpMailer{host: host, port: fmt.Sprintf("%d", port), username: username, password: password, from: from}
}

func (m *smtpMailer) SendPasswordReset(_ context.Context, toEmail, toName, resetURL string) error {
	subject := "Reset your Book Reader password"
	body := fmt.Sprintf(
		"Hi %s,\r\n\r\nWe received a request to reset your password. Click the link below to choose a new one. This link expires in 1 hour and can only be used once.\r\n\r\n%s\r\n\r\nIf you didn't request this, you can safely ignore this email.\r\n",
		toName, resetURL,
	)
	msg := fmt.Appendf(nil, "From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.from, toEmail, subject, body)

	addr := m.host + ":" + m.port
	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	return smtp.SendMail(addr, auth, m.from, []string{toEmail}, msg)
}
