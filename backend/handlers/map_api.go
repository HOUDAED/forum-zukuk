package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"forum-zukuk/backend/database"
)

type ActivityResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	City        string  `json:"city"`
	Address     string  `json:"address"`
	Category    string  `json:"category"`
	Mood        string  `json:"mood"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Description string  `json:"description"`
	Distance    string  `json:"distance"`
	Rating      string  `json:"rating"`
	People      string  `json:"people"`
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	Photo       string  `json:"photo"`
}

type CreateActivityRequest struct {
	Title       string  `json:"title"`
	City        string  `json:"city"`
	Category    string  `json:"category"`
	Mood        string  `json:"mood"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Description string  `json:"description"`
	MaxPlaces   int     `json:"maxPlaces"`
	Photo       string  `json:"photo"`
}

func GetActivities(c *gin.Context) {
	rows, err := database.ZUKUKDB.Query(`
		SELECT a.id, a.name, a.description, a.address, a.latitude, a.longitude,
		       a.rating, a.max_places, a.mood, a.photo, ac.name
		FROM activities a
		LEFT JOIN activity_categories ac ON a.category_id = ac.id
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lecture activités."})
		return
	}
	defer rows.Close()

	activities := []ActivityResponse{}
	for rows.Next() {
		var id int
		var name, description, address, mood, photo, categoryName sql.NullString
		var lat, lng, rating float64
		var maxPlaces int
		if err := rows.Scan(&id, &name, &description, &address, &lat, &lng, &rating, &maxPlaces, &mood, &photo, &categoryName); err != nil {
			continue
		}

		categoryCode := mapCategoryNameToCode(categoryName.String)
		meta := categoryMeta(categoryCode)
		location := address.String
		if location == "" {
			location = "France"
		}
		cityName := location
		if comma := strings.Index(location, ","); comma > 0 {
			cityName = strings.TrimSpace(location[:comma])
		}

		activities = append(activities, ActivityResponse{
			ID:          id,
			Title:       name.String,
			City:        cityName,
			Address:     location,
			Category:    categoryCode,
			Mood:        mood.String,
			Lat:         lat,
			Lng:         lng,
			Description: description.String,
			Distance:    "France",
			Rating:      fmt.Sprintf("%.1f", rating),
			People:      strconv.Itoa(maxPlaces),
			Icon:        meta.Icon,
			Color:       meta.Color,
			Photo:       photo.String,
		})
	}

	c.JSON(http.StatusOK, activities)
}

func CreateActivity(c *gin.Context) {
	var req CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides."})
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Category) == "" || strings.TrimSpace(req.Mood) == "" || req.Lat == 0 || req.Lng == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tous les champs obligatoires doivent être remplis."})
		return
	}

	categoryName := mapCategoryCodeToName(req.Category)
	if categoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Catégorie invalide."})
		return
	}

	var categoryID int
	if err := database.ZUKUKDB.QueryRow(`SELECT id FROM activity_categories WHERE name = ?`, categoryName).Scan(&categoryID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Catégorie introuvable."})
		return
	}

	res, err := database.ZUKUKDB.Exec(`
		INSERT INTO activities (
			created_by, category_id, name, description, address, latitude, longitude,
			schedule, rating, max_places, mood, photo
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, nil, categoryID, strings.TrimSpace(req.Title), req.Description, strings.TrimSpace(req.City), req.Lat, req.Lng, "", 0, req.MaxPlaces, strings.TrimSpace(req.Mood), req.Photo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'ajouter l'activité."})
		return
	}

	newID, _ := res.LastInsertId()
	meta := categoryMeta(req.Category)
	if strings.TrimSpace(req.City) == "" {
		req.City = "France"
	}
	cityName := req.City
	if comma := strings.Index(req.City, ","); comma > 0 {
		cityName = strings.TrimSpace(req.City[:comma])
	}

	activity := ActivityResponse{
		ID:          int(newID),
		Title:       req.Title,
		City:        cityName,
		Address:     req.City,
		Category:    req.Category,
		Mood:        req.Mood,
		Lat:         req.Lat,
		Lng:         req.Lng,
		Description: req.Description,
		Distance:    "France",
		Rating:      "0.0",
		People:      strconv.Itoa(req.MaxPlaces),
		Icon:        meta.Icon,
		Color:       meta.Color,
		Photo:       req.Photo,
	}

	c.JSON(http.StatusCreated, activity)
}

type categoryMetaData struct {
	Icon  string
	Color string
}

func categoryMeta(code string) categoryMetaData {
	switch code {
	case "sport":
		return categoryMetaData{"⚽", "blue"}
	case "relax":
		return categoryMetaData{"✦", "purple"}
	case "talk":
		return categoryMetaData{"☷", "green"}
	case "creative":
		return categoryMetaData{"☯", "pink"}
	case "social":
		return categoryMetaData{"☕", "orange"}
	case "nature":
		return categoryMetaData{"♧", "teal"}
	default:
		return categoryMetaData{"★", "gray"}
	}
}

func mapCategoryNameToCode(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "activités sportives":
		return "sport"
	case "relaxation & bien-être":
		return "relax"
	case "groupes de parole":
		return "talk"
	case "activités créatives":
		return "creative"
	case "lieux sociaux":
		return "social"
	case "activités nature":
		return "nature"
	default:
		return "other"
	}
}

func mapCategoryCodeToName(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "sport":
		return "Activités sportives"
	case "relax":
		return "Relaxation & bien-être"
	case "talk":
		return "Groupes de parole"
	case "creative":
		return "Activités créatives"
	case "social":
		return "Lieux sociaux"
	case "nature":
		return "Activités nature"
	default:
		return ""
	}
}
