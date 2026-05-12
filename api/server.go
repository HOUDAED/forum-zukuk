package main

import (
	"fmt"
	"log"
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

	// ── CORS (Version Ultra-Robuste) ────────────────────────────────────────
	router.Use(func(c *gin.Context) {
		reqOrigin := c.Request.Header.Get("Origin")
		if reqOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", reqOrigin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

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
	api.GET("/board", handlers.GetBoardData)
	api.GET("/quote", handlers.GetRandomQuote)
	api.GET("/categories", handlers.GetCategories)
	api.GET("/posts", handlers.GetPosts)
	api.GET("/posts/:id", handlers.GetPost)
	api.GET("/network", handlers.GetNetwork)
	api.GET("/network/:id", handlers.GetPublicProfile)
	api.GET("/discussion/:id", handlers.GetDiscussionByID)

	// 📍 CARTE : Lecture publique des activités (Maintenant en standard Gin)
	api.GET("/activities", handlers.GetActivitiesHandler)

	// ── Endpoints protégés ───────────────────────────────────────────────────
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		// Profil & Posts...
		protected.GET("/me", handlers.GetMe)
		protected.PUT("/me", handlers.UpdateProfile)
		protected.DELETE("/me", handlers.DeleteAccount)
		protected.POST("/me/avatar", handlers.UploadAvatar)
		protected.GET("/me/activity", handlers.GetMyActivity)
		protected.GET("/me/connections", handlers.GetConnectionHistory)
		protected.POST("/me/pause", handlers.PauseAccount)

		protected.POST("/mood", handlers.UpdateMood)
		protected.GET("/me/mood-history", handlers.GetMoodHistory)
		protected.GET("/me/mood-stats", handlers.GetMoodStats)

		protected.POST("/posts", handlers.CreatePost)
		protected.PUT("/posts/:id", handlers.UpdatePost)
		protected.DELETE("/posts/:id", handlers.DeletePost)
		protected.POST("/posts/:id/like", handlers.ToggleLike)
		protected.GET("/posts/:id/liked", handlers.GetLikeStatus)
		protected.POST("/posts/:id/comments", handlers.AddComment)
		protected.PUT("/comments/:id", handlers.UpdateComment)
		protected.DELETE("/comments/:id", handlers.DeleteComment)

		protected.GET("/notifications", handlers.GetNotifications)
		protected.POST("/notifications/read", handlers.MarkNotificationsRead)
		protected.POST("/discussion/:id/like", handlers.LikeDiscussion)
		protected.GET("/me/dashboard-stats", handlers.GetUserDashboardStats)

		// 📍 CARTE : Routes protégées (Création, Inscription, Suppression, Liste Inscrits)
		protected.POST("/activities", handlers.CreateActivityHandler)
		protected.POST("/activities/:id/join", handlers.ToggleJoinHandler)
		protected.DELETE("/activities/:id", handlers.DeleteActivityHandler)
		protected.GET("/activities/:id/participants", handlers.GetActivityParticipantsHandler)

		// Paramètres...
		settings := protected.Group("/settings")
		{
			settings.GET("/", handlers.GetSettings)
			settings.PUT("/", handlers.UpdateSettings)
			settings.GET("/sessions", handlers.GetSessions)
			settings.POST("/sessions/revoke", handlers.RevokeAllSessions)
			settings.GET("/export", handlers.ExportMyData)
			settings.POST("/pause", handlers.PauseAccount)
		}
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