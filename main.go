package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"forum-zukuk/backend/database"
	"forum-zukuk/backend/handlers"
	"forum-zukuk/backend/middleware"
)

func main() {
	database.Init()

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		origin := os.Getenv("ALLOWED_ORIGIN")
		if origin == "" {
			origin = "*"
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

	api := router.Group("/api")
	limiter := middleware.RateLimit(10, time.Minute)

	auth := api.Group("/auth")
	{
		auth.POST("/register", limiter, handlers.Register)
		auth.POST("/login", limiter, handlers.Login)
		auth.POST("/logout", middleware.RequireAuth(), handlers.Logout)
	}

	api.GET("/health", handlers.Health)

	protected := api.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/me", handlers.GetMe)
		protected.PUT("/me", handlers.UpdateProfile)
		protected.DELETE("/me", handlers.DeleteAccount)
		protected.POST("/me/avatar", handlers.UploadAvatar)
	}

	router.Static("/uploads", filepath.Join(".", "database", "uploads"))
	router.Static("/frontend", filepath.Join(".", "frontend"))

	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(".", "frontend", "html", "index.html"))
	})
	router.GET("/login", func(c *gin.Context) {
		c.File(filepath.Join(".", "frontend", "html", "login.html"))
	})
	router.GET("/register", func(c *gin.Context) {
		c.File(filepath.Join(".", "frontend", "html", "register.html"))
	})
	router.GET("/profile", func(c *gin.Context) {
		c.File(filepath.Join(".", "frontend", "html", "profile.html"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Serveur Zukuk demarre sur http://localhost:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erreur demarrage serveur: %v", err)
	}
}
