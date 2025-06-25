package main

import (
	"fmt"
	"log"
	"time"

	"actor-model-observability/internal/config"
	"actor-model-observability/internal/database"
	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/models"
	"github.com/google/uuid"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Populate sample data
	if err := populateData(db); err != nil {
		log.Fatalf("Failed to populate data: %v", err)
	}

	fmt.Println("Successfully populated database with sample data")
}

func populateData(db *database.PostgresDB) error {
	// Clear existing trips
	if _, err := db.Exec("DELETE FROM trips"); err != nil {
		return fmt.Errorf("failed to clear trips: %w", err)
	}

	// Get existing users, drivers, and passengers
	users, err := getUsers(db)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	drivers, err := getDrivers(db)
	if err != nil {
		return fmt.Errorf("failed to get drivers: %w", err)
	}

	passengers, err := getPassengers(db)
	if err != nil {
		return fmt.Errorf("failed to get passengers: %w", err)
	}

	if len(users) == 0 || len(drivers) == 0 || len(passengers) == 0 {
		return fmt.Errorf("insufficient base data: users=%d, drivers=%d, passengers=%d", len(users), len(drivers), len(passengers))
	}

	// Create sample trips
	trips := generateSampleTrips(drivers, passengers)

	// Insert trips
	for _, trip := range trips {
		if err := insertTrip(db, trip); err != nil {
			return fmt.Errorf("failed to insert trip: %w", err)
		}
	}

	fmt.Printf("Inserted %d sample trips\n", len(trips))
	return nil
}

func getUsers(db *database.PostgresDB) ([]models.User, error) {
	rows, err := db.Query("SELECT id, name, email, phone, user_type, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.UserType, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func getDrivers(db *database.PostgresDB) ([]models.Driver, error) {
	rows, err := db.Query("SELECT id, user_id, license_number, vehicle_type, vehicle_plate, status, current_latitude, current_longitude, rating, total_trips, created_at, updated_at FROM drivers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []models.Driver
	for rows.Next() {
		var driver models.Driver
		if err := rows.Scan(&driver.ID, &driver.UserID, &driver.LicenseNumber, &driver.VehicleType, &driver.VehiclePlate, &driver.Status, &driver.CurrentLatitude, &driver.CurrentLongitude, &driver.Rating, &driver.TotalTrips, &driver.CreatedAt, &driver.UpdatedAt); err != nil {
			return nil, err
		}
		drivers = append(drivers, driver)
	}
	return drivers, nil
}

func getPassengers(db *database.PostgresDB) ([]models.Passenger, error) {
	rows, err := db.Query("SELECT id, user_id, rating, total_trips, created_at, updated_at FROM passengers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passengers []models.Passenger
	for rows.Next() {
		var passenger models.Passenger
		if err := rows.Scan(&passenger.ID, &passenger.UserID, &passenger.Rating, &passenger.TotalTrips, &passenger.CreatedAt, &passenger.UpdatedAt); err != nil {
			return nil, err
		}
		passengers = append(passengers, passenger)
	}
	return passengers, nil
}

func generateSampleTrips(drivers []models.Driver, passengers []models.Passenger) []models.Trip {
	locations := []struct {
		pickup, destination string
		fare                float64
	}{
		{"123 Main St, Downtown", "456 Oak Ave, Uptown", 15.50},
		{"789 Pine Rd, Westside", "321 Elm St, Eastside", 22.75},
		{"555 Broadway, Theater District", "777 Park Ave, Central", 18.25},
		{"999 Market St, Financial", "111 Beach Rd, Coastal", 35.00},
		{"222 University Ave, Campus", "444 Mall Dr, Shopping", 12.80},
		{"666 Airport Blvd, Terminal", "888 Hotel Circle, Resort", 28.90},
		{"333 Hospital Way, Medical", "555 Home St, Residential", 16.40},
		{"777 Stadium Dr, Sports", "999 Restaurant Row, Dining", 21.60},
		{"111 Tech Park, Innovation", "333 Coffee St, Cafe District", 14.20},
		{"444 Convention Center, Events", "666 Train Station, Transit", 19.75},
	}

	statuses := []string{"completed", "requested", "in_progress", "cancelled"}

	var trips []models.Trip
	baseTime := time.Now().Add(-24 * time.Hour) // Start from 24 hours ago

	for i := 0; i < 50; i++ {
		loc := locations[i%len(locations)]
		status := statuses[i%len(statuses)]
		driver := drivers[i%len(drivers)]
		passenger := passengers[i%len(passengers)]

		// Vary the time
		tripTime := baseTime.Add(time.Duration(i*30) * time.Minute)

		pickupAddr := loc.pickup
		destAddr := loc.destination
		fareAmount := loc.fare + float64(i%5)

		trip := models.Trip{
			ID:                   uuid.New(),
			PassengerID:          passenger.ID,
			DriverID:             &driver.ID,
			PickupAddress:        &pickupAddr,
			DestinationAddress:   &destAddr,
			PickupLatitude:       37.7749 + float64(i%10)*0.01,  // Vary around SF
			PickupLongitude:      -122.4194 + float64(i%10)*0.01,
			DestinationLatitude:  37.7849 + float64(i%10)*0.01,
			DestinationLongitude: -122.4094 + float64(i%10)*0.01,
			Status:               models.TripStatus(status),
			FareAmount:           &fareAmount,
			RequestedAt:          tripTime,
			CreatedAt:            tripTime,
			UpdatedAt:            tripTime.Add(time.Duration(i%60) * time.Minute),
		}

		// Set completion time for completed trips
		if status == "completed" {
			completedAt := tripTime.Add(time.Duration(15+i%30) * time.Minute)
			trip.CompletedAt = &completedAt
		}

		trips = append(trips, trip)
	}

	return trips
}

func insertTrip(db *database.PostgresDB, trip models.Trip) error {
	query := `
		INSERT INTO trips (
			id, passenger_id, driver_id, pickup_address, destination_address,
			pickup_latitude, pickup_longitude, destination_latitude, destination_longitude,
			status, fare_amount, requested_at, completed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := db.Exec(query,
		trip.ID, trip.PassengerID, trip.DriverID, trip.PickupAddress, trip.DestinationAddress,
		trip.PickupLatitude, trip.PickupLongitude, trip.DestinationLatitude, trip.DestinationLongitude,
		trip.Status, trip.FareAmount, trip.RequestedAt, trip.CompletedAt, trip.CreatedAt, trip.UpdatedAt,
	)

	return err
}