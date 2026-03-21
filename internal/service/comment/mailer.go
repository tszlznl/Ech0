package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wneessen/go-mail"
)

type GoMailSender struct{}

func NewGoMailSender() *GoMailSender {
	return &GoMailSender{}
}

func (s *GoMailSender) Send(ctx context.Context, cfg MailerConfig, msg MailMessage) error {
	host := strings.TrimSpace(cfg.Host)
	to := strings.TrimSpace(msg.To)
	from := strings.TrimSpace(cfg.Username)
	if host == "" || to == "" || from == "" {
		return fmt.Errorf("missing mail configuration")
	}

	opts := []mail.Option{
		mail.WithPort(defaultPort(cfg.Port)),
		mail.WithTimeout(10 * time.Second),
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
	}

	if strings.TrimSpace(cfg.Username) != "" {
		opts = append(opts, mail.WithUsername(strings.TrimSpace(cfg.Username)))
	}
	if cfg.Password != "" {
		opts = append(opts, mail.WithPassword(cfg.Password))
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		return err
	}
	if err := client.DialWithContext(ctx); err != nil {
		return err
	}
	defer client.Close()

	m := mail.NewMsg()
	if err := m.From(from); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(strings.TrimSpace(msg.Subject))
	m.SetBodyString(mail.TypeTextPlain, msg.TextBody)
	return client.Send(m)
}

func defaultPort(port int) int {
	if port > 0 {
		return port
	}
	return 587
}
