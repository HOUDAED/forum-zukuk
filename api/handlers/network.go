package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)

// GetNetwork – liste des membres avec stats et humeur courante
func GetNetwork(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))

	sql := `
		SELECT
			u.id,
			u.pseudo,
			COALESCE(u.avatar_url, '') AS avatar_url,
			COALESCE(mh.mood, '')      AS current_mood,
			COUNT(DISTINCT p.id)       AS posts_count,
			COUNT(DISTINCT com.id)     AS comments_count,
			COUNT(DISTINCT pl.id)      AS likes_given
		FROM users u
		LEFT JOIN (
			SELECT user_id, mood
			FROM mood_history
			WHERE id IN (SELECT MAX(id) FROM mood_history GROUP BY user_id)
		) mh ON mh.user_id = u.id
		LEFT JOIN posts    p   ON p.user_id   = u.id AND p.deleted_at   IS NULL
		LEFT JOIN comments com ON com.user_id = u.id AND com.deleted_at IS NULL
		LEFT JOIN post_likes pl ON pl.user_id = u.id
		WHERE u.deleted_at IS NULL`

	args := []interface{}{}
	if search != "" {
		sql += ` AND LOWER(u.pseudo) LIKE LOWER(?)`
		args = append(args, "%"+search+"%")
	}
	sql += ` GROUP BY u.id ORDER BY posts_count DESC LIMIT 100`

	rows, err := database.ZUKUKDB.Query(sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	type Member struct {
		ID            int    `json:"id"`
		Pseudo        string `json:"pseudo"`
		AvatarURL     string `json:"avatar_url"`
		CurrentMood   string `json:"current_mood"`
		PostsCount    int    `json:"posts_count"`
		CommentsCount int    `json:"comments_count"`
		LikesGiven    int    `json:"likes_given"`
	}

	members := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Pseudo, &m.AvatarURL, &m.CurrentMood,
			&m.PostsCount, &m.CommentsCount, &m.LikesGiven); err == nil {
			members = append(members, m)
		}
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

// GetPublicProfile – profil public d'un utilisateur avec ses posts non-anonymes
// GetPublicProfile – profil public d'un utilisateur avec ses posts non-anonymes
func GetPublicProfile(c *gin.Context) {
	idStr := c.Param("id")

	type PublicUser struct {
		ID            int    `json:"id"`
		Pseudo        string `json:"pseudo"`
		AvatarURL     string `json:"avatar_url"`
		CurrentMood   string `json:"current_mood"`
		Bio           string `json:"bio"` // 🔴 1. AJOUT DE LA BIO DANS LA STRUCTURE
		PostsCount    int    `json:"posts_count"`
		CommentsCount int    `json:"comments_count"`
		LikesGiven    int    `json:"likes_given"`
	}

	var u PublicUser
	err := database.ZUKUKDB.QueryRow(`
		SELECT
			u.id, u.pseudo, COALESCE(u.avatar_url,'') AS avatar_url,
			COALESCE(mh.mood,'') AS current_mood,
			COALESCE(u.bio, '') AS bio, -- 🔴 2. AJOUT DU SELECT DE LA BIO
			COUNT(DISTINCT p.id)    AS posts_count,
			COUNT(DISTINCT com.id)  AS comments_count,
			COUNT(DISTINCT pl.id)   AS likes_given
		FROM users u
		LEFT JOIN (
			SELECT user_id, mood
			FROM mood_history
			WHERE id IN (SELECT MAX(id) FROM mood_history GROUP BY user_id)
		) mh ON mh.user_id = u.id
		LEFT JOIN posts    p   ON p.user_id   = u.id AND p.deleted_at   IS NULL
		LEFT JOIN comments com ON com.user_id = u.id AND com.deleted_at IS NULL
		LEFT JOIN post_likes pl ON pl.user_id = u.id
		WHERE u.id = ? AND u.deleted_at IS NULL
		GROUP BY u.id`, idStr,
	).Scan(&u.ID, &u.Pseudo, &u.AvatarURL, &u.CurrentMood, &u.Bio, // 🔴 3. AJOUT DU SCAN DE LA BIO
		&u.PostsCount, &u.CommentsCount, &u.LikesGiven)
		
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profil introuvable."})
		return
	}

	// Posts publics (non-anonymes) de cet utilisateur
	type PostSummary struct {
		ID            int    `json:"id"`
		Title         string `json:"title"`
		Category      string `json:"category"`
		LikesCount    int    `json:"likes_count"`
		CommentsCount int    `json:"comments_count"`
		CreatedAt     string `json:"created_at"`
	}

	rows, _ := database.ZUKUKDB.Query(`
		SELECT p.id, p.title, COALESCE(cat.name,'') AS category,
			COUNT(DISTINCT pl.id) AS likes_count,
			COUNT(DISTINCT com.id) AS comments_count,
			p.created_at
		FROM posts p
		LEFT JOIN post_categories cat ON cat.id = p.category_id
		LEFT JOIN post_likes pl ON pl.post_id = p.id
		LEFT JOIN comments com ON com.post_id = p.id AND com.deleted_at IS NULL
		WHERE p.user_id = ? AND p.deleted_at IS NULL AND p.is_anonymous = 0
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT 20`, u.ID)

	posts := []PostSummary{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var p PostSummary
			if err := rows.Scan(&p.ID, &p.Title, &p.Category,
				&p.LikesCount, &p.CommentsCount, &p.CreatedAt); err == nil {
				posts = append(posts, p)
			}
		}
	}

	// Commentaires publics
	type CommentSummary struct {
		ID        int    `json:"id"`
		Content   string `json:"content"`
		PostID    int    `json:"post_id"`
		PostTitle string `json:"post_title"`
		CreatedAt string `json:"created_at"`
	}

	rowsC, _ := database.ZUKUKDB.Query(`
		SELECT c.id, c.content, c.post_id, p.title, c.created_at
		FROM comments c
		JOIN posts p ON p.id = c.post_id AND p.deleted_at IS NULL
		WHERE c.user_id = ? AND c.deleted_at IS NULL AND c.is_anonymous = 0
		ORDER BY c.created_at DESC
		LIMIT 20`, u.ID)

	comments := []CommentSummary{}
	if rowsC != nil {
		defer rowsC.Close()
		for rowsC.Next() {
			var cm CommentSummary
			if err := rowsC.Scan(&cm.ID, &cm.Content, &cm.PostID, &cm.PostTitle, &cm.CreatedAt); err == nil {
				comments = append(comments, cm)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"user": u, "posts": posts, "comments": comments})
}