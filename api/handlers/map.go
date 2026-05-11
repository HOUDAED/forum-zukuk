package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"votre-projet/database" // Remplace par le chemin réel de ton package database
)

// Activity correspond exactement aux colonnes de ta table SQL
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

// GetActivitiesHandler récupère les lieux avec calcul des participants en temps réel
func GetActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	userID := getCurrentUserID(r)

	query := `
		SELECT a.id, a.created_by, a.category_id, a.name, a.description, a.address, 
		       a.latitude, a.longitude, a.schedule, a.rating, a.max_places, a.created_at,
		       (SELECT COUNT(*) FROM activity_participants WHERE activity_id = a.id) as participants_count,
		       EXISTS(SELECT 1 FROM activity_participants WHERE activity_id = a.id AND user_id = ?) as is_participating
		FROM activities a 
		ORDER BY a.created_at DESC`

	rows, err := database.ZUKUKDB.Query(query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// CreateActivityHandler insère une nouvelle activité proposée par un utilisateur
func CreateActivityHandler(w http.ResponseWriter, r *http.Request) {
	var a Activity
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	userID := getCurrentUserID(r)
	if userID == 0 {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	query := `INSERT INTO activities (created_by, category_id, name, description, address, latitude, longitude, schedule, max_places) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := database.ZUKUKDB.Exec(query, userID, a.CategoryID, a.Name, a.Description, a.Address, a.Latitude, a.Longitude, a.Schedule, a.MaxPlaces)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	a.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// ToggleJoinHandler gère l'inscription et la désinscription
func ToggleJoinHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actID, _ := strconv.ParseInt(vars["id"], 10, 64)
	userID := getCurrentUserID(r)

	if userID == 0 {
		http.Error(w, "Action impossible : non connecté", http.StatusUnauthorized)
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
	cookie, err := r.Cookie("user_id")
	if err != nil { return 0 }
	id, _ := strconv.ParseInt(cookie.Value, 10, 64)
	return id
}