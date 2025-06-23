package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository"

	"github.com/lib/pq"
)

// DriverRepositoryImpl implements the DriverRepository interface using PostgreSQL
type DriverRepositoryImpl struct {
	db *sql.DB
}

// NewDriverRepository creates a new instance of DriverRepositoryImpl
func NewDriverRepository(db *sql.DB) repository.DriverRepository {
	return &DriverRepositoryImpl{db: db}
}

// Create creates a new driver in the database
func (r *DriverRepositoryImpl) Create(ctx context.Context, driver *models.Driver) error {
	query := `
		INSERT INTO drivers (id, user_id, license_number, vehicle_make, vehicle_model, 
			vehicle_year, vehicle_plate, status, current_lat, current_lng, rating, total_trips, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		driver.ID,
		driver.UserID,
		driver.LicenseNumber,
		driver.VehicleType,
		driver.VehiclePlate,
		driver.Status,
		driver.CurrentLatitude,
		driver.CurrentLongitude,
		driver.Rating,
		driver.TotalTrips,
		driver.CreatedAt,
		driver.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "drivers_license_number_key" {
					return &models.ValidationError{
						Field:   "license_number",
						Message: "license number already exists",
					}
				}
				if pqErr.Constraint == "drivers_vehicle_plate_key" {
					return &models.ValidationError{
						Field:   "vehicle_plate",
						Message: "vehicle plate already exists",
					}
				}
			case "23503": // foreign_key_violation
				if pqErr.Constraint == "drivers_user_id_fkey" {
					return &models.ValidationError{
						Field:   "user_id",
						Message: "user does not exist",
					}
				}
			}
		}
		return fmt.Errorf("failed to create driver: %w", err)
	}
	
	return nil
}

// GetByID retrieves a driver by ID
func (r *DriverRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	query := `
		SELECT id, user_id, license_number, vehicle_make, vehicle_model, vehicle_year, 
			vehicle_plate, status, current_lat, current_lng, rating, total_trips, created_at, updated_at
		FROM drivers
		WHERE id = $1
	`
	
	driver := &models.Driver{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&driver.ID,
		&driver.UserID,
		&driver.LicenseNumber,
		&driver.VehicleType,
		&driver.VehiclePlate,
		&driver.Status,
		&driver.CurrentLatitude,
		&driver.CurrentLongitude,
		&driver.Rating,
		&driver.TotalTrips,
		&driver.CreatedAt,
		&driver.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "driver",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get driver by ID: %w", err)
	}
	
	return driver, nil
}

// GetByUserID retrieves a driver by user ID
func (r *DriverRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*models.Driver, error) {
	query := `
		SELECT id, user_id, license_number, vehicle_make, vehicle_model, vehicle_year, 
			vehicle_plate, status, current_lat, current_lng, rating, total_trips, created_at, updated_at
		FROM drivers
		WHERE user_id = $1
	`
	
	driver := &models.Driver{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&driver.ID,
		&driver.UserID,
		&driver.LicenseNumber,
		&driver.VehicleType,
		&driver.VehiclePlate,
		&driver.Status,
		&driver.CurrentLatitude,
		&driver.CurrentLongitude,
		&driver.Rating,
		&driver.TotalTrips,
		&driver.CreatedAt,
		&driver.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "driver",
				ID:       userID,
			}
		}
		return nil, fmt.Errorf("failed to get driver by user ID: %w", err)
	}
	
	return driver, nil
}

// Update updates an existing driver
func (r *DriverRepositoryImpl) Update(ctx context.Context, driver *models.Driver) error {
	query := `
		UPDATE drivers
		SET license_number = $2, vehicle_make = $3, vehicle_model = $4, vehicle_year = $5,
			vehicle_plate = $6, status = $7, current_lat = $8, current_lng = $9, 
			rating = $10, total_trips = $11, updated_at = $12
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		driver.ID,
		driver.LicenseNumber,
		driver.VehicleType,
		driver.VehiclePlate,
		driver.Status,
		driver.CurrentLatitude,
		driver.CurrentLongitude,
		driver.Rating,
		driver.TotalTrips,
		driver.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "drivers_license_number_key" {
					return &models.ValidationError{
						Field:   "license_number",
						Message: "license number already exists",
					}
				}
				if pqErr.Constraint == "drivers_vehicle_plate_key" {
					return &models.ValidationError{
						Field:   "vehicle_plate",
						Message: "vehicle plate already exists",
					}
				}
			}
		}
		return fmt.Errorf("failed to update driver: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "driver",
			ID:       driver.ID.String(),
		}
	}
	
	return nil
}

