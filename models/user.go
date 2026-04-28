package models

import "time"

// User représente un utilisateur du système
type User struct {
	ID         int64     `json:"id"`
	Pseudo     string    `json:"pseudo"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	GoogleID   string    `json:"-"`
	SchoolID   string    `json:"-"`
	Provider   string    `json:"provider"`
	SchoolName string    `json:"school_name,omitempty"`
}

// OTPRecord représente un enregistrement OTP temporaire
type OTPRecord struct {
	Code    string
	Expires time.Time
}

// LoginRequest représente une demande de connexion
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest représente une demande d'inscription
type RegisterRequest struct {
	Pseudo          string `json:"pseudo" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// VerifyOTPRequest représente une vérification OTP
type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// TokenClaims représente les claims du JWT
type TokenClaims struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Pseudo string `json:"pseudo"`
}

// AuthResponse représente la réponse d'authentification
type AuthResponse struct {
	Message string        `json:"message"`
	Token   string        `json:"token,omitempty"`
	Email   string        `json:"email,omitempty"`
	User    *UserResponse `json:"user,omitempty"`
}

// UserResponse représente les données utilisateur retournées
type UserResponse struct {
	ID     int64  `json:"id"`
	Pseudo string `json:"pseudo"`
	Email  string `json:"email"`
}

// ErrorResponse représente une réponse d'erreur
type ErrorResponse struct {
	Error string `json:"error"`
}
