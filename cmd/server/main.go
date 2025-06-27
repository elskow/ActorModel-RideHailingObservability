// Package main Actor Model Ride Hailing Observability API
//
// This is a ride hailing application built with the Actor Model pattern,
// featuring comprehensive observability with OpenTelemetry, Prometheus, and Jaeger.
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /api/v1
//	Version: 1.0.0
//	License: MIT https://opensource.org/licenses/MIT
//	Contact: Developer <dev@example.com> https://github.com/your-repo
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- api_key:
//
//	SecurityDefinitions:
//	api_key:
//	     type: apiKey
//	     name: KEY
//	     in: header
//
// swagger:meta
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"actor-model-observability/internal/actor"
	"actor-model-observability/internal/config"
	"actor-model-observability/internal/database"
	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/models"
	"actor-model-observability/internal/observability"
	"actor-model-observability/internal/repository/postgres"
	"actor-model-observability/internal/router"
	"actor-model-observability/internal/service"
	"actor-model-observability/internal/traditional"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	// Set up actor system event handlers for observability
	actorSystem.SetEventHandlers(
		// onActorStarted
		func(actorID string) {
			// Get actor reference to determine type
			actorRef, err := actorSystem.GetActor(actorID)
			if err != nil {
				logger.WithError(err).WithField("actor_id", actorID).Error("Failed to get actor reference for observability")
				return
			}

			// Create actor instance record
			actorInstance := &models.ActorInstance{
				ID:            uuid.New(),
				ActorType:     models.ActorType(actorRef.Type),
				ActorID:       actorID,
				EntityID:      nil, // Will be set when actor processes specific entities
				Status:        models.ActorStatusActive,
				LastHeartbeat: time.Now(),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			if err := observabilityRepo.CreateActorInstance(context.Background(), actorInstance); err != nil {
				logger.WithError(err).WithFields(logging.Fields{
					"actor_id":   actorID,
					"actor_type": actorRef.Type,
				}).Error("Failed to create actor instance record")
			}

			// Create event log for actor started
			eventData, _ := json.Marshal(map[string]interface{}{
				"actor_id":   actorID,
				"actor_type": actorRef.Type,
				"status":     "started",
			})
			eventLog := &models.EventLog{
				ID:            uuid.New(),
				EventType:     "actor_started",
				EventCategory: models.EventCategorySystem,
				ActorType:     &actorInstance.ActorType,
				ActorID:       &actorID,
				EventData:     eventData,
				Severity:      models.EventSeverityInfo,
				Message:       fmt.Sprintf("Actor %s of type %s started", actorID, actorRef.Type),
				Timestamp:     time.Now(),
				CreatedAt:     time.Now(),
			}
			if err := observabilityRepo.CreateEventLog(context.Background(), eventLog); err != nil {
				logger.WithError(err).Error("Failed to create actor started event log")
			}
		},
		// onActorStopped
		func(actorID string) {
			// Create event log for actor stopped
			eventData, _ := json.Marshal(map[string]interface{}{
				"actor_id": actorID,
				"status":   "stopped",
			})
			eventLog := &models.EventLog{
				ID:            uuid.New(),
				EventType:     "actor_stopped",
				EventCategory: models.EventCategorySystem,
				ActorID:       &actorID,
				EventData:     eventData,
				Severity:      models.EventSeverityInfo,
				Message:       fmt.Sprintf("Actor %s stopped", actorID),
				Timestamp:     time.Now(),
				CreatedAt:     time.Now(),
			}
			if err := observabilityRepo.CreateEventLog(context.Background(), eventLog); err != nil {
				logger.WithError(err).Error("Failed to create actor stopped event log")
			}
			logger.WithField("actor_id", actorID).Debug("Actor stopped - observability tracking")
		},
		// onActorFailed
		func(actorID string, err error) {
			// Create event log for actor failed
			eventData, _ := json.Marshal(map[string]interface{}{
				"actor_id": actorID,
				"error":    err.Error(),
				"status":   "failed",
			})
			eventLog := &models.EventLog{
				ID:            uuid.New(),
				EventType:     "actor_failed",
				EventCategory: models.EventCategoryError,
				ActorID:       &actorID,
				EventData:     eventData,
				Severity:      models.EventSeverityError,
				Message:       fmt.Sprintf("Actor %s failed: %s", actorID, err.Error()),
				Timestamp:     time.Now(),
				CreatedAt:     time.Now(),
			}
			if createErr := observabilityRepo.CreateEventLog(context.Background(), eventLog); createErr != nil {
				logger.WithError(createErr).Error("Failed to create actor failed event log")
			}
			logger.WithError(err).WithField("actor_id", actorID).Error("Actor failed - observability tracking")
		},
		// onMessage
		func(from, to, messageType string) {
			// Create event log for message sent
			eventData, _ := json.Marshal(map[string]interface{}{
				"from":         from,
				"to":           to,
				"message_type": messageType,
			})
			eventLog := &models.EventLog{
				ID:            uuid.New(),
				EventType:     "message_sent",
				EventCategory: models.EventCategoryBusiness,
				ActorID:       &from,
				EventData:     eventData,
				Severity:      models.EventSeverityDebug,
				Message:       fmt.Sprintf("Message %s sent from %s to %s", messageType, from, to),
				Timestamp:     time.Now(),
				CreatedAt:     time.Now(),
			}
			if err := observabilityRepo.CreateEventLog(context.Background(), eventLog); err != nil {
				logger.WithError(err).Error("Failed to create message sent event log")
			}
			logger.WithFields(logging.Fields{
				"from":         from,
				"to":           to,
				"message_type": messageType,
			}).Debug("Message sent - observability tracking")
		},
	)

	// Initialize OpenTelemetry monitor
	otelMonitor, err := observability.NewOTelMonitor(&cfg.OpenTelemetry, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize OpenTelemetry monitor")
	}

	// Initialize observability collector
	metricsCollector := observability.NewMetricsCollector(db, redisClient.Client, cfg, logger)

	// Initialize traditional monitoring
	traditionalMonitor := traditional.NewTraditionalMonitor(logger, otelMonitor)

	// Initialize services
	rideService := service.NewRideService(
		userRepo,
		driverRepo,
		passengerRepo,
		tripRepo,
		actorSystem,
		metricsCollector,
		traditionalMonitor,
		logger,
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

	// Perform graceful shutdown with proper error handling
	performGracefulShutdown(server, actorSystem, metricsCollector, traditionalMonitor, db, redisClient, logger)

	logger.Info("Application shutdown completed")
}

// performGracefulShutdown handles the graceful shutdown of all services
func performGracefulShutdown(
	server *http.Server,
	actorSystem *actor.ActorSystem,
	metricsCollector *observability.MetricsCollector,
	traditionalMonitor *traditional.TraditionalMonitor,
	db *database.PostgresDB,
	redisClient *database.RedisClient,
	logger *logging.Logger,
) {
	// Create a context with timeout for the entire shutdown process
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer shutdownCancel()

	// Channel to collect shutdown errors
	errorChan := make(chan error, 6)
	var shutdownWg sync.WaitGroup

	// Shutdown HTTP server first
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		logger.Info("Shutting down HTTP server...")

		// Create a shorter timeout for HTTP server shutdown
		serverCtx, serverCancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer serverCancel()

		if err := server.Shutdown(serverCtx); err != nil {
			errorChan <- fmt.Errorf("HTTP server shutdown error: %w", err)
			logger.WithError(err).Error("Server forced to shutdown")
		} else {
			logger.Info("HTTP server shutdown completed")
		}
	}()

	// Stop background services concurrently
	logger.Info("Stopping background services...")

	// Stop traditional monitor
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		logger.Info("Stopping traditional monitor...")

		// Create a timeout context for traditional monitor
		monitorCtx, monitorCancel := context.WithTimeout(shutdownCtx, 8*time.Second)
		defer monitorCancel()

		// Create a channel to handle the stop operation with timeout
		done := make(chan error, 1)
		go func() {
			done <- traditionalMonitor.Stop()
		}()

		select {
		case err := <-done:
			if err != nil {
				errorChan <- fmt.Errorf("traditional monitor stop error: %w", err)
				logger.WithError(err).Error("Failed to stop traditional monitor")
			} else {
				logger.Info("Traditional monitor stopped")
			}
		case <-monitorCtx.Done():
			errorChan <- fmt.Errorf("traditional monitor stop timeout")
			logger.Error("Traditional monitor stop timed out")
		}
	}()

	// Stop metrics collector
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		logger.Info("Stopping metrics collector...")

		if err := metricsCollector.Stop(); err != nil {
			errorChan <- fmt.Errorf("metrics collector stop error: %w", err)
			logger.WithError(err).Error("Failed to stop metrics collector")
		} else {
			logger.Info("Metrics collector stopped")
		}
	}()

	// Stop actor system
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		logger.Info("Stopping actor system...")

		if err := actorSystem.Stop(); err != nil {
			errorChan <- fmt.Errorf("actor system stop error: %w", err)
			logger.WithError(err).Error("Failed to stop actor system")
		} else {
			logger.Info("Actor system stopped")
		}
	}()

	// Close database connection
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		logger.Info("Closing database connection...")

		if err := db.Close(); err != nil {
			errorChan <- fmt.Errorf("database close error: %w", err)
			logger.WithError(err).Error("Failed to close database connection")
		} else {
			logger.Info("Database connection closed")
		}
	}()

	// Close Redis connection
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		logger.Info("Closing Redis connection...")

		if err := redisClient.Close(); err != nil {
			errorChan <- fmt.Errorf("Redis close error: %w", err)
			logger.WithError(err).Error("Failed to close Redis connection")
		} else {
			logger.Info("Redis connection closed")
		}
	}()

	// Wait for all shutdown operations to complete or timeout
	done := make(chan struct{})
	go func() {
		shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All services stopped successfully")
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout reached, forcing exit")
	}

	// Close error channel and log any collected errors
	close(errorChan)
	errorCount := 0
	for err := range errorChan {
		logger.WithError(err).Error("Shutdown error occurred")
		errorCount++
	}

	if errorCount > 0 {
		logger.WithFields(logging.Fields{
			"error_count": errorCount,
		}).Warn("Shutdown completed with errors")
	} else {
		logger.Info("Shutdown completed successfully")
	}
}
