package handlers

import (
	"hotelhub/broker/middleware"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"

	"github.com/gin-gonic/gin"
)

// PluginHandler handles requests for registered plugins
type PluginHandler struct {
	repository *repository.RepositoryStrategy
}

// NewPluginHandler creates a new instance of PluginHandler
func NewPluginHandler(repository *repository.RepositoryStrategy) BaseHandler {
	return &PluginHandler{
		repository: repository,
	}
}

// RegisterRoutes registers the plugin routes with the provided Gin engine
func (handler *PluginHandler) RegisterRoutes(engine *gin.Engine, jwt *services.JwtService) {
	engine.GET("/api/plugins", middleware.JwtMiddleware(jwt, handler.repository, ""), handler.getAllPlugins())
	engine.GET("/api/plugins/:slug", middleware.JwtMiddleware(jwt, handler.repository, ""), handler.getPlugin())
}

// getAllPlugins returns all registered plugins
func (handler *PluginHandler) getAllPlugins() gin.HandlerFunc {
	return func(context *gin.Context) {
		plugins, err := handler.repository.PluginRepository.GetAllPlugins(context.Request.Context())
		if err != nil {
			context.JSON(500, gin.H{
				"error": "Failed to retrieve plugins",
			})
			return
		}

		context.JSON(200, gin.H{
			"plugins": plugins,
			"count":   len(plugins),
		})
	}
}

// getPlugin returns a single plugin by slug
func (handler *PluginHandler) getPlugin() gin.HandlerFunc {
	return func(context *gin.Context) {
		slug := context.Param("slug")

		plugin, err := handler.repository.PluginRepository.GetPlugin(context.Request.Context(), slug)
		if err != nil {
			context.JSON(404, gin.H{
				"error": "Plugin not found",
			})
			return
		}

		context.JSON(200, plugin)
	}
}
