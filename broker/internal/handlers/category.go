package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"broker/internal/models"
	"broker/internal/plugins"
)

// ListPluginsByCategory returns all plugins filtered by category
func ListPluginsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category parameter is required"})
		return
	}

	allPlugins := plugins.Global.List()
	var filtered []*models.Plugin

	for _, p := range allPlugins {
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
		"count":    len(filtered),
		"plugins":  filtered,
	})
}

// GetCategories returns a list of all unique plugin categories
func GetCategories(c *gin.Context) {
	allPlugins := plugins.Global.List()
	categoryMap := make(map[string]int)

	for _, p := range allPlugins {
		if p.Category != "" {
			categoryMap[p.Category]++
		}
	}

	var categories []gin.H
	for cat, count := range categoryMap {
		categories = append(categories, gin.H{
			"category": cat,
			"count":    count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"total":      len(categories),
	})
}
