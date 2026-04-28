package utils

import (
	"fmt"
	"os"
	"time"

	"forum-zukuk/models"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken génère un JWT token
func GenerateToken(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your_secret_key"
	}

	claims := jwt.MapClaims{
		"id":     user.ID,
		"email":  user.Email,
		"pseudo": user.Pseudo,
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération du token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken vérifie et parse un JWT token
func VerifyToken(tokenString string) (*models.TokenClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your_secret_key"
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token invalide: %w", err)
	}

	claims := token.Claims.(*jwt.MapClaims)
	return &models.TokenClaims{
		ID:     int64((*claims)["id"].(float64)),
		Email:  (*claims)["email"].(string),
		Pseudo: (*claims)["pseudo"].(string),
	}, nil
}
