package main

import (
	"art-guard-api/config"
	"art-guard-api/handlers"
	"art-guard-api/middleware"
	"art-guard-api/models"
	"log"

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

	// Auto-migrate the schema
	db.AutoMigrate(&models.User{})

	// Create handlers
	authHandler := &handlers.AuthHandler{DB: db}

	// Create Gin router
	router := gin.Default()

	// Routes
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			c.JSON(200, gin.H{"message": "Protected route", "user_id": userID})
		})
	}

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Start server
	router.Run(":8080")
}
