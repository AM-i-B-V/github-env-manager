package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Logger returns a gin.HandlerFunc for logging HTTP requests
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.WithFields(logrus.Fields{
			"client_ip":   param.ClientIP,
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"method":      param.Method,
			"path":        param.Path,
			"protocol":    param.Request.Proto,
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"user_agent":  param.Request.UserAgent(),
			"error":       param.ErrorMessage,
		}).Info("HTTP Request")
		return ""
	})
}

// CORS returns a gin.HandlerFunc for handling CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Session-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID returns a gin.HandlerFunc for adding request IDs
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// AuthMiddleware returns a gin.HandlerFunc for authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// TODO: Implement proper session validation
		// For now, we'll just check if the session ID exists
		c.Set("session_id", sessionID)
		c.Next()
	}
}
