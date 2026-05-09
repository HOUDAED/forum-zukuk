package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)

type Mood struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
}

type Discussion struct {
	ID       int    `json:"id"`
	Author   string `json:"author"`
	Time     string `json:"time"`
	Title    string `json:"title"`
	Tag      string `json:"tag"`
	TagColor string `json:"tagColor"`
	Likes    int    `json:"likes"`
	Comments int    `json:"comments"`
}

type Stat struct {
	Title string `json:"title"`
	Color string `json:"color"`
}

type BoardResponse struct {
	Moods       []Mood       `json:"moods"`
	Discussions []Discussion `json:"discussions"`
	Stats       []Stat       `json:"stats"`
	Quotes      []string     `json:"quotes"`
}

// validMoods contient les humeurs autorisées (identique au CHECK de la table)
var validMoods = map[string]bool{
	"Bien": true, "Calme": true, "Triste": true, "Anxieux": true, "Colère": true,
}

// GetBoardData — rétrocompatibilité (board.js v1)
func GetBoardData(c *gin.Context) {
	moods := []Mood{
		{Name: "Bien", Emoji: "😊"},
		{Name: "Calme", Emoji: "😌"},
		{Name: "Triste", Emoji: "😢"},
		{Name: "Anxieux", Emoji: "😰"},
		{Name: "Colère", Emoji: "😡"},
	}

	discussions := []Discussion{
		{ID: 1, Author: "Marie_123", Time: "2h", Title: "J'ai du mal à gérer mon stress au travail", Tag: "Stress", TagColor: "bg-[#ff5722]", Likes: 12, Comments: 8},
		{ID: 2, Author: "Thomas_zen", Time: "5h", Title: "Techniques de respiration qui m'ont aidé", Tag: "Bien-être", TagColor: "bg-[#00c853]", Likes: 24, Comments: 15},
		{ID: 3, Author: "Sophie_22", Time: "1 jour", Title: "Se sentir seul à l'université", Tag: "Solitude", TagColor: "bg-[#9c27b0]", Likes: 18, Comments: 22},
	}

	stats := []Stat{
		{Title: "Posts partagés", Color: "bg-blue-100"},
		{Title: "Réponses reçues", Color: "bg-red-50"},
		{Title: "Jours actifs", Color: "bg-purple-100"},
	}

	quotes := []string{
		"Tu n'es pas seul. Chaque jour est une nouvelle opportunité.",
		"Prendre soin de soi n'est pas un luxe, c'est une nécessité.",
		"Chaque petit pas vers la guérison est une victoire.",
		"Il est tout à fait normal de ne pas se sentir bien tous les jours.",
		"Ta santé mentale est une priorité absolue.",
		"Respire. Tu fais de ton mieux et c'est largement suffisant.",
		"N'oublie pas d'être aussi indulgent avec toi-même qu'avec les autres.",
	}

	c.JSON(http.StatusOK, BoardResponse{Moods: moods, Discussions: discussions, Stats: stats, Quotes: quotes})
}

// GetRandomQuote — citation aléatoire
func GetRandomQuote(c *gin.Context) {
	quotes := []string{
		"Tu n'es pas seul. Chaque jour est une nouvelle opportunité.",
		"Prendre soin de soi n'est pas un luxe, c'est une nécessité.",
		"Chaque petit pas vers la guérison est une victoire.",
		"Il est tout à fait normal de ne pas se sentir bien tous les jours.",
		"Ta santé mentale est une priorité absolue.",
		"Respire. Tu fais de ton mieux et c'est largement suffisant.",
		"N'oublie pas d'être aussi indulgent avec toi-même qu'avec les autres.",
	}
	randomIdx := int(time.Now().UnixNano()) % len(quotes)
	c.JSON(http.StatusOK, gin.H{"quote": quotes[randomIdx]})
}

