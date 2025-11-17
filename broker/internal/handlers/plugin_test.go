package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"
    "time"

    "broker/internal/jwt"
    "broker/internal/middleware"
    "broker/internal/plugins"

    "github.com/gin-gonic/gin"
)

func setupPluginRouter(jm *jwt.Manager) *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    v1 := router.Group("/api/v1")
    {
        v1.POST("/route", middleware.RequireAuth(jm), RegisterPlugin)
        v1.GET("/routes", ListPlugins)
        v1.PUT("/route/:slug", middleware.RequireAuth(jm), UpdatePlugin)
        v1.DELETE("/route/:slug", middleware.RequireAuth(jm), DeletePlugin)
    }
    return router
}

func makeTempPersistPath(t *testing.T) string {
    dir := t.TempDir()
    return filepath.Join(dir, "plugins.json")
}

func TestRegisterPlugin_Unauthorized(t *testing.T) {
    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)

    payload := map[string]interface{}{
        "slug": "test-plugin",
        "name": "Test Plugin",
    }
    b, _ := json.Marshal(payload)

    req, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b))
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    if w.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401 Unauthorized when no token provided, got %d", w.Code)
    }
}

func TestRegisterPlugin_SuccessAndPersistence(t *testing.T) {
    // Reset global registry and set temp persist path
    plugins.Global = plugins.NewRegistry()
    persistPath := makeTempPersistPath(t)
    if err := plugins.Global.SetPersistPath(persistPath); err != nil {
        t.Fatalf("failed to set persist path: %v", err)
    }

    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)

    payload := map[string]interface{}{
        "description": "Toont een welkomstscherm",
        "version": "1.0.2",
        "slug": "kiosk",
        "name": "Kiosk Plug-in",
        "host": "http://localhost:8080",
        "base-api-route": "/kiosk",
        "settings-route": "/kiosk/settings",
        "api-routes": []string{"/status", "/reset", "/welcome"},
        "enabled": true,
    }
    b, _ := json.Marshal(payload)

    token, _, err := jm.GenerateToken("admin@example.com")
    if err != nil {
        t.Fatalf("failed to create token: %v", err)
    }

    req, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201 Created, got %d; body=%s", w.Code, w.Body.String())
    }

    // Check persisted file exists
    if _, err := os.Stat(persistPath); err != nil {
        t.Fatalf("expected persistence file at %s, got error: %v", persistPath, err)
    }

    // Ensure plugin is listed
    req2, _ := http.NewRequest(http.MethodGet, "/api/v1/routes", nil)
    w2 := httptest.NewRecorder()
    router.ServeHTTP(w2, req2)
    if w2.Code != http.StatusOK {
        t.Fatalf("expected 200 OK from list, got %d", w2.Code)
    }
}

func TestRegisterPlugin_DuplicateSlug(t *testing.T) {
    // Reset global registry and set temp persist path
    plugins.Global = plugins.NewRegistry()
    persistPath := makeTempPersistPath(t)
    if err := plugins.Global.SetPersistPath(persistPath); err != nil {
        t.Fatalf("failed to set persist path: %v", err)
    }

    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)

    payload := map[string]interface{}{
        "slug": "dup-plugin",
        "name": "Dup Plugin",
        "host": "http://localhost:8080",
    }
    b, _ := json.Marshal(payload)

    token, _, err := jm.GenerateToken("admin@example.com")
    if err != nil {
        t.Fatalf("failed to create token: %v", err)
    }

    // First request
    req1, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b))
    req1.Header.Set("Authorization", "Bearer "+token)
    req1.Header.Set("Content-Type", "application/json")
    w1 := httptest.NewRecorder()
    router.ServeHTTP(w1, req1)
    if w1.Code != http.StatusCreated {
        t.Fatalf("expected 201 Created for first registration, got %d; body=%s", w1.Code, w1.Body.String())
    }

    // Second request with same slug
    req2, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b))
    req2.Header.Set("Authorization", "Bearer "+token)
    req2.Header.Set("Content-Type", "application/json")
    w2 := httptest.NewRecorder()
    router.ServeHTTP(w2, req2)
    if w2.Code != http.StatusConflict {
        t.Fatalf("expected 409 Conflict for duplicate slug, got %d; body=%s", w2.Code, w2.Body.String())
    }
}

func TestUpdatePlugin_Success(t *testing.T) {
    plugins.Global = plugins.NewRegistry()
    persistPath := makeTempPersistPath(t)
    if err := plugins.Global.SetPersistPath(persistPath); err != nil {
        t.Fatalf("failed to set persist path: %v", err)
    }

    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)
    token, _, _ := jm.GenerateToken("admin@example.com")

    // First, register a plugin
    payload := map[string]interface{}{
        "slug":            "test-plugin",
        "name":            "Test Plugin",
        "host":            "http://localhost:8080",
        "version":         "1.0.0",
        "base-api-route":  "/test",
    }
    b, _ := json.Marshal(payload)
    req1, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b))
    req1.Header.Set("Authorization", "Bearer "+token)
    req1.Header.Set("Content-Type", "application/json")
    w1 := httptest.NewRecorder()
    router.ServeHTTP(w1, req1)
    if w1.Code != http.StatusCreated {
        t.Fatalf("setup: expected 201, got %d", w1.Code)
    }

    // Now update it
    updatePayload := map[string]interface{}{
        "slug":            "test-plugin",
        "name":            "Updated Test Plugin",
        "host":            "http://localhost:8080",
        "version":         "2.0.0",
        "base-api-route":  "/test",
    }
    b2, _ := json.Marshal(updatePayload)
    req2, _ := http.NewRequest(http.MethodPut, "/api/v1/route/test-plugin", bytes.NewReader(b2))
    req2.Header.Set("Authorization", "Bearer "+token)
    req2.Header.Set("Content-Type", "application/json")
    w2 := httptest.NewRecorder()
    router.ServeHTTP(w2, req2)

    if w2.Code != http.StatusOK {
        t.Fatalf("expected 200 OK, got %d; body=%s", w2.Code, w2.Body.String())
    }
}

