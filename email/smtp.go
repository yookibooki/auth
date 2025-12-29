package email

import (
	"fmt"
	"net/smtp"
)

type Sender interface {
	Send(to, subject, body string) error
}

type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPSender(host string, port int, username, password, from string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.from, to, subject, body)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
