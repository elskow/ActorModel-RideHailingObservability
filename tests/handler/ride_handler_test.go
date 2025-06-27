package handler

import (
	"actor-model-observability/tests/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"actor-model-observability/internal/handlers"
	"actor-model-observability/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRideHandler_RequestRide_Success tests successful ride request
func TestRideHandler_RequestRide_Success(t *testing.T) {
	handler, mockService := utils.SetupRideHandler()

	// Setup test data
	passengerID := uuid.New()
	tripID := uuid.New()
	pickup := models.Location{Latitude: 37.7749, Longitude: -122.4194}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	// Expected trip response
	expectedTrip := &models.Trip{
		ID:                   tripID,
		PassengerID:          passengerID,
		PickupLatitude:       pickup.Latitude,
		PickupLongitude:      pickup.Longitude,
		DestinationLatitude:  dropoff.Latitude,
		DestinationLongitude: dropoff.Longitude,
		Status:               models.TripStatusRequested,
		RequestedAt:          time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Setup mock expectations
	mockService.On("RequestRide", mock.Anything, passengerID.String(), pickup, dropoff, "", "").Return(expectedTrip, nil)

	// Create request
	requestBody := handlers.RequestRideRequest{
		PassengerID:    passengerID,
		PickupLat:      pickup.Latitude,
		PickupLng:      pickup.Longitude,
		DestinationLat: dropoff.Latitude,
		DestinationLng: dropoff.Longitude,
		RideType:       "standard",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.RequestRide(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response handlers.RequestRideResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, tripID, response.TripID)
	assert.Equal(t, string(models.TripStatusRequested), response.Status)
	assert.Greater(t, response.EstimatedFare, 0.0)
	assert.Equal(t, "Ride request created successfully", response.Message)

	mockService.AssertExpectations(t)
}

// TestRideHandler_RequestRide_InvalidJSON tests request with invalid JSON
func TestRideHandler_RequestRide_InvalidJSON(t *testing.T) {
	handler, _ := utils.SetupRideHandler()

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/request", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.RequestRide(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestRideHandler_RequestRide_MissingFields tests request with missing required fields
func TestRideHandler_RequestRide_MissingFields(t *testing.T) {
	handler, _ := utils.SetupRideHandler()

	// Create request with missing fields
	requestBody := map[string]interface{}{
		"pickup_lat": 37.7749,
		// Missing other required fields
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.RequestRide(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestRideHandler_RequestRide_ServiceError tests service error handling
func TestRideHandler_RequestRide_ServiceError(t *testing.T) {
	handler, mockService := utils.SetupRideHandler()

	// Setup test data
	passengerID := uuid.New()
	pickup := models.Location{Latitude: 37.7749, Longitude: -122.4194}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	// Setup mock to return error
	mockService.On("RequestRide", mock.Anything, passengerID.String(), pickup, dropoff, "", "").Return((*models.Trip)(nil), &models.NotFoundError{Resource: "Passenger", ID: passengerID.String()})

	// Create request
	requestBody := handlers.RequestRideRequest{
		PassengerID:    passengerID,
		PickupLat:      pickup.Latitude,
		PickupLng:      pickup.Longitude,
		DestinationLat: dropoff.Latitude,
		DestinationLng: dropoff.Longitude,
		RideType:       "standard",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.RequestRide(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Resource not found", response.Error)

	mockService.AssertExpectations(t)
}

// TestRideHandler_CancelRide_Success tests successful ride cancellation
func TestRideHandler_CancelRide_Success(t *testing.T) {
	handler, mockService := utils.SetupRideHandler()

	// Setup test data
	tripID := uuid.New()
	passengerID := uuid.New()
	reason := "Changed my mind"

	// Setup mock expectations
	mockService.On("CancelRide", mock.Anything, tripID.String(), reason).Return(nil)

	// Create request
	requestBody := handlers.CancelRideRequest{
		TripID:      tripID,
		PassengerID: passengerID,
		Reason:      reason,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.CancelRide(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.CancelRideResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, tripID, response.TripID)
	assert.Equal(t, "cancelled", response.Status)
	assert.Equal(t, "Ride cancelled successfully", response.Message)

	mockService.AssertExpectations(t)
}

// TestRideHandler_GetRideStatus_Success tests successful ride status retrieval
func TestRideHandler_GetRideStatus_Success(t *testing.T) {
	handler, mockService := utils.SetupRideHandler()

	// Setup test data
	tripID := uuid.New()
	passengerID := uuid.New()

	expectedTrip := &models.Trip{
		ID:          tripID,
		PassengerID: passengerID,
		Status:      models.TripStatusInProgress,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Setup mock expectations
	mockService.On("GetTripStatus", mock.Anything, tripID.String()).Return(expectedTrip, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/rides/%s/status", tripID.String()), nil)
	w := httptest.NewRecorder()

	// Setup Gin context with URL parameter
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: tripID.String()}}

	// Execute handler
	handler.GetRideStatus(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Trip
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, tripID, response.ID)
	assert.Equal(t, passengerID, response.PassengerID)
	assert.Equal(t, models.TripStatusInProgress, response.Status)

	mockService.AssertExpectations(t)
}

// TestRideHandler_GetRideStatus_InvalidUUID tests invalid trip ID
func TestRideHandler_GetRideStatus_InvalidUUID(t *testing.T) {
	handler, _ := utils.SetupRideHandler()

	// Create request with invalid UUID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rides/invalid-uuid/status", nil)
	w := httptest.NewRecorder()

	// Setup Gin context with invalid UUID parameter
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	// Execute handler
	handler.GetRideStatus(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid trip ID", response.Error)
}

// TestRideHandler_GetRideStatus_NotFound tests trip not found
func TestRideHandler_GetRideStatus_NotFound(t *testing.T) {
	handler, mockService := utils.SetupRideHandler()

	// Setup test data
	tripID := uuid.New()

	// Setup mock to return not found error
	mockService.On("GetTripStatus", mock.Anything, tripID.String()).Return((*models.Trip)(nil), &models.NotFoundError{Resource: "Trip", ID: tripID.String()})

	// Create request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/rides/%s/status", tripID.String()), nil)
	w := httptest.NewRecorder()

	// Setup Gin context with URL parameter
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: tripID.String()}}

	// Execute handler
	handler.GetRideStatus(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Trip not found", response.Error)

	mockService.AssertExpectations(t)
}

// TestRideHandler_ListRides_Success tests successful ride listing
func TestRideHandler_ListRides_Success(t *testing.T) {
	handler, mockService := utils.SetupRideHandler()

	// Setup test data
	trip1 := &models.Trip{
		ID:          uuid.New(),
		PassengerID: uuid.New(),
		Status:      models.TripStatusCompleted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	trip2 := &models.Trip{
		ID:          uuid.New(),
		PassengerID: uuid.New(),
		Status:      models.TripStatusInProgress,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	expectedTrips := []*models.Trip{trip1, trip2}
	expectedTotal := int64(2)

	// Setup mock expectations
	mockService.On("ListRides", mock.Anything, (*string)(nil), (*string)(nil), (*string)(nil), 10, 0).Return(expectedTrips, expectedTotal, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rides", nil)
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.ListRides(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, expectedTotal, response.Total)
	assert.False(t, response.HasMore)

	// Verify the data is an array (it will be unmarshaled as []interface{})
	data, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)

	mockService.AssertExpectations(t)
}

// TestRideHandler_ListRides_InvalidLimit tests invalid limit parameter
func TestRideHandler_ListRides_InvalidLimit(t *testing.T) {
	handler, _ := utils.SetupRideHandler()

	// Create request with invalid limit
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rides?limit=invalid", nil)
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.ListRides(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid limit", response.Error)
}

// TestRideHandler_ListRides_InvalidOffset tests invalid offset parameter
func TestRideHandler_ListRides_InvalidOffset(t *testing.T) {
	handler, _ := utils.SetupRideHandler()

	// Create request with invalid offset
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rides?offset=-1", nil)
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.ListRides(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid offset", response.Error)
}