// Delete deletes a driver by ID
func (r *DriverRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM drivers WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete driver: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "driver",
			ID:       id,
		}
	}
	
	return nil
}

// GetOnlineDrivers retrieves all online drivers
func (r *DriverRepositoryImpl) GetOnlineDrivers(ctx context.Context) ([]*models.Driver, error) {
	query := `
		SELECT id, user_id, license_number, vehicle_make, vehicle_model, vehicle_year, 
			vehicle_plate, status, current_lat, current_lng, rating, total_trips, created_at, updated_at
		FROM drivers
		WHERE status = 'online'
		ORDER BY rating DESC, total_trips DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get online drivers: %w", err)
	}
	defer rows.Close()
	
	var drivers []*models.Driver
	for rows.Next() {
		driver := &models.Driver{}
		err := rows.Scan(
			&driver.ID,
			&driver.UserID,
			&driver.LicenseNumber,
			&driver.VehicleType,
			&driver.VehiclePlate,
			&driver.Status,
			&driver.CurrentLatitude,
		&driver.CurrentLongitude,
			&driver.Rating,
			&driver.TotalTrips,
			&driver.CreatedAt,
			&driver.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}
		drivers = append(drivers, driver)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating drivers: %w", err)
	}
	
	return drivers, nil
}

// GetDriversInRadius retrieves drivers within a specified radius
func (r *DriverRepositoryImpl) GetDriversInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]*models.Driver, error) {
	// Using Haversine formula to calculate distance
	query := `
		SELECT id, user_id, license_number, vehicle_make, vehicle_model, vehicle_year, 
			vehicle_plate, status, current_lat, current_lng, rating, total_trips, created_at, updated_at,
			(
				6371 * acos(
					cos(radians($1)) * cos(radians(current_lat)) * 
					cos(radians(current_lng) - radians($2)) + 
					sin(radians($1)) * sin(radians(current_lat))
				)
			) AS distance
		FROM drivers
		WHERE status = 'online'
			AND current_lat IS NOT NULL 
			AND current_lng IS NOT NULL
		HAVING distance <= $3
		ORDER BY distance ASC, rating DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, lat, lng, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("failed to get drivers in radius: %w", err)
	}
	defer rows.Close()
	
	var drivers []*models.Driver
	for rows.Next() {
		driver := &models.Driver{}
		var distance float64
		err := rows.Scan(
			&driver.ID,
			&driver.UserID,
			&driver.LicenseNumber,
			&driver.VehicleType,
			&driver.VehiclePlate,
			&driver.Status,
			&driver.CurrentLatitude,
		&driver.CurrentLongitude,
			&driver.Rating,
			&driver.TotalTrips,
			&driver.CreatedAt,
			&driver.UpdatedAt,
			&distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}
		drivers = append(drivers, driver)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating drivers: %w", err)
	}
	
	return drivers, nil
}

// UpdateLocation updates a driver's current location
func (r *DriverRepositoryImpl) UpdateLocation(ctx context.Context, driverID string, lat, lng float64) error {
	query := `
		UPDATE drivers
		SET current_lat = $2, current_lng = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, driverID, lat, lng)
	if err != nil {
		return fmt.Errorf("failed to update driver location: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "driver",
			ID:       driverID,
		}
	}
	
	return nil
}

// UpdateStatus updates a driver's status
func (r *DriverRepositoryImpl) UpdateStatus(ctx context.Context, driverID string, status models.DriverStatus) error {
	query := `
		UPDATE drivers
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, driverID, status)
	if err != nil {
		return fmt.Errorf("failed to update driver status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "driver",
			ID:       driverID,
		}
	}
	
	return nil
}

// List retrieves a list of drivers with pagination
func (r *DriverRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*models.Driver, error) {
	query := `
		SELECT id, user_id, license_number, vehicle_make, vehicle_model, vehicle_year, 
			vehicle_plate, status, current_lat, current_lng, rating, total_trips, created_at, updated_at
		FROM drivers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list drivers: %w", err)
	}
	defer rows.Close()
	
	var drivers []*models.Driver
	for rows.Next() {
		driver := &models.Driver{}
		err := rows.Scan(
			&driver.ID,
			&driver.UserID,
			&driver.LicenseNumber,
			&driver.VehicleType,
			&driver.VehiclePlate,
			&driver.Status,
			&driver.CurrentLatitude,
		&driver.CurrentLongitude,
			&driver.Rating,
			&driver.TotalTrips,
			&driver.CreatedAt,
			&driver.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}
		drivers = append(drivers, driver)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating drivers: %w", err)
	}
	
	return drivers, nil
}

// calculateDistance calculates the distance between two points using Haversine formula
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers
	
	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180
	
	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad
	
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadius * c
}