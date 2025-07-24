package main

import (
	"art-guard-api/config"
	"art-guard-api/handlers"
	"art-guard-api/middleware"
	"art-guard-api/models"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	db := config.ConnectDB()

	// Auto-migrate the schema with new fields
	db.AutoMigrate(&models.User{}, &models.Artwork{}, &models.Portfolio{})

	// Create handlers
	authHandler := &handlers.AuthHandler{DB: db}
	artworkHandler := &handlers.ArtworkHandler{DB: db}

	// Create Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // React dev servers
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Auth routes (public)
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected API routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// User profile
		api.GET("/profile", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			c.JSON(200, gin.H{"message": "Protected route", "user_id": userID})
		})

		// Artwork routes
		artworks := api.Group("/artworks")
		{
			artworks.POST("/upload", artworkHandler.Upload)
			artworks.GET("", artworkHandler.GetArtworks)
			artworks.GET("/:id", artworkHandler.GetArtwork)
			artworks.PUT("/:id", artworkHandler.UpdateArtwork)
			artworks.DELETE("/:id", artworkHandler.DeleteArtwork)
			artworks.GET("/:id/file", artworkHandler.ServeFile)
		}
	}

	// Start server
	router.Run(":8080")
}
