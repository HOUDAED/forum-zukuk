package utils

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

func SendOTPEmail(email, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Votre code de vérification")
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>Code de vérification</h2>
		<p>Votre code est: <strong>%s</strong></p>
		<p>Ce code expire dans 5 minutes.</p>
	`, code))

	d := gomail.NewDialer(
		os.Getenv("EMAIL_HOST"),
		getEmailPort(),
		os.Getenv("EMAIL_USER"),
		os.Getenv("EMAIL_PASSWORD"),
	)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("erreur lors de l'envoi de l'email OTP: %w", err)
	}

	fmt.Printf("[MAIL] Code OTP envoyé à %s\n", email)
	return nil
}

func SendWelcomeEmail(pseudo, email, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Bienvenue sur Forum Zukuk")
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>Bienvenue %s ! 👋</h2>
		<p>Votre code de vérification est: <strong>%s</strong></p>
		<p>Ce code expire dans 10 minutes.</p>
		<p>Merci d'avoir rejoint notre communauté !</p>
	`, pseudo, code))

	d := gomail.NewDialer(
		os.Getenv("EMAIL_HOST"),
		getEmailPort(),
		os.Getenv("EMAIL_USER"),
		os.Getenv("EMAIL_PASSWORD"),
	)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("erreur lors de l'envoi de l'email de bienvenue: %w", err)
	}

	fmt.Printf("[MAIL] Email de bienvenue envoyé à %s\n", email)
	return nil
}

func getEmailPort() int {
	return 587
}
