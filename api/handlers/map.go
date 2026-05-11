package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"github.com/gorilla/mux"
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
	IsParticipating   bool    `json:"is_participating"` // 🔴 LE NOM EXACT
}

// GetActivitiesHandler récupère les lieux
func GetActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	userID := getCurrentUserID(r)

	// 1. On récupère la date et l'heure exactes d'aujourd'hui, au même format que ta BDD (ex: 2026-05-11T20:00)
	now := time.Now().Format("2006-01-02T15:04")

	// 2. 🔴 LE FILTRE MAGIQUE : "WHERE a.schedule >= ?" permet de zapper le passé
	// Et "ORDER BY a.schedule ASC" trie du plus imminent au plus lointain !
	query := `
		SELECT a.id, a.created_by, a.category_id, a.name, a.description, a.address, 
		       a.latitude, a.longitude, a.schedule, a.rating, a.max_places, a.created_at,
		       (SELECT COUNT(*) FROM activity_participants WHERE activity_id = a.id) as participants_count,
		       EXISTS(SELECT 1 FROM activity_participants WHERE activity_id = a.id AND user_id = ?) as is_participating
		FROM activities a 
		WHERE a.schedule >= ? 
		ORDER BY a.schedule ASC`

	// On passe 'now' dans la requête pour remplacer le deuxième point d'interrogation
	rows, err := database.ZUKUKDB.Query(query, userID, now)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du chargement de la carte."})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// CreateActivityHandler insère une nouvelle activité
func CreateActivityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var a Activity
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Vérifie tes informations. L'adresse doit être sélectionnée dans la liste Google."})
		return
	}

	userID := getCurrentUserID(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Tu dois te connecter pour publier une activité."})
		return
	}

	query := `INSERT INTO activities (created_by, category_id, name, description, address, latitude, longitude, schedule, max_places) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := database.ZUKUKDB.Exec(query, userID, a.CategoryID, a.Name, a.Description, a.Address, a.Latitude, a.Longitude, a.Schedule, a.MaxPlaces)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur du serveur lors de la sauvegarde."})
		return
	}

	id, _ := res.LastInsertId()
	a.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// ToggleJoinHandler gère l'inscription et la désinscription
func ToggleJoinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	actID, _ := strconv.ParseInt(vars["id"], 10, 64)
	userID := getCurrentUserID(r)

	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Tu dois te connecter pour t'inscrire."})
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

	json.NewEncoder(w).Encode(map[string]interface{}{"joined": !exists, "count": count})
}

func getCurrentUserID(r *http.Request) int64 {
	// On boucle sur tous les cookies pour trouver le jeton de session Zukuk valide, 
	// peu importe son nom ("session", "token", etc.)
	for _, cookie := range r.Cookies() {
		var userID int64
		// On interroge ta vraie table `sessions` (définie dans db.go)
		err := database.ZUKUKDB.QueryRow(`
			SELECT user_id FROM sessions 
			WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`, 
			cookie.Value,
		).Scan(&userID)
		
		if err == nil && userID > 0 {
			return userID // Succès : L'utilisateur est identifié !
		}
	}
	return 0 // Non connecté
}