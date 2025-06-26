package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"actor-model-observability/internal/actor"
	"actor-model-observability/internal/config"
	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/models"
	"actor-model-observability/internal/observability"
	"actor-model-observability/internal/service"
	"actor-model-observability/internal/traditional"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// BenchmarkSetup holds common benchmark setup
type BenchmarkSetup struct {
	actorSystem        *actor.ActorSystem
	metricsCollector   *observability.MetricsCollector
	traditionalMonitor *traditional.TraditionalMonitor
	rideService        *service.RideService
	actorRouter        *gin.Engine
	traditionalRouter  *gin.Engine
}

// setupBenchmark initializes the test environment
func setupBenchmark(b *testing.B) *BenchmarkSetup {
	b.Helper()

	// Create configuration for observability (pkg/config)
	pkgCfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Mode: "test",
		},
		Actor: config.ActorConfig{
			MaxActors:           1000,
			SupervisionStrategy: "restart",
		},
		OpenTelemetry: config.OpenTelemetryConfig{
			ServiceName:    "benchmark-test",
			TracingEnabled: false,
			MetricsEnabled: false,
			OTLPEndpoint:   "http://localhost:4317",
		},
		Metrics: config.MetricsConfig{
			CollectInterval: 1 * time.Second,
			FlushInterval:   5 * time.Second,
			RetentionPeriod: 1 * time.Hour,
			BatchSize:       10,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}

	// Initialize actor system
	actorSystem := actor.NewActorSystem("benchmark-system")
	// Start the actor system with a background context
	if err := actorSystem.Start(context.Background()); err != nil {
		b.Fatalf("Failed to start actor system: %v", err)
	}

	// Initialize logger for benchmark
	logger, err := logging.NewLogger(&pkgCfg.Logging)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	// Initialize observability (with nil dependencies for benchmark)
	metricsCollector := observability.NewMetricsCollector(nil, nil, pkgCfg, logger)

	// Initialize OTel monitor for traditional monitor
	otelMonitor, err := observability.NewOTelMonitor(&pkgCfg.OpenTelemetry, logger)
	if err != nil {
		b.Fatalf("Failed to create OTel monitor: %v", err)
	}

	// Initialize traditional monitor
	traditionalMonitor := traditional.NewTraditionalMonitor(logger, otelMonitor)

	// Create mock repositories
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}
	tripRepo := &MockTripRepository{}

	// Setup mock expectations for benchmark tests
	// Mock passenger repository
	passengerRepo.On("GetByID", mock.Anything, mock.AnythingOfType("string")).Return(&models.Passenger{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Rating: 4.5,
	}, nil)

	// Mock driver repository
	lat := 40.7128
	lng := -74.0060
	mockDrivers := []*models.Driver{
		{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			VehicleType:      "sedan",
			VehiclePlate:     "ABC123",
			Status:           "online",
			CurrentLatitude:  &lat,
			CurrentLongitude: &lng,
			Rating:           4.8,
		},
	}
	driverRepo.On("GetOnlineDrivers", mock.Anything).Return(mockDrivers, nil)

	// Mock trip repository
	tripRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(nil)

	// Initialize ride service
	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystem, metricsCollector, traditionalMonitor, logger, true,
	)

	// Setup routers
	gin.SetMode(gin.TestMode)
	actorRouter := setupActorRouter(rideService, actorSystem, metricsCollector)
	traditionalRouter := setupTraditionalRouter(rideService, traditionalMonitor)

	return &BenchmarkSetup{
		actorSystem:        actorSystem,
		metricsCollector:   metricsCollector,
		traditionalMonitor: traditionalMonitor,
		rideService:        rideService,
		actorRouter:        actorRouter,
		traditionalRouter:  traditionalRouter,
	}
}

// setupActorRouter creates a router for actor-based endpoints
func setupActorRouter(rideService *service.RideService, actorSystem *actor.ActorSystem, metricsCollector *observability.MetricsCollector) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Actor-based endpoints
	v1 := router.Group("/api/v1")
	{
		v1.POST("/rides/request", func(c *gin.Context) {
			var req models.RideRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Simulate actor-based ride request processing
			pickup := models.Location{Latitude: req.PickupLat, Longitude: req.PickupLng}
			dropoff := models.Location{Latitude: req.DestinationLat, Longitude: req.DestinationLng}
			trip, err := rideService.RequestRide(c.Request.Context(), req.PassengerID.String(), pickup, dropoff, "", "")

			// Record metrics
			metricsCollector.RecordMessage("passenger", "matching", "ride_request", req, time.Now())

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, trip)
		})

		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy", "mode": "actor"})
		})
	}

	return router
}

