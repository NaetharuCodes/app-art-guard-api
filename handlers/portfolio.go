package handlers

import (
	"art-guard-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PortfolioHandler struct {
	DB *gorm.DB
}

// GET PORTFOLIO
// =============
func (h *PortfolioHandler) GetPortfolio(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Check that the user is authenticated.
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Query the DB to get all portfolio items for this user
	var portfolioItems []models.Portfolio

	err := h.DB.Where("User_id = ?", userID).Preload("Artwork").Order("position ASC").Find(&portfolioItems).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"portfolio": portfolioItems,
		"count":     len(portfolioItems),
	})
}

// ADD TO PORTFOLIO
// =============
func (h *PortfolioHandler) AddToPortfolio(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse the request body to get artwork_id
	var request struct {
		ArtworkID uint `json:"artwork_id" binding:"required"`
		Position  int  `json:"position" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var artwork models.Artwork
	err := h.DB.Where("id = ? AND user_id = ?", request.ArtworkID, userID).First(&artwork).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artwork not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Create or update portfolio entry at this position
	portfolioItem := models.Portfolio{
		UserID:        userID,
		ArtworkID:     request.ArtworkID,
		Position:      request.Position,
		PortfolioName: "Main Portfolio",
	}

	err = h.DB.Where("user_id = ? AND position = ?", userID, request.Position).
		Assign(portfolioItem).
		FirstOrCreate(&portfolioItem).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Added to portfolio successfully",
		"portfolio_item": portfolioItem,
	})
}
