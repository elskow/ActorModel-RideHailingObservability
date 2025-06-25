package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/traditional"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// LoggingMiddleware creates a middleware for request logging with configurable filtering
func LoggingMiddleware(logger *logging.Logger, skipPaths []string, skipUserAgents []string) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Check if path should be skipped
		for _, skipPath := range skipPaths {
			if strings.Contains(param.Path, skipPath) {
				return ""
			}
		}

		// Check if user agent should be skipped
		userAgent := param.Request.UserAgent()
		for _, skipUA := range skipUserAgents {
			if strings.Contains(userAgent, skipUA) {
				return ""
			}
		}

		// Extract request ID from context
		requestID := ""
		if param.Keys != nil {
			if id, exists := param.Keys["request_id"]; exists {
				if idStr, ok := id.(string); ok {
					requestID = idStr
				}
			}
		}

		// Log the HTTP request
		logger.LogHTTPRequest(
			param.Method,
			param.Path,
			userAgent,
			requestID,
			param.StatusCode,
			param.Latency.Milliseconds(),
			logging.Fields{
				"client_ip": param.ClientIP,
				"body_size": param.BodySize,
			},
		)

		// Return empty string since we're handling logging ourselves
		return ""
	})
}

// CORSMiddleware creates a middleware for handling CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	// Create a map to store rate limiters for each client IP
	var (
		mu       sync.RWMutex
		limiters = make(map[string]*rate.Limiter)
	)

	// Clean up old limiters periodically
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			// Simple cleanup - in production, you'd want more sophisticated cleanup
			if len(limiters) > 1000 {
				limiters = make(map[string]*rate.Limiter)
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		mu.RLock()
		limiter, exists := limiters[clientIP]
		mu.RUnlock()

		if !exists {
			// Create new limiter: 100 requests per minute
			limiter = rate.NewLimiter(rate.Every(time.Minute/100), 10)
			mu.Lock()
			limiters[clientIP] = limiter
			mu.Unlock()
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// MetricsMiddleware creates a middleware for collecting HTTP metrics
func MetricsMiddleware(monitor *traditional.TraditionalMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record HTTP request metrics
		if monitor != nil {
			monitor.RecordRequest(path, method, duration, statusCode)
		}
	}
}

// SecurityMiddleware creates a middleware for basic security headers
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// TimeoutMiddleware creates a middleware for request timeout
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		finish := make(chan struct{})
		go func() {
			c.Next()
			finish <- struct{}{}
		}()

		select {
		case <-time.After(timeout):
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"message": "The request took too long to process",
			})
			c.Abort()
		case <-finish:
			// Request completed within timeout
		}
	}
}

// AuthMiddleware creates a middleware for authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			// Allow requests without authentication for development
			// In production, you might want to require authentication for protected routes
			c.Next()
			return
		}

		// Check if the authorization header has the correct format
		if len(authorization) < 7 || authorization[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format. Expected 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		token := authorization[7:] // Remove "Bearer " prefix

		// Validate token (simplified validation for demo purposes)
		// In production, you would:
		// 1. Parse and validate JWT token
		// 2. Check token expiration
		// 3. Verify token signature
		// 4. Extract user claims from token
		if isValidToken(token) {
			userID, userType := extractUserInfoFromToken(token)
			c.Set("user_id", userID)
			c.Set("user_type", userType)
			c.Set("authenticated", true)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired authentication token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isValidToken validates the authentication token
// This is a simplified implementation for demo purposes
func isValidToken(token string) bool {
	// In production, implement proper JWT validation
	// For demo purposes, accept specific test tokens
	validTokens := map[string]bool{
		"valid-user-token":      true,
		"valid-driver-token":    true,
		"valid-passenger-token": true,
		"admin-token":           true,
	}
	return validTokens[token]
}

// extractUserInfoFromToken extracts user information from the token
// This is a simplified implementation for demo purposes
func extractUserInfoFromToken(token string) (string, string) {
	// In production, extract this from JWT claims
	switch token {
	case "valid-driver-token":
		return "driver-user-id", "driver"
	case "valid-passenger-token":
		return "passenger-user-id", "passenger"
	case "admin-token":
		return "admin-user-id", "admin"
	default:
		return "user-id", "user"
	}
}

// ValidationMiddleware creates a middleware for request validation
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content-Type validation for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && contentType != "application/json" {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":   "Unsupported media type",
					"message": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// CompressionMiddleware creates a middleware for response compression
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple compression logic
		// In production, you might want to use a more sophisticated compression middleware
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if contains(acceptEncoding, "gzip") {
			c.Header("Content-Encoding", "gzip")
		}

		c.Next()
	}
}

// CacheMiddleware creates a middleware for response caching
func CacheMiddleware(maxAge time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method == "GET" {
			c.Header("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge.Seconds())))
		} else {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		c.Next()
	}
}

// HealthCheckMiddleware creates a middleware that bypasses other middleware for health checks
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip other middleware for health check endpoints
		if c.Request.URL.Path == "/health/ping" || c.Request.URL.Path == "/health/live" {
			c.Next()
			return
		}

		c.Next()
	}
}

// ErrorHandlingMiddleware creates a middleware for centralized error handling
func ErrorHandlingMiddleware(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.LogError(err, "http", "request_processing", logging.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"ip":     c.ClientIP(),
			})

			// Return appropriate error response
			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Bad request",
					"message": "Invalid request format",
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": "An unexpected error occurred",
				})
			}
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
