package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"actor-model-observability/internal/actor"
	"actor-model-observability/internal/models"
	"actor-model-observability/internal/observability"
	"actor-model-observability/internal/repository"
	"actor-model-observability/internal/traditional"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RideService handles ride-related business logic
type RideService struct {
	userRepo       repository.UserRepository
	driverRepo     repository.DriverRepository
	passengerRepo  repository.PassengerRepository
	tripRepo       repository.TripRepository
	actorSystem    *actor.ActorSystem
	metricsCollector *observability.MetricsCollector
	traditionalMonitor *traditional.TraditionalMonitor
	logger         *logrus.Entry
	useActorModel  bool
}

// NewRideService creates a new ride service
func NewRideService(
	userRepo repository.UserRepository,
	driverRepo repository.DriverRepository,
	passengerRepo repository.PassengerRepository,
	tripRepo repository.TripRepository,
	actorSystem *actor.ActorSystem,
	metricsCollector *observability.MetricsCollector,
	traditionalMonitor *traditional.TraditionalMonitor,
	useActorModel bool,
) *RideService {
	return &RideService{
		userRepo:           userRepo,
		driverRepo:         driverRepo,
		passengerRepo:      passengerRepo,
		tripRepo:           tripRepo,
		actorSystem:        actorSystem,
		metricsCollector:   metricsCollector,
		traditionalMonitor: traditionalMonitor,
		logger:             logrus.WithField("service", "ride"),
		useActorModel:      useActorModel,
	}
}

// RequestRide handles ride requests
func (rs *RideService) RequestRide(ctx context.Context, passengerID string, pickup, dropoff models.Location, pickupAddr, dropoffAddr string) (*models.Trip, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if rs.useActorModel {
			rs.metricsCollector.RecordEvent("ride_request", "ride_service", "Ride request processed", map[string]interface{}{
				"passenger_id": passengerID,
				"duration_ms":  duration.Milliseconds(),
				"method":       "actor_model",
			})
		} else {
			rs.traditionalMonitor.RecordRequest("/api/rides", "POST", duration, 200)
			rs.traditionalMonitor.RecordBusinessMetrics("ride_requests_total", 1, map[string]string{
				"passenger_id": passengerID,
			})
		}
	}()

	// Validate passenger
	passenger, err := rs.passengerRepo.GetByID(ctx, passengerID)
	if err != nil {
		return nil, fmt.Errorf("passenger not found: %w", err)
	}

	// Parse passengerID string to UUID
	passengerUUID, err := uuid.Parse(passengerID)
	if err != nil {
		return nil, fmt.Errorf("invalid passenger ID: %w", err)
	}

	// Create trip
	trip := &models.Trip{
		ID:                   uuid.New(),
		PassengerID:          passengerUUID,
		PickupLatitude:       pickup.Latitude,
		PickupLongitude:      pickup.Longitude,
		DestinationLatitude:  dropoff.Latitude,
		DestinationLongitude: dropoff.Longitude,
		PickupAddress:        &pickupAddr,
		DestinationAddress:   &dropoffAddr,
		Status:               models.TripStatusRequested,
		RequestedAt:          time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := rs.tripRepo.Create(ctx, trip); err != nil {
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}

	if rs.useActorModel {
		return rs.requestRideActorModel(ctx, passenger, trip, pickup, dropoff, pickupAddr, dropoffAddr)
	} else {
		return rs.requestRideTraditional(ctx, passenger, trip, pickup, dropoff, pickupAddr, dropoffAddr)
	}
}

