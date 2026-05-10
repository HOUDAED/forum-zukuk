package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)

// ─── Request types ────────────────────────────────────────────────────────────

type PostRequest struct {
	Title       string `json:"title"    binding:"required"`
	Content     string `json:"content"  binding:"required"`
	CategoryID  *int   `json:"category_id"`
	IsAnonymous int    `json:"is_anonymous"`
}

type CommentRequest struct {
	Content     string `json:"content" binding:"required"`
	IsAnonymous int    `json:"is_anonymous"`
}

// ─── Response types ──────────────────────────────────────────────────────────

type PostItem struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	IsAnonymous   int       `json:"is_anonymous"`
	CreatedAt     time.Time `json:"created_at"`
	Author        string    `json:"author"`
	AvatarURL     string    `json:"avatar_url"`
	Category      string    `json:"category"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount int       `json:"comments_count"`
	UserID        int       `json:"user_id"`
}

type CommentItem struct {
	ID          int       `json:"id"`
	Content     string    `json:"content"`
	IsAnonymous int       `json:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at"`
	Author      string    `json:"author"`
	AvatarURL   string    `json:"avatar_url"`
	UserID      int       `json:"user_id"`
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func scanPost(rows interface {
	Scan(...interface{}) error
}) (PostItem, error) {
	var p PostItem
	err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.IsAnonymous, &p.CreatedAt,
		&p.Author, &p.AvatarURL, &p.Category, &p.LikesCount, &p.CommentsCount, &p.UserID)
	return p, err
}

const postSelectSQL = `
	SELECT
		p.id, p.title, p.content, p.is_anonymous, p.created_at,
		CASE WHEN p.is_anonymous = 1 THEN 'Anonyme' ELSE u.pseudo END AS author,
		CASE WHEN p.is_anonymous = 1 THEN '' ELSE u.avatar_url END AS avatar_url,
		COALESCE(cat.name,'') AS category,
		COUNT(DISTINCT pl.id) AS likes_count,
		COUNT(DISTINCT com.id) AS comments_count,
		p.user_id
	FROM posts p
	INNER JOIN users u ON u.id = p.user_id
	LEFT JOIN post_categories cat ON cat.id = p.category_id
	LEFT JOIN post_likes pl ON pl.post_id = p.id
	LEFT JOIN comments com ON com.post_id = p.id AND com.deleted_at IS NULL`

// ─── GetPosts : list + search + filter + sort ─────────────────────────────────

