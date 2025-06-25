package handlers

import (
	"net/http"
	"strconv"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userRepo      repository.UserRepository
	driverRepo    repository.DriverRepository
	passengerRepo repository.PassengerRepository
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(
	userRepo repository.UserRepository,
	driverRepo repository.DriverRepository,
	passengerRepo repository.PassengerRepository,
) *UserHandler {
	return &UserHandler{
		userRepo:      userRepo,
		driverRepo:    driverRepo,
		passengerRepo: passengerRepo,
	}
}

// CreateUserRequest represents the request payload for user creation
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Name     string `json:"name" binding:"required"`
	UserType string `json:"user_type" binding:"required,oneof=passenger driver"`
}

// CreateDriverRequest represents additional driver-specific fields
type CreateDriverRequest struct {
	CreateUserRequest
	LicenseNumber string `json:"license_number" binding:"required"`
	VehicleType   string `json:"vehicle_type" binding:"required"`
	VehiclePlate  string `json:"vehicle_plate" binding:"required"`
}

// UpdateUserRequest represents the request payload for user updates
type UpdateUserRequest struct {
	Email *string `json:"email,omitempty" binding:"omitempty,email"`
	Phone *string `json:"phone,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// UpdateDriverLocationRequest represents the request for updating driver location
type UpdateDriverLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

// UpdateDriverStatusRequest represents the request for updating driver status
type UpdateDriverStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=online offline busy"`
}

// CreateUser handles user creation
// @Summary Create a new user
// @Description Create a new user (passenger or driver)
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation details"
// @Success 201 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Create user model
	user := &models.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Phone:    req.Phone,
		Name:     req.Name,
		UserType: models.UserType(req.UserType),
	}

	// Validate user
	if err := user.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: err.Error(),
		})
		return
	}

	// Create user in database
	err := h.userRepo.Create(c.Request.Context(), user)
	if err != nil {
		switch err.(type) {
		case *models.ValidationError:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Validation error",
				Message: err.Error(),
			})
		// case *models.DatabaseError: // Commented out as this type doesn't exist yet
		//	dbErr := err.(*models.DatabaseError)
		//	if dbErr.Code == "unique_violation" {
		//		c.JSON(http.StatusConflict, ErrorResponse{
		//			Error:   "User already exists",
		//			Message: "A user with this email or phone already exists",
		//		})
		//	} else {
		//		c.JSON(http.StatusInternalServerError, ErrorResponse{
		//			Error:   "Database error",
		//			Message: "Failed to create user",
		//		})
		//	}
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to create user",
			})
		}
		return
	}

	// If user is a passenger, create passenger record
	if user.UserType == models.UserTypePassenger {
		passenger := &models.Passenger{
			ID:     uuid.New(),
			UserID: user.ID,
			Rating: 5.0, // Default rating
		}

		if err := h.passengerRepo.Create(c.Request.Context(), passenger); err != nil {
			// If passenger creation fails, we should ideally rollback the user creation
			// For simplicity, we'll just log the error and continue
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "User created but failed to create passenger profile",
			})
			return
		}
	}

	c.JSON(http.StatusCreated, user)
}

// CreateDriver handles driver creation with additional driver-specific fields
// @Summary Create a new driver
// @Description Create a new driver with vehicle information
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateDriverRequest true "Driver creation details"
// @Success 201 {object} models.Driver
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/drivers [post]
func (h *UserHandler) CreateDriver(c *gin.Context) {
	var req CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Create user model
	user := &models.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Phone:    req.Phone,
		Name:     req.Name,
		UserType: models.UserTypeDriver,
	}

	// Validate user
	if err := user.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: err.Error(),
		})
		return
	}

	// Create user in database
	err := h.userRepo.Create(c.Request.Context(), user)
	if err != nil {
		switch err.(type) {
		case *models.ValidationError:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Validation error",
				Message: err.Error(),
			})
		// case *models.DatabaseError: // Commented out as this type doesn't exist yet
		//	dbErr := err.(*models.DatabaseError)
		//	if dbErr.Code == "unique_violation" {
		//		c.JSON(http.StatusConflict, ErrorResponse{
		//			Error:   "User already exists",
		//			Message: "A user with this email or phone already exists",
		//		})
		//	} else {
		//		c.JSON(http.StatusInternalServerError, ErrorResponse{
		//			Error:   "Database error",
		//			Message: "Failed to create user",
		//		})
		//	}
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to create user",
			})
		}
		return
	}

	// Create driver record
	driver := &models.Driver{
		ID:            uuid.New(),
		UserID:        user.ID,
		LicenseNumber: req.LicenseNumber,
		VehicleType:   req.VehicleType,
		VehiclePlate:  req.VehiclePlate,
		Status:        models.DriverStatusOffline,
		Rating:        5.0, // Default rating
	}

	if err := driver.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: err.Error(),
		})
		return
	}

	if err := h.driverRepo.Create(c.Request.Context(), driver); err != nil {
		// If driver creation fails, we should ideally rollback the user creation
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "User created but failed to create driver profile",
		})
		return
	}

	c.JSON(http.StatusCreated, driver)
}

