package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"forum-zukuk/api/utils"
	"forum-zukuk/database"
)

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "L'email est requis."})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !utils.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email invalide."})
		return
	}

	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne."})
		return
	}

	err = database.CreateResetToken(email, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Si cet email existe, un lien de réinitialisation a été envoyé."})
		return
	}

	go utils.SendPasswordResetEmail(email, token)

	c.JSON(http.StatusOK, gin.H{"message": "Si cet email existe, un lien de réinitialisation a été envoyé."})
}

func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tous les champs sont requis."})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Les mots de passe ne correspondent pas."})
		return
	}
	if !utils.ValidatePassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.PasswordPolicyMessage()})
		return
	}

	userID, err := database.ValidateResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Lien invalide ou expiré."})
		return
	}

	inHistory, err := database.IsPasswordInHistory(userID, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification."})
		return
	}
	if inHistory {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ce mot de passe a déjà été utilisé récemment."})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne."})
		return
	}

	if err := database.UpdatePassword(userID, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mot de passe mis à jour avec succès. Vous pouvez vous connecter."})
}
