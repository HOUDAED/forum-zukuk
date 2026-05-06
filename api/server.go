package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/database"
	"forum-zukuk/api/handlers"
	"forum-zukuk/api/middleware"
)

func main() {
	database.Init()

	router := gin.Default()

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

	router.Static("/uploads", filepath.Join("..", "database", "uploads"))

	api := router.Group("/api")
	limiter := middleware.RateLimit(10, time.Minute)

	auth := api.Group("/auth")
	{
		auth.POST("/register", limiter, handlers.Register)
		auth.POST("/login", limiter, handlers.Login)
		auth.POST("/logout", middleware.RequireAuth(), handlers.Logout)
		auth.POST("/forgot-password", limiter, handlers.ForgotPassword)
		auth.POST("/reset-password", handlers.ResetPassword)
	}

	api.GET("/health", handlers.Health)
	api.GET("/board", handlers.GetBoardData)
	api.GET("/quote", handlers.GetRandomQuote)
	api.GET("/discussion/:id", handlers.GetDiscussionByID)

	protected := api.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/me", handlers.GetMe)
		protected.PUT("/me", handlers.UpdateProfile)
		protected.DELETE("/me", handlers.DeleteAccount)
		protected.POST("/me/avatar", handlers.UploadAvatar)
		protected.POST("/mood", handlers.UpdateMood)
		protected.GET("/me/activity", handlers.GetMyActivity)
		protected.GET("/me/connections", handlers.GetConnectionHistory)
		protected.POST("/discussion", handlers.CreateDiscussion)
		protected.POST("/discussion/:id/like", handlers.LikeDiscussion)
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
