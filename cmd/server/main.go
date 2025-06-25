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

	logger.WithFields(logging.Fields{
		"version": "1.0.0",
		"mode":    cfg.Server.Mode,
	}).Info("Starting Actor Model Observability Application")

	// Initialize database
	db, err := database.NewPostgresConnection(&cfg.Database, logger)
	if err != nil {
		logger.WithFields(logging.Fields{
			"host": cfg.Database.Host,
			"port": cfg.Database.Port,
			"name": cfg.Database.DBName,
		}).WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.WithError(err).Error("Failed to close database connection")
		}
	}()

	// Test database connection
	if err := db.HealthCheck(context.Background()); err != nil {
		logger.WithError(err).Fatal("Database health check failed")
	}
	logger.Info("Database connection established successfully")

	// Initialize Redis
	redisClient, err := database.NewRedisConnection(&cfg.Redis, logger)
	if err != nil {
		logger.WithFields(logging.Fields{
			"host": cfg.Redis.Host,
			"port": cfg.Redis.Port,
			"db":   cfg.Redis.DB,
		}).WithError(err).Fatal("Failed to connect to Redis")
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.WithError(err).Error("Failed to close Redis connection")
		}
	}()

	// Test Redis connection
	if err := redisClient.HealthCheck(context.Background()); err != nil {
		logger.WithError(err).Fatal("Redis health check failed")
	}
	logger.Info("Redis connection established successfully")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.DB)
	driverRepo := postgres.NewDriverRepository(db.DB)
	passengerRepo := postgres.NewPassengerRepository(db.DB)
	tripRepo := postgres.NewTripRepository(db.DB)
	observabilityRepo := postgres.NewObservabilityRepository(db.DB)
	traditionalRepo := postgres.NewTraditionalRepository(db.DB)

	// Initialize actor system
	actorSystem := actor.NewActorSystem("main-system")

	// Initialize observability collector
	metricsCollector := observability.NewMetricsCollector(
		db,
		redisClient.Client,
		cfg,
	)

	// Initialize traditional monitor
	traditionalMonitor, err := traditional.NewTraditionalMonitor(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize traditional monitor: %v", err)
	}

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
	logger.Info("Starting background services")

	// Start actor system
	if err := actorSystem.Start(context.Background()); err != nil {
		logger.WithError(err).Fatal("Failed to start actor system")
	}

	// Start metrics collector
	if err := metricsCollector.Start(context.Background()); err != nil {
		logger.WithError(err).Fatal("Failed to start metrics collector")
	}

	// Start traditional monitor
	if err := traditionalMonitor.Start(context.Background()); err != nil {
		logger.WithError(err).Fatal("Failed to start traditional monitor")
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.WithFields(logging.Fields{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		}).Info("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithFields(logging.Fields{
			"port": cfg.Server.Port,
		}).WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("HTTP server shutdown completed")
	}

	// Stop background services
	logger.Info("Stopping background services")

	// Stop traditional monitor
	if err := traditionalMonitor.Stop(); err != nil {
		logger.WithError(err).Error("Failed to stop traditional monitor")
	} else {
		logger.Info("Traditional monitor stopped")
	}

	// Stop metrics collector
	if err := metricsCollector.Stop(); err != nil {
		logger.WithError(err).Error("Failed to stop metrics collector")
	} else {
		logger.Info("Metrics collector stopped")
	}

	// Stop actor system
	if err := actorSystem.Stop(); err != nil {
		logger.WithError(err).Error("Failed to stop actor system")
	} else {
		logger.Info("Actor system stopped")
	}

	logger.Info("Application shutdown completed")
}
