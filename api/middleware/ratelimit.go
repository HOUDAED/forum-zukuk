package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = map[string]*visitor{}
	mu       sync.Mutex
)

// RequireAuth vérifie la session ET bloque les comptes en pause.
// Si l'utilisateur a paused_until dans le futur → 403 avec date de retour.
// Si paused_until est dépassé → on le lève automatiquement (reprise silencieuse).
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié."})
			c.Abort()
			return
		}

		var userID int
		var pausedUntil *time.Time

		err = database.ZUKUKDB.QueryRow(`
			SELECT s.user_id, u.paused_until
			FROM sessions s
			JOIN users u ON u.id = s.user_id
			WHERE s.token = ?
			  AND s.expires_at > ?
			  AND u.deleted_at IS NULL`,
			token, time.Now().UTC(),
		).Scan(&userID, &pausedUntil)

		if err != nil {
			c.SetCookie("session_token", "", -1, "/", "", false, true)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expirée."})
			c.Abort()
			return
		}

		// Compte en pause ?
		if pausedUntil != nil {
			if time.Now().UTC().Before(*pausedUntil) {
				// Pause encore active → bloquer
				c.JSON(http.StatusForbidden, gin.H{
					"error":       "Ton compte est en pause.",
					"paused_until": pausedUntil,
					"code":        "ACCOUNT_PAUSED",
				})
				c.Abort()
				return
			}
			// Pause expirée → lever silencieusement
			database.ZUKUKDB.Exec(`
				UPDATE users SET paused_until = NULL, updated_at = CURRENT_TIMESTAMP
				WHERE id = ?`, userID)
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip  := c.ClientIP()
		now := time.Now()

		mu.Lock()
		v, ok := visitors[ip]
		if !ok || now.Sub(v.lastSeen) > window {
			visitors[ip] = &visitor{lastSeen: now, count: 1}
			mu.Unlock()
			c.Next()
			return
		}
		v.count++
		v.lastSeen = now
		if v.count > maxRequests {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Trop de tentatives. Veuillez patienter."})
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}