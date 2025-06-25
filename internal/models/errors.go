package models

import (
	"errors"
	"fmt"
)

// User validation errors
var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPhone    = errors.New("invalid phone number")
	ErrInvalidName     = errors.New("invalid name")
	ErrInvalidUserType = errors.New("invalid user type")
	ErrInvalidUserID   = errors.New("invalid user ID")
)

// Driver validation errors
var (
	ErrInvalidLicenseNumber = errors.New("invalid license number")
	ErrInvalidVehicleType   = errors.New("invalid vehicle type")
	ErrInvalidVehiclePlate  = errors.New("invalid vehicle plate")
	ErrInvalidDriverStatus  = errors.New("invalid driver status")
	ErrInvalidRating        = errors.New("invalid rating")
)

// Trip validation errors
var (
	ErrInvalidPassengerID         = errors.New("invalid passenger ID")
	ErrInvalidDriverID            = errors.New("invalid driver ID")
	ErrInvalidPickupLocation      = errors.New("invalid pickup location")
	ErrInvalidDestinationLocation = errors.New("invalid destination location")
	ErrInvalidFareAmount          = errors.New("invalid fare amount")
	ErrInvalidDistance            = errors.New("invalid distance")
	ErrInvalidDuration            = errors.New("invalid duration")
	ErrInvalidTripStatus          = errors.New("invalid trip status")
	ErrInvalidStatusTransition    = errors.New("invalid status transition")
	ErrInvalidRideType            = errors.New("invalid ride type")
)

// Business logic errors
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrDriverNotFound        = errors.New("driver not found")
	ErrPassengerNotFound     = errors.New("passenger not found")
	ErrTripNotFound          = errors.New("trip not found")
	ErrDriverNotAvailable    = errors.New("driver not available")
	ErrNoDriversAvailable    = errors.New("no drivers available")
	ErrTripAlreadyAssigned   = errors.New("trip already assigned")
	ErrTripNotAssigned       = errors.New("trip not assigned")
	ErrTripAlreadyCompleted  = errors.New("trip already completed")
	ErrTripAlreadyCancelled  = errors.New("trip already cancelled")
	ErrUnauthorizedOperation = errors.New("unauthorized operation")
)

// Actor system errors
var (
	ErrActorNotFound         = errors.New("actor not found")
	ErrActorAlreadyExists    = errors.New("actor already exists")
	ErrActorNotActive        = errors.New("actor not active")
	ErrMessageDeliveryFailed = errors.New("message delivery failed")
	ErrMessageTimeout        = errors.New("message timeout")
	ErrInvalidMessage        = errors.New("invalid message")
	ErrActorSystemShutdown   = errors.New("actor system shutdown")
)

// Observability errors
var (
	ErrTraceNotFound      = errors.New("trace not found")
	ErrInvalidTraceID     = errors.New("invalid trace ID")
	ErrInvalidSpanID      = errors.New("invalid span ID")
	ErrMetricNotFound     = errors.New("metric not found")
	ErrInvalidMetricValue = errors.New("invalid metric value")
)

// Database errors
var (
	ErrDatabaseConnection  = errors.New("database connection failed")
	ErrDatabaseTransaction = errors.New("database transaction failed")
	ErrDuplicateEntry      = errors.New("duplicate entry")
	ErrRecordNotFound      = errors.New("record not found")
)

// Configuration errors
var (
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrMissingConfiguration = errors.New("missing configuration")
)

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}
