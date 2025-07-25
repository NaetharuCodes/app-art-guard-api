package handlers

import (
	"art-guard-api/models"
	"art-guard-api/services"
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

	// Populate image URLs for each artwork
	cfService := services.NewCloudflareImagesService()
	for i := range portfolioItems {
		portfolioItems[i].Artwork.ImageURL = cfService.GetImageURL(portfolioItems[i].Artwork.CloudflareImageID, "large") // Use large as default
		portfolioItems[i].Artwork.ImageVariants.Thumbnail = cfService.GetImageURL(portfolioItems[i].Artwork.CloudflareImageID, "thumbnail")
		portfolioItems[i].Artwork.ImageVariants.Medium = cfService.GetImageURL(portfolioItems[i].Artwork.CloudflareImageID, "medium")
		portfolioItems[i].Artwork.ImageVariants.Large = cfService.GetImageURL(portfolioItems[i].Artwork.CloudflareImageID, "large")
		portfolioItems[i].Artwork.ImageVariants.Original = cfService.GetImageURL(portfolioItems[i].Artwork.CloudflareImageID, "large") // Large is the "original" quality
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

// REMOVE FROM PORTFOLIO
// =====================
func (h *PortfolioHandler) RemoveFromPortfolio(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	portfolioID := c.Param("id")

	// Find the portfolio item to ensure it belongs to this user
	var portfolioItem models.Portfolio
	err := h.DB.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolioItem).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Delete the portfolio item
	err = h.DB.Delete(&portfolioItem).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Removed from portfolio successfully",
	})
}
