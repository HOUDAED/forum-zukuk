package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"forum-zukuk/api/handlers"
	"forum-zukuk/api/middleware"
	"forum-zukuk/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: Fichier .env non trouvé")
	}

	database.Init()

	router := gin.Default()

	// ── CORS ────────────────────────────────────────────────────────────────
	router.Use(func(c *gin.Context) {
		origin := os.Getenv("FRONTEND_ORIGIN")
		if origin == "" {
			origin = "http://localhost:3000"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// ── Static uploads ───────────────────────────────────────────────────────
	router.Static("/uploads", filepath.Join("..", "database", "uploads"))

	api := router.Group("/api")
	limiter := middleware.RateLimit(10, time.Minute)

	// ── Auth ─────────────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", limiter, handlers.Register)
		auth.POST("/login", limiter, handlers.Login)
		auth.POST("/logout", middleware.RequireAuth(), handlers.Logout)
		auth.POST("/forgot-password", limiter, handlers.ForgotPassword)
		auth.POST("/reset-password", handlers.ResetPassword)
	}

	// ── Endpoints publics ────────────────────────────────────────────────────
	api.GET("/health", handlers.Health)
	api.GET("/board", handlers.GetBoardData) // rétrocompatibilité
	api.GET("/quote", handlers.GetRandomQuote)
	api.GET("/categories", handlers.GetCategories)

	// Posts (lecture publique)
	api.GET("/posts", handlers.GetPosts)
	api.GET("/posts/:id", handlers.GetPost)

	// Réseau / profils publics
	api.GET("/network", handlers.GetNetwork)
	api.GET("/network/:id", handlers.GetPublicProfile)

	// Rétrocompatibilité board.js v1
	api.GET("/discussion/:id", handlers.GetDiscussionByID)

	// ── Endpoints protégés ───────────────────────────────────────────────────
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		// ── Profil ─────────────────────────────────────────────────────────
		protected.GET("/me", handlers.GetMe)
		protected.PUT("/me", handlers.UpdateProfile)
		protected.DELETE("/me", handlers.DeleteAccount)
		protected.POST("/me/avatar", handlers.UploadAvatar)
		protected.GET("/me/activity", handlers.GetMyActivity)
		protected.GET("/me/connections", handlers.GetConnectionHistory)
		protected.POST("/me/pause", handlers.PauseAccount)

		// ── Humeur ─────────────────────────────────────────────────────────
		protected.POST("/mood", handlers.UpdateMood)               // enregistre + persiste
		protected.GET("/me/mood-history", handlers.GetMoodHistory) // 30 derniers jours
		protected.GET("/me/mood-stats", handlers.GetMoodStats)
		// répartition 30 jours

		// ── Posts CRUD ─────────────────────────────────────────────────────
		protected.POST("/posts", handlers.CreatePost)
		protected.PUT("/posts/:id", handlers.UpdatePost)
		protected.DELETE("/posts/:id", handlers.DeletePost)

		// Like toggle + vérification
		protected.POST("/posts/:id/like", handlers.ToggleLike)
		protected.GET("/posts/:id/liked", handlers.GetLikeStatus)

		// ── Commentaires ───────────────────────────────────────────────────
		protected.POST("/posts/:id/comments", handlers.AddComment)
		protected.PUT("/comments/:id", handlers.UpdateComment)
		protected.DELETE("/comments/:id", handlers.DeleteComment)

		// ── Notifications ──────────────────────────────────────────────────
		protected.GET("/notifications", handlers.GetNotifications)
		protected.POST("/notifications/read", handlers.MarkNotificationsRead)

		// Rétrocompatibilité like board.js v1
		protected.POST("/discussion/:id/like", handlers.LikeDiscussion)
		protected.GET("/me/dashboard-stats", handlers.GetUserDashboardStats)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Serveur API Zukuk démarré sur http://localhost:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erreur démarrage serveur API : %v", err)
	}
}
