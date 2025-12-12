package handlers

import (
	"hotelhub/broker/middleware"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// EventLogHandler handles requests for event logs
type EventLogHandler struct {
	repository *repository.RepositoryStrategy
}

// NewEventLogHandler creates a new instance of EventLogHandler
func NewEventLogHandler(repository *repository.RepositoryStrategy) BaseHandler {
	return &EventLogHandler{
		repository: repository,
	}
}

// RegisterRoutes registers the event log routes with the provided Gin engine
func (handler *EventLogHandler) RegisterRoutes(engine *gin.Engine, jwt *services.JwtService) {
	engine.GET("/api/event-logs", middleware.JwtMiddleware(jwt, handler.repository, ""), handler.getEventLogs())
	engine.GET("/api/event-logs/count", middleware.JwtMiddleware(jwt, handler.repository, ""), handler.getEventLogsCount())
}

// getEventLogs returns recent event logs
// Query parameter: limit (default: 100)
func (handler *EventLogHandler) getEventLogs() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Get limit from query parameter, default to 100
		limitStr := context.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 100
		}

		// Maximum limit of 1000 to prevent performance issues
		if limit > 1000 {
			limit = 1000
		}

		// Fetch event logs from repository
		events, err := handler.repository.EventLogRepository.GetRecent(context.Request.Context(), limit)
		if err != nil {
			context.JSON(500, gin.H{
				"error": "Failed to retrieve event logs",
			})
			return
		}

		context.JSON(200, gin.H{
			"events": events,
			"count":  len(events),
			"limit":  limit,
		})
	}
}

// getEventLogsCount returns the total count of event logs
func (handler *EventLogHandler) getEventLogsCount() gin.HandlerFunc {
	return func(context *gin.Context) {
		count, err := handler.repository.EventLogRepository.Count(context.Request.Context())
		if err != nil {
			context.JSON(500, gin.H{
				"error": "Failed to retrieve event log count",
			})
			return
		}

		context.JSON(200, gin.H{
			"count": count,
		})
	}
}
