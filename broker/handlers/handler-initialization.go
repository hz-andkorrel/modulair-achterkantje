package handlers

import (
	"hotelhub/broker/domain"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// InitializationHandler is used to handle a status check, implementing the BaseHandler interface.
// It requires a RepositoryStrategy to access the data layer if needed.
type InitializationHandler struct {
	repository *repository.RepositoryStrategy
}

// NewInitializationHandler creates a new instance of StatusHandler.
// It takes a RepositoryStrategy as a parameter to access the data layer if needed.
func NewInitializationHandler(repository *repository.RepositoryStrategy) BaseHandler {
	return &InitializationHandler{
		repository: repository,
	}
}

// RegisterRoutes registers the status check route with the provided Gin engine.
func (handler *InitializationHandler) RegisterRoutes(engine *gin.Engine, jwt *services.JwtService) {
	engine.POST("/initialize", handler.Initialize(jwt))
}

// The initialize method handles the system initialization process.
// It checks if the system is already initialized, and will fail if so.
// If not initialized, it creates an admin user and sets the necessary metadata.
func (handler *InitializationHandler) Initialize(jwtService *services.JwtService) gin.HandlerFunc {
	type InitializationRequest struct {
		HotelName     string `json:"hotel_name" binding:"required"`
		AdminEmail    string `json:"admin_email" binding:"required,email"`
		AdminPassword string `json:"admin_password" binding:"required,min=8"`
		AdminName     string `json:"admin_name" binding:"required"`
	}

	return func(context *gin.Context) {
		initialized := handler.repository.MetaRepository.Get("initialized")
		if initialized != "" {
			context.JSON(400, gin.H{"error": "System is already initialized"})
			return
		}

		var request InitializationRequest
		if err := context.ShouldBindJSON(&request); err != nil {
			context.JSON(400, gin.H{"error": err.Error()})
			return
		}

		password, err := bcrypt.GenerateFromPassword([]byte(request.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			context.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}

		user := domain.User{
			Email:        request.AdminEmail,
			PasswordHash: string(password),
			Name:         request.AdminName,
			Role:         "admin",
			Enabled:      true,
		}

		createdUser := handler.repository.UserRepository.Create(&user)
		createdUser = handler.repository.UserRepository.Update(user.Email, user.PasswordHash)
		if createdUser == nil {
			context.JSON(500, gin.H{"error": "Failed to create admin user"})
			return
		}

		handler.repository.MetaRepository.Set("hotel", request.HotelName)
		handler.repository.MetaRepository.Set("initialized", "true")
		token, time := jwtService.GenerateToken(user.Email)
		handler.repository.JwtRepository.Add(token, user.Email, time)

		context.JSON(200, gin.H{
			"token":      token,
			"role":       user.Role,
			"expires_at": time,
		})
	}
}
