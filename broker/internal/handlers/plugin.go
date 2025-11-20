package handlers

import (
    "net/http"
    "strings"

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

    if p.Host == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "host is required"})
        return
    }

    // Normalize paths
    p.BaseAPIRoute = normalizePath(p.BaseAPIRoute)
    p.SettingsRoute = normalizePath(p.SettingsRoute)
    for i := range p.APIRoutes {
        p.APIRoutes[i] = normalizePath(p.APIRoutes[i])
    }

    // Check for route conflicts
    if err := checkRouteConflicts(&p, ""); err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "route conflict", "details": err.Error()})
        return
    }

    if err := plugins.Global.Register(&p); err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "could not register plugin", "details": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "plugin registered", "plugin": p})
}

// normalizePath ensures paths have a leading slash and no trailing slash
func normalizePath(path string) string {
    if path == "" {
        return ""
    }
    path = strings.TrimSpace(path)
    if !strings.HasPrefix(path, "/") {
        path = "/" + path
    }
    path = strings.TrimSuffix(path, "/")
    return path
}

// checkRouteConflicts validates that base-api-route doesn't conflict with broker routes or other plugins
func checkRouteConflicts(p *models.Plugin, excludeSlug string) error {
    if p.BaseAPIRoute == "" {
        return nil
    }

    // Reserved broker routes
    reservedRoutes := []string{"/api/v1/status", "/api/v1/route", "/api/v1/routes"}
    for _, reserved := range reservedRoutes {
        if strings.HasPrefix(reserved, p.BaseAPIRoute) || strings.HasPrefix(p.BaseAPIRoute, reserved) {
            return http.ErrAbortHandler // Using a standard error; could define custom
        }
    }

    // Check against other plugins
    existingRoutes := plugins.Global.GetAllBaseRoutes()
    for _, route := range existingRoutes {
        // Skip if this is an update and the route belongs to the plugin being updated
        if excludeSlug != "" {
            if existing := plugins.Global.Get(excludeSlug); existing != nil && existing.BaseAPIRoute == route {
                continue
            }
        }
        if route == p.BaseAPIRoute {
            return http.ErrAbortHandler
        }
    }

    return nil
}

// ListPlugins returns all registered plugins (GET /api/v1/routes)
func ListPlugins(c *gin.Context) {
    list := plugins.Global.List()
    c.JSON(http.StatusOK, gin.H{"plugins": list})
}

// UpdatePlugin handles plugin update PUT /api/v1/route/:slug
func UpdatePlugin(c *gin.Context) {
    slug := c.Param("slug")
    if slug == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
        return
    }

    var p models.Plugin
    if err := c.ShouldBindJSON(&p); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload", "details": err.Error()})
        return
    }

    // Ensure slug in URL matches slug in body
    if p.Slug != "" && p.Slug != slug {
        c.JSON(http.StatusBadRequest, gin.H{"error": "slug in URL and body must match"})
        return
    }
    p.Slug = slug

    // Normalize paths
    p.BaseAPIRoute = normalizePath(p.BaseAPIRoute)
    p.SettingsRoute = normalizePath(p.SettingsRoute)
    for i := range p.APIRoutes {
        p.APIRoutes[i] = normalizePath(p.APIRoutes[i])
    }

    // Check for route conflicts (excluding this plugin's current routes)
    if err := checkRouteConflicts(&p, slug); err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "route conflict", "details": err.Error()})
        return
    }

    if err := plugins.Global.Update(&p); err != nil {
        if err.Error() == "plugin not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update plugin", "details": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "plugin updated", "plugin": p})
}

// DeletePlugin handles plugin deletion DELETE /api/v1/route/:slug
func DeletePlugin(c *gin.Context) {
    slug := c.Param("slug")
    if slug == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
        return
    }

    if err := plugins.Global.Delete(slug); err != nil {
        if err.Error() == "plugin not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete plugin", "details": err.Error()})
        }
        return
    }

    c.JSON(http.StatusNoContent, nil)
}
