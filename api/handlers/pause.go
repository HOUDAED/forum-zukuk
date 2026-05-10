package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)

type PauseRequest struct {
	DurationDays int `json:"duration_days" binding:"required"`
}

// PauseAccount masque le compte pendant N jours.
// Le profil disparaît du réseau, l'utilisateur est déconnecté.
// Il peut se reconnecter à tout moment pour reprendre.
func PauseAccount(c *gin.Context) {
	userID := c.GetInt("userID")

	var req PauseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Durée requise."})
		return
	}

	// Valider la durée (7, 30, 90 ou 180 jours)
	allowed := map[int]bool{7: true, 30: true, 90: true, 180: true}
	if !allowed[req.DurationDays] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Durée invalide (7, 30, 90 ou 180 jours)."})
		return
	}

	pauseUntil := time.Now().UTC().AddDate(0, 0, req.DurationDays)

	// Mettre à jour paused_until en base
	_, err := database.ZUKUKDB.Exec(`
		UPDATE users SET paused_until = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL`,
		pauseUntil, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur activation de la pause."})
		return
	}

	// Invalider toutes les sessions → l'utilisateur est déconnecté
	database.ZUKUKDB.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	c.SetCookie("session_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Pause activée.",
		"paused_until": pauseUntil,
	})
}

// ResumAccount lève la pause manuellement (appelé au login si paused_until est dépassé,
// ou explicitement par l'utilisateur depuis son profil).
func ResumeAccount(c *gin.Context) {
	userID := c.GetInt("userID")

	database.ZUKUKDB.Exec(`
		UPDATE users SET paused_until = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, userID)

	c.JSON(http.StatusOK, gin.H{"message": "Compte réactivé."})
}
