package handlers

import (
	"hotelhub/broker/repository"
	"hotelhub/broker/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler is responsible for handling authentication-related routes.
// It requires access to the repository for user data.
type AuthHandler struct {
	repository *repository.RepositoryStrategy
}

// NewAuthHandler creates a new instance of AuthHandler with the provided repository.
func NewAuthHandler(repository *repository.RepositoryStrategy) BaseHandler {
	return &AuthHandler{
		repository: repository,
	}
}

// RegisterRoutes registers the authentication routes with the provided Gin engine.
// The auth endpoint supports login (POST) and logout (DELETE) operations.
func (handler *AuthHandler) RegisterRoutes(engine *gin.Engine, jwtService *services.JwtService) {
	engine.POST("/auth", handler.login(jwtService))
	engine.DELETE("/auth", handler.logout())
}

// The login handler processes user login requests.
// It should validate user credentials and issue a JWT token upon successful authentication.
func (handler *AuthHandler) login(jwtService *services.JwtService) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Retrieve and validate user credentials from the request
		username := context.PostForm("username")
		password := context.PostForm("password")

		if username == "" || password == "" {
			context.JSON(400, gin.H{"error": "Username and password are required"})
			return
		}

		user := handler.repository.UserRepository.Get(username)
		if user == nil || !user.ValidatePassword(password) {
			context.JSON(401, gin.H{"error": "Invalid credentials"})
			return
		}

		// If valid, generate a JWT token using jwtService
		token, time := jwtService.GenerateToken(user.Id)

		// Return the token in the response
		context.JSON(200, gin.H{
			"token":      token,
			"expires_at": time,
		})
	}
}

// The logout handler processes user logout requests.
// It should invalidate the user's session or token as needed.
func (handler *AuthHandler) logout() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Invalidate the user's token

		// Return a success response
		context.JSON(418, gin.H{"message": "Logout successful"})
	}
}
