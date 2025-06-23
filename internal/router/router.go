package router

import (
	"context"
	"net/http"
	"time"

	"actor-model-observability/internal/actor"
	"actor-model-observability/internal/config"
	"actor-model-observability/internal/handlers"
	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/middleware"
	"actor-model-observability/internal/repository"
	"actor-model-observability/internal/service"
	"actor-model-observability/internal/traditional"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RouterConfig holds dependencies for router setup
type RouterConfig struct {
	Config              *config.Config
	Logger              *logging.Logger
	UserRepo            repository.UserRepository
	DriverRepo          repository.DriverRepository
	PassengerRepo       repository.PassengerRepository
	TripRepo            repository.TripRepository
	ObservabilityRepo   repository.ObservabilityRepository
	TraditionalRepo     repository.TraditionalRepository
	ActorSystem         *actor.ActorSystem
	TraditionalMonitor  *traditional.TraditionalMonitor
	RideService         *service.RideService
}

// SetupRouter configures and returns the Gin router with all routes and middleware
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Config.Server.Mode)

	// Create router
	router := gin.New()

	// Add middleware
	setupMiddleware(router, cfg)

	// Setup routes
	setupRoutes(router, cfg)

	return router
}

// setupMiddleware configures all middleware
func setupMiddleware(router *gin.Engine, cfg *RouterConfig) {
	// Recovery middleware
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		cfg.Logger.LogPanic(recovered, "http", "request_handling", logging.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		})
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "An unexpected error occurred",
		})
	}))

	// Request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	// Logging middleware
	router.Use(middleware.LoggingMiddleware(cfg.Logger))

	// CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Rate limiting middleware (if enabled)
	if cfg.Config.Server.Mode == "release" {
		router.Use(middleware.RateLimitMiddleware())
	}

	// Metrics middleware
	router.Use(middleware.MetricsMiddleware(cfg.TraditionalMonitor))
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine, cfg *RouterConfig) {
	// Create handlers
	userHandler := handlers.NewUserHandler(
		cfg.UserRepo,
		cfg.DriverRepo,
		cfg.PassengerRepo,
	)

	rideHandler := handlers.NewRideHandler(
		cfg.RideService,
	)

	observabilityHandler := handlers.NewObservabilityHandler(
		cfg.ObservabilityRepo,
		cfg.TraditionalRepo,
	)

	// Health check endpoints
	setupHealthRoutes(router, cfg)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// User management routes
		userRoutes := v1.Group("/users")
		{
			userRoutes.POST("", userHandler.CreateUser)
			userRoutes.GET("/:id", userHandler.GetUser)
			userRoutes.PUT("/:id", userHandler.UpdateUser)
			userRoutes.GET("", userHandler.ListUsers)
		}

		// Driver-specific routes
		driverRoutes := v1.Group("/drivers")
		{
			driverRoutes.POST("", userHandler.CreateDriver)
			driverRoutes.PUT("/:id/location", userHandler.UpdateDriverLocation)
			driverRoutes.PUT("/:id/status", userHandler.UpdateDriverStatus)
			// TODO: Implement GetOnlineDrivers method
			// driverRoutes.GET("/online", userHandler.GetOnlineDrivers)
		}

		// Passenger-specific routes
		// TODO: Implement CreatePassenger method
		// passengerRoutes := v1.Group("/passengers")
		// {
		//	passengerRoutes.POST("", userHandler.CreatePassenger)
		// }

		// Ride management routes
		rideRoutes := v1.Group("/rides")
		{
			rideRoutes.POST("/request", rideHandler.RequestRide)
			rideRoutes.POST("/:id/cancel", rideHandler.CancelRide)
			rideRoutes.GET("/:id/status", rideHandler.GetRideStatus)
			rideRoutes.GET("", rideHandler.ListRides)
		}

		// Observability routes (Actor model)
		observabilityRoutes := v1.Group("/observability")
		{
			actorRoutes := observabilityRoutes.Group("/actors")
			{
				actorRoutes.GET("", observabilityHandler.GetActorInstances)
			}

			messageRoutes := observabilityRoutes.Group("/messages")
			{
				messageRoutes.GET("", observabilityHandler.GetActorMessages)
			}

			metricsRoutes := observabilityRoutes.Group("/metrics")
			{
				metricsRoutes.GET("", observabilityHandler.GetSystemMetrics)
			}

			traceRoutes := observabilityRoutes.Group("/traces")
			{
				traceRoutes.GET("", observabilityHandler.GetDistributedTraces)
			}

			eventRoutes := observabilityRoutes.Group("/events")
			{
				eventRoutes.GET("", observabilityHandler.GetEventLogs)
			}
		}

		// Traditional monitoring routes
		traditionalRoutes := v1.Group("/traditional")
		{
			traditionalMetricsRoutes := traditionalRoutes.Group("/metrics")
			{
				traditionalMetricsRoutes.GET("", observabilityHandler.GetTraditionalMetrics)
			}

			traditionalLogsRoutes := traditionalRoutes.Group("/logs")
			{
				traditionalLogsRoutes.GET("", observabilityHandler.GetTraditionalLogs)
			}

			healthRoutes := traditionalRoutes.Group("/health")
			{
				healthRoutes.GET("", observabilityHandler.GetServiceHealth)
				healthRoutes.GET("/:service_name/latest", observabilityHandler.GetLatestServiceHealth)
			}
		}

		// System information routes
		systemRoutes := v1.Group("/system")
		{
			systemRoutes.GET("/info", getSystemInfo(cfg))
			systemRoutes.GET("/stats", getSystemStats(cfg))
		}
	}

	// Admin routes (if needed)
	setupAdminRoutes(router, cfg)

	// Swagger documentation (in development mode)
	if cfg.Config.Server.Mode == "debug" {
		setupSwaggerRoutes(router)
	}
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(router *gin.Engine, cfg *RouterConfig) {
	health := router.Group("/health")
	{
		// Basic health check
		health.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().UTC(),
				"service":   "actor-model-observability",
			})
		})

		// Detailed health check
		health.GET("/ready", func(c *gin.Context) {
			ctx := c.Request.Context()
			status := gin.H{
				"status":    "ok",
				"timestamp": time.Now().UTC(),
				"service":   "actor-model-observability",
				"checks":    gin.H{},
			}

			checks := status["checks"].(gin.H)
			allHealthy := true

			// Check actor system
			if cfg.ActorSystem != nil {
				// TODO: Implement IsRunning method or use alternative status check
				checks["actor_system"] = "ok" // Assume running if not nil
			}

			// Check traditional monitor
			if cfg.TraditionalMonitor != nil {
				// TODO: Implement IsRunning method or use alternative status check
				checks["traditional_monitor"] = "ok" // Assume running if not nil
			}

			// Check database connectivity (if repositories implement health check)
			if healthChecker, ok := cfg.UserRepo.(interface{ HealthCheck(ctx context.Context) error }); ok {
				if err := healthChecker.HealthCheck(ctx); err != nil {
					checks["database"] = "error"
					allHealthy = false
				} else {
					checks["database"] = "ok"
				}
			}

			if !allHealthy {
				status["status"] = "error"
				c.JSON(http.StatusServiceUnavailable, status)
			} else {
				c.JSON(http.StatusOK, status)
			}
		})

		// Liveness probe
		health.GET("/live", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "alive",
				"timestamp": time.Now().UTC(),
			})
		})
	}
}

