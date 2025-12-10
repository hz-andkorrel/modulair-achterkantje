package handlers

import (
	"hotelhub/broker/domain"
	"hotelhub/broker/middleware"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"
	"net/mail"

	"github.com/gin-gonic/gin"
)

// UserHandler is responsible for handling authentication-related routes.
// It requires access to the repository for user data.
type UserHandler struct {
	repository *repository.RepositoryStrategy
}

// NewUserHandler creates a new instance of AuthHandler with the provided repository.
func NewUserHandler(repository *repository.RepositoryStrategy) BaseHandler {
	return &UserHandler{
		repository: repository,
	}
}

// RegisterRoutes registers the authentication routes with the provided Gin engine.
// The auth endpoint supports login (POST) and logout (DELETE) operations.
func (handler *UserHandler) RegisterRoutes(engine *gin.Engine, jwtService *services.JwtService) {
	engine.POST("/user", middleware.JwtMiddleware(jwtService, handler.repository, "admin"), handler.register())
}

// A user should be created upon registration by an administrator.
// During registration, the users email, name and role are set.
// The password should be set using the password reset functionality.
// The function validates the existence and format of the email,
// the existence of the name, the existence and validity of the role (either "admin" or "user"),
// and whether the user was correctly created in the repository.
// If everything goes well, it returns the created user's email with a 201 status code.
func (handler *UserHandler) register() gin.HandlerFunc {
	return func(context *gin.Context) {
		var user *domain.User

		err := context.ShouldBindJSON(&user)
		if err != nil {
			context.JSON(400, gin.H{"error": "Invalid request payload"})
			return
		}

		if user.Email == "" || user.Name == "" || user.Role == "" {
			context.JSON(400, gin.H{"error": "Missing required fields"})
			return
		}

		_, err = mail.ParseAddress(user.Email)
		if err != nil {
			context.JSON(400, gin.H{"error": "Invalid email address"})
			return
		}

		if user.Role != "admin" && user.Role != "user" {
			context.JSON(400, gin.H{"error": "Invalid role"})
			return
		}

		createdUser := handler.repository.UserRepository.Create(user)
		if createdUser == nil {
			context.JSON(500, gin.H{"error": "Failed to create user"})
			return
		}

		context.JSON(201, createdUser.Email)
	}
}