// requestRideActorModel handles ride request using actor model
func (rs *RideService) requestRideActorModel(ctx context.Context, passenger *models.Passenger, trip *models.Trip, pickup, dropoff models.Location, pickupAddr, dropoffAddr string) (*models.Trip, error) {
	// Check if passenger actor already exists
	passengerActorID := fmt.Sprintf("passenger-%s", passenger.ID.String())
	_, err := rs.actorSystem.GetActor(passengerActorID)
	if err != nil {
		// Create new passenger actor
		pa, err := actor.NewPassengerActor(passenger, rs.actorSystem)
		if err != nil {
			return nil, fmt.Errorf("failed to create passenger actor: %w", err)
		}

		if err := pa.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start passenger actor: %w", err)
		}

		// Register actor in system
		// Create a wrapper function to handle messages for this passenger actor
		handler := func(msg actor.Message) error {
			// Since handleMessage is unexported, we need to use the actor's Send method
			// or create a public method. For now, let's use a simple approach.
			return nil // TODO: Implement proper message handling
		}
		if _, err := rs.actorSystem.SpawnActor("passenger", passengerActorID, 100, handler, actor.SupervisionRestart); err != nil {
			return nil, fmt.Errorf("failed to spawn passenger actor: %w", err)
		}

		// Store the passenger actor reference for later use
		_ = pa // Suppress unused variable warning
	}

	// Send ride request message to trip matching service
	payload := actor.RequestRidePayload{
		PickupLat:   pickup.Latitude,
		PickupLng:   pickup.Longitude,
		DropoffLat:  dropoff.Latitude,
		DropoffLng:  dropoff.Longitude,
		PickupAddr:  pickupAddr,
		DropoffAddr: dropoffAddr,
		RequestedAt: time.Now(),
	}

	// Record message in observability system
	rs.metricsCollector.RecordMessage(passengerActorID, "trip-matcher", actor.MsgTypeRequestRide, payload, time.Now())

	// In a real system, this would be sent to a trip matching actor
	// For demo purposes, we'll simulate the matching process
	go rs.simulateRideMatching(ctx, trip)

	rs.logger.WithFields(logrus.Fields{
		"trip_id":      trip.ID,
		"passenger_id": passenger.ID,
		"method":       "actor_model",
	}).Info("Ride request processed via actor model")

	return trip, nil
}

// requestRideTraditional handles ride request using traditional approach
func (rs *RideService) requestRideTraditional(ctx context.Context, passenger *models.Passenger, trip *models.Trip, pickup, dropoff models.Location, pickupAddr, dropoffAddr string) (*models.Trip, error) {
	// Traditional centralized approach
	start := time.Now()

	// Find available drivers
	drivers, err := rs.findNearbyDrivers(ctx, pickup, 5.0) // 5km radius
	if err != nil {
		rs.traditionalMonitor.RecordDatabaseOperation("SELECT", "drivers", time.Since(start), false)
		return nil, fmt.Errorf("failed to find nearby drivers: %w", err)
	}
	rs.traditionalMonitor.RecordDatabaseOperation("SELECT", "drivers", time.Since(start), true)

	if len(drivers) == 0 {
		return nil, fmt.Errorf("no available drivers found")
	}

	// Select best driver (closest for simplicity)
	bestDriver := rs.selectBestDriver(drivers, pickup)

	// Update trip with matched driver
	trip.DriverID = &bestDriver.ID
	trip.Status = models.TripStatusMatched
	trip.MatchedAt = &[]time.Time{time.Now()}[0]

	updateStart := time.Now()
	if err := rs.tripRepo.Update(ctx, trip); err != nil {
		rs.traditionalMonitor.RecordDatabaseOperation("UPDATE", "trips", time.Since(updateStart), false)
		return nil, fmt.Errorf("failed to update trip: %w", err)
	}
	rs.traditionalMonitor.RecordDatabaseOperation("UPDATE", "trips", time.Since(updateStart), true)

	// Update driver status
	bestDriver.Status = models.DriverStatusBusy
	driverUpdateStart := time.Now()
	if err := rs.driverRepo.Update(ctx, bestDriver); err != nil {
		rs.traditionalMonitor.RecordDatabaseOperation("UPDATE", "drivers", time.Since(driverUpdateStart), false)
		return nil, fmt.Errorf("failed to update driver status: %w", err)
	}
	rs.traditionalMonitor.RecordDatabaseOperation("UPDATE", "drivers", time.Since(driverUpdateStart), true)

	rs.logger.WithFields(logrus.Fields{
		"trip_id":      trip.ID,
		"passenger_id": passenger.ID,
		"driver_id":    bestDriver.ID,
		"method":       "traditional",
	}).Info("Ride request processed via traditional approach")

	return trip, nil
}

