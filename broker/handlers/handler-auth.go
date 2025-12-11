package handlers

import (
	"hotelhub/broker/middleware"
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
	engine.PUT("/auth", handler.resetPassword())
	engine.DELETE("/auth", middleware.JwtMiddleware(jwtService, handler.repository, "user", "access"), handler.logout())
}

// The login handler processes user login requests.
// It should validate user credentials and issue a JWT token upon successful authentication.
func (handler *AuthHandler) login(jwtService *services.JwtService) gin.HandlerFunc {
	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	return func(context *gin.Context) {
		var request LoginRequest
		err := context.ShouldBindJSON(&request)
		if err != nil || request.Username == "" || request.Password == "" {
			context.JSON(400, gin.H{"error": "Username and password are required"})
			return
		}

		user := handler.repository.UserRepository.Get(request.Username)
		if user == nil || !user.ValidatePassword(request.Password) {
			context.JSON(401, gin.H{"error": "Invalid credentials"})
			return
		}

		token, time := jwtService.GenerateToken(user.Email)
		handler.repository.JwtRepository.Add(token, user.Email, time)
		context.JSON(200, gin.H{
			"token":      token,
			"role":       user.Role,
			"expires_at": time,
		})
	}
}

func (handler *AuthHandler) resetPassword() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Implementation for password reset would go here.
	}
}

// The logout handler processes user logout requests.
// It invalidates the provided JWT token by setting it as revoked in the repository.
// The handler returns a JSON object with a result field indicating success or failure.
func (handler *AuthHandler) logout() gin.HandlerFunc {
	return func(context *gin.Context) {
		authHeader := context.GetHeader("Authorization")
		token := authHeader[len("Bearer "):]

		status := handler.repository.JwtRepository.Delete(token)
		if !status {
			context.JSON(400, gin.H{"result": false})
			return
		}

		context.JSON(200, gin.H{"result": true})
	}
}
