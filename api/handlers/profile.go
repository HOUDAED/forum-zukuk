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

type ConnectionLog struct {
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Status    string    `json:"status"` // Nouveau champ
	CreatedAt time.Time `json:"created_at"`
}

type ActivityPost struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type ActivityComment struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	PostTitle string    `json:"post_title"`
	CreatedAt time.Time `json:"created_at"`
}

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

func GetConnectionHistory(c *gin.Context) {
	userID := c.GetInt("userID")
	
	// On récupère les 10 dernières pour bien voir les échecs
	rows, err := database.ZUKUKDB.Query(`
		SELECT ip_address, user_agent, status, created_at 
		FROM connection_history 
		WHERE user_id = ? 
		ORDER BY created_at DESC LIMIT 10`, 
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	var history []ConnectionLog
	for rows.Next() {
		var log ConnectionLog
		if err := rows.Scan(&log.IPAddress, &log.UserAgent, &log.Status, &log.CreatedAt); err == nil {
			history = append(history, log)
		}
	}
	
	c.JSON(http.StatusOK, history)
}


func GetMyActivity(c *gin.Context) {
	userID := c.GetInt("userID")

	// Récupérer les publications
	var posts []ActivityPost
	rowsP, err := database.ZUKUKDB.Query(`
		SELECT id, title, created_at 
		FROM posts 
		WHERE user_id = ? AND deleted_at IS NULL 
		ORDER BY created_at DESC`, userID)
	if err == nil {
		defer rowsP.Close()
		for rowsP.Next() {
			var p ActivityPost
			if err := rowsP.Scan(&p.ID, &p.Title, &p.CreatedAt); err == nil {
				posts = append(posts, p)
			}
		}
	}

	// Récupérer les commentaires avec le titre du post associé
	var comments []ActivityComment
	rowsC, err := database.ZUKUKDB.Query(`
		SELECT c.id, c.content, c.created_at, p.title 
		FROM comments c 
		JOIN posts p ON c.post_id = p.id 
		WHERE c.user_id = ? AND c.deleted_at IS NULL 
		ORDER BY c.created_at DESC`, userID)
	if err == nil {
		defer rowsC.Close()
		for rowsC.Next() {
			var com ActivityComment
			if err := rowsC.Scan(&com.ID, &com.Content, &com.CreatedAt, &com.PostTitle); err == nil {
				comments = append(comments, com)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":    posts,
		"comments": comments,
	})
}