package handlers

import (
    "log"
    "net/http"
    "sync"

    "hotelhub/broker/repository"
    "hotelhub/broker/services"

    "github.com/gin-gonic/gin"
)

// PluginRegistration represents the minimal payload a plugin sends when registering with the broker.
type PluginRegistration struct {
    Slug          string   `json:"slug"`
    Name          string   `json:"name"`
    Version       string   `json:"version"`
    Description   string   `json:"description"`
    Host          string   `json:"host"`
    BaseAPIRoute  string   `json:"base-api-route"`
    SettingsRoute string   `json:"settings-route,omitempty"`
    APIRoutes     []string `json:"api-routes,omitempty"`
    Enabled       bool     `json:"enabled"`
}

// in-memory registry for plugins (simple, non-persistent)
var (
    registeredPlugins   = make(map[string]PluginRegistration)
    registeredPluginsMu sync.RWMutex
)

type PluginHandler struct {
    repository *repository.RepositoryStrategy
}

func NewPluginHandler(repository *repository.RepositoryStrategy) BaseHandler {
    return &PluginHandler{repository: repository}
}

func (h *PluginHandler) RegisterRoutes(engine *gin.Engine, jwt *services.JwtService) {
    // Expose a public endpoint for plugin registration
    engine.POST("/api/v1/route", h.registerPlugin())
}

func (h *PluginHandler) registerPlugin() gin.HandlerFunc {
    return func(c *gin.Context) {
        var reg PluginRegistration
        if err := c.ShouldBindJSON(&reg); err != nil {
            log.Printf("Plugin registration: invalid payload: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
            return
        }

        // store in memory
        registeredPluginsMu.Lock()
        registeredPlugins[reg.Slug] = reg
        registeredPluginsMu.Unlock()

        log.Printf("Plugin registered: %s (%s) at %s", reg.Name, reg.Slug, reg.Host)

        c.JSON(http.StatusOK, gin.H{"status": "registered"})
    }
}
