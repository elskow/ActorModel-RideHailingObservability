package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"actor-model-observability/internal/handlers"
	"actor-model-observability/internal/models"
)

// Mock repository implementations
// Mock repositories are defined in benchmark_test.go

// Test UserHandler.GetUser endpoint
func TestUserHandler_GetUser_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Test with valid user ID
	userID := "b412752c-e710-4647-8bdf-252abe290fa1"
	parsedUUID, _ := uuid.Parse(userID)

	// Setup mock expectations
	expectedUser := &models.User{
		ID:       parsedUUID,
		Email:    "test@example.com",
		Phone:    "+1234567890",
		Name:     "Test User",
		UserType: models.UserTypePassenger,
	}
	userRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users/:id", userHandler.GetUser)

	req := httptest.NewRequest("GET", "/users/"+userID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "id")
	assert.Contains(t, response, "email")
	assert.Contains(t, response, "name")
}

func TestUserHandler_GetUser_InvalidUUID(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users/:id", userHandler.GetUser)

	// Test with invalid UUID
	req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Test with non-existent user ID
	nonExistentID := uuid.New().String()

	// Setup mock expectations - return nil user and error for not found
	userRepo.On("GetByID", mock.Anything, nonExistentID).Return((*models.User)(nil), assert.AnError)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users/:id", userHandler.GetUser)

	req := httptest.NewRequest("GET", "/users/"+nonExistentID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Expect error response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Test UserHandler.GetOnlineDrivers endpoint
func TestUserHandler_GetOnlineDrivers_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	lat := 40.7128
	lon := -74.0060
	mockDrivers := []*models.Driver{
		{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			Status:           models.DriverStatusOnline,
			LicenseNumber:    "DL123456",
			VehicleType:      "sedan",
			VehiclePlate:     "ABC123",
			CurrentLatitude:  &lat,
			CurrentLongitude: &lon,
			Rating:           4.5,
		},
	}
	driverRepo.On("GetOnlineDrivers", mock.Anything).Return(mockDrivers, nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/drivers/online", userHandler.GetOnlineDrivers)

	req := httptest.NewRequest("GET", "/drivers/online", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// The response should be an array of drivers
	assert.IsType(t, []interface{}{}, response)
}

// Test UserHandler.CreatePassenger endpoint
func TestUserHandler_CreatePassenger_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	passengerRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Passenger")).Return(nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/passengers", userHandler.CreatePassenger)

	// Create request body
	requestBody := map[string]interface{}{
		"email":     "test@example.com",
		"phone":     "+1234567890",
		"name":      "Test User",
		"user_type": "passenger",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/passengers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "id")
	assert.Contains(t, response, "email")
	assert.Contains(t, response, "name")
	assert.Contains(t, response, "user_type")
	assert.Equal(t, "passenger", response["user_type"])
}

func TestUserHandler_CreatePassenger_InvalidJSON(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/passengers", userHandler.CreatePassenger)

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/passengers", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestUserHandler_CreatePassenger_MissingFields(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/passengers", userHandler.CreatePassenger)

	// Create request body with missing required fields
	requestBody := map[string]interface{}{
		"preferred_language": "en",
		// Missing user_id and payment_method
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/passengers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// Test UserHandler.CreateUser endpoint
func TestUserHandler_CreateUser_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	passengerRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Passenger")).Return(nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/users", userHandler.CreateUser)

	// Create request body
	requestBody := map[string]interface{}{
		"email":     "test@example.com",
		"phone":     "+1234567890",
		"name":      "Test User",
		"user_type": "passenger",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "id")
	assert.Contains(t, response, "email")
	assert.Contains(t, response, "name")
	assert.Equal(t, "test@example.com", response["email"])
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/users", userHandler.CreateUser)

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestUserHandler_CreateUser_MissingFields(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/users", userHandler.CreateUser)

	// Create request body with missing required fields
	requestBody := map[string]interface{}{
		"name": "Test User",
		// Missing email, phone, user_type
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// Test UserHandler.UpdateUser endpoint
func TestUserHandler_UpdateUser_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	existingUser := &models.User{
		ID:       uuid.MustParse("b412752c-e710-4647-8bdf-252abe290fa1"),
		Email:    "test@example.com",
		Phone:    "+1234567890",
		Name:     "Test User",
		UserType: "passenger",
	}
	userRepo.On("GetByID", mock.Anything, "b412752c-e710-4647-8bdf-252abe290fa1").Return(existingUser, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/users/:user_id", userHandler.UpdateUser)

	// Create request body
	requestBody := map[string]interface{}{
		"name":  "Updated User",
		"phone": "+9876543210",
	}

	body, _ := json.Marshal(requestBody)
	userID := "b412752c-e710-4647-8bdf-252abe290fa1"
	req := httptest.NewRequest("PUT", "/users/"+userID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "id")
	assert.Contains(t, response, "name")
}

func TestUserHandler_UpdateUser_InvalidUUID(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/users/:user_id", userHandler.UpdateUser)

	// Create request body
	requestBody := map[string]interface{}{
		"name": "Updated User",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/users/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// Test UserHandler.ListUsers endpoint
func TestUserHandler_ListUsers_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	users := []*models.User{
		{
			ID:       uuid.New(),
			Email:    "user1@example.com",
			Phone:    "+1234567890",
			Name:     "User 1",
			UserType: "passenger",
		},
		{
			ID:       uuid.New(),
			Email:    "user2@example.com",
			Phone:    "+0987654321",
			Name:     "User 2",
			UserType: "driver",
		},
	}
	userRepo.On("List", mock.Anything, 10, 0).Return(users, nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users", userHandler.ListUsers)

	req := httptest.NewRequest("GET", "/users?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// The response should contain pagination metadata
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "limit")
	assert.Contains(t, response, "offset")
	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
}

func TestUserHandler_ListUsers_InvalidLimit(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users", userHandler.ListUsers)

	req := httptest.NewRequest("GET", "/users?limit=invalid&offset=0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// Test UserHandler.GetDriver endpoint
func TestUserHandler_GetDriver_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	expectedDriver := &models.Driver{
		ID:            uuid.New(),
		UserID:        uuid.MustParse("b412752c-e710-4647-8bdf-252abe290fa1"),
		LicenseNumber: "DL123456789",
		VehicleType:   "sedan",
		VehiclePlate:  "ABC123",
		Rating:        4.5,
		Status:        "online",
	}
	driverRepo.On("GetByUserID", mock.Anything, "b412752c-e710-4647-8bdf-252abe290fa1").Return(expectedDriver, nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users/:id/driver", userHandler.GetDriver)

	// Test with valid user ID
	userID := "b412752c-e710-4647-8bdf-252abe290fa1"
	req := httptest.NewRequest("GET", "/users/"+userID+"/driver", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "id")
	assert.Contains(t, response, "user_id")
	assert.Contains(t, response, "status")
}

func TestUserHandler_GetDriver_InvalidUUID(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users/:id/driver", userHandler.GetDriver)

	// Test with invalid UUID
	req := httptest.NewRequest("GET", "/users/invalid-uuid/driver", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// Integration test to verify the actual issue with GetUser endpoint
func TestUserHandler_GetUser_RealScenario(t *testing.T) {
	// This test simulates the actual scenario described in the issue
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	expectedUser := &models.User{
		ID:       uuid.MustParse("b412752c-e710-4647-8bdf-252abe290fa1"),
		Email:    "test@example.com",
		Name:     "Test User",
		Phone:    "+1234567890",
		UserType: "passenger",
	}
	userRepo.On("GetByID", mock.Anything, "b412752c-e710-4647-8bdf-252abe290fa1").Return(expectedUser, nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/users/:id", userHandler.GetUser)

	// Test with the exact user ID from the issue
	userID := "b412752c-e710-4647-8bdf-252abe290fa1"
	req := httptest.NewRequest("GET", "/api/v1/users/"+userID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test UserHandler.UpdateDriverLocation endpoint
func TestUserHandler_UpdateDriverLocation_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	driverRepo.On("UpdateLocation", mock.Anything, "550e8400-e29b-41d4-a716-446655440001", -6.2088, 106.8456).Return(nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/location", userHandler.UpdateDriverLocation)

	// Create request body
	requestBody := map[string]interface{}{
		"latitude":  -6.2088,
		"longitude": 106.8456,
	}

	body, _ := json.Marshal(requestBody)
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/location", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Equal(t, "Driver location updated successfully", response["message"])
}

func TestUserHandler_UpdateDriverLocation_InvalidUUID(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/location", userHandler.UpdateDriverLocation)

	// Create request body
	requestBody := map[string]interface{}{
		"latitude":  -6.2088,
		"longitude": 106.8456,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/drivers/invalid-uuid/location", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid driver ID", response["error"])
}

func TestUserHandler_UpdateDriverLocation_InvalidJSON(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/location", userHandler.UpdateDriverLocation)

	// Invalid JSON
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/location", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid request payload", response["error"])
}

func TestUserHandler_UpdateDriverLocation_MissingFields(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/location", userHandler.UpdateDriverLocation)

	// Missing longitude
	requestBody := map[string]interface{}{
		"latitude": -6.2088,
	}

	body, _ := json.Marshal(requestBody)
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/location", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid request payload", response["error"])
}

func TestUserHandler_UpdateDriverLocation_InvalidCoordinates(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/location", userHandler.UpdateDriverLocation)

	// Invalid latitude (out of range)
	requestBody := map[string]interface{}{
		"latitude":  95.0, // Invalid: > 90
		"longitude": 106.8456,
	}

	body, _ := json.Marshal(requestBody)
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/location", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid request payload", response["error"])
}

// Test UserHandler.UpdateDriverStatus endpoint
func TestUserHandler_UpdateDriverStatus_Success(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	// Setup mock expectations
	driverRepo.On("UpdateStatus", mock.Anything, "550e8400-e29b-41d4-a716-446655440001", models.DriverStatus("online")).Return(nil)

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/status", userHandler.UpdateDriverStatus)

	// Create request body
	requestBody := map[string]interface{}{
		"status": "online",
	}

	body, _ := json.Marshal(requestBody)
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Equal(t, "Driver status updated successfully", response["message"])
}

func TestUserHandler_UpdateDriverStatus_InvalidUUID(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/status", userHandler.UpdateDriverStatus)

	// Create request body
	requestBody := map[string]interface{}{
		"status": "online",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/drivers/invalid-uuid/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid driver ID", response["error"])
}

func TestUserHandler_UpdateDriverStatus_InvalidJSON(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/status", userHandler.UpdateDriverStatus)

	// Invalid JSON
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/status", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid request payload", response["error"])
}

func TestUserHandler_UpdateDriverStatus_MissingFields(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/status", userHandler.UpdateDriverStatus)

	// Missing status field
	requestBody := map[string]interface{}{}

	body, _ := json.Marshal(requestBody)
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid request payload", response["error"])
}

func TestUserHandler_UpdateDriverStatus_InvalidStatus(t *testing.T) {
	// Setup
	userRepo := &MockUserRepository{}
	driverRepo := &MockDriverRepository{}
	passengerRepo := &MockPassengerRepository{}

	userHandler := handlers.NewUserHandler(userRepo, driverRepo, passengerRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/drivers/:id/status", userHandler.UpdateDriverStatus)

	// Invalid status value
	requestBody := map[string]interface{}{
		"status": "invalid_status",
	}

	body, _ := json.Marshal(requestBody)
	driverID := "550e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("PUT", "/drivers/"+driverID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid request payload", response["error"])
}