func TestUpdatePlugin_NotFound(t *testing.T) {
    plugins.Global = plugins.NewRegistry()
    persistPath := makeTempPersistPath(t)
    if err := plugins.Global.SetPersistPath(persistPath); err != nil {
        t.Fatalf("failed to set persist path: %v", err)
    }

    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)
    token, _, _ := jm.GenerateToken("admin@example.com")

    payload := map[string]interface{}{
        "slug": "nonexistent",
        "name": "Nonexistent",
        "host": "http://localhost:8080",
    }
    b, _ := json.Marshal(payload)
    req, _ := http.NewRequest(http.MethodPut, "/api/v1/route/nonexistent", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    if w.Code != http.StatusNotFound {
        t.Fatalf("expected 404 Not Found, got %d; body=%s", w.Code, w.Body.String())
    }
}

func TestDeletePlugin_Success(t *testing.T) {
    plugins.Global = plugins.NewRegistry()
    persistPath := makeTempPersistPath(t)
    if err := plugins.Global.SetPersistPath(persistPath); err != nil {
        t.Fatalf("failed to set persist path: %v", err)
    }

    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)
    token, _, _ := jm.GenerateToken("admin@example.com")

    // Register a plugin first
    payload := map[string]interface{}{
        "slug": "delete-me",
        "name": "Delete Me",
        "host": "http://localhost:8080",
    }
    b, _ := json.Marshal(payload)
    req1, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b))
    req1.Header.Set("Authorization", "Bearer "+token)
    req1.Header.Set("Content-Type", "application/json")
    w1 := httptest.NewRecorder()
    router.ServeHTTP(w1, req1)
    if w1.Code != http.StatusCreated {
        t.Fatalf("setup: expected 201, got %d", w1.Code)
    }

    // Delete it
    req2, _ := http.NewRequest(http.MethodDelete, "/api/v1/route/delete-me", nil)
    req2.Header.Set("Authorization", "Bearer "+token)
    w2 := httptest.NewRecorder()
    router.ServeHTTP(w2, req2)

    if w2.Code != http.StatusNoContent {
        t.Fatalf("expected 204 No Content, got %d; body=%s", w2.Code, w2.Body.String())
    }

    // Verify it's gone
    if plugins.Global.Get("delete-me") != nil {
        t.Fatal("plugin should have been deleted")
    }
}

func TestDeletePlugin_NotFound(t *testing.T) {
    plugins.Global = plugins.NewRegistry()
    persistPath := makeTempPersistPath(t)
    if err := plugins.Global.SetPersistPath(persistPath); err != nil {
        t.Fatalf("failed to set persist path: %v", err)
    }

    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)
    token, _, _ := jm.GenerateToken("admin@example.com")

    req, _ := http.NewRequest(http.MethodDelete, "/api/v1/route/nonexistent", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    if w.Code != http.StatusNotFound {
        t.Fatalf("expected 404 Not Found, got %d; body=%s", w.Code, w.Body.String())
    }
}

func TestRegisterPlugin_RouteConflict(t *testing.T) {
    plugins.Global = plugins.NewRegistry()
    persistPath := makeTempPersistPath(t)
    if err := plugins.Global.SetPersistPath(persistPath); err != nil {
        t.Fatalf("failed to set persist path: %v", err)
    }

    jm, _ := jwt.NewManager(1*time.Minute, "test")
    router := setupPluginRouter(jm)
    token, _, _ := jm.GenerateToken("admin@example.com")

    // Register first plugin
    payload1 := map[string]interface{}{
        "slug":           "plugin1",
        "name":           "Plugin 1",
        "host":           "http://localhost:8080",
        "base-api-route": "/custom",
    }
    b1, _ := json.Marshal(payload1)
    req1, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b1))
    req1.Header.Set("Authorization", "Bearer "+token)
    req1.Header.Set("Content-Type", "application/json")
    w1 := httptest.NewRecorder()
    router.ServeHTTP(w1, req1)
    if w1.Code != http.StatusCreated {
        t.Fatalf("setup: expected 201, got %d", w1.Code)
    }

    // Try to register second plugin with same base-api-route
    payload2 := map[string]interface{}{
        "slug":           "plugin2",
        "name":           "Plugin 2",
        "host":           "http://localhost:9090",
        "base-api-route": "/custom",
    }
    b2, _ := json.Marshal(payload2)
    req2, _ := http.NewRequest(http.MethodPost, "/api/v1/route", bytes.NewReader(b2))
    req2.Header.Set("Authorization", "Bearer "+token)
    req2.Header.Set("Content-Type", "application/json")
    w2 := httptest.NewRecorder()
    router.ServeHTTP(w2, req2)

    if w2.Code != http.StatusConflict {
        t.Fatalf("expected 409 Conflict for duplicate route, got %d; body=%s", w2.Code, w2.Body.String())
    }
}
