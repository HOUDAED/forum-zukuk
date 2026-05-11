package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)


type UserSettings struct {
	AnonByDefault  bool   `json:"anon_by_default"`
	ShowMood       bool   `json:"show_mood"`
	NotifLikes     bool   `json:"notif_likes"`
	NotifComments  bool   `json:"notif_comments"`
	Theme          string `json:"theme"`          // "light" | "dark" | "glass"
	ReduceAnim     bool   `json:"reduce_anim"`
	ContentDensity string `json:"content_density"` // "normal" | "compact"
}

// GET SETTINGS


func GetSettings(c *gin.Context) {
	userID := c.GetInt("userID")

	var s UserSettings
	var anonByDefault, showMood, notifLikes, notifComments, reduceAnim int

	err := database.ZUKUKDB.QueryRow(`
		SELECT anon_by_default, show_mood, notif_likes, notif_comments,
		       theme, reduce_anim, content_density
		FROM user_settings WHERE user_id = ?`, userID,
	).Scan(&anonByDefault, &showMood, &notifLikes, &notifComments,
		&s.Theme, &reduceAnim, &s.ContentDensity)

	if err != nil {
		// Première visite : créer les paramètres par défaut
		database.ZUKUKDB.Exec(`
			INSERT OR IGNORE INTO user_settings (user_id)
			VALUES (?)`, userID)
		s = UserSettings{
			ShowMood:       true,
			NotifLikes:     true,
			NotifComments:  true,
			Theme:          "glass", // On met ton thème emblématique par défaut !
			ContentDensity: "normal",
		}
	} else {
		s.AnonByDefault = anonByDefault == 1
		s.ShowMood      = showMood == 1
		s.NotifLikes    = notifLikes == 1
		s.NotifComments = notifComments == 1
		s.ReduceAnim    = reduceAnim == 1
	}

	c.JSON(http.StatusOK, s)
}


// UPDATE SETTINGS


func UpdateSettings(c *gin.Context) {
	userID := c.GetInt("userID")
	var req UserSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides."})
		return
	}

	// Validation du thème
	if req.Theme != "light" && req.Theme != "dark" && req.Theme != "glass" {
		req.Theme = "glass"
	}
	if req.ContentDensity != "normal" && req.ContentDensity != "compact" {
		req.ContentDensity = "normal"
	}

	b := func(v bool) int {
		if v { return 1 }
		return 0
	}

	_, err := database.ZUKUKDB.Exec(`
		INSERT INTO user_settings
			(user_id, anon_by_default, show_mood, notif_likes, notif_comments,
			 theme, reduce_anim, content_density, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
			anon_by_default  = excluded.anon_by_default,
			show_mood        = excluded.show_mood,
			notif_likes      = excluded.notif_likes,
			notif_comments   = excluded.notif_comments,
			theme            = excluded.theme,
			reduce_anim      = excluded.reduce_anim,
			content_density  = excluded.content_density,
			updated_at       = CURRENT_TIMESTAMP`,
		userID,
		b(req.AnonByDefault), b(req.ShowMood),
		b(req.NotifLikes), b(req.NotifComments),
		req.Theme, b(req.ReduceAnim), req.ContentDensity,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur enregistrement paramètres."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Paramètres enregistrés."})
}

// SESSIONS


type SessionInfo struct {
	Token     string    `json:"token"`     // tronqué pour sécurité
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsCurrent bool      `json:"is_current"`
}

