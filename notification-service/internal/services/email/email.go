package email

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/ocenb/music-go/notification-service/internal/config"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		cfg: cfg,
	}
}

func (s *EmailService) SendVerificationEmail(email, verificationLink string) error {
	const maxRetries = 3
	var lastErr error

	for i := range maxRetries {
		if i > 0 {
			time.Sleep(time.Second * time.Duration(i*2))
		}

		err := s.sendEmail(email, verificationLink)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to send email after %d retries: %w", maxRetries, lastErr)
}

func (s *EmailService) sendEmail(to, verificationLink string) error {
	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	subject := "Verify your email"
	htmlBody := fmt.Sprintf(`
		<html>
			<body>
				<p>Please verify your email by clicking the link below:</p>
				<p><a href="%s">Verify Email</a></p>
				<p>If you didn't request this, please ignore this email.</p>
			</body>
		</html>
	`, verificationLink)

	message := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, htmlBody)

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	return smtp.SendMail(addr, auth, s.cfg.SMTPUsername, []string{to}, []byte(message))
}
