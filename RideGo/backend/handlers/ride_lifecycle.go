package handlers

import (
	"net/http"

	"github.com/Demiladeolorunsola/ridego/database"
	"github.com/Demiladeolorunsola/ridego/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StartRide allows the assigned driver to start an accepted ride.
func StartRide(c *gin.Context) {
	transitionRide(c, "accepted", "in_progress")
}

// CompleteRide allows the assigned driver to complete an ongoing ride.
func CompleteRide(c *gin.Context) {
	transitionRide(c, "in_progress", "completed")
}

// transitionRide safely changes a ride's status.
func transitionRide(c *gin.Context, currentStatus, nextStatus string) {
	userID := c.GetUint("user_id")
	rideID := c.Param("id")

	// Find the authenticated user's driver profile.
	var driver models.Driver
	if err := database.DB.Where("user_id = ?", userID).
		First(&driver).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You must be registered as a driver",
		})
		return
	}

	// Find the requested ride.
	var ride models.Ride
	if err := database.DB.First(&ride, rideID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Ride not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve ride",
		})
		return
	}

	// Only the driver assigned to this ride can change its status.
	if ride.DriverID == nil || *ride.DriverID != driver.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not assigned to this ride",
		})
		return
	}

	// Only allow the expected status transition.
	if ride.Status != currentStatus {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Ride status must be " + currentStatus +
				" before it can become " + nextStatus,
			"current_status": ride.Status,
		})
		return
	}

	// Update conditionally to prevent competing requests from
	// completing or starting the same ride more than once.
	result := database.DB.Model(&models.Ride{}).
		Where(
			"id = ? AND driver_id = ? AND status = ?",
			ride.ID, driver.ID, currentStatus,
		).
		Update("status", nextStatus)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update ride status",
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
		"message": "Ride status updated successfully",
		"ride_id": ride.ID,
		"status":  nextStatus,
	})
}
