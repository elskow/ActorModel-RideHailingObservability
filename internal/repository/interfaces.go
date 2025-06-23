package repository

import (
	"context"

	"actor-model-observability/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

// DriverRepository defines the interface for driver data operations
type DriverRepository interface {
	Create(ctx context.Context, driver *models.Driver) error
	GetByID(ctx context.Context, id string) (*models.Driver, error)
	GetByUserID(ctx context.Context, userID string) (*models.Driver, error)
	Update(ctx context.Context, driver *models.Driver) error
	Delete(ctx context.Context, id string) error
	GetOnlineDrivers(ctx context.Context) ([]*models.Driver, error)
	GetDriversInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]*models.Driver, error)
	UpdateLocation(ctx context.Context, driverID string, lat, lng float64) error
	UpdateStatus(ctx context.Context, driverID string, status models.DriverStatus) error
	List(ctx context.Context, limit, offset int) ([]*models.Driver, error)
}

// PassengerRepository defines the interface for passenger data operations
type PassengerRepository interface {
	Create(ctx context.Context, passenger *models.Passenger) error
	GetByID(ctx context.Context, id string) (*models.Passenger, error)
	GetByUserID(ctx context.Context, userID string) (*models.Passenger, error)
	Update(ctx context.Context, passenger *models.Passenger) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Passenger, error)
}

// TripRepository defines the interface for trip data operations
type TripRepository interface {
	Create(ctx context.Context, trip *models.Trip) error
	GetByID(ctx context.Context, id string) (*models.Trip, error)
	Update(ctx context.Context, trip *models.Trip) error
	Delete(ctx context.Context, id string) error
	GetByPassengerID(ctx context.Context, passengerID string, limit, offset int) ([]*models.Trip, error)
	GetByDriverID(ctx context.Context, driverID string, limit, offset int) ([]*models.Trip, error)
	GetActiveTrips(ctx context.Context) ([]*models.Trip, error)
	GetTripsByStatus(ctx context.Context, status models.TripStatus, limit, offset int) ([]*models.Trip, error)
	GetTripsByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Trip, error)
	List(ctx context.Context, limit, offset int) ([]*models.Trip, error)
}

// ObservabilityRepository defines the interface for observability data operations
type ObservabilityRepository interface {
	// Actor Instances
	CreateActorInstance(ctx context.Context, instance *models.ActorInstance) error
	GetActorInstance(ctx context.Context, id string) (*models.ActorInstance, error)
	UpdateActorInstance(ctx context.Context, instance *models.ActorInstance) error
	ListActorInstances(ctx context.Context, actorType string, limit, offset int) ([]*models.ActorInstance, error)

	// Actor Messages
	CreateActorMessage(ctx context.Context, message *models.ActorMessage) error
	GetActorMessage(ctx context.Context, id string) (*models.ActorMessage, error)
	ListActorMessages(ctx context.Context, fromActor, toActor string, limit, offset int) ([]*models.ActorMessage, error)
	GetMessagesByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.ActorMessage, error)

	// System Metrics
	CreateSystemMetric(ctx context.Context, metric *models.SystemMetric) error
	GetSystemMetric(ctx context.Context, id string) (*models.SystemMetric, error)
	ListSystemMetrics(ctx context.Context, metricType string, limit, offset int) ([]*models.SystemMetric, error)
	GetMetricsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.SystemMetric, error)

	// Distributed Traces
	CreateDistributedTrace(ctx context.Context, trace *models.DistributedTrace) error
	GetDistributedTrace(ctx context.Context, id string) (*models.DistributedTrace, error)
	GetTracesByTraceID(ctx context.Context, traceID string) ([]*models.DistributedTrace, error)
	ListDistributedTraces(ctx context.Context, operation string, limit, offset int) ([]*models.DistributedTrace, error)

	// Event Logs
	CreateEventLog(ctx context.Context, log *models.EventLog) error
	GetEventLog(ctx context.Context, id string) (*models.EventLog, error)
	ListEventLogs(ctx context.Context, eventType, source string, limit, offset int) ([]*models.EventLog, error)
	GetEventLogsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.EventLog, error)
}

// TraditionalRepository defines the interface for traditional monitoring data operations
type TraditionalRepository interface {
	// Traditional Metrics
	CreateTraditionalMetric(ctx context.Context, metric *models.TraditionalMetric) error
	GetTraditionalMetric(ctx context.Context, id string) (*models.TraditionalMetric, error)
	ListTraditionalMetrics(ctx context.Context, name, metricType string, limit, offset int) ([]*models.TraditionalMetric, error)
	GetTraditionalMetricsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.TraditionalMetric, error)

	// Traditional Logs
	CreateTraditionalLog(ctx context.Context, log *models.TraditionalLog) error
	GetTraditionalLog(ctx context.Context, id string) (*models.TraditionalLog, error)
	ListTraditionalLogs(ctx context.Context, level, source string, limit, offset int) ([]*models.TraditionalLog, error)
	GetTraditionalLogsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.TraditionalLog, error)
}