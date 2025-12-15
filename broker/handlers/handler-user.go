package handlers

import (
	"hotelhub/broker/domain"
	"hotelhub/broker/middleware"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"
	"net/mail"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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

// RegisterRoutes registers the user-related routes with the provided Gin engine.
// It sets up the routes for user registration and password update.
func (handler *UserHandler) RegisterRoutes(engine *gin.Engine, jwtService *services.JwtService) {
	engine.POST("/user", middleware.JwtMiddleware(jwtService, handler.repository, "admin", "access"), handler.register())
	engine.PUT("/user", middleware.JwtMiddleware(jwtService, handler.repository, "user", "reset"), handler.updatePassword())
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

// updatePassword allows a user to update their password.
// It expects a JSON payload with the new password.
// The function validates the presence of the password,
// hashes it using bcrypt, and updates the user's password in the repository.
// The token used for authentication is then invalidated.
// If successful, it returns a success message with a 200 status code.
func (handler *UserHandler) updatePassword() gin.HandlerFunc {
	type PasswordUpdateRequest struct {
		Password string `json:"password"`
	}

	return func(context *gin.Context) {
		var request PasswordUpdateRequest
		err := context.ShouldBindJSON(&request)
		if err != nil {
			context.JSON(400, gin.H{"error": "Invalid request payload"})
			return
		}

		if request.Password == "" {
			context.JSON(400, gin.H{"error": "Password cannot be empty"})
			return
		}

		subject := context.GetString("user_id")
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			context.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}

		updatedUser := handler.repository.UserRepository.Update(subject, string(passwordHash))
		if updatedUser == nil {
			context.JSON(500, gin.H{"error": "Failed to update password"})
			return
		}

		token := context.GetHeader("Authorization")[len("Bearer "):]
		handler.repository.JwtRepository.Delete(token)
		context.JSON(200, gin.H{"message": "Password updated successfully"})
	}
}
