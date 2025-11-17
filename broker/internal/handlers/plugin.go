package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "broker/internal/models"
    "broker/internal/plugins"
)

// RegisterPlugin handles plugin registration POST /api/v1/route
func RegisterPlugin(c *gin.Context) {
    var p models.Plugin
    if err := c.ShouldBindJSON(&p); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload", "details": err.Error()})
        return
    }

    if p.Slug == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
        return
    }

    if err := plugins.Global.Register(&p); err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "could not register plugin", "details": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "plugin registered", "plugin": p})
}

// ListPlugins returns all registered plugins (GET /api/v1/routes)
func ListPlugins(c *gin.Context) {
    list := plugins.Global.List()
    c.JSON(http.StatusOK, gin.H{"plugins": list})
}