func GetPosts(c *gin.Context) {
	query     := strings.TrimSpace(c.Query("query"))
	sort      := c.DefaultQuery("sort", "date")
	order     := c.DefaultQuery("order", "desc")
	minDate   := c.Query("min_date")
	maxDate   := c.Query("max_date")
	minLikes  := c.DefaultQuery("min_likes", "0")
	maxLikes  := c.Query("max_likes")
	minCom   := c.DefaultQuery("min_comments", "0")
	maxCom   := c.Query("max_comments")
	limitStr  := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 || limit > 100 { limit = 20 }
	if offset < 0 { offset = 0 }

	sortCol := map[string]string{
		"date": "p.created_at", "likes": "likes_count", "comments": "comments_count",
	}[sort]
	if sortCol == "" { sortCol = "p.created_at" }
	if order != "asc" { order = "desc" }

	where := []string{
		"p.deleted_at IS NULL",
		"u.deleted_at IS NULL", // Masque les comptes supprimés
		"(u.paused_until IS NULL OR u.paused_until < CURRENT_TIMESTAMP)", // Masque les comptes en pause
	}
	args  := []interface{}{}

	if query != "" {
		where = append(where, "(LOWER(p.title) LIKE LOWER(?) OR LOWER(p.content) LIKE LOWER(?))")
		like := "%" + query + "%"
		args = append(args, like, like)
	}
	if minDate != "" { where = append(where, "date(p.created_at) >= ?"); args = append(args, minDate) }
	if maxDate != "" { where = append(where, "date(p.created_at) <= ?"); args = append(args, maxDate) }

	having := []string{}
	if v, err := strconv.Atoi(minLikes);  err == nil && v > 0 { having = append(having, fmt.Sprintf("likes_count >= %d", v)) }
	if maxLikes != "" { if v, err := strconv.Atoi(maxLikes);  err == nil { having = append(having, fmt.Sprintf("likes_count <= %d", v)) } }
	if v, err := strconv.Atoi(minCom); err == nil && v > 0 { having = append(having, fmt.Sprintf("comments_count >= %d", v)) }
	if maxCom != "" { if v, err := strconv.Atoi(maxCom); err == nil { having = append(having, fmt.Sprintf("comments_count <= %d", v)) } }

	whereStr  := "WHERE " + strings.Join(where, " AND ")
	havingStr := ""
	if len(having) > 0 { havingStr = "HAVING " + strings.Join(having, " AND ") }

	sql := fmt.Sprintf(`%s %s GROUP BY p.id %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		postSelectSQL, whereStr, havingStr, sortCol, order)

	args = append(args, limit, offset)
	rows, err := database.ZUKUKDB.Query(sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	posts := []PostItem{}
	for rows.Next() {
		if p, err := scanPost(rows); err == nil { posts = append(posts, p) }
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// ─── GetPost : single post + comments ────────────────────────────────────────

func GetPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}

	// 1. Filtrer le post lui-même
	sql := postSelectSQL + ` 
		WHERE p.id = ? 
		  AND p.deleted_at IS NULL 
		  AND u.deleted_at IS NULL 
		  AND (u.paused_until IS NULL OR u.paused_until < CURRENT_TIMESTAMP)
		GROUP BY p.id`
		
	row := database.ZUKUKDB.QueryRow(sql, id)
	p, err := scanPost(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post introuvable ou auteur inactif."})
		return
	}

	// 2. Filtrer les commentaires
	rows, err := database.ZUKUKDB.Query(`
		SELECT c.id, c.content, c.is_anonymous, c.created_at,
			CASE WHEN c.is_anonymous = 1 THEN 'Anonyme' ELSE u.pseudo END AS author,
			CASE WHEN c.is_anonymous = 1 THEN '' ELSE u.avatar_url END AS avatar_url,
			c.user_id
		FROM comments c
		INNER JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ? 
		  AND c.deleted_at IS NULL
		  AND u.deleted_at IS NULL
		  AND (u.paused_until IS NULL OR u.paused_until < CURRENT_TIMESTAMP)
		ORDER BY c.created_at ASC`, id)
		
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	comments := []CommentItem{}
	for rows.Next() {
		var cm CommentItem
		if err := rows.Scan(&cm.ID, &cm.Content, &cm.IsAnonymous, &cm.CreatedAt,
			&cm.Author, &cm.AvatarURL, &cm.UserID); err == nil {
			comments = append(comments, cm)
		}
	}
	c.JSON(http.StatusOK, gin.H{"post": p, "comments": comments})
}

// ─── CreatePost ───────────────────────────────────────────────────────────────

func CreatePost(c *gin.Context) {
	userID := c.GetInt("userID")
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Titre et contenu requis."})
		return
	}
	req.Title   = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	if len(req.Title) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le titre doit faire au moins 3 caractères."})
		return
	}
	if len(req.Content) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le contenu doit faire au moins 10 caractères."})
		return
	}

	result, err := database.ZUKUKDB.Exec(`
		INSERT INTO posts (user_id, category_id, title, content, is_anonymous)
		VALUES (?, ?, ?, ?, ?)`,
		userID, req.CategoryID, req.Title, req.Content, req.IsAnonymous)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création post."})
		return
	}
	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"message": "Post créé.", "id": newID})
}

// ─── UpdatePost ───────────────────────────────────────────────────────────────

func UpdatePost(c *gin.Context) {
	userID := c.GetInt("userID")
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	var ownerID int
	if err := database.ZUKUKDB.QueryRow(`SELECT user_id FROM posts WHERE id = ? AND deleted_at IS NULL`, postID).Scan(&ownerID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post introuvable."})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Non autorisé."})
		return
	}
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides."})
		return
	}
	req.Title   = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	if len(req.Title) < 3 || len(req.Content) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Titre (3+) et contenu (10+) requis."})
		return
	}
	database.ZUKUKDB.Exec(`UPDATE posts SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		req.Title, req.Content, postID)
	c.JSON(http.StatusOK, gin.H{"message": "Post mis à jour."})
}

// ─── DeletePost ───────────────────────────────────────────────────────────────

