package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverStatus represents the status of a driver
type DriverStatus string

const (
	DriverStatusOnline  DriverStatus = "online"
	DriverStatusOffline DriverStatus = "offline"
	DriverStatusBusy    DriverStatus = "busy"
)

// Driver represents a driver in the system
type Driver struct {
	ID               uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	User             *User        `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	LicenseNumber    string       `json:"license_number" gorm:"uniqueIndex;not null"`
	VehicleType      string       `json:"vehicle_type" gorm:"not null"`
	VehiclePlate     string       `json:"vehicle_plate" gorm:"not null"`
	Status           DriverStatus `json:"status" gorm:"default:'offline';check:status IN ('online', 'offline', 'busy')"`
	CurrentLatitude  *float64     `json:"current_latitude" gorm:"type:decimal(10,8)"`
	CurrentLongitude *float64     `json:"current_longitude" gorm:"type:decimal(11,8)"`
	Rating           float64      `json:"rating" gorm:"type:decimal(3,2);default:5.00"`
	TotalTrips       int          `json:"total_trips" gorm:"default:0"`
	CreatedAt        time.Time    `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time    `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for Driver
func (Driver) TableName() string {
	return "drivers"
}

// IsOnline returns true if the driver is online
func (d *Driver) IsOnline() bool {
	return d.Status == DriverStatusOnline
}

// IsAvailable returns true if the driver is available for new trips
func (d *Driver) IsAvailable() bool {
	return d.Status == DriverStatusOnline
}

// IsBusy returns true if the driver is currently on a trip
func (d *Driver) IsBusy() bool {
	return d.Status == DriverStatusBusy
}

// HasLocation returns true if the driver has a current location
func (d *Driver) HasLocation() bool {
	return d.CurrentLatitude != nil && d.CurrentLongitude != nil
}

// GetLocation returns the driver's current location
func (d *Driver) GetLocation() (lat, lng float64, ok bool) {
	if !d.HasLocation() {
		return 0, 0, false
	}
	return *d.CurrentLatitude, *d.CurrentLongitude, true
}

// SetLocation sets the driver's current location
func (d *Driver) SetLocation(lat, lng float64) {
	d.CurrentLatitude = &lat
	d.CurrentLongitude = &lng
	d.UpdatedAt = time.Now()
}

// SetStatus sets the driver's status
func (d *Driver) SetStatus(status DriverStatus) {
	d.Status = status
	d.UpdatedAt = time.Now()
}

// GoOnline sets the driver status to online
func (d *Driver) GoOnline() {
	d.SetStatus(DriverStatusOnline)
}

// GoOffline sets the driver status to offline
func (d *Driver) GoOffline() {
	d.SetStatus(DriverStatusOffline)
}

// SetBusy sets the driver status to busy
func (d *Driver) SetBusy() {
	d.SetStatus(DriverStatusBusy)
}

// CompleteTrip increments the total trips and updates rating
func (d *Driver) CompleteTrip(newRating float64) {
	// Calculate new average rating
	totalRating := d.Rating * float64(d.TotalTrips)
	totalRating += newRating
	d.TotalTrips++
	d.Rating = totalRating / float64(d.TotalTrips)
	d.UpdatedAt = time.Now()

	// Set status back to online after completing trip
	d.GoOnline()
}

// Validate validates the driver data
func (d *Driver) Validate() error {
	if d.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if d.LicenseNumber == "" {
		return ErrInvalidLicenseNumber
	}
	if d.VehicleType == "" {
		return ErrInvalidVehicleType
	}
	if d.VehiclePlate == "" {
		return ErrInvalidVehiclePlate
	}
	if d.Status != DriverStatusOnline && d.Status != DriverStatusOffline && d.Status != DriverStatusBusy {
		return ErrInvalidDriverStatus
	}
	if d.Rating < 0 || d.Rating > 5 {
		return ErrInvalidRating
	}
	return nil
}
