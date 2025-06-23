package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"actor-model-observability/internal/actor"
	"actor-model-observability/internal/config"
	"actor-model-observability/internal/database"
	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/observability"
	"actor-model-observability/internal/repository/postgres"
	"actor-model-observability/internal/router"
	"actor-model-observability/internal/service"
	"actor-model-observability/internal/traditional"
	pkgconfig "actor-model-observability/pkg/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("Starting Actor Model Observability Application", logging.Fields{
		"version": "1.0.0",
		"mode":    cfg.Server.Mode,
	})

	// Initialize database
	db, err := database.NewPostgresConnection(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", err, logging.Fields{
			"host": cfg.Database.Host,
			"port": cfg.Database.Port,
			"name": cfg.Database.DBName,
		})
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", err, logging.Fields{})
		}
	}()

	// Test database connection
	if err := db.HealthCheck(context.Background()); err != nil {
		logger.Fatal("Database health check failed", err, logging.Fields{})
	}
	logger.Info("Database connection established successfully", logging.Fields{})

	// Initialize Redis
	redisClient, err := database.NewRedisConnection(&cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", err, logging.Fields{
			"host": cfg.Redis.Host,
			"port": cfg.Redis.Port,
			"db":   cfg.Redis.DB,
		})
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis connection", err, logging.Fields{})
		}
	}()

	// Test Redis connection
	if err := redisClient.HealthCheck(context.Background()); err != nil {
		logger.Fatal("Redis health check failed", err, logging.Fields{})
	}
	logger.Info("Redis connection established successfully", logging.Fields{})

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.DB)
	driverRepo := postgres.NewDriverRepository(db.DB)
	passengerRepo := postgres.NewPassengerRepository(db.DB)
	tripRepo := postgres.NewTripRepository(db.DB)
	observabilityRepo := postgres.NewObservabilityRepository(db.DB)
	traditionalRepo := postgres.NewTraditionalRepository(db.DB)

	// Initialize actor system
	actorSystem := actor.NewActorSystem("main-system")

	// Load pkg config for metrics collector
	pkgCfg, err := pkgconfig.Load()
	if err != nil {
		logger.Fatal("Failed to load pkg config", err, logging.Fields{})
	}

	// Initialize observability collector
	metricsCollector := observability.NewMetricsCollector(
		db,
		redisClient.Client,
		pkgCfg,
	)

	// Initialize traditional monitor
	traditionalMonitor := traditional.NewTraditionalMonitor(
		db,
		redisClient.Client,
		cfg,
	)

	// Initialize services
	rideService := service.NewRideService(
		userRepo,
		driverRepo,
		passengerRepo,
		tripRepo,
		actorSystem,
		metricsCollector,
		traditionalMonitor,
		true, // useActorModel
	)

	// Set Gin mode based on server mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize router
	routerConfig := &router.RouterConfig{
		UserRepo:           userRepo,
		DriverRepo:         driverRepo,
		PassengerRepo:      passengerRepo,
		TripRepo:           tripRepo,
		ObservabilityRepo:  observabilityRepo,
		TraditionalRepo:    traditionalRepo,
		RideService:        rideService,
		ActorSystem:        actorSystem,
		TraditionalMonitor: traditionalMonitor,
		Logger:             logger,
		Config:             cfg,
	}

	ginEngine := router.SetupRouter(routerConfig)

	// Create HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        ginEngine,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start background services
	logger.Info("Starting background services", logging.Fields{})

	// Start actor system
	if err := actorSystem.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start actor system", err, logging.Fields{})
	}

	// Start metrics collector
	if err := metricsCollector.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start metrics collector", err, logging.Fields{})
	}

	// Start traditional monitor
	if err := traditionalMonitor.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start traditional monitor", err, logging.Fields{})
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", logging.Fields{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", err, logging.Fields{
				"port": cfg.Server.Port,
			})
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...", logging.Fields{})

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err, logging.Fields{})
	} else {
		logger.Info("HTTP server shutdown completed", logging.Fields{})
	}

	// Stop background services
	logger.Info("Stopping background services", logging.Fields{})

	// Stop traditional monitor
	if err := traditionalMonitor.Stop(); err != nil {
		logger.Error("Failed to stop traditional monitor", err, logging.Fields{})
	} else {
		logger.Info("Traditional monitor stopped", logging.Fields{})
	}

	// Stop metrics collector
	if err := metricsCollector.Stop(); err != nil {
		logger.Error("Failed to stop metrics collector", err, logging.Fields{})
	} else {
		logger.Info("Metrics collector stopped", logging.Fields{})
	}

	// Stop actor system
	if err := actorSystem.Stop(); err != nil {
		logger.Error("Failed to stop actor system", err, logging.Fields{})
	} else {
		logger.Info("Actor system stopped", logging.Fields{})
	}

	logger.Info("Application shutdown completed", logging.Fields{})
}