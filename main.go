package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"forum-zukuk/handlers"
	"forum-zukuk/middleware"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Pas de fichier .env trouvé, utilisation des variables d'environnement par défaut")
	}
}

func main() {
	router := gin.Default()

	// Middleware CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	loginLimiter := middleware.NewRateLimiter(5, 15*60)    // 5 requêtes par 15 minutes
	registerLimiter := middleware.NewRateLimiter(3, 60*60) // 3 requêtes par heure

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", loginLimiter.Limit(), handlers.Login)
			auth.POST("/verify-otp", handlers.VerifyOTP)
			auth.POST("/register", registerLimiter.Limit(), handlers.Register)
			auth.GET("/google", handlers.GoogleOAuth)
			auth.POST("/google-callback", handlers.GoogleCallback)
			auth.GET("/school", handlers.SchoolOAuth)
			auth.POST("/school-callback", handlers.SchoolCallback)
		}

		api.GET("/health", handlers.Health)
	}

	router.Static("/auth", filepath.Join(".", "auth"))
	router.Static("/sign", filepath.Join(".", "sign"))
	router.Static("/public", filepath.Join(".", "public"))

	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(".", "auth", "auth.html"))
	})

	router.GET("/auth.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/auth/auth.html")
	})

	router.GET("/sign.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/sign/sign.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Serveur démarré sur http://localhost:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}