// CancelRide handles ride cancellation
func (rs *RideService) CancelRide(ctx context.Context, tripID, reason string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if rs.useActorModel {
			rs.metricsCollector.RecordEvent("ride_cancel", "ride_service", "Ride cancelled", map[string]interface{}{
				"trip_id":     tripID,
				"reason":      reason,
				"duration_ms": duration.Milliseconds(),
				"method":      "actor_model",
			})
		} else {
			rs.traditionalMonitor.RecordRequest("/api/rides/cancel", "POST", duration, 200)
		}
	}()

	// Get trip
	trip, err := rs.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	if rs.useActorModel {
		return rs.cancelRideActorModel(ctx, trip, reason)
	} else {
		return rs.cancelRideTraditional(ctx, trip, reason)
	}
}

// cancelRideActorModel handles cancellation using actor model
func (rs *RideService) cancelRideActorModel(ctx context.Context, trip *models.Trip, reason string) error {
	// Send cancellation message to relevant actors
	payload := actor.CancelRidePayload{
		TripID:      trip.ID.String(),
		Reason:      reason,
		CancelledAt: time.Now(),
	}

	message := actor.NewBaseMessage(actor.MsgTypeCancelRide, payload, "ride-service")

	// Notify passenger actor
	passengerActorID := fmt.Sprintf("passenger-%s", trip.PassengerID.String())
	if err := rs.actorSystem.SendMessage(passengerActorID, message); err != nil {
		rs.logger.WithError(err).Warn("Failed to notify passenger actor of cancellation")
	}

	// Notify driver actor if assigned
	if trip.DriverID != nil {
		driverActorID := fmt.Sprintf("driver-%s", trip.DriverID.String())
		if err := rs.actorSystem.SendMessage(driverActorID, message); err != nil {
			rs.logger.WithError(err).Warn("Failed to notify driver actor of cancellation")
		}
	}

	// Update trip status
	trip.Status = models.TripStatusCancelled
	trip.CancelledAt = &[]time.Time{time.Now()}[0]

	if err := rs.tripRepo.Update(ctx, trip); err != nil {
		return fmt.Errorf("failed to update trip: %w", err)
	}

	// Record message
	rs.metricsCollector.RecordMessage("ride-service", passengerActorID, actor.MsgTypeCancelRide, payload, time.Now())

	return nil
}

// cancelRideTraditional handles cancellation using traditional approach
func (rs *RideService) cancelRideTraditional(ctx context.Context, trip *models.Trip, reason string) error {
	// Traditional centralized cancellation
	start := time.Now()

	// Update trip status
	trip.Status = models.TripStatusCancelled
	trip.CancelledAt = &[]time.Time{time.Now()}[0]

	if err := rs.tripRepo.Update(ctx, trip); err != nil {
		rs.traditionalMonitor.RecordDatabaseOperation("UPDATE", "trips", time.Since(start), false)
		return fmt.Errorf("failed to update trip: %w", err)
	}
	rs.traditionalMonitor.RecordDatabaseOperation("UPDATE", "trips", time.Since(start), true)

	// Free up driver if assigned
	if trip.DriverID != nil {
		driverStart := time.Now()
		driver, err := rs.driverRepo.GetByID(ctx, trip.DriverID.String())
		if err == nil {
			driver.Status = models.DriverStatusOnline
			rs.driverRepo.Update(ctx, driver)
		}
		rs.traditionalMonitor.RecordDatabaseOperation("UPDATE", "drivers", time.Since(driverStart), err == nil)
	}

	return nil
}

