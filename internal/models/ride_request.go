package models

import (
	"time"

	"github.com/google/uuid"
)

// RideRequest represents a ride request in the system
type RideRequest struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PassengerID    uuid.UUID `json:"passenger_id" gorm:"type:uuid;not null;index"`
	PickupLat      float64   `json:"pickup_lat" gorm:"type:decimal(10,8);not null"`
	PickupLng      float64   `json:"pickup_lng" gorm:"type:decimal(11,8);not null"`
	DestinationLat float64   `json:"destination_lat" gorm:"type:decimal(10,8);not null"`
	DestinationLng float64   `json:"destination_lng" gorm:"type:decimal(11,8);not null"`
	RideType       string    `json:"ride_type" gorm:"default:'standard';check:ride_type IN ('standard', 'premium')"`
	Status         string    `json:"status" gorm:"default:'pending';check:status IN ('pending', 'matched', 'cancelled')"`
	CreatedAt      time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for RideRequest
func (RideRequest) TableName() string {
	return "ride_requests"
}

// Validate validates the ride request data
func (r *RideRequest) Validate() error {
	if r.PassengerID == uuid.Nil {
		return ErrInvalidPassengerID
	}
	if r.PickupLat < -90 || r.PickupLat > 90 {
		return ErrInvalidPickupLocation
	}
	if r.PickupLng < -180 || r.PickupLng > 180 {
		return ErrInvalidPickupLocation
	}
	if r.DestinationLat < -90 || r.DestinationLat > 90 {
		return ErrInvalidDestinationLocation
	}
	if r.DestinationLng < -180 || r.DestinationLng > 180 {
		return ErrInvalidDestinationLocation
	}
	if r.RideType != "standard" && r.RideType != "premium" {
		return ErrInvalidRideType
	}
	return nil
}

// GetPickupLocation returns the pickup location coordinates
func (r *RideRequest) GetPickupLocation() (lat, lng float64) {
	return r.PickupLat, r.PickupLng
}

// GetDestinationLocation returns the destination location coordinates
func (r *RideRequest) GetDestinationLocation() (lat, lng float64) {
	return r.DestinationLat, r.DestinationLng
}
