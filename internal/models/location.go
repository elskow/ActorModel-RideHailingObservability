package models

// Location represents a geographical location with latitude and longitude coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// IsValid checks if the location coordinates are valid
func (l *Location) IsValid() bool {
	return l.Latitude >= -90 && l.Latitude <= 90 &&
		l.Longitude >= -180 && l.Longitude <= 180
}

// DistanceTo calculates the distance to another location in kilometers
// using the Haversine formula
func (l *Location) DistanceTo(other Location) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := l.Latitude * (3.14159265359 / 180)
	lon1Rad := l.Longitude * (3.14159265359 / 180)
	lat2Rad := other.Latitude * (3.14159265359 / 180)
	lon2Rad := other.Longitude * (3.14159265359 / 180)

	// Calculate differences
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Haversine formula
	a := sin(dLat/2)*sin(dLat/2) + cos(lat1Rad)*cos(lat2Rad)*sin(dLon/2)*sin(dLon/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return earthRadius * c
}

// Helper functions for math operations
func sin(x float64) float64 {
	// Simple sine approximation for small angles
	return x - (x*x*x)/6 + (x*x*x*x*x)/120
}

func cos(x float64) float64 {
	// Simple cosine approximation for small angles
	return 1 - (x*x)/2 + (x*x*x*x)/24
}

func sqrt(x float64) float64 {
	// Newton's method for square root
	if x == 0 {
		return 0
	}
	guess := x / 2
	for i := 0; i < 10; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}

func atan2(y, x float64) float64 {
	// Simple atan2 approximation
	if x > 0 {
		return atan(y / x)
	}
	if x < 0 && y >= 0 {
		return atan(y/x) + 3.14159265359
	}
	if x < 0 && y < 0 {
		return atan(y/x) - 3.14159265359
	}
	if x == 0 && y > 0 {
		return 3.14159265359 / 2
	}
	if x == 0 && y < 0 {
		return -3.14159265359 / 2
	}
	return 0 // x == 0 && y == 0
}

func atan(x float64) float64 {
	// Simple arctangent approximation using Taylor series
	if x > 1 {
		return 3.14159265359/2 - atan(1/x)
	}
	if x < -1 {
		return -3.14159265359/2 - atan(1/x)
	}
	// For |x| <= 1, use Taylor series
	return x - (x*x*x)/3 + (x*x*x*x*x)/5 - (x*x*x*x*x*x*x)/7
}
