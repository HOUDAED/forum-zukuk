package models

import "time"

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

type OTPRecord struct {
	Code    string
	Expires time.Time
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Pseudo          string `json:"pseudo" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type TokenClaims struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Pseudo string `json:"pseudo"`
}

type AuthResponse struct {
	Message string        `json:"message"`
	Token   string        `json:"token,omitempty"`
	Email   string        `json:"email,omitempty"`
	User    *UserResponse `json:"user,omitempty"`
}

type UserResponse struct {
	ID     int64  `json:"id"`
	Pseudo string `json:"pseudo"`
	Email  string `json:"email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
