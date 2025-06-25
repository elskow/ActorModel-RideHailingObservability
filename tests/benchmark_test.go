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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			ServiceName:     "benchmark-test",
			TracingEnabled:  false,
			MetricsEnabled:  false,
			OTLPEndpoint:    "http://localhost:4317",
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

// Mock repositories for testing
type MockUserRepository struct{}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = uuid.New()
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	userID, _ := uuid.Parse(id)
	return &models.User{ID: userID, Email: "test@example.com", Name: "Test User"}, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return &models.User{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Email: email, Name: "Test User"}, nil
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return &models.User{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Phone: phone, Name: "Test User"}, nil
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return []*models.User{}, nil
}

type MockDriverRepository struct{}

func (m *MockDriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	driver.ID = uuid.New()
	return nil
}

func (m *MockDriverRepository) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	driverID, _ := uuid.Parse(id)
	lat := -6.2088
	lng := 106.8456
	return &models.Driver{
		ID:               driverID,
		UserID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		Status:           models.DriverStatusOnline,
		CurrentLatitude:  &lat,
		CurrentLongitude: &lng,
	}, nil
}

func (m *MockDriverRepository) Update(ctx context.Context, driver *models.Driver) error {
	return nil
}

func (m *MockDriverRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockDriverRepository) GetAvailableDrivers(ctx context.Context, location *models.Location, radius float64) ([]*models.Driver, error) {
	return []*models.Driver{
		{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Status: models.DriverStatusOnline, CurrentLatitude: &location.Latitude, CurrentLongitude: &location.Longitude},
		{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), Status: models.DriverStatusOnline, CurrentLatitude: &location.Latitude, CurrentLongitude: &location.Longitude},
	}, nil
}

func (m *MockDriverRepository) UpdateLocation(ctx context.Context, driverID string, lat, lng float64) error {
	return nil
}

func (m *MockDriverRepository) UpdateStatus(ctx context.Context, driverID string, status models.DriverStatus) error {
	return nil
}

func (m *MockDriverRepository) GetByUserID(ctx context.Context, userID string) (*models.Driver, error) {
	userUUID, _ := uuid.Parse(userID)
	lat := -6.2088
	lng := 106.8456
	return &models.Driver{
		ID:               uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserID:           userUUID,
		Status:           models.DriverStatusOnline,
		CurrentLatitude:  &lat,
		CurrentLongitude: &lng,
	}, nil
}

func (m *MockDriverRepository) GetDriversInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]*models.Driver, error) {
	return []*models.Driver{
		{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Status: models.DriverStatusOnline, CurrentLatitude: &lat, CurrentLongitude: &lng},
	}, nil
}

func (m *MockDriverRepository) GetOnlineDrivers(ctx context.Context) ([]*models.Driver, error) {
	lat := -6.2088
	lng := 106.8456
	return []*models.Driver{
		{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Status: models.DriverStatusOnline, CurrentLatitude: &lat, CurrentLongitude: &lng},
	}, nil
}

func (m *MockDriverRepository) List(ctx context.Context, limit, offset int) ([]*models.Driver, error) {
	return []*models.Driver{}, nil
}

type MockPassengerRepository struct{}

func (m *MockPassengerRepository) Create(ctx context.Context, passenger *models.Passenger) error {
	passenger.ID = uuid.New()
	return nil
}

func (m *MockPassengerRepository) GetByID(ctx context.Context, id string) (*models.Passenger, error) {
	passengerID, _ := uuid.Parse(id)
	return &models.Passenger{ID: passengerID, UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")}, nil
}

func (m *MockPassengerRepository) Update(ctx context.Context, passenger *models.Passenger) error {
	return nil
}

func (m *MockPassengerRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockPassengerRepository) GetByUserID(ctx context.Context, userID string) (*models.Passenger, error) {
	return &models.Passenger{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), UserID: uuid.MustParse(userID)}, nil
}

func (m *MockPassengerRepository) List(ctx context.Context, limit, offset int) ([]*models.Passenger, error) {
	return []*models.Passenger{}, nil
}

type MockTripRepository struct{}

func (m *MockTripRepository) Create(ctx context.Context, trip *models.Trip) error {
	trip.ID = uuid.New()
	trip.Status = models.TripStatusRequested
	trip.CreatedAt = time.Now()
	return nil
}

func (m *MockTripRepository) GetByID(ctx context.Context, id string) (*models.Trip, error) {
	parsedID, _ := uuid.Parse(id)
	passengerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	driverID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	return &models.Trip{
		ID:                   parsedID,
		PassengerID:          passengerID,
		DriverID:             &driverID,
		Status:               models.TripStatusRequested,
		PickupLatitude:       -6.2088,
		PickupLongitude:      106.8456,
		DestinationLatitude:  -6.1944,
		DestinationLongitude: 106.8229,
	}, nil
}

func (m *MockTripRepository) Update(ctx context.Context, trip *models.Trip) error {
	return nil
}

func (m *MockTripRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockTripRepository) GetActiveTrips(ctx context.Context) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

func (m *MockTripRepository) GetTripsByPassenger(ctx context.Context, passengerID string) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

func (m *MockTripRepository) GetTripsByDriver(ctx context.Context, driverID string) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

func (m *MockTripRepository) UpdateStatus(ctx context.Context, tripID string, status models.TripStatus) error {
	return nil
}

func (m *MockTripRepository) GetByDriverID(ctx context.Context, driverID string, limit, offset int) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

func (m *MockTripRepository) GetTripsByStatus(ctx context.Context, status models.TripStatus, limit, offset int) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

func (m *MockTripRepository) GetTripsByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

func (m *MockTripRepository) GetByPassengerID(ctx context.Context, passengerID string, limit, offset int) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

func (m *MockTripRepository) List(ctx context.Context, limit, offset int) ([]*models.Trip, error) {
	return []*models.Trip{}, nil
}

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
	actorResult := testing.Benchmark(BenchmarkActorRideRequest)

	// Test traditional performance
	traditionalResult := testing.Benchmark(BenchmarkTraditionalRideRequest)

	// Compare results
	actorNsPerOp := actorResult.NsPerOp()
	traditionalNsPerOp := traditionalResult.NsPerOp()

	t.Logf("Actor model: %d ns/op", actorNsPerOp)
	t.Logf("Traditional: %d ns/op", traditionalNsPerOp)

	if actorNsPerOp < traditionalNsPerOp {
		performanceGain := float64(traditionalNsPerOp-actorNsPerOp) / float64(traditionalNsPerOp) * 100
		t.Logf("Actor model is %.2f%% faster", performanceGain)
	} else {
		performanceLoss := float64(actorNsPerOp-traditionalNsPerOp) / float64(traditionalNsPerOp) * 100
		t.Logf("Actor model is %.2f%% slower", performanceLoss)
	}

	// Ensure both approaches work
	assert.True(t, actorResult.N > 0, "Actor benchmark should run")
	assert.True(t, traditionalResult.N > 0, "Traditional benchmark should run")
}
