package utils

import "regexp"

func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return emailRegex.MatchString(email)
}

func ValidatePassword(password string) bool {
	return len(password) >= 8
}

func ValidatePseudo(pseudo string) bool {
	length := len(pseudo)
	return length >= 3 && length <= 20
}

func ValidateOTPCode(code string) bool {
	otpRegex := regexp.MustCompile(`^\d{6}$`)
	return otpRegex.MatchString(code)
}
