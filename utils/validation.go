package utils

import "regexp"

// ValidateEmail valide le format d'une adresse email
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword valide les critères du mot de passe
func ValidatePassword(password string) bool {
	return len(password) >= 8
}

// ValidatePseudo valide le pseudo
func ValidatePseudo(pseudo string) bool {
	length := len(pseudo)
	return length >= 3 && length <= 20
}

// ValidateOTPCode valide le format du code OTP
func ValidateOTPCode(code string) bool {
	otpRegex := regexp.MustCompile(`^\d{6}$`)
	return otpRegex.MatchString(code)
}