func GetSessions(c *gin.Context) {
	userID := c.GetInt("userID")
	currentToken, _ := c.Cookie("session_token")

	rows, err := database.ZUKUKDB.Query(`
		SELECT token, created_at, expires_at
		FROM sessions
		WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	sessions := []SessionInfo{}
	for rows.Next() {
		var s SessionInfo
		var token string
		if err := rows.Scan(&token, &s.CreatedAt, &s.ExpiresAt); err == nil {
			if len(token) > 8 {
				s.Token = token[:8] + "…"
			} else {
				s.Token = token
			}
			s.IsCurrent = (token == currentToken)
			sessions = append(sessions, s)
		}
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func RevokeAllSessions(c *gin.Context) {
	userID := c.GetInt("userID")
	currentToken, _ := c.Cookie("session_token")

	database.ZUKUKDB.Exec(`
		DELETE FROM sessions
		WHERE user_id = ? AND token != ?`, userID, currentToken)

	c.JSON(http.StatusOK, gin.H{"message": "Toutes les autres sessions ont été révoquées."})
}

// EXPORT DE DONNÉES (RGPD)


func ExportMyData(c *gin.Context) {
	userID := c.GetInt("userID")

	var pseudo, email, avatarURL string
	var createdAt time.Time
	database.ZUKUKDB.QueryRow(`
		SELECT pseudo, email, avatar_url, created_at
		FROM users WHERE id = ?`, userID,
	).Scan(&pseudo, &email, &avatarURL, &createdAt)

	// Ajouter le domaine à l'avatar si besoin
	if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
		avatarURL = "http://localhost:8081" + avatarURL
	}

	type ExportPost struct {
		ID        int       `json:"id"`
		Title     string    `json:"title"`
		Content   string    `json:"content"`
		Category  string    `json:"category"`
		CreatedAt time.Time `json:"created_at"`
	}
	posts := []ExportPost{}
	rp, _ := database.ZUKUKDB.Query(`
		SELECT p.id, p.title, p.content, COALESCE(cat.name,''), p.created_at
		FROM posts p
		LEFT JOIN post_categories cat ON cat.id = p.category_id
		WHERE p.user_id = ? AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC`, userID)
	if rp != nil {
		defer rp.Close()
		for rp.Next() {
			var ep ExportPost
			rp.Scan(&ep.ID, &ep.Title, &ep.Content, &ep.Category, &ep.CreatedAt)
			posts = append(posts, ep)
		}
	}

	type ExportComment struct {
		ID        int       `json:"id"`
		Content   string    `json:"content"`
		PostTitle string    `json:"post_title"`
		CreatedAt time.Time `json:"created_at"`
	}
	comments := []ExportComment{}
	rc, _ := database.ZUKUKDB.Query(`
		SELECT c.id, c.content, COALESCE(p.title,''), c.created_at
		FROM comments c
		LEFT JOIN posts p ON p.id = c.post_id
		WHERE c.user_id = ? AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC`, userID)
	if rc != nil {
		defer rc.Close()
		for rc.Next() {
			var ec ExportComment
			rc.Scan(&ec.ID, &ec.Content, &ec.PostTitle, &ec.CreatedAt)
			comments = append(comments, ec)
		}
	}

	type ExportMood struct {
		Mood string    `json:"mood"`
		Day  time.Time `json:"recorded_at"`
	}
	moods := []ExportMood{}
	rm, _ := database.ZUKUKDB.Query(`
		SELECT mood, recorded_at FROM mood_history
		WHERE user_id = ? ORDER BY recorded_at DESC`, userID)
	if rm != nil {
		defer rm.Close()
		for rm.Next() {
			var em ExportMood
			rm.Scan(&em.Mood, &em.Day)
			moods = append(moods, em)
		}
	}

	type ExportConn struct {
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}
	conns := []ExportConn{}
	rconn, _ := database.ZUKUKDB.Query(`
		SELECT status, created_at FROM connection_history
		WHERE user_id = ? ORDER BY created_at DESC LIMIT 50`, userID)
	if rconn != nil {
		defer rconn.Close()
		for rconn.Next() {
			var ec ExportConn
			rconn.Scan(&ec.Status, &ec.CreatedAt)
			conns = append(conns, ec)
		}
	}

	export := gin.H{
		"exported_at": time.Now().UTC(),
		"user": gin.H{
			"pseudo":     pseudo,
			"email":      email,
			"avatar_url": avatarURL,
			"created_at": createdAt,
		},
		"posts":              posts,
		"comments":           comments,
		"mood_history":       moods,
		"connection_history": conns,
	}

	filename := "zukuk-export-" + strings.ReplaceAll(pseudo, " ", "_") + ".json"
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.JSON(http.StatusOK, export)
}