// simulateRideMatching simulates the ride matching process for actor model
func (rs *RideService) simulateRideMatching(ctx context.Context, trip *models.Trip) {
	// Simulate matching delay
	time.Sleep(2 * time.Second)

	// Find nearby drivers
	pickup := models.Location{Latitude: trip.PickupLatitude, Longitude: trip.PickupLongitude}
	drivers, err := rs.findNearbyDrivers(ctx, pickup, 5.0)
	if err != nil || len(drivers) == 0 {
		rs.logger.WithField("trip_id", trip.ID).Warn("No drivers found for matching")
		return
	}

	// Select best driver
	bestDriver := rs.selectBestDriver(drivers, pickup)

	// Update trip
	trip.DriverID = &bestDriver.ID
	trip.Status = models.TripStatusMatched
	trip.MatchedAt = &[]time.Time{time.Now()}[0]
	rs.tripRepo.Update(ctx, trip)

	// Update driver status
	bestDriver.Status = models.DriverStatusBusy
	rs.driverRepo.Update(ctx, bestDriver)

	// Send matched notification to passenger actor
	payload := actor.RideMatchedPayload{
		TripID:       trip.ID.String(),
		DriverID:     bestDriver.ID.String(),
		DriverName:   "Driver Name", // Would get from user table
		DriverPhone:  "123-456-7890",
		VehicleInfo:  bestDriver.VehicleType + " " + bestDriver.VehiclePlate,
		DriverLat:    *bestDriver.CurrentLatitude,
		DriverLng:    *bestDriver.CurrentLongitude,
		EstimatedETA: 5 * time.Minute,
		MatchedAt:    time.Now(),
	}

	message := actor.NewBaseMessage(actor.MsgTypeRideMatched, payload, "trip-matcher")
	passengerActorID := fmt.Sprintf("passenger-%s", trip.PassengerID.String())
	rs.actorSystem.SendMessage(passengerActorID, message)

	// Record the matching event
	rs.metricsCollector.RecordMessage("trip-matcher", passengerActorID, actor.MsgTypeRideMatched, payload, time.Now())
}

// findNearbyDrivers finds drivers within a specified radius
func (rs *RideService) findNearbyDrivers(ctx context.Context, location models.Location, radiusKm float64) ([]*models.Driver, error) {
	// This is a simplified implementation
	// In a real system, you would use spatial queries or geospatial indexing
	allDrivers, err := rs.driverRepo.GetOnlineDrivers(ctx)
	if err != nil {
		return nil, err
	}

	var nearbyDrivers []*models.Driver
	for _, driver := range allDrivers {
		distance := rs.calculateDistance(location.Latitude, location.Longitude, *driver.CurrentLatitude, *driver.CurrentLongitude)
		if distance <= radiusKm {
			nearbyDrivers = append(nearbyDrivers, driver)
		}
	}

	return nearbyDrivers, nil
}

// selectBestDriver selects the best driver based on distance and rating
func (rs *RideService) selectBestDriver(drivers []*models.Driver, pickup models.Location) *models.Driver {
	if len(drivers) == 0 {
		return nil
	}

	bestDriver := drivers[0]
	bestScore := rs.calculateDriverScore(bestDriver, pickup)

	for _, driver := range drivers[1:] {
		score := rs.calculateDriverScore(driver, pickup)
		if score > bestScore {
			bestDriver = driver
			bestScore = score
		}
	}

	return bestDriver
}

// calculateDriverScore calculates a score for driver selection
func (rs *RideService) calculateDriverScore(driver *models.Driver, pickup models.Location) float64 {
	distance := rs.calculateDistance(pickup.Latitude, pickup.Longitude, *driver.CurrentLatitude, *driver.CurrentLongitude)
	
	// Score based on distance (closer is better) and rating (higher is better)
	// Normalize distance to 0-1 scale (assuming max 10km)
	distanceScore := math.Max(0, 1.0-(distance/10.0))
	
	// Normalize rating to 0-1 scale (assuming max 5.0)
	ratingScore := driver.Rating / 5.0
	
	// Weighted score: 70% distance, 30% rating
	return (distanceScore * 0.7) + (ratingScore * 0.3)
}

// calculateDistance calculates the distance between two points using Haversine formula
func (rs *RideService) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
}

// calculateEstimatedFare calculates estimated fare based on distance
func (rs *RideService) calculateEstimatedFare(pickup, dropoff models.Location) float64 {
	distance := rs.calculateDistance(pickup.Latitude, pickup.Longitude, dropoff.Latitude, dropoff.Longitude)
	
	// Simple fare calculation: base fare + distance rate
	baseFare := 5.0
	distanceRate := 2.0 // per km
	
	return baseFare + (distance * distanceRate)
}

// GetTripStatus returns the current status of a trip
func (rs *RideService) GetTripStatus(ctx context.Context, tripID string) (*models.Trip, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if !rs.useActorModel {
			rs.traditionalMonitor.RecordRequest("/api/trips/status", "GET", duration, 200)
		}
	}()

	return rs.tripRepo.GetByID(ctx, tripID)
}