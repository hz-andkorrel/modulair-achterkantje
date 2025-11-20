package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"broker/internal/jwt"
	"broker/internal/middleware"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// create jwt manager for tests
	jm, _ := jwt.NewManager(10*60, "test-issuer")

	v1 := router.Group("/api/v1")
	{
		v1.GET("/status", middleware.OptionalAuth(jm, nil), GetStatus)
	}
	return router
}

func TestStatus_NoAuth(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}

func TestStatus_InvalidToken(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/status", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for invalid token (optional auth), got %d", w.Code)
	}
}

func TestStatus_WithValidToken(t *testing.T) {
	// create a jwt manager with short ttl for signing
	jm, _ := jwt.NewManager(10*60, "test-issuer")
	token, _, err := jm.GenerateToken("admin@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// setup router using the same manager so token verifies
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/status", middleware.OptionalAuth(jm, nil), GetStatus)
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK with valid token, got %d", w.Code)
	}
}
