package handlers

import (
	"art-guard-api/models"
	"art-guard-api/services"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArtworkHandler struct {
	DB *gorm.DB
}

type UploadResponse struct {
	Message string         `json:"message"`
	Artwork models.Artwork `json:"artwork"`
}

// Upload handles file upload and artwork creation
func (h *ArtworkHandler) Upload(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the file directly (no ParseMultipartForm)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Simple file type validation by extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isValid := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	// Get form data using PostForm
	title := c.PostForm("title")
	if title == "" {
		title = strings.TrimSuffix(header.Filename, ext)
	}

	description := c.PostForm("description")
	software := c.PostForm("software")
	tags := c.PostForm("tags")
	aiProtection := c.PostForm("ai_protection") == "true"

	// Prepare metadata for Cloudflare
	metadata := map[string]string{
		"title":   title,
		"user_id": fmt.Sprintf("%d", userID),
	}
	if description != "" {
		metadata["description"] = description
	}

	// Initialize Cloudflare service
	cfService := services.NewCloudflareImagesService()

	// Upload to Cloudflare
	cfResp, err := cfService.UploadImage(file, header.Filename, metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to Cloudflare: " + err.Error()})
		return
	}

	// Create artwork record (you can add image dimensions later if needed)
	artwork := models.Artwork{
		UserID:              userID,
		Title:               title,
		Description:         description,
		Filename:            header.Filename,
		CloudflareImageID:   cfResp.Result.ID,
		FileSize:            header.Size,
		Format:              strings.ToUpper(strings.TrimPrefix(ext, ".")),
		Software:            software,
		Tags:                tags,
		CopyrightRegistered: true,
		AIProtectionEnabled: aiProtection,
	}

	if err := h.DB.Create(&artwork).Error; err != nil {
		cfService.DeleteImage(cfResp.Result.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save artwork"})
		return
	}

	c.JSON(http.StatusCreated, UploadResponse{
		Message: "Artwork uploaded successfully",
		Artwork: artwork,
	})
}

// GetArtworks returns all artworks for the authenticated user
func (h *ArtworkHandler) GetArtworks(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var artworks []models.Artwork
	if err := h.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&artworks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve artworks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"artworks": artworks,
		"count":    len(artworks),
	})
}

// GetArtwork returns a specific artwork by ID
func (h *ArtworkHandler) GetArtwork(c *gin.Context) {
	userID := c.GetUint("user_id")
	artworkID := c.Param("id")

	var artwork models.Artwork
	if err := h.DB.Where("id = ? AND user_id = ?", artworkID, userID).First(&artwork).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artwork not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve artwork"})
		}
		return
	}

	c.JSON(http.StatusOK, artwork)
}

// UpdateArtwork updates artwork metadata
func (h *ArtworkHandler) UpdateArtwork(c *gin.Context) {
	userID := c.GetUint("user_id")
	artworkID := c.Param("id")

	var artwork models.Artwork
	if err := h.DB.Where("id = ? AND user_id = ?", artworkID, userID).First(&artwork).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artwork not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve artwork"})
		}
		return
	}

	// Parse JSON body
	var updateData struct {
		Title               *string `json:"title"`
		Description         *string `json:"description"`
		Software            *string `json:"software"`
		Tags                *string `json:"tags"`
		AIProtectionEnabled *bool   `json:"ai_protection_enabled"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if updateData.Title != nil {
		artwork.Title = *updateData.Title
	}
	if updateData.Description != nil {
		artwork.Description = *updateData.Description
	}
	if updateData.Software != nil {
		artwork.Software = *updateData.Software
	}
	if updateData.Tags != nil {
		artwork.Tags = *updateData.Tags
	}
	if updateData.AIProtectionEnabled != nil {
		artwork.AIProtectionEnabled = *updateData.AIProtectionEnabled
	}

	if err := h.DB.Save(&artwork).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update artwork"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Artwork updated successfully",
		"artwork": artwork,
	})
}

// DeleteArtwork deletes an artwork and its file
func (h *ArtworkHandler) DeleteArtwork(c *gin.Context) {
	userID := c.GetUint("user_id")
	artworkID := c.Param("id")

	var artwork models.Artwork
	if err := h.DB.Where("id = ? AND user_id = ?", artworkID, userID).First(&artwork).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artwork not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve artwork"})
		}
		return
	}

	// Initialize Cloudflare service
	cfService := services.NewCloudflareImagesService()

	// Delete from Cloudflare
	if err := cfService.DeleteImage(artwork.CloudflareImageID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to delete Cloudflare image %s: %v\n", artwork.CloudflareImageID, err)
	}

	// Delete from database
	if err := h.DB.Delete(&artwork).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete artwork"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Artwork deleted successfully"})
}

// ServeFile serves the actual image file
func (h *ArtworkHandler) ServeFile(c *gin.Context) {
	userID := c.GetUint("user_id")
	artworkID := c.Param("id")

	var artwork models.Artwork
	if err := h.DB.Where("id = ? AND user_id = ?", artworkID, userID).First(&artwork).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artwork not found"})
		return
	}

	// Get Cloudflare service and redirect to image URL
	cfService := services.NewCloudflareImagesService()
	imageURL := cfService.GetImageURL(artwork.CloudflareImageID, "public")

	c.Redirect(http.StatusFound, imageURL)
}
