package models

import (
	"time"

	"github.com/google/uuid"
)

// Passenger represents a passenger in the system
type Passenger struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Rating     float64   `json:"rating" gorm:"type:decimal(3,2);default:5.00"`
	TotalTrips int       `json:"total_trips" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for Passenger
func (Passenger) TableName() string {
	return "passengers"
}

// CompleteTrip increments the total trips and updates rating
func (p *Passenger) CompleteTrip(newRating float64) {
	// Calculate new average rating
	totalRating := p.Rating * float64(p.TotalTrips)
	totalRating += newRating
	p.TotalTrips++
	p.Rating = totalRating / float64(p.TotalTrips)
	p.UpdatedAt = time.Now()
}

// Validate validates the passenger data
func (p *Passenger) Validate() error {
	if p.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if p.Rating < 0 || p.Rating > 5 {
		return ErrInvalidRating
	}
	return nil
}
