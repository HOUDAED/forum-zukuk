package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendPasswordResetEmail(toEmail, token string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	appURL := os.Getenv("FRONTEND_ORIGIN")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", appURL, token)

	subject := "Subject: Réinitialisation de votre mot de passe Zukuk\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h2>Bonjour,</h2>
		<p>Vous avez demandé à réinitialiser votre mot de passe.</p>
		<p>Cliquez sur le lien ci-dessous (valable 15 minutes) :</p>
		<a href="%s">Réinitialiser mon mot de passe</a>
		<p>Si vous n'avez rien demandé, ignorez cet email.</p>
	`, resetLink)

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := smtpHost + ":" + smtpPort

	return smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, msg)
}