// setupTraditionalRouter creates a router for traditional endpoints
func setupTraditionalRouter(rideService *service.RideService, traditionalMonitor *traditional.TraditionalMonitor) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Traditional endpoints
	v1 := router.Group("/api/v1/traditional")
	{
		v1.POST("/rides/request", func(c *gin.Context) {
			var req models.RideRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Simulate traditional ride request processing
			start := time.Now()
			pickup := models.Location{Latitude: req.PickupLat, Longitude: req.PickupLng}
			dropoff := models.Location{Latitude: req.DestinationLat, Longitude: req.DestinationLng}
			trip, err := rideService.RequestRide(c.Request.Context(), req.PassengerID.String(), pickup, dropoff, "", "")
			processingTime := time.Since(start)

			// Record metrics
			traditionalMonitor.RecordRequest("/rides/request", "POST", processingTime, http.StatusOK)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, trip)
		})

		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy", "mode": "traditional"})
		})
	}

	return router
}

// Mock repositories are now defined in test_helpers.go

// Benchmark tests

// BenchmarkActorRideRequest benchmarks actor-based ride requests
func BenchmarkActorRideRequest(b *testing.B) {
	setup := setupBenchmark(b)
	server := httptest.NewServer(setup.actorRouter)
	defer server.Close()

	rideRequest := models.RideRequest{
		PassengerID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		PickupLat:      -6.2088,
		PickupLng:      106.8456,
		DestinationLat: -6.1944,
		DestinationLng: 106.8229,
	}

	payload, _ := json.Marshal(rideRequest)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Post(server.URL+"/api/v1/rides/request", "application/json", strings.NewReader(string(payload)))
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkTraditionalRideRequest benchmarks traditional ride requests
func BenchmarkTraditionalRideRequest(b *testing.B) {
	setup := setupBenchmark(b)
	server := httptest.NewServer(setup.traditionalRouter)
	defer server.Close()

	rideRequest := models.RideRequest{
		PassengerID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		PickupLat:      -6.2088,
		PickupLng:      106.8456,
		DestinationLat: -6.1944,
		DestinationLng: 106.8229,
	}

	payload, _ := json.Marshal(rideRequest)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Post(server.URL+"/api/v1/traditional/rides/request", "application/json", strings.NewReader(string(payload)))
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkActorSystemMessagePassing benchmarks actor message passing
func BenchmarkActorSystemMessagePassing(b *testing.B) {
	setup := setupBenchmark(b)
	ctx := context.Background()

	// Start actor system
	setup.actorSystem.Start(ctx)
	defer setup.actorSystem.Stop()

	// Create test actors with simple handlers
	passengerHandler := func(msg actor.Message) error { return nil }
	driverHandler := func(msg actor.Message) error { return nil }

	passengerRef, _ := setup.actorSystem.SpawnActor("passenger", "passenger-1", 100, passengerHandler, actor.SupervisionRestart)
	driverRef, _ := setup.actorSystem.SpawnActor("driver", "driver-1", 100, driverHandler, actor.SupervisionRestart)

	message := actor.NewBaseMessage("ride_request", map[string]interface{}{"passenger_id": "passenger-1"}, passengerRef.ID)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			setup.actorSystem.SendMessage(driverRef.ID, message)
		}
	})
}

// BenchmarkObservabilityOverhead benchmarks observability collection overhead
func BenchmarkObservabilityOverhead(b *testing.B) {
	setup := setupBenchmark(b)
	ctx := context.Background()

	// Start metrics collection
	setup.metricsCollector.Start(ctx)
	defer setup.metricsCollector.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate metric collection
			setup.metricsCollector.RecordMessage(
				"sender",
				"receiver",
				"test_message",
				"test_payload",
				time.Now(),
			)
		}
	})
}

// BenchmarkTraditionalMonitoringOverhead benchmarks traditional monitoring overhead
func BenchmarkTraditionalMonitoringOverhead(b *testing.B) {
	setup := setupBenchmark(b)
	ctx := context.Background()

	// Start traditional monitoring
	setup.traditionalMonitor.Start(ctx)
	defer setup.traditionalMonitor.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate traditional metric recording
			setup.traditionalMonitor.RecordRequest(
				"/api/test",
				"POST",
				time.Millisecond*10,
				http.StatusOK,
			)
		}
	})
}

