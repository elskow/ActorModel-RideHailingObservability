package service

import (
	"context"

	"actor-model-observability/internal/models"
)

// RideServiceInterface defines the interface for ride service operations
type RideServiceInterface interface {
	RequestRide(ctx context.Context, passengerID string, pickup, dropoff models.Location, pickupAddr, dropoffAddr string) (*models.Trip, error)
	CancelRide(ctx context.Context, tripID, reason string) error
	GetTripStatus(ctx context.Context, tripID string) (*models.Trip, error)
	ListRides(ctx context.Context, passengerID, driverID *string, status *string, limit, offset int) ([]*models.Trip, int64, error)
}

// Ensure RideService implements RideServiceInterface
var _ RideServiceInterface = (*RideService)(nil)
