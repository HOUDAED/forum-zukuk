package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"forum-zukuk/database"
	"forum-zukuk/api/utils"
)

type RegisterRequest struct {
	Pseudo          string `json:"pseudo" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Password   string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Pseudo          string `json:"pseudo"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	CurrentPassword string `json:"currentPassword"`
	AvatarURL       string `json:"avatar_url"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tous les champs sont requis."})
		return
	}

	req.Pseudo = strings.TrimSpace(req.Pseudo)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if !utils.ValidatePseudo(req.Pseudo) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pseudo invalide (3-20 caractères)."})
		return
	}
	if !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email invalide."})
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

	var count int
	database.ZUKUKDB.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ? AND deleted_at IS NULL`, req.Email).Scan(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cet email est déjà utilisé."})
		return
	}
	database.ZUKUKDB.QueryRow(`SELECT COUNT(*) FROM users WHERE pseudo = ? AND deleted_at IS NULL`, req.Pseudo).Scan(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Ce pseudo est déjà pris."})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne."})
		return
	}
	if _, err := database.ZUKUKDB.Exec(`
		INSERT INTO users (pseudo, email, password_hash, email_verified)
		VALUES (?, ?, ?, 1)`, req.Pseudo, req.Email, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création compte."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Compte créé.", "email": req.Email})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identifiant et mot de passe requis."})
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identifiant requis."})
		return
	}

	var userID int
	var hash string
	err := database.ZUKUKDB.QueryRow(`
		SELECT id, password_hash FROM users
		WHERE (email = ? OR pseudo = ?) AND deleted_at IS NULL`,
		strings.ToLower(identifier), identifier,
	).Scan(&userID, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants incorrects."})
		return
	}

	token, err := createSession(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création session."})
		return
	}
	setSessionCookie(c, token)
	c.JSON(http.StatusOK, gin.H{"message": "Connecté."})
}

func Logout(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err == nil {
		database.ZUKUKDB.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	}
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Déconnecté."})
}

func GetMe(c *gin.Context) {
	userID := c.GetInt("userID")
	var pseudo, email, avatarURL string
	var createdAt time.Time
	err := database.ZUKUKDB.QueryRow(`
		SELECT pseudo, email, avatar_url, created_at
		FROM users WHERE id = ? AND deleted_at IS NULL`, userID,
	).Scan(&pseudo, &email, &avatarURL, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur introuvable."})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": userID, "pseudo": pseudo, "email": email,
		"avatar_url": avatarURL, "created_at": createdAt,
	})
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides."})
		return
	}

	if req.Pseudo != "" {
		req.Pseudo = strings.TrimSpace(req.Pseudo)
		if !utils.ValidatePseudo(req.Pseudo) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pseudo invalide (3-20 caractères)."})
			return
		}
		var count int
		database.ZUKUKDB.QueryRow(`SELECT COUNT(*) FROM users WHERE pseudo = ? AND id != ? AND deleted_at IS NULL`, req.Pseudo, userID).Scan(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Ce pseudo est déjà pris."})
			return
		}
		if _, err := database.ZUKUKDB.Exec(`UPDATE users SET pseudo = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, req.Pseudo, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour pseudo."})
			return
		}
	}

	if req.Email != "" {
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if !utils.ValidateEmail(req.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email invalide."})
			return
		}
		var count int
		database.ZUKUKDB.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ? AND id != ? AND deleted_at IS NULL`, req.Email, userID).Scan(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Cet email est déjà utilisé."})
			return
		}
		if _, err := database.ZUKUKDB.Exec(`UPDATE users SET email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, req.Email, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour email."})
			return
		}
	}

	if req.AvatarURL != "" {
		if _, err := database.ZUKUKDB.Exec(`UPDATE users SET avatar_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, strings.TrimSpace(req.AvatarURL), userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour avatar."})
			return
		}
	}

	if req.Password != "" {
		if req.Password != req.ConfirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Les mots de passe ne correspondent pas."})
			return
		}
		if !utils.ValidatePassword(req.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"error": utils.PasswordPolicyMessage()})
			return
		}
		var currentHash string
		if err := database.ZUKUKDB.QueryRow(`SELECT password_hash FROM users WHERE id = ? AND deleted_at IS NULL`, userID).Scan(&currentHash); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur introuvable."})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mot de passe actuel incorrect."})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne."})
			return
		}
		if _, err := database.ZUKUKDB.Exec(`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, string(hash), userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour mot de passe."})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profil mis à jour."})
}

func DeleteAccount(c *gin.Context) {
	userID := c.GetInt("userID")
	if err := database.DeleteUserAndData(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur suppression compte."})
		return
	}
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Compte supprimé."})
}

func Health(c *gin.Context) {
	if err := database.ZUKUKDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "db": "unreachable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "connected"})
}

func createSession(userID int) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := fmt.Sprintf("%x", b)
	expires := time.Now().UTC().Add(7 * 24 * time.Hour)
	_, err := database.ZUKUKDB.Exec(`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`, token, userID, expires)
	return token, err
}

func setSessionCookie(c *gin.Context, token string) {
	secure := os.Getenv("GIN_MODE") == "release"
	c.SetCookie("session_token", token, 7*24*3600, "/", "", secure, true)
}