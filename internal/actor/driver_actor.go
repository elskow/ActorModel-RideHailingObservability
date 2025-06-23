package actor

import (
	"encoding/json"
	"fmt"
	"time"

	"actor-model-observability/internal/models"
	"github.com/sirupsen/logrus"
)

// DriverActor handles driver-related operations
type DriverActor struct {
	*BaseActor
	driver *models.Driver
	logger *logrus.Entry
}

// Driver message types
const (
	MsgTypeGoOnline       = "go_online"
	MsgTypeGoOffline      = "go_offline"
	MsgTypeRideRequest    = "ride_request"
	MsgTypeAcceptRide     = "accept_ride"
	MsgTypeRejectRide     = "reject_ride"
	MsgTypeStartRide      = "start_ride"
	MsgTypeCompleteRide   = "complete_ride"
	MsgTypeDriverLocation = "driver_location"
	MsgTypePassengerRated = "passenger_rated"
)

// Driver message payloads
type GoOnlinePayload struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
}

type GoOfflinePayload struct {
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

type RideRequestPayload struct {
	TripID         string        `json:"trip_id"`
	PassengerID    string        `json:"passenger_id"`
	PassengerName  string        `json:"passenger_name"`
	PickupLat      float64       `json:"pickup_lat"`
	PickupLng      float64       `json:"pickup_lng"`
	DropoffLat     float64       `json:"dropoff_lat"`
	DropoffLng     float64       `json:"dropoff_lng"`
	PickupAddr     string        `json:"pickup_address"`
	DropoffAddr    string        `json:"dropoff_address"`
	EstimatedFare  float64       `json:"estimated_fare"`
	EstimatedTime  time.Duration `json:"estimated_time"`
	RequestTimeout time.Duration `json:"request_timeout"`
	RequestedAt    time.Time     `json:"requested_at"`
}

type AcceptRidePayload struct {
	TripID     string    `json:"trip_id"`
	AcceptedAt time.Time `json:"accepted_at"`
}

type RejectRidePayload struct {
	TripID     string    `json:"trip_id"`
	Reason     string    `json:"reason"`
	RejectedAt time.Time `json:"rejected_at"`
}

type StartRidePayload struct {
	TripID    string    `json:"trip_id"`
	StartedAt time.Time `json:"started_at"`
	StartLat  float64   `json:"start_lat"`
	StartLng  float64   `json:"start_lng"`
}

type CompleteRidePayload struct {
	TripID      string        `json:"trip_id"`
	EndLat      float64       `json:"end_lat"`
	EndLng      float64       `json:"end_lng"`
	Distance    float64       `json:"distance"`
	Duration    time.Duration `json:"duration"`
	Fare        float64       `json:"fare"`
	CompletedAt time.Time     `json:"completed_at"`
}

type DriverLocationPayload struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	Timestamp time.Time `json:"timestamp"`
}

type PassengerRatedPayload struct {
	TripID      string  `json:"trip_id"`
	PassengerID string  `json:"passenger_id"`
	Rating      float64 `json:"rating"`
	Comment     string  `json:"comment"`
	RatedAt     time.Time `json:"rated_at"`
}

// NewDriverActor creates a new driver actor
func NewDriverActor(driver *models.Driver, system *ActorSystem) (*DriverActor, error) {
	if driver == nil {
		return nil, fmt.Errorf("driver cannot be nil")
	}

	actorID := fmt.Sprintf("driver-%s", driver.ID)
	logger := logrus.WithFields(logrus.Fields{
		"actor_type": "driver",
		"actor_id":   actorID,
		"driver_id":  driver.ID,
		"user_id":    driver.UserID,
	})

	da := &DriverActor{
		driver: driver,
		logger: logger,
	}

	// Create base actor with message handler
	baseActor := NewBaseActor(actorID, "driver", 100, da.handleMessage)
	da.BaseActor = baseActor

	return da, nil
}

// handleMessage processes incoming messages
func (da *DriverActor) handleMessage(message Message) error {
	da.logger.WithFields(logrus.Fields{
		"message_id":   message.GetID(),
		"message_type": message.GetType(),
		"sender":       message.GetSender(),
	}).Debug("Processing driver message")

	switch message.GetType() {
	case MsgTypeRideRequest:
		return da.handleRideRequest(message)
	case MsgTypePassengerRated:
		return da.handlePassengerRated(message)
	default:
		da.logger.WithField("message_type", message.GetType()).Warn("Unknown message type")
		return fmt.Errorf("unknown message type: %s", message.GetType())
	}
}

// handleRideRequest processes incoming ride requests
func (da *DriverActor) handleRideRequest(message Message) error {
	var payload RideRequestPayload
	if err := da.unmarshalPayload(message.GetPayload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ride request payload: %w", err)
	}

	da.logger.WithFields(logrus.Fields{
		"trip_id":        payload.TripID,
		"passenger_id":   payload.PassengerID,
		"passenger_name": payload.PassengerName,
		"pickup_addr":    payload.PickupAddr,
		"dropoff_addr":   payload.DropoffAddr,
		"estimated_fare": payload.EstimatedFare,
		"estimated_time": payload.EstimatedTime,
	}).Info("Received ride request")

	// Check if driver is available
	if da.driver.Status != models.DriverStatusOnline {
		da.logger.Warn("Driver not online, rejecting ride request")
		return da.rejectRideRequest(payload.TripID, "driver not available")
	}

	// Here you could implement:
	// 1. Auto-accept logic based on driver preferences
	// 2. Distance/fare validation
	// 3. Driver notification system
	// 4. Timeout handling for request expiration

	// For demo purposes, we'll auto-accept if driver is online
	return da.acceptRideRequest(payload.TripID)
}

// handlePassengerRated processes passenger rating notifications
func (da *DriverActor) handlePassengerRated(message Message) error {
	var payload PassengerRatedPayload
	if err := da.unmarshalPayload(message.GetPayload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal passenger rated payload: %w", err)
	}

	da.logger.WithFields(logrus.Fields{
		"trip_id":      payload.TripID,
		"passenger_id": payload.PassengerID,
		"rating":       payload.Rating,
		"comment":      payload.Comment,
	}).Info("Received passenger rating")

	// Update driver's rating
	totalRating := da.driver.Rating*float64(da.driver.TotalTrips) + payload.Rating
	da.driver.TotalTrips++
	da.driver.Rating = totalRating / float64(da.driver.TotalTrips)

	// Here you could:
	// 1. Update driver rating in database
	// 2. Send notification to driver
	// 3. Trigger rating-based rewards/penalties
	// 4. Update driver profile

	return nil
}

// GoOnline sets the driver status to online
func (da *DriverActor) GoOnline(system *ActorSystem, lat, lng float64) error {
	if da.driver.Status == models.DriverStatusOnline {
		return fmt.Errorf("driver is already online")
	}

	da.driver.Status = models.DriverStatusOnline
	da.driver.CurrentLatitude = &lat
	da.driver.CurrentLongitude = &lng

	payload := GoOnlinePayload{
		Lat:       lat,
		Lng:       lng,
		Timestamp: time.Now(),
	}

	message := NewBaseMessage(MsgTypeGoOnline, payload, da.GetID())

	// Notify location service
	if err := system.SendMessage("location-service", message); err != nil {
		return fmt.Errorf("failed to send go online message: %w", err)
	}

	da.logger.WithFields(logrus.Fields{
		"lat": lat,
		"lng": lng,
	}).Info("Driver went online")

	return nil
}

// GoOffline sets the driver status to offline
func (da *DriverActor) GoOffline(system *ActorSystem, reason string) error {
	if da.driver.Status == models.DriverStatusOffline {
		return fmt.Errorf("driver is already offline")
	}

	da.driver.Status = models.DriverStatusOffline

	payload := GoOfflinePayload{
		Reason:    reason,
		Timestamp: time.Now(),
	}

	message := NewBaseMessage(MsgTypeGoOffline, payload, da.GetID())

	// Notify location service
	if err := system.SendMessage("location-service", message); err != nil {
		return fmt.Errorf("failed to send go offline message: %w", err)
	}

	da.logger.WithField("reason", reason).Info("Driver went offline")

	return nil
}

// acceptRideRequest sends an accept ride message
func (da *DriverActor) acceptRideRequest(tripID string) error {
	da.driver.Status = models.DriverStatusBusy

	// Here you would send to trip management service
	// For now, just log the acceptance
	da.logger.WithField("trip_id", tripID).Info("Ride request accepted")

	return nil
}

// rejectRideRequest sends a reject ride message
func (da *DriverActor) rejectRideRequest(tripID, reason string) error {
	// Here you would send to trip management service
	// For now, just log the rejection
	da.logger.WithFields(logrus.Fields{
		"trip_id": tripID,
		"reason":  reason,
	}).Info("Ride request rejected")

	return nil
}

// StartRide notifies that the ride has started
func (da *DriverActor) StartRide(system *ActorSystem, tripID string, startLat, startLng float64) error {
	payload := StartRidePayload{
		TripID:    tripID,
		StartedAt: time.Now(),
		StartLat:  startLat,
		StartLng:  startLng,
	}

	message := NewBaseMessage(MsgTypeStartRide, payload, da.GetID())

	// Send to trip management service
	if err := system.SendMessage("trip-manager", message); err != nil {
		return fmt.Errorf("failed to send start ride message: %w", err)
	}

	da.logger.WithFields(logrus.Fields{
		"trip_id":   tripID,
		"start_lat": startLat,
		"start_lng": startLng,
	}).Info("Ride started")

	return nil
}

// CompleteRide notifies that the ride has been completed
func (da *DriverActor) CompleteRide(system *ActorSystem, tripID string, endLat, endLng, distance, fare float64, duration time.Duration) error {
	da.driver.Status = models.DriverStatusOnline // Back to online after completing ride
	da.driver.TotalTrips++

	payload := CompleteRidePayload{
		TripID:      tripID,
		EndLat:      endLat,
		EndLng:      endLng,
		Distance:    distance,
		Duration:    duration,
		Fare:        fare,
		CompletedAt: time.Now(),
	}

	message := NewBaseMessage(MsgTypeCompleteRide, payload, da.GetID())

	// Send to trip management service
	if err := system.SendMessage("trip-manager", message); err != nil {
		return fmt.Errorf("failed to send complete ride message: %w", err)
	}

	da.logger.WithFields(logrus.Fields{
		"trip_id":  tripID,
		"end_lat":  endLat,
		"end_lng":  endLng,
		"distance": distance,
		"fare":     fare,
		"duration": duration,
	}).Info("Ride completed")

	return nil
}

// UpdateLocation sends location updates
func (da *DriverActor) UpdateLocation(system *ActorSystem, lat, lng, heading, speed float64) error {
	da.driver.CurrentLatitude = &lat
	da.driver.CurrentLongitude = &lng

	payload := DriverLocationPayload{
		Lat:       lat,
		Lng:       lng,
		Heading:   heading,
		Speed:     speed,
		Timestamp: time.Now(),
	}

	message := NewBaseMessage(MsgTypeDriverLocation, payload, da.GetID())

	// Send to location service
	if err := system.SendMessage("location-service", message); err != nil {
		return fmt.Errorf("failed to send location update: %w", err)
	}

	return nil
}

// GetDriver returns the driver model
func (da *DriverActor) GetDriver() *models.Driver {
	return da.driver
}

// unmarshalPayload is a helper function to unmarshal message payloads
func (da *DriverActor) unmarshalPayload(payload interface{}, target interface{}) error {
	// Convert payload to JSON bytes first
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Unmarshal into target struct
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}