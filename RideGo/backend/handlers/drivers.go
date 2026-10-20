package handlers

import (
	"net/http"

	"github.com/Demiladeolorunsola/ridego/database"
	"github.com/Demiladeolorunsola/ridego/models"
	"github.com/gin-gonic/gin"
)

type AvailabilityInput struct {
	Available *bool `json:"available" binding:"required"`
}

func UpdateDriverAvailability(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input AvailabilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide available as true or false",
		})
		return
	}

	var driver models.Driver
	if err := database.DB.Where("user_id = ?", userID).
		First(&driver).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You must register as a driver first",
		})
		return
	}

	result := database.DB.Model(&driver).
		Update("available", *input.Available)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update driver availability",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Driver availability updated successfully",
		"driver_id": driver.ID,
		"available": *input.Available,
	})
}
