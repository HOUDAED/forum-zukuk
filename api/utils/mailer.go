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
	fromEmail := os.Getenv("SMTP_FROM")

	if fromEmail == "" {
		fromEmail = smtpUser
	}

	appURL := os.Getenv("FRONTEND_ORIGIN")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", appURL, token)

	headers := fmt.Sprintf(
		"From: Zukuk <%s>\r\nTo: %s\r\nSubject: =?UTF-8?Q?R=C3=A9initialisation_de_votre_mot_de_passe_Zukuk?=\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		fromEmail, toEmail,
	)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#f8fafc;font-family:'Segoe UI',sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f8fafc;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background:white;border-radius:16px;overflow:hidden;box-shadow:0 4px 20px rgba(0,0,0,0.06);">
          <tr>
            <td style="background:linear-gradient(135deg,#818cf8,#a78bfa);padding:32px;text-align:center;">
              <h1 style="color:white;margin:0;font-size:28px;font-weight:700;letter-spacing:0.05em;">Zukuk 💙</h1>
              <p style="color:rgba(255,255,255,0.85);margin:8px 0 0;font-size:14px;">Tu n'es pas seul.</p>
            </td>
          </tr>
          <tr>
            <td style="padding:40px 48px;">
              <h2 style="color:#1e293b;font-size:20px;margin:0 0 16px;">Réinitialisation de ton mot de passe</h2>
              <p style="color:#64748b;line-height:1.7;margin:0 0 24px;">
                Bonjour,<br><br>
                Tu as demandé à réinitialiser ton mot de passe sur Zukuk.<br>
                Clique sur le bouton ci-dessous — ce lien est valable <strong>15 minutes</strong>.
              </p>
              <div style="text-align:center;margin:32px 0;">
                <a href="%s"
                   style="background:linear-gradient(135deg,#818cf8,#a78bfa);color:white;padding:14px 32px;
                          text-decoration:none;border-radius:12px;font-weight:600;font-size:16px;
                          display:inline-block;">
                  Réinitialiser mon mot de passe
                </a>
              </div>
              <p style="color:#94a3b8;font-size:13px;line-height:1.6;margin:24px 0 0;border-top:1px solid #f1f5f9;padding-top:24px;">
                Si tu n'as pas fait cette demande, ignore cet e-mail. Ton compte reste sécurisé.<br>
                Ce lien expirera automatiquement dans 15 minutes.
              </p>
            </td>
          </tr>
          <tr>
            <td style="background:#f8fafc;padding:20px 48px;text-align:center;border-top:1px solid #f1f5f9;">
              <p style="color:#94a3b8;font-size:12px;margin:0;">© 2026 Zukuk — Prends soin de toi 💙</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, resetLink)

	msg := []byte(headers + body)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := smtpHost + ":" + smtpPort

	return smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, msg)
}
