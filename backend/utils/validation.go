package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return emailRegex.MatchString(email)
}

func ValidatePseudo(pseudo string) bool {
	length := len(strings.TrimSpace(pseudo))
	return length >= 3 && length <= 20
}

func ValidatePassword(password string) bool {
	if len(password) >= 12 && hasLower(password) && hasUpper(password) && hasDigit(password) && hasSpecial(password) {
		return true
	}
	if len(password) >= 14 && hasLower(password) && hasUpper(password) && hasDigit(password) {
		return true
	}
	return len(strings.Fields(password)) >= 7
}

func PasswordPolicyMessage() string {
	return "Le mot de passe doit contenir au moins 12 caracteres avec majuscule, minuscule, chiffre et caractere special, ou 14 caracteres avec majuscule, minuscule et chiffre, ou une phrase de passe d'au moins 7 mots."
}

func hasLower(value string) bool {
	for _, r := range value {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasUpper(value string) bool {
	for _, r := range value {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasDigit(value string) bool {
	for _, r := range value {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSpecial(value string) bool {
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
