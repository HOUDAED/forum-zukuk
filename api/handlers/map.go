package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"forum-zukuk/database"
)

type Activity struct {
	ID                int64   `json:"id"`
	CreatedBy         *int64  `json:"created_by"`
	CategoryID        int     `json:"category_id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Address           string  `json:"address"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Schedule          string  `json:"schedule"`
	Rating            float64 `json:"rating"`
	MaxPlaces         int     `json:"max_places"`
	CreatedAt         string  `json:"created_at"`
	ParticipantsCount int     `json:"participants_count"`
	IsParticipating   bool    `json:"is_participating"`
}

type Participant struct {
	Pseudo    string `json:"pseudo"`
	AvatarURL string `json:"avatar_url"`
}

// ── UTILITAIRE ──────────────────────────────────────────────────────────
func getCurrentUserIDForMap(c *gin.Context) int64 {
	for _, cookie := range c.Request.Cookies() {
		var userID int64
		err := database.ZUKUKDB.QueryRow(`
			SELECT user_id FROM sessions 
			WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`, 
			cookie.Value,
		).Scan(&userID)
		
		if err == nil && userID > 0 {
			return userID
		}
	}
	return 0
}

// ── HANDLERS ────────────────────────────────────────────────────────────
func GetActivitiesHandler(c *gin.Context) {
	userID := getCurrentUserIDForMap(c)
	now := time.Now().Format("2006-01-02T15:04")

	query := `
		SELECT a.id, a.created_by, a.category_id, a.name, a.description, a.address, 
		       a.latitude, a.longitude, a.schedule, a.rating, a.max_places, a.created_at,
		       (SELECT COUNT(*) FROM activity_participants WHERE activity_id = a.id) as participants_count,
		       EXISTS(SELECT 1 FROM activity_participants WHERE activity_id = a.id AND user_id = ?) as is_participating
		FROM activities a 
		WHERE a.schedule >= ? 
		ORDER BY a.schedule ASC`

	rows, err := database.ZUKUKDB.Query(query, userID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors du chargement de la carte."})
		return
	}
	defer rows.Close()

	var list []Activity
	for rows.Next() {
		var a Activity
		rows.Scan(&a.ID, &a.CreatedBy, &a.CategoryID, &a.Name, &a.Description, &a.Address, 
		          &a.Latitude, &a.Longitude, &a.Schedule, &a.Rating, &a.MaxPlaces, &a.CreatedAt,
		          &a.ParticipantsCount, &a.IsParticipating)
		list = append(list, a)
	}

	if list == nil {
		list = []Activity{}
	}
	c.JSON(http.StatusOK, list)
}

func CreateActivityHandler(c *gin.Context) {
	var a Activity
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vérifie tes informations. L'adresse doit être sélectionnée dans la liste Google."})
		return
	}

	userID := getCurrentUserIDForMap(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tu dois te connecter pour publier une activité."})
		return
	}

	query := `INSERT INTO activities (created_by, category_id, name, description, address, latitude, longitude, schedule, max_places) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := database.ZUKUKDB.Exec(query, userID, a.CategoryID, a.Name, a.Description, a.Address, a.Latitude, a.Longitude, a.Schedule, a.MaxPlaces)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur du serveur lors de la sauvegarde."})
		return
	}

	id, _ := res.LastInsertId()
	a.ID = id
	c.JSON(http.StatusCreated, a)
}

func ToggleJoinHandler(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := getCurrentUserIDForMap(c)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tu dois te connecter pour t'inscrire."})
		return
	}

	var exists bool
	database.ZUKUKDB.QueryRow("SELECT EXISTS(SELECT 1 FROM activity_participants WHERE activity_id=? AND user_id=?)", actID, userID).Scan(&exists)

	if exists {
		database.ZUKUKDB.Exec("DELETE FROM activity_participants WHERE activity_id=? AND user_id=?", actID, userID)
	} else {
		database.ZUKUKDB.Exec("INSERT INTO activity_participants (activity_id, user_id) VALUES (?, ?)", actID, userID)
	}

	var count int
	database.ZUKUKDB.QueryRow("SELECT COUNT(*) FROM activity_participants WHERE activity_id=?", actID).Scan(&count)

	c.JSON(http.StatusOK, gin.H{"joined": !exists, "count": count})
}

func GetActivityParticipantsHandler(c *gin.Context) {
	userID := getCurrentUserIDForMap(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Connexion requise"})
		return
	}

	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var creatorID int64
	err := database.ZUKUKDB.QueryRow("SELECT created_by FROM activities WHERE id = ?", actID).Scan(&creatorID)
	
	if err != nil || creatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Seul l'organisateur peut voir la liste."})
		return
	}

	query := `
		SELECT u.pseudo, u.avatar_url 
		FROM users u
		JOIN activity_participants ap ON u.id = ap.user_id
		WHERE ap.activity_id = ?
		ORDER BY ap.joined_at ASC`

	rows, err := database.ZUKUKDB.Query(query, actID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur SQL"})
		return
	}
	defer rows.Close()

	var list []Participant
	for rows.Next() {
		var p Participant
		rows.Scan(&p.Pseudo, &p.AvatarURL)
		list = append(list, p)
	}
    
	if list == nil { 
		list = []Participant{} 
	}
	c.JSON(http.StatusOK, list)
}

func DeleteActivityHandler(c *gin.Context) {
	userID := getCurrentUserIDForMap(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var creatorID int64
	err := database.ZUKUKDB.QueryRow("SELECT created_by FROM activities WHERE id = ?", actID).Scan(&creatorID)
	if err != nil || creatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Non autorisé."})
		return
	}

	database.ZUKUKDB.Exec("DELETE FROM activities WHERE id = ?", actID)
	c.JSON(http.StatusOK, gin.H{"message": "Supprimée"})
}