package models

import (
	"time"

	"github.com/google/uuid"
)

// TripStatus represents the status of a trip
type TripStatus string

const (
	TripStatusRequested      TripStatus = "requested"
	TripStatusMatched        TripStatus = "matched"
	TripStatusAccepted       TripStatus = "accepted"
	TripStatusDriverArrived  TripStatus = "driver_arrived"
	TripStatusInProgress     TripStatus = "in_progress"
	TripStatusCompleted      TripStatus = "completed"
	TripStatusCancelled      TripStatus = "cancelled"
)

// Trip represents a trip in the system
type Trip struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PassengerID          uuid.UUID  `json:"passenger_id" gorm:"type:uuid;not null;index"`
	Passenger            *Passenger `json:"passenger,omitempty" gorm:"foreignKey:PassengerID"`
	DriverID             *uuid.UUID `json:"driver_id" gorm:"type:uuid;index"`
	Driver               *Driver    `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
	PickupLatitude       float64    `json:"pickup_latitude" gorm:"type:decimal(10,8);not null"`
	PickupLongitude      float64    `json:"pickup_longitude" gorm:"type:decimal(11,8);not null"`
	PickupAddress        *string    `json:"pickup_address"`
	DestinationLatitude  float64    `json:"destination_latitude" gorm:"type:decimal(10,8);not null"`
	DestinationLongitude float64    `json:"destination_longitude" gorm:"type:decimal(11,8);not null"`
	DestinationAddress   *string    `json:"destination_address"`
	Status               TripStatus `json:"status" gorm:"default:'requested';check:status IN ('requested', 'matched', 'accepted', 'driver_arrived', 'in_progress', 'completed', 'cancelled')"`
	FareAmount           *float64   `json:"fare_amount" gorm:"type:decimal(10,2)"`
	DistanceKm           *float64   `json:"distance_km" gorm:"type:decimal(8,2)"`
	DurationMinutes      *int       `json:"duration_minutes"`
	RequestedAt          time.Time  `json:"requested_at" gorm:"default:CURRENT_TIMESTAMP"`
	MatchedAt            *time.Time `json:"matched_at"`
	AcceptedAt           *time.Time `json:"accepted_at"`
	PickupAt             *time.Time `json:"pickup_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	CancelledAt          *time.Time `json:"cancelled_at"`
	CreatedAt            time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for Trip
func (Trip) TableName() string {
	return "trips"
}

// IsActive returns true if the trip is in an active state
func (t *Trip) IsActive() bool {
	return t.Status != TripStatusCompleted && t.Status != TripStatusCancelled
}

// IsCompleted returns true if the trip is completed
func (t *Trip) IsCompleted() bool {
	return t.Status == TripStatusCompleted
}

// IsCancelled returns true if the trip is cancelled
func (t *Trip) IsCancelled() bool {
	return t.Status == TripStatusCancelled
}

// HasDriver returns true if the trip has been assigned a driver
func (t *Trip) HasDriver() bool {
	return t.DriverID != nil
}

// GetPickupLocation returns the pickup location coordinates
func (t *Trip) GetPickupLocation() (lat, lng float64) {
	return t.PickupLatitude, t.PickupLongitude
}

// GetDestinationLocation returns the destination location coordinates
func (t *Trip) GetDestinationLocation() (lat, lng float64) {
	return t.DestinationLatitude, t.DestinationLongitude
}

// SetStatus sets the trip status and updates the corresponding timestamp
func (t *Trip) SetStatus(status TripStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
	
	now := time.Now()
	switch status {
	case TripStatusMatched:
		t.MatchedAt = &now
	case TripStatusAccepted:
		t.AcceptedAt = &now
	case TripStatusDriverArrived:
		// Driver arrived timestamp can be set separately
	case TripStatusInProgress:
		t.PickupAt = &now
	case TripStatusCompleted:
		t.CompletedAt = &now
	case TripStatusCancelled:
		t.CancelledAt = &now
	}
}

// AssignDriver assigns a driver to the trip
func (t *Trip) AssignDriver(driverID uuid.UUID) {
	t.DriverID = &driverID
	t.SetStatus(TripStatusMatched)
}

// Accept marks the trip as accepted by the driver
func (t *Trip) Accept() {
	t.SetStatus(TripStatusAccepted)
}

// DriverArrived marks that the driver has arrived at pickup location
func (t *Trip) DriverArrived() {
	t.SetStatus(TripStatusDriverArrived)
}

// StartTrip marks the trip as in progress
func (t *Trip) StartTrip() {
	t.SetStatus(TripStatusInProgress)
}

// CompleteTrip marks the trip as completed with fare and duration
func (t *Trip) CompleteTrip(fareAmount float64, distanceKm float64, durationMinutes int) {
	t.FareAmount = &fareAmount
	t.DistanceKm = &distanceKm
	t.DurationMinutes = &durationMinutes
	t.SetStatus(TripStatusCompleted)
}

// Cancel marks the trip as cancelled
func (t *Trip) Cancel() {
	t.SetStatus(TripStatusCancelled)
}

// GetDuration returns the trip duration if completed
func (t *Trip) GetDuration() time.Duration {
	if t.CompletedAt != nil && t.PickupAt != nil {
		return t.CompletedAt.Sub(*t.PickupAt)
	}
	if t.PickupAt != nil {
		return time.Since(*t.PickupAt)
	}
	return 0
}

// GetWaitTime returns the time between request and pickup
func (t *Trip) GetWaitTime() time.Duration {
	if t.PickupAt != nil {
		return t.PickupAt.Sub(t.RequestedAt)
	}
	if t.IsActive() {
		return time.Since(t.RequestedAt)
	}
	return 0
}

// CanTransitionTo checks if the trip can transition to the given status
func (t *Trip) CanTransitionTo(newStatus TripStatus) bool {
	switch t.Status {
	case TripStatusRequested:
		return newStatus == TripStatusMatched || newStatus == TripStatusCancelled
	case TripStatusMatched:
		return newStatus == TripStatusAccepted || newStatus == TripStatusCancelled
	case TripStatusAccepted:
		return newStatus == TripStatusDriverArrived || newStatus == TripStatusCancelled
	case TripStatusDriverArrived:
		return newStatus == TripStatusInProgress || newStatus == TripStatusCancelled
	case TripStatusInProgress:
		return newStatus == TripStatusCompleted || newStatus == TripStatusCancelled
	case TripStatusCompleted, TripStatusCancelled:
		return false // Terminal states
	default:
		return false
	}
}

// Validate validates the trip data
func (t *Trip) Validate() error {
	if t.PassengerID == uuid.Nil {
		return ErrInvalidPassengerID
	}
	if t.PickupLatitude < -90 || t.PickupLatitude > 90 {
		return ErrInvalidPickupLocation
	}
	if t.PickupLongitude < -180 || t.PickupLongitude > 180 {
		return ErrInvalidPickupLocation
	}
	if t.DestinationLatitude < -90 || t.DestinationLatitude > 90 {
		return ErrInvalidDestinationLocation
	}
	if t.DestinationLongitude < -180 || t.DestinationLongitude > 180 {
		return ErrInvalidDestinationLocation
	}
	if t.FareAmount != nil && *t.FareAmount < 0 {
		return ErrInvalidFareAmount
	}
	if t.DistanceKm != nil && *t.DistanceKm < 0 {
		return ErrInvalidDistance
	}
	if t.DurationMinutes != nil && *t.DurationMinutes < 0 {
		return ErrInvalidDuration
	}
	return nil
}