package handlers

import (
	"net/http"
	"strconv"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorResponse and PaginatedResponse are defined in observability_handler.go

// RideHandler handles ride-related HTTP requests
type RideHandler struct {
	rideService *service.RideService
}

// NewRideHandler creates a new RideHandler instance
func NewRideHandler(rideService *service.RideService) *RideHandler {
	return &RideHandler{
		rideService: rideService,
	}
}

// RequestRideRequest represents the request payload for ride requests
type RequestRideRequest struct {
	PassengerID   uuid.UUID `json:"passenger_id" binding:"required"`
	PickupLat     float64   `json:"pickup_lat" binding:"required,min=-90,max=90"`
	PickupLng     float64   `json:"pickup_lng" binding:"required,min=-180,max=180"`
	DestinationLat float64  `json:"destination_lat" binding:"required,min=-90,max=90"`
	DestinationLng float64  `json:"destination_lng" binding:"required,min=-180,max=180"`
	RideType      string    `json:"ride_type" binding:"required,oneof=standard premium"`
}

// RequestRideResponse represents the response for ride requests
type RequestRideResponse struct {
	TripID        uuid.UUID `json:"trip_id"`
	Status        string    `json:"status"`
	EstimatedFare float64   `json:"estimated_fare"`
	Message       string    `json:"message"`
}

// CancelRideRequest represents the request payload for ride cancellation
type CancelRideRequest struct {
	TripID      uuid.UUID `json:"trip_id" binding:"required"`
	PassengerID uuid.UUID `json:"passenger_id" binding:"required"`
	Reason      string    `json:"reason"`
}

// CancelRideResponse represents the response for ride cancellation
type CancelRideResponse struct {
	TripID  uuid.UUID `json:"trip_id"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
}

// RequestRide handles ride request creation
// @Summary Request a ride
// @Description Create a new ride request for a passenger
// @Tags rides
// @Accept json
// @Produce json
// @Param request body RequestRideRequest true "Ride request details"
// @Param approach query string false "Processing approach" Enums(actor, traditional) default(actor)
// @Success 201 {object} RequestRideResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rides/request [post]
func (h *RideHandler) RequestRide(c *gin.Context) {
	var req RequestRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Get processing approach from query parameter
	approach := c.DefaultQuery("approach", "actor")
	if approach != "actor" && approach != "traditional" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid approach",
			Message: "Approach must be either 'actor' or 'traditional'",
		})
		return
	}

	// Create pickup and dropoff locations
	pickup := models.Location{
		Latitude:  req.PickupLat,
		Longitude: req.PickupLng,
	}
	dropoff := models.Location{
		Latitude:  req.DestinationLat,
		Longitude: req.DestinationLng,
	}

	// Request ride using the specified approach
	var trip *models.Trip
	var err error

	if approach == "actor" {
		trip, err = h.rideService.RequestRide(c.Request.Context(), req.PassengerID.String(), pickup, dropoff, "", "")
	} else {
		// For traditional approach, we'll use the same method for now
		trip, err = h.rideService.RequestRide(c.Request.Context(), req.PassengerID.String(), pickup, dropoff, "", "")
	}

	if err != nil {
		// Handle different types of errors
		switch err.(type) {
		case *models.ValidationError:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Validation error",
				Message: err.Error(),
			})
		case *models.NotFoundError:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Resource not found",
				Message: err.Error(),
			})
		// case *models.BusinessLogicError: // Commented out as this type doesn't exist yet
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Error:   "Business logic error",
				Message: err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to process ride request",
			})
		}
		return
	}

	// Return successful response
	c.JSON(http.StatusCreated, RequestRideResponse{
		TripID:        trip.ID,
		Status:        string(trip.Status),
		EstimatedFare: 0.0, // TODO: Calculate estimated fare
		Message:       "Ride request created successfully",
	})
}

// CancelRide handles ride cancellation
// @Summary Cancel a ride
// @Description Cancel an existing ride request
// @Tags rides
// @Accept json
// @Produce json
// @Param request body CancelRideRequest true "Ride cancellation details"
// @Param approach query string false "Processing approach" Enums(actor, traditional) default(actor)
// @Success 200 {object} CancelRideResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rides/cancel [post]
func (h *RideHandler) CancelRide(c *gin.Context) {
	var req CancelRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Get processing approach from query parameter
	approach := c.DefaultQuery("approach", "actor")
	if approach != "actor" && approach != "traditional" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid approach",
			Message: "Approach must be either 'actor' or 'traditional'",
		})
		return
	}

	// Cancel ride using the specified approach
	var err error

	if approach == "actor" {
		err = h.rideService.CancelRide(c.Request.Context(), req.TripID.String(), req.Reason)
	} else {
		// For traditional approach, we'll use the same method for now
		err = h.rideService.CancelRide(c.Request.Context(), req.TripID.String(), req.Reason)
	}

	if err != nil {
		// Handle different types of errors
		switch err.(type) {
		case *models.ValidationError:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Validation error",
				Message: err.Error(),
			})
		case *models.NotFoundError:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Resource not found",
				Message: err.Error(),
			})
		// case *models.BusinessLogicError: // Commented out as this type doesn't exist yet
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Error:   "Business logic error",
				Message: err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to cancel ride",
			})
		}
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, CancelRideResponse{
		TripID:  req.TripID,
		Status:  "cancelled",
		Message: "Ride cancelled successfully",
	})
}

// GetRideStatus handles ride status retrieval
// @Summary Get ride status
// @Description Get the current status of a ride
// @Tags rides
// @Produce json
// @Param trip_id path string true "Trip ID"
// @Success 200 {object} models.Trip
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rides/{trip_id}/status [get]
func (h *RideHandler) GetRideStatus(c *gin.Context) {
	tripIDStr := c.Param("trip_id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid trip ID",
			Message: "Trip ID must be a valid UUID",
		})
		return
	}

	trip, err := h.rideService.GetTripStatus(c.Request.Context(), tripID.String())
	if err != nil {
		switch err.(type) {
		case *models.NotFoundError:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Trip not found",
				Message: err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to get trip status",
			})
		}
		return
	}

	c.JSON(http.StatusOK, trip)
}

// ListRides handles ride listing with pagination
// @Summary List rides
// @Description Get a paginated list of rides with optional filtering
// @Tags rides
// @Produce json
// @Param passenger_id query string false "Filter by passenger ID"
// @Param driver_id query string false "Filter by driver ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.Trip}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rides [get]
func (h *RideHandler) ListRides(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	// TODO: Use these when implementing ListRides
	// passengerIDStr := c.Query("passenger_id")
	// driverIDStr := c.Query("driver_id")
	// status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// TODO: Parse optional UUID parameters when implementing ListRides
	// var passengerID, driverID *uuid.UUID
	// ... UUID parsing logic ...

	// Get rides from service
	// TODO: Implement ListRides method in RideService
	// rides, err := h.rideService.ListRides(c.Request.Context(), passengerID, driverID, status, limit, offset)
	rides := []*models.Trip{} // Placeholder empty slice

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   rides,
		Limit:  limit,
		Offset: offset,
		Total:  int64(len(rides)), // Convert to int64
	})
}