func DeletePost(c *gin.Context) {
	userID := c.GetInt("userID")
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	var ownerID int
	if err := database.ZUKUKDB.QueryRow(`SELECT user_id FROM posts WHERE id = ? AND deleted_at IS NULL`, postID).Scan(&ownerID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post introuvable."})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Non autorisé."})
		return
	}
	tx, _ := database.ZUKUKDB.Begin()
	defer tx.Rollback()
	tx.Exec(`UPDATE posts    SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, postID)
	tx.Exec(`UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE post_id = ?`, postID)
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Post supprimé."})
}

// ─── ToggleLike ───────────────────────────────────────────────────────────────

func ToggleLike(c *gin.Context) {
	userID := c.GetInt("userID")
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	liked, count, err := database.TogglePostLike(postID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"liked": liked, "count": count})
}

// ─── GetLikeStatus : check if current user liked a post ──────────────────────

func GetLikeStatus(c *gin.Context) {
	userID := c.GetInt("userID")
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	var count int
	database.ZUKUKDB.QueryRow(`SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND user_id = ?`, postID, userID).Scan(&count)
	c.JSON(http.StatusOK, gin.H{"liked": count > 0})
}

// ─── AddComment ───────────────────────────────────────────────────────────────

func AddComment(c *gin.Context) {
	userID := c.GetInt("userID")
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	var req CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contenu requis."})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if len(req.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Commentaire vide."})
		return
	}

	var exists int
	database.ZUKUKDB.QueryRow(`SELECT COUNT(*) FROM posts WHERE id = ? AND deleted_at IS NULL`, postID).Scan(&exists)
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post introuvable."})
		return
	}

	result, err := database.ZUKUKDB.Exec(`
		INSERT INTO comments (post_id, user_id, content, is_anonymous) VALUES (?, ?, ?, ?)`,
		postID, userID, req.Content, req.IsAnonymous)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur ajout commentaire."})
		return
	}
	newID, _ := result.LastInsertId()

	var cm CommentItem
	database.ZUKUKDB.QueryRow(`
		SELECT c.id, c.content, c.is_anonymous, c.created_at,
			CASE WHEN c.is_anonymous = 1 THEN 'Anonyme' ELSE COALESCE(u.pseudo,'?') END,
			CASE WHEN c.is_anonymous = 1 THEN '' ELSE COALESCE(u.avatar_url,'') END,
			c.user_id
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.id = ?`, newID,
	).Scan(&cm.ID, &cm.Content, &cm.IsAnonymous, &cm.CreatedAt, &cm.Author, &cm.AvatarURL, &cm.UserID)

	c.JSON(http.StatusCreated, cm)
}

// ─── UpdateComment ────────────────────────────────────────────────────────────

func UpdateComment(c *gin.Context) {
	userID    := c.GetInt("userID")
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	var ownerID int
	if err := database.ZUKUKDB.QueryRow(`SELECT user_id FROM comments WHERE id = ? AND deleted_at IS NULL`, commentID).Scan(&ownerID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Commentaire introuvable."})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Non autorisé."})
		return
	}
	var req CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contenu requis."})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if len(req.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Commentaire vide."})
		return
	}
	database.ZUKUKDB.Exec(`UPDATE comments SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		req.Content, commentID)
	c.JSON(http.StatusOK, gin.H{"message": "Commentaire mis à jour.", "content": req.Content})
}

// ─── DeleteComment ────────────────────────────────────────────────────────────

func DeleteComment(c *gin.Context) {
	userID    := c.GetInt("userID")
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	var ownerID int
	if err := database.ZUKUKDB.QueryRow(`SELECT user_id FROM comments WHERE id = ? AND deleted_at IS NULL`, commentID).Scan(&ownerID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Commentaire introuvable."})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Non autorisé."})
		return
	}
	database.ZUKUKDB.Exec(`UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, commentID)
	c.JSON(http.StatusOK, gin.H{"message": "Commentaire supprimé."})
}

// ─── GetCategories ────────────────────────────────────────────────────────────

func GetCategories(c *gin.Context) {
	rows, err := database.ZUKUKDB.Query(`SELECT id, name FROM post_categories ORDER BY name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()
	type Cat struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	cats := []Cat{}
	for rows.Next() {
		var cat Cat
		rows.Scan(&cat.ID, &cat.Name)
		cats = append(cats, cat)
	}
	c.JSON(http.StatusOK, cats)
}