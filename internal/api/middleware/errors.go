package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleError logs the full error server-side and returns a sanitized response.
// Use this for all 500-level errors to prevent information disclosure.
func HandleError(c *gin.Context, status int, errCode string, err error) {
	log.Printf("[ERROR] %s: %v", errCode, err)
	c.JSON(status, gin.H{
		"error":   errCode,
		"message": "An internal error occurred",
	})
}

// RecoveryMiddleware recovers from panics and returns a generic 500 response
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err any) {
		log.Printf("[PANIC] %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "An internal error occurred",
		})
	})
}