// GetUser handles user retrieval by ID
// @Summary Get user by ID
// @Description Retrieve a user by their ID
// @Tags users
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid UUID",
		})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.String())
	if err != nil {
		switch err.(type) {
		case *models.NotFoundError:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to get user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles user updates
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body UpdateUserRequest true "User update details"
// @Success 200 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid UUID",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Get existing user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID.String())
	if err != nil {
		switch err.(type) {
		case *models.NotFoundError:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to get user",
			})
		}
		return
	}

	// Update fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Name != nil {
		user.Name = *req.Name
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: err.Error(),
		})
		return
	}

	// Update user in database
	err = h.userRepo.Update(c.Request.Context(), user)
	if err != nil {
		switch err.(type) {
		case *models.ValidationError:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Validation error",
				Message: err.Error(),
			})
		// case *models.DatabaseError: // Commented out as this type doesn't exist yet
		//	dbErr := err.(*models.DatabaseError)
		//	if dbErr.Code == "unique_violation" {
		//		c.JSON(http.StatusConflict, ErrorResponse{
		//			Error:   "Conflict",
		//			Message: "A user with this email or phone already exists",
		//		})
		//	} else {
		//		c.JSON(http.StatusInternalServerError, ErrorResponse{
		//			Error:   "Database error",
		//			Message: "Failed to update user",
		//		})
		//	}
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to update user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetDriver handles driver retrieval by user ID
// @Summary Get driver by user ID
// @Description Retrieve driver information by user ID
// @Tags users
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} models.Driver
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/driver [get]
func (h *UserHandler) GetDriver(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid UUID",
		})
		return
	}

	_, err = h.driverRepo.GetByUserID(c.Request.Context(), userID.String())
	if err != nil {
		switch err.(type) {
		case *models.NotFoundError:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Driver not found",
				Message: err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Message: "Failed to get driver",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver location updated successfully"})
}

// UpdateDriverLocation handles driver location updates
// @Summary Update driver location
// @Description Update the current location of a driver
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body UpdateDriverLocationRequest true "Location update details"
// @Success 200 {object} models.Driver
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/driver/location [put]
func (h *UserHandler) UpdateDriverLocation(c *gin.Context) {
	driverIDStr := c.Param("id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid driver ID",
			Message: "Driver ID must be a valid UUID",
		})
		return
	}

	var req UpdateDriverLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Update location directly using driver ID
	err = h.driverRepo.UpdateLocation(c.Request.Context(), driverID.String(), req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to update driver location",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver location updated successfully"})
}

// UpdateDriverStatus handles driver status updates
// @Summary Update driver status
// @Description Update the status of a driver (online, offline, busy)
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body UpdateDriverStatusRequest true "Status update details"
// @Success 200 {object} models.Driver
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/driver/status [put]
func (h *UserHandler) UpdateDriverStatus(c *gin.Context) {
	driverIDStr := c.Param("id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid driver ID",
			Message: "Driver ID must be a valid UUID",
		})
		return
	}

	var req UpdateDriverStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Update status directly using driver ID
	newStatus := models.DriverStatus(req.Status)
	err = h.driverRepo.UpdateStatus(c.Request.Context(), driverID.String(), newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to update driver status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver status updated successfully"})
}

// ListUsers handles user listing with pagination
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Produce json
// @Param user_type query string false "Filter by user type" Enums(passenger, driver)
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.User}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	userType := c.Query("user_type")

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

	// Validate user type if provided
	if userType != "" && userType != "passenger" && userType != "driver" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user type",
			Message: "User type must be either 'passenger' or 'driver'",
		})
		return
	}

	// Get users from repository
	// TODO: Filter by userType when implementing user filtering
	users, err := h.userRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list users",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   users,
		Limit:  limit,
		Offset: offset,
		Total:  int64(len(users)), // Convert to int64
	})
}