// UpdateMood — enregistre l'humeur du jour en base (1 entrée par jour par utilisateur)
func UpdateMood(c *gin.Context) {
	userID := c.GetInt("userID")

	var mood Mood
	if err := json.NewDecoder(c.Request.Body).Decode(&mood); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides."})
		return
	}

	if !validMoods[mood.Name] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Humeur non reconnue."})
		return
	}

	// Mettre à jour l'entrée du jour si elle existe, sinon insérer
	// L'index UNIQUE est sur (user_id, date(recorded_at))
	res, err := database.ZUKUKDB.Exec(`
		UPDATE mood_history
		SET mood = ?, recorded_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND date(recorded_at) = date('now')`,
		mood.Name, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour humeur."})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		// Aucune entrée aujourd'hui → créer
		if _, err := database.ZUKUKDB.Exec(`
			INSERT INTO mood_history (user_id, mood) VALUES (?, ?)`,
			userID, mood.Name,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur enregistrement humeur."})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "mood": mood})
}

// GetMoodHistory — retourne les 30 derniers jours d'humeur de l'utilisateur connecté
func GetMoodHistory(c *gin.Context) {
	userID := c.GetInt("userID")

	rows, err := database.ZUKUKDB.Query(`
		SELECT mood, date(recorded_at) AS day, recorded_at
		FROM mood_history
		WHERE user_id = ?
		ORDER BY recorded_at DESC
		LIMIT 30`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	type MoodEntry struct {
		Mood       string    `json:"mood"`
		Day        string    `json:"day"`
		RecordedAt time.Time `json:"recorded_at"`
	}

	entries := []MoodEntry{}
	for rows.Next() {
		var e MoodEntry
		if err := rows.Scan(&e.Mood, &e.Day, &e.RecordedAt); err == nil {
			entries = append(entries, e)
		}
	}

	c.JSON(http.StatusOK, gin.H{"history": entries})
}

// GetMoodStats — résumé statistique de l'humeur (30 jours)
func GetMoodStats(c *gin.Context) {
	userID := c.GetInt("userID")

	rows, err := database.ZUKUKDB.Query(`
		SELECT mood, COUNT(*) AS total
		FROM mood_history
		WHERE user_id = ?
		  AND recorded_at >= datetime('now', '-30 days')
		GROUP BY mood
		ORDER BY total DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	type MoodStat struct {
		Mood  string `json:"mood"`
		Total int    `json:"total"`
	}

	stats := []MoodStat{}
	for rows.Next() {
		var s MoodStat
		if err := rows.Scan(&s.Mood, &s.Total); err == nil {
			stats = append(stats, s)
		}
	}

	// Total de jours actifs sur 30 jours
	var activeDays int
	database.ZUKUKDB.QueryRow(`
		SELECT COUNT(DISTINCT date(recorded_at))
		FROM mood_history
		WHERE user_id = ? AND recorded_at >= datetime('now', '-30 days')`,
		userID,
	).Scan(&activeDays)

	c.JSON(http.StatusOK, gin.H{
		"stats":       stats,
		"active_days": activeDays,
	})
}

// ─── Rétrocompatibilité board.js v1 ─────────────────────────────────────────

func CreateDiscussion(c *gin.Context) {
	var discussion Discussion
	if err := json.NewDecoder(c.Request.Body).Decode(&discussion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides."})
		return
	}
	discussion.Time     = "À l'instant"
	discussion.Likes    = 0
	discussion.Comments = 0
	c.JSON(http.StatusOK, discussion)
}

func LikeDiscussion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "discussionID": id, "message": "Like ajouté."})
}

func GetDiscussionByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide."})
		return
	}

	discussions := []Discussion{
		{ID: 1, Author: "Marie_123", Time: "2h", Title: "J'ai du mal à gérer mon stress au travail", Tag: "Stress", TagColor: "bg-[#ff5722]", Likes: 12, Comments: 8},
		{ID: 2, Author: "Thomas_zen", Time: "5h", Title: "Techniques de respiration qui m'ont aidé", Tag: "Bien-être", TagColor: "bg-[#00c853]", Likes: 24, Comments: 15},
		{ID: 3, Author: "Sophie_22", Time: "1 jour", Title: "Se sentir seul à l'université", Tag: "Solitude", TagColor: "bg-[#9c27b0]", Likes: 18, Comments: 22},
	}

	for _, d := range discussions {
		if d.ID == id {
			c.JSON(http.StatusOK, d)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Discussion introuvable."})
}