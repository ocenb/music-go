package smtp

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/ocenb/music-go/notification-service/internal/config"
)

type Client struct {
	cfg  config.SMTPConfig
	auth smtp.Auth
}

func New(cfg config.SMTPConfig) *Client {
	return &Client{
		cfg:  cfg,
		auth: smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host),
	}
}

func (c *Client) Send(_ context.Context, to, htmlBody string) error {
	subject := "Notification"
	message := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, htmlBody)

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	if err := smtp.SendMail(addr, c.auth, c.cfg.Username, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("smtp send mail: %w", err)
	}

	return nil
}
