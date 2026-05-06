package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)

func UploadAvatar(c *gin.Context) {
	userID := c.GetInt("userID")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Photo de profil requise."})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedAvatarExtension(ext) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format accepté : jpg, jpeg, png, webp ou gif."})
		return
	}

	uploadDir := filepath.Join("..", "database", "uploads", "avatars")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur stockage avatar."})
		return
	}

	filename := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().UnixNano(), ext)
	destination := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, destination); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur enregistrement avatar."})
		return
	}

	avatarURL := "/uploads/avatars/" + filename
	if _, err := database.ZUKUKDB.Exec(`UPDATE users SET avatar_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, avatarURL, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour avatar."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Photo de profil mise à jour.", "avatar_url": avatarURL})
}

func allowedAvatarExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return true
	default:
		return false
	}
}