// setupAdminRoutes configures admin endpoints
func setupAdminRoutes(router *gin.Engine, cfg *RouterConfig) {
	admin := router.Group("/admin")
	{
		// Actor system management
		actorAdmin := admin.Group("/actors")
		{
			actorAdmin.POST("/start", func(c *gin.Context) {
				if cfg.ActorSystem != nil {
					if err := cfg.ActorSystem.Start(c.Request.Context()); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Failed to start actor system",
						})
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"status": "started"})
			})

			actorAdmin.POST("/stop", func(c *gin.Context) {
				if cfg.ActorSystem != nil {
					cfg.ActorSystem.Stop()
				}
				c.JSON(http.StatusOK, gin.H{"status": "stopped"})
			})

			actorAdmin.GET("/stats", func(c *gin.Context) {
				if cfg.ActorSystem == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error": "Actor system not available",
					})
					return
				}

				stats := cfg.ActorSystem.GetMetrics()
				c.JSON(http.StatusOK, stats)
			})
		}

		// Traditional monitor management
		traditionalAdmin := admin.Group("/traditional")
		{
			traditionalAdmin.POST("/start", func(c *gin.Context) {
				if cfg.TraditionalMonitor != nil {
					if err := cfg.TraditionalMonitor.Start(c.Request.Context()); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Failed to start traditional monitor",
						})
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"status": "started"})
			})

			traditionalAdmin.POST("/stop", func(c *gin.Context) {
				if cfg.TraditionalMonitor != nil {
					cfg.TraditionalMonitor.Stop()
				}
				c.JSON(http.StatusOK, gin.H{"status": "stopped"})
			})
		}
	}
}

// setupSwaggerRoutes configures Swagger documentation routes
func setupSwaggerRoutes(router *gin.Engine) {
	// This would typically use swaggo/gin-swagger
	// For now, just a placeholder
	router.GET("/swagger/*any", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Swagger documentation would be available here",
			"note":    "Install swaggo/gin-swagger for full documentation",
		})
	})
}

// getSystemInfo returns system information
func getSystemInfo(cfg *RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := gin.H{
			"service":   "actor-model-observability",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
			"config": gin.H{
				"server_mode": cfg.Config.Server.Mode,
				"log_level":   cfg.Config.Logging.Level,
			},
		}

		if cfg.ActorSystem != nil {
			info["actor_system"] = gin.H{
				"running":     true, // TODO: Implement proper status check
				"max_actors":  cfg.Config.Actor.MaxActors,
				"supervision": cfg.Config.Actor.SupervisionStrategy,
			}
		}

		if cfg.TraditionalMonitor != nil {
			info["traditional_monitor"] = gin.H{
				"running": true, // TODO: Implement proper status check
			}
		}

		c.JSON(http.StatusOK, info)
	}
}

// getSystemStats returns system statistics
func getSystemStats(cfg *RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := gin.H{
			"timestamp": time.Now().UTC(),
		}

		if cfg.ActorSystem != nil {
			stats["actor_system"] = cfg.ActorSystem.GetMetrics()
		}

		// Add more system statistics as needed
		c.JSON(http.StatusOK, stats)
	}
}