// BenchmarkConcurrentActorOperations benchmarks concurrent actor operations
func BenchmarkConcurrentActorOperations(b *testing.B) {
	setup := setupBenchmark(b)
	ctx := context.Background()

	setup.actorSystem.Start(ctx)
	defer setup.actorSystem.Stop()

	// Create multiple actors
	numActors := 100
	actors := make([]*actor.ActorRef, numActors)
	for i := 0; i < numActors; i++ {
		actorID := fmt.Sprintf("actor-%d", i)
		actorHandler := func(msg actor.Message) error { return nil }
		actorRef, _ := setup.actorSystem.SpawnActor("passenger", actorID, 100, actorHandler, actor.SupervisionRestart)
		actors[i] = actorRef
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Send message to random actor
			actorIndex := b.N % numActors
			message := actor.NewBaseMessage("test_message", map[string]interface{}{"data": "test"}, "benchmark")
			setup.actorSystem.SendMessage(actors[actorIndex].ID, message)
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	setup := setupBenchmark(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create and process ride request
		passengerID := fmt.Sprintf("passenger-%d", i)
		pickup := models.Location{
			Latitude:  -6.2088 + float64(i%100)/10000,
			Longitude: 106.8456 + float64(i%100)/10000,
		}
		dropoff := models.Location{
			Latitude:  -6.1944 + float64(i%100)/10000,
			Longitude: 106.8229 + float64(i%100)/10000,
		}

		// Process with actor model
		_, err := setup.rideService.RequestRide(context.Background(), passengerID, pickup, dropoff, "", "")
		assert.NoError(b, err)
	}
}

// BenchmarkScalability tests scalability characteristics
func BenchmarkScalability(b *testing.B) {
	userCounts := []int{1, 10, 50, 100, 500}

	for _, userCount := range userCounts {
		b.Run(fmt.Sprintf("Users_%d", userCount), func(b *testing.B) {
			setup := setupBenchmark(b)
			server := httptest.NewServer(setup.actorRouter)
			defer server.Close()

			rideRequest := models.RideRequest{
				PassengerID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				PickupLat:      -6.2088,
				PickupLng:      106.8456,
				DestinationLat: -6.1944,
				DestinationLng: 106.8229,
			}

			payload, _ := json.Marshal(rideRequest)

			b.ResetTimer()
			b.SetParallelism(userCount)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					resp, err := http.Post(server.URL+"/api/v1/rides/request", "application/json", strings.NewReader(string(payload)))
					if err != nil {
						b.Error(err)
						continue
					}
					resp.Body.Close()
				}
			})
		})
	}
}

// TestBenchmarkComparison runs a comparison test between actor and traditional approaches
func TestBenchmarkComparison(t *testing.T) {
	// Test actor performance
	actorResult := testing.Benchmark(func(b *testing.B) {
		setup := setupBenchmark(b)
		server := httptest.NewServer(setup.actorRouter)
		defer server.Close()

		rideRequest := models.RideRequest{
			PassengerID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			PickupLat:      -6.2088,
			PickupLng:      106.8456,
			DestinationLat: -6.1944,
			DestinationLng: 106.8229,
		}

		payload, _ := json.Marshal(rideRequest)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := http.Post(server.URL+"/api/v1/rides/request", "application/json", strings.NewReader(string(payload)))
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
		}
	})

	// Test traditional performance
	traditionalResult := testing.Benchmark(func(b *testing.B) {
		setup := setupBenchmark(b)
		server := httptest.NewServer(setup.traditionalRouter)
		defer server.Close()

		rideRequest := models.RideRequest{
			PassengerID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			PickupLat:      -6.2088,
			PickupLng:      106.8456,
			DestinationLat: -6.1944,
			DestinationLng: 106.8229,
		}

		payload, _ := json.Marshal(rideRequest)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := http.Post(server.URL+"/api/v1/traditional/rides/request", "application/json", strings.NewReader(string(payload)))
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
		}
	})

	// Compare results
	actorNsPerOp := actorResult.NsPerOp()
	traditionalNsPerOp := traditionalResult.NsPerOp()

	if actorNsPerOp < traditionalNsPerOp {
		performanceGain := float64(traditionalNsPerOp-actorNsPerOp) / float64(traditionalNsPerOp) * 100
		_ = performanceGain // Performance gain calculated but not logged
	} else {
		performanceLoss := float64(actorNsPerOp-traditionalNsPerOp) / float64(traditionalNsPerOp) * 100
		_ = performanceLoss // Performance loss calculated but not logged
	}

	// Ensure both approaches work
	assert.True(t, actorResult.N > 0, "Actor benchmark should run")
	assert.True(t, traditionalResult.N > 0, "Traditional benchmark should run")
}
