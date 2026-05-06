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

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié."})
			c.Abort()
			return
		}

		var userID int
		err = database.ZUKUKDB.QueryRow(`
			SELECT s.user_id
			FROM sessions s
			JOIN users u ON u.id = s.user_id
			WHERE s.token = ? AND s.expires_at > ? AND u.deleted_at IS NULL`,
			token, time.Now().UTC(),
		).Scan(&userID)
		if err != nil {
			c.SetCookie("session_token", "", -1, "/", "", false, true)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expirée."})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
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