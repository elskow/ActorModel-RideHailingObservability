package actor

import (
	"encoding/json"
	"fmt"
	"time"

	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/models"
)

// PassengerActor handles passenger-related operations
type PassengerActor struct {
	*BaseActor
	passenger *models.Passenger
	logger    *logging.Logger
}

// Passenger message types
const (
	MsgTypeRequestRide    = "request_ride"
	MsgTypeCancelRide     = "cancel_ride"
	MsgTypeRideMatched    = "ride_matched"
	MsgTypeRideStarted    = "ride_started"
	MsgTypeRideCompleted  = "ride_completed"
	MsgTypeRideCancelled  = "ride_cancelled"
	MsgTypeUpdateLocation = "update_location"
	MsgTypeRateDriver     = "rate_driver"
)

// Passenger message payloads
type RequestRidePayload struct {
	PickupLat   float64   `json:"pickup_lat"`
	PickupLng   float64   `json:"pickup_lng"`
	DropoffLat  float64   `json:"dropoff_lat"`
	DropoffLng  float64   `json:"dropoff_lng"`
	PickupAddr  string    `json:"pickup_address"`
	DropoffAddr string    `json:"dropoff_address"`
	RequestedAt time.Time `json:"requested_at"`
}

type CancelRidePayload struct {
	TripID      string    `json:"trip_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

type RideMatchedPayload struct {
	TripID       string        `json:"trip_id"`
	DriverID     string        `json:"driver_id"`
	DriverName   string        `json:"driver_name"`
	DriverPhone  string        `json:"driver_phone"`
	VehicleInfo  string        `json:"vehicle_info"`
	DriverLat    float64       `json:"driver_lat"`
	DriverLng    float64       `json:"driver_lng"`
	EstimatedETA time.Duration `json:"estimated_eta"`
	MatchedAt    time.Time     `json:"matched_at"`
}

type RideStartedPayload struct {
	TripID    string    `json:"trip_id"`
	StartedAt time.Time `json:"started_at"`
}

type RideCompletedPayload struct {
	TripID      string        `json:"trip_id"`
	Fare        float64       `json:"fare"`
	Distance    float64       `json:"distance"`
	Duration    time.Duration `json:"duration"`
	CompletedAt time.Time     `json:"completed_at"`
}

type RideCancelledPayload struct {
	TripID      string    `json:"trip_id"`
	Reason      string    `json:"reason"`
	CancelledBy string    `json:"cancelled_by"`
	CancelledAt time.Time `json:"cancelled_at"`
}

type UpdateLocationPayload struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
}

type RateDriverPayload struct {
	TripID   string    `json:"trip_id"`
	DriverID string    `json:"driver_id"`
	Rating   float64   `json:"rating"`
	Comment  string    `json:"comment"`
	RatedAt  time.Time `json:"rated_at"`
}

// NewPassengerActor creates a new passenger actor
func NewPassengerActor(passenger *models.Passenger, system *ActorSystem) (*PassengerActor, error) {
	if passenger == nil {
		return nil, fmt.Errorf("passenger cannot be nil")
	}

	actorID := fmt.Sprintf("passenger-%s", passenger.ID)
	logger := logging.GetGlobalLogger().WithActor(actorID, "passenger").WithFields(logging.Fields{
		"passenger_id": passenger.ID,
		"user_id":      passenger.UserID,
	})

	pa := &PassengerActor{
		passenger: passenger,
		logger:    logger,
	}

	// Create base actor with message handler
	baseActor := NewBaseActor(actorID, "passenger", 100, pa.handleMessage)
	pa.BaseActor = baseActor

	return pa, nil
}

// handleMessage processes incoming messages
func (pa *PassengerActor) handleMessage(message Message) error {
	pa.logger.WithMessage(message.GetID(), message.GetType(), message.GetSender(), pa.GetID()).Debug("Processing passenger message")

	switch message.GetType() {
	case MsgTypeRideMatched:
		return pa.handleRideMatched(message)
	case MsgTypeRideStarted:
		return pa.handleRideStarted(message)
	case MsgTypeRideCompleted:
		return pa.handleRideCompleted(message)
	case MsgTypeRideCancelled:
		return pa.handleRideCancelled(message)
	default:
		pa.logger.WithField("message_type", message.GetType()).Warn("Unknown message type")
		return fmt.Errorf("unknown message type: %s", message.GetType())
	}
}

// handleRideMatched processes ride matched notifications
func (pa *PassengerActor) handleRideMatched(message Message) error {
	var payload RideMatchedPayload
	if err := pa.unmarshalPayload(message.GetPayload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ride matched payload: %w", err)
	}

	pa.logger.WithFields(logging.Fields{
		"trip_id":       payload.TripID,
		"driver_id":     payload.DriverID,
		"driver_name":   payload.DriverName,
		"estimated_eta": payload.EstimatedETA,
	}).Info("Ride matched with driver")

	// Here you could:
	// 1. Send push notification to passenger
	// 2. Update passenger's current trip status
	// 3. Start tracking driver location
	// 4. Log the event for analytics

	return nil
}

// handleRideStarted processes ride started notifications
func (pa *PassengerActor) handleRideStarted(message Message) error {
	var payload RideStartedPayload
	if err := pa.unmarshalPayload(message.GetPayload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ride started payload: %w", err)
	}

	pa.logger.WithFields(logging.Fields{
		"trip_id":    payload.TripID,
		"started_at": payload.StartedAt,
	}).Info("Ride started")

	// Here you could:
	// 1. Start trip timer
	// 2. Begin real-time location tracking
	// 3. Send notification to passenger
	// 4. Update trip status in database

	return nil
}

// handleRideCompleted processes ride completion notifications
func (pa *PassengerActor) handleRideCompleted(message Message) error {
	var payload RideCompletedPayload
	if err := pa.unmarshalPayload(message.GetPayload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ride completed payload: %w", err)
	}

	pa.logger.WithFields(logging.Fields{
		"trip_id":      payload.TripID,
		"fare":         payload.Fare,
		"distance":     payload.Distance,
		"duration":     payload.Duration,
		"completed_at": payload.CompletedAt,
	}).Info("Ride completed")

	// Update passenger statistics
	pa.passenger.TotalTrips++

	// Here you could:
	// 1. Process payment
	// 2. Send receipt to passenger
	// 3. Prompt for driver rating
	// 4. Update passenger's trip history
	// 5. Calculate loyalty points

	return nil
}

// handleRideCancelled processes ride cancellation notifications
func (pa *PassengerActor) handleRideCancelled(message Message) error {
	var payload RideCancelledPayload
	if err := pa.unmarshalPayload(message.GetPayload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ride cancelled payload: %w", err)
	}

	pa.logger.WithFields(logging.Fields{
		"trip_id":      payload.TripID,
		"reason":       payload.Reason,
		"cancelled_by": payload.CancelledBy,
		"cancelled_at": payload.CancelledAt,
	}).Info("Ride cancelled")

	// Here you could:
	// 1. Refund any pre-authorization
	// 2. Send cancellation notification
	// 3. Suggest alternative rides
	// 4. Update cancellation statistics

	return nil
}

// RequestRide sends a ride request message to the trip matching system
func (pa *PassengerActor) RequestRide(system *ActorSystem, pickup, dropoff models.Location, pickupAddr, dropoffAddr string) error {
	payload := RequestRidePayload{
		PickupLat:   pickup.Latitude,
		PickupLng:   pickup.Longitude,
		DropoffLat:  dropoff.Latitude,
		DropoffLng:  dropoff.Longitude,
		PickupAddr:  pickupAddr,
		DropoffAddr: dropoffAddr,
		RequestedAt: time.Now(),
	}

	message := NewBaseMessage(MsgTypeRequestRide, payload, pa.GetID())

	// Send to trip matching service
	if err := system.SendMessage("trip-matcher", message); err != nil {
		return fmt.Errorf("failed to send ride request: %w", err)
	}

	pa.logger.WithFields(logging.Fields{
		"pickup_lat":   pickup.Latitude,
		"pickup_lng":   pickup.Longitude,
		"dropoff_lat":  dropoff.Latitude,
		"dropoff_lng":  dropoff.Longitude,
		"pickup_addr":  pickupAddr,
		"dropoff_addr": dropoffAddr,
	}).Info("Ride request sent")

	return nil
}

// CancelRide sends a ride cancellation message
func (pa *PassengerActor) CancelRide(system *ActorSystem, tripID, reason string) error {
	payload := CancelRidePayload{
		TripID:      tripID,
		Reason:      reason,
		CancelledAt: time.Now(),
	}

	message := NewBaseMessage(MsgTypeCancelRide, payload, pa.GetID())

	// Send to trip management service
	if err := system.SendMessage("trip-manager", message); err != nil {
		return fmt.Errorf("failed to send cancellation request: %w", err)
	}

	pa.logger.WithFields(logging.Fields{
		"trip_id": tripID,
		"reason":  reason,
	}).Info("Ride cancellation sent")

	return nil
}

// RateDriver sends a driver rating message
func (pa *PassengerActor) RateDriver(system *ActorSystem, tripID, driverID string, rating float64, comment string) error {
	if rating < 1.0 || rating > 5.0 {
		return fmt.Errorf("rating must be between 1.0 and 5.0")
	}

	payload := RateDriverPayload{
		TripID:   tripID,
		DriverID: driverID,
		Rating:   rating,
		Comment:  comment,
		RatedAt:  time.Now(),
	}

	message := NewBaseMessage(MsgTypeRateDriver, payload, pa.GetID())

	// Send to rating service
	if err := system.SendMessage("rating-service", message); err != nil {
		return fmt.Errorf("failed to send driver rating: %w", err)
	}

	pa.logger.WithFields(logging.Fields{
		"trip_id":   tripID,
		"driver_id": driverID,
		"rating":    rating,
		"comment":   comment,
	}).Info("Driver rating sent")

	return nil
}

// GetPassenger returns the passenger model
func (pa *PassengerActor) GetPassenger() *models.Passenger {
	return pa.passenger
}

// unmarshalPayload is a helper function to unmarshal message payloads
func (pa *PassengerActor) unmarshalPayload(payload interface{}, target interface{}) error {
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
