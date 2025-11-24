package handlers

import "github.com/gin-gonic/gin"

// BaseHandler defines a common interface for all handlers to implement.
// The constructor should return BaseHandler and setup anything required for the handler.
// The GetHandler method returns the actual Gin handler function.
type BaseHandler interface {
	GetHandler() gin.HandlerFunc
}
