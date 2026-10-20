package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"size:100;not null" json:"name"`
	Email    string `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`
	Phone    string `gorm:"size:20" json:"phone"`
	Role     string `gorm:"size:20;not null;default:passenger" json:"role"`
}

type Driver struct {
	gorm.Model
	UserID    uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	LicenseNo string `gorm:"size:100;not null" json:"license_no"`
	Available bool   `gorm:"default:false" json:"available"`
}

type Vehicle struct {
	gorm.Model
	DriverID     uint   `gorm:"not null;index" json:"driver_id"`
	Make         string `gorm:"size:100;not null" json:"make"`
	VehicleModel string `gorm:"size:100;not null" json:"model"`
	PlateNo      string `gorm:"size:30;uniqueIndex;not null" json:"plate_no"`
	Color        string `gorm:"size:50" json:"color"`
}

type Ride struct {
	gorm.Model
	PassengerID    uint      `gorm:"not null;index" json:"passenger_id"`
	DriverID       *uint     `gorm:"index" json:"driver_id"`
	PickupLocation string    `gorm:"size:255;not null" json:"pickup_location"`
	Destination    string    `gorm:"size:255;not null" json:"destination"`
	DistanceKM     float64   `gorm:"not null;default:0" json:"distance_km"`
	DurationMin    int       `gorm:"not null;default:0" json:"duration_min"`
	Fare           float64   `gorm:"not null;default:0" json:"fare"`
	Status         string    `gorm:"size:30;not null;default:requested" json:"status"`
	RequestedAt    time.Time `json:"requested_at"`
}
