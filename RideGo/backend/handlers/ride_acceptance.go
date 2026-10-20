package handlers

import (
	"net/http"

	"github.com/Demiladeolorunsola/ridego/database"
	"github.com/Demiladeolorunsola/ridego/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AcceptRide(c *gin.Context) {
	userID := c.GetUint("user_id")
	rideID := c.Param("id")

	// Find the driver's profile.
	var driver models.Driver
	if err := database.DB.Where("user_id = ?", userID).
		First(&driver).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You must register as a driver first",
		})
		return
	}

	// The driver must be online.
	if !driver.Available {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You must be online to accept rides",
		})
		return
	}

	// The driver must have a registered vehicle.
	var vehicle models.Vehicle
	if err := database.DB.Where("driver_id = ?", driver.ID).
		First(&vehicle).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Register a vehicle before accepting rides",
		})
		return
	}

	// Atomically assign the ride only if it is still requested.
	result := database.DB.Model(&models.Ride{}).
		Where("id = ? AND status = ? AND driver_id IS NULL",
			rideID, "requested").
		Updates(map[string]interface{}{
			"driver_id": driver.ID,
			"status":    "accepted",
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not accept ride",
		})
		return
	}

	if result.RowsAffected == 0 {
		// Distinguish a missing ride from one already assigned or cancelled.
		var ride models.Ride
		err := database.DB.First(&ride, rideID).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Ride not found",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not retrieve ride",
			})
			return
		}

		c.JSON(http.StatusConflict, gin.H{
			"error": "Ride is no longer available for acceptance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Ride accepted successfully",
		"ride_id":   rideID,
		"driver_id": driver.ID,
		"status":    "accepted",
	})
}
