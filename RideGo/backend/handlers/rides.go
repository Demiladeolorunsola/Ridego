package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/Demiladeolorunsola/ridego/database"
	"github.com/Demiladeolorunsola/ridego/models"
	"github.com/gin-gonic/gin"
)

type CreateRideInput struct {
	PickupLocation string  `json:"pickup_location" binding:"required"`
	Destination    string  `json:"destination" binding:"required"`
	DistanceKM     float64 `json:"distance_km" binding:"required,gt=0"`
	DurationMin    int     `json:"duration_min" binding:"required,gt=0"`
}

const (
	BaseFare      = 500.0
	FarePerKM     = 200.0
	FarePerMinute = 20.0
)

func CreateRide(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input CreateRideInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide valid pickup, destination, distance, and duration",
		})
		return
	}

	pickup := strings.TrimSpace(input.PickupLocation)
	destination := strings.TrimSpace(input.Destination)

	if pickup == "" || destination == "" ||
		strings.EqualFold(pickup, destination) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pickup and destination must be valid and different",
		})
		return
	}

	var passenger models.User
	if err := database.DB.First(&passenger, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Passenger account not found",
		})
		return
	}

	if passenger.Role != "passenger" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Only passengers can request rides",
		})
		return
	}

	fare := BaseFare +
		input.DistanceKM*FarePerKM +
		float64(input.DurationMin)*FarePerMinute

	ride := models.Ride{
		PassengerID:    passenger.ID,
		PickupLocation: pickup,
		Destination:    destination,
		DistanceKM:     input.DistanceKM,
		DurationMin:    input.DurationMin,
		Fare:           fare,
		Status:         "requested",
		RequestedAt:    time.Now(),
	}

	if err := database.DB.Create(&ride).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create ride request",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ride requested successfully",
		"ride":    ride,
	})
}

func GetMyRides(c *gin.Context) {
	userID := c.GetUint("user_id")

	var rides []models.Ride
	if err := database.DB.
		Where("passenger_id = ?", userID).
		Order("created_at DESC").
		Find(&rides).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve rides",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(rides),
		"rides": rides,
	})
}

// CancelRide lets a passenger cancel their own requested ride.
func CancelRide(c *gin.Context) {
	userID := c.GetUint("user_id")
	rideID := c.Param("id")

	var ride models.Ride
	if err := database.DB.First(&ride, rideID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ride not found",
		})
		return
	}

	if ride.PassengerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You cannot cancel another passenger's ride",
		})
		return
	}

	if ride.Status != "requested" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Only rides awaiting a driver can be cancelled",
		})
		return
	}

	result := database.DB.Model(&models.Ride{}).
		Where("id = ? AND passenger_id = ? AND status = ?",
			ride.ID, userID, "requested").
		Update("status", "cancelled")

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not cancel ride",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Ride status changed; refresh and try again",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ride cancelled successfully",
		"ride_id": ride.ID,
		"status":  "cancelled",
	})
}
