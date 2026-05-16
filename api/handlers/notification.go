package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)

// ─── GetNotifications ─────────────────────────────────────────────────────────
// ─── GetNotifications ─────────────────────────────────────────────────────────
func GetNotifications(c *gin.Context) {
	userID := c.GetInt("userID")

	// 🚨 MOUCHARD 1 : Qui fait la demande ?
	fmt.Println("➡️ Requête de notifications reçue pour le userID :", userID)

	rows, err := database.ZUKUKDB.Query(`
        SELECT n.id, n.type, n.post_id, n.is_read, n.created_at,
               u.pseudo AS actor_name,
               u.avatar_url AS actor_avatar
        FROM notifications n
        JOIN users u ON u.id = n.actor_id
        WHERE n.user_id = ?
        ORDER BY n.created_at DESC
        LIMIT 20`, userID)

	if err != nil {
		fmt.Println("❌ Erreur SQL Notifications :", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur."})
		return
	}
	defer rows.Close()

	type NotifItem struct {
		ID          int    `json:"id"`
		Type        string `json:"type"`
		PostID      int    `json:"post_id"`
		IsRead      int    `json:"is_read"`
		CreatedAt   string `json:"created_at"`
		ActorName   string `json:"actor_name"`
		ActorAvatar string `json:"actor_avatar"`
	}

	notifs := []NotifItem{}
	lignesTrouvees := 0

	for rows.Next() {
		lignesTrouvees++
		var n NotifItem
		var avatar sql.NullString
		var createdAt sql.NullString

		err := rows.Scan(&n.ID, &n.Type, &n.PostID, &n.IsRead, &createdAt, &n.ActorName, &avatar)
		if err != nil {
			// 🚨 MOUCHARD 2 : Erreur de lecture
			fmt.Println("🔥 ERREUR LECTURE LIGNE", lignesTrouvees, ":", err)
		} else {
			n.CreatedAt = createdAt.String
			n.ActorAvatar = avatar.String
			notifs = append(notifs, n)
		}
	}

	// 🚨 MOUCHARD 3 : Le bilan
	fmt.Printf("✅ Bilan : %d lignes trouvées en base, %d ajoutées au JSON.\n", lignesTrouvees, len(notifs))

	c.JSON(http.StatusOK, gin.H{"notifications": notifs})
}

// ─── MarkNotificationsRead ────────────────────────────────────────────────────
func MarkNotificationsRead(c *gin.Context) {
	userID := c.GetInt("userID")

	// On passe toutes les notifications non lues à "lues" (is_read = 1)
	database.ZUKUKDB.Exec(`UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0`, userID)

	c.JSON(http.StatusOK, gin.H{"message": "Notifications marquées comme lues."})
}