package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"forum-zukuk/models"
	"forum-zukuk/utils"
)

var (
	users   []*models.User
	tempOTP = make(map[string]*models.OTPRecord)
	mu      sync.RWMutex
)

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Email et mot de passe requis."})
		return
	}

	if !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Email invalide."})
		return
	}

	mu.RLock()
	var user *models.User
	for _, u := range users {
		if u.Email == req.Email {
			user = u
			break
		}
	}
	mu.RUnlock()

	if user == nil || !validatePasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Identifiants incorrects."})
		return
	}

	otp := generateOTP()
	mu.Lock()
	tempOTP[req.Email] = &models.OTPRecord{
		Code:    otp,
		Expires: time.Now().Add(5 * time.Minute),
	}
	mu.Unlock()

	if err := utils.SendOTPEmail(req.Email, otp); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Erreur lors de l'envoi de l'email."})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Message: "OTP_SENT",
		Email:   req.Email,
	})
}

func VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Email et code requis."})
		return
	}

	mu.Lock()
	record, exists := tempOTP[req.Email]
	if exists {
		delete(tempOTP, req.Email)
	}
	mu.Unlock()

	if !exists || record.Code != req.Code || time.Now().After(record.Expires) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Code invalide ou expiré."})
		return
	}

	mu.RLock()
	var user *models.User
	for _, u := range users {
		if u.Email == req.Email {
			user = u
			break
		}
	}
	mu.RUnlock()

	if user == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Utilisateur non trouvé."})
		return
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Erreur lors de la génération du token."})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Message: "SUCCESS",
		Token:   token,
		User: &models.UserResponse{
			ID:     user.ID,
			Pseudo: user.Pseudo,
			Email:  user.Email,
		},
	})
}

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Tous les champs sont requis."})
		return
	}

	if !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Email invalide."})
		return
	}

	if !utils.ValidatePassword(req.Password) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Le mot de passe doit contenir au moins 8 caractères."})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Les mots de passe ne correspondent pas."})
		return
	}

	if !utils.ValidatePseudo(req.Pseudo) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Le pseudo doit contenir entre 3 et 20 caractères."})
		return
	}

	mu.RLock()
	for _, u := range users {
		if u.Email == req.Email {
			mu.RUnlock()
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Cet email est déjà utilisé."})
			return
		}
	}

	for _, u := range users {
		if u.Pseudo == req.Pseudo {
			mu.RUnlock()
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Ce pseudo est déjà pris."})
			return
		}
	}
	mu.RUnlock()

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Erreur lors du traitement du mot de passe."})
		return
	}

	newUser := &models.User{
		ID:         time.Now().UnixMilli(),
		Pseudo:     req.Pseudo,
		Email:      req.Email,
		Password:   hashedPassword,
		IsVerified: false,
		CreatedAt:  time.Now(),
	}

	mu.Lock()
	users = append(users, newUser)
	mu.Unlock()

	otp := generateOTP()
	mu.Lock()
	tempOTP[req.Email] = &models.OTPRecord{
		Code:    otp,
		Expires: time.Now().Add(10 * time.Minute),
	}
	mu.Unlock()

	if err := utils.SendWelcomeEmail(req.Pseudo, req.Email, otp); err != nil {
		fmt.Printf("Erreur lors de l'envoi de l'email de bienvenue: %v\n", err)
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Message: "Utilisateur créé, OTP envoyé.",
	})
}

// GoogleOAuth endpoint
func GoogleOAuth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":   "OAuth Google non configuré",
		"message": "Pour intégrer Google OAuth, voir la documentation",
		"docs":    "https://developers.google.com/identity/protocols/oauth2",
	})
}

// GoogleCallback traite la réponse Google OAuth
func GoogleCallback(c *gin.Context) {
	var req struct {
		GoogleID string `json:"googleId"`
		Email    string `json:"email"`
		Pseudo   string `json:"pseudo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Données invalides."})
		return
	}

	mu.Lock()
	var user *models.User
	for _, u := range users {
		if u.GoogleID == req.GoogleID {
			user = u
			break
		}
	}

	if user == nil {
		user = &models.User{
			ID:         time.Now().UnixMilli(),
			GoogleID:   req.GoogleID,
			Email:      req.Email,
			Pseudo:     req.Pseudo,
			IsVerified: true,
			Provider:   "google",
			CreatedAt:  time.Now(),
		}
		users = append(users, user)
	}
	mu.Unlock()

	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Erreur de génération de token."})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Message: "SUCCESS",
		Token:   token,
		User: &models.UserResponse{
			ID:     user.ID,
			Pseudo: user.Pseudo,
			Email:  user.Email,
		},
	})
}

// SchoolOAuth endpoint
func SchoolOAuth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":   "Portail École non configuré",
		"message": "À intégrer avec votre système de portail scolaire",
	})
}

// SchoolCallback traite la réponse du portail scolaire
func SchoolCallback(c *gin.Context) {
	var req struct {
		SchoolID   string `json:"schoolId"`
		Email      string `json:"email"`
		Pseudo     string `json:"pseudo"`
		SchoolName string `json:"schoolName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Données invalides."})
		return
	}

	mu.Lock()
	var user *models.User
	for _, u := range users {
		if u.SchoolID == req.SchoolID {
			user = u
			break
		}
	}

	if user == nil {
		user = &models.User{
			ID:         time.Now().UnixMilli(),
			SchoolID:   req.SchoolID,
			Email:      req.Email,
			Pseudo:     req.Pseudo,
			IsVerified: true,
			Provider:   "school",
			SchoolName: req.SchoolName,
			CreatedAt:  time.Now(),
		}
		users = append(users, user)
	}
	mu.Unlock()

	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Erreur de génération de token."})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Message: "SUCCESS",
		Token:   token,
		User: &models.UserResponse{
			ID:     user.ID,
			Pseudo: user.Pseudo,
			Email:  user.Email,
		},
	})
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hash), err
}

func validatePasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
