package handlers

import (
	"net/http"
	"strings"

	"github.com/Demiladeolorunsola/ridego/database"
	"github.com/Demiladeolorunsola/ridego/models"
	"github.com/gin-gonic/gin"
)

type RegisterVehicleInput struct {
	Make         string `json:"make" binding:"required"`
	VehicleModel string `json:"model" binding:"required"`
	PlateNo      string `json:"plate_no" binding:"required"`
	Color        string `json:"color"`
}

func RegisterVehicle(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input RegisterVehicleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Make, model, and plate number are required",
		})
		return
	}

	input.Make = strings.TrimSpace(input.Make)
	input.VehicleModel = strings.TrimSpace(input.VehicleModel)
	input.PlateNo = strings.ToUpper(strings.TrimSpace(input.PlateNo))
	input.Color = strings.TrimSpace(input.Color)

	if input.Make == "" || input.VehicleModel == "" || input.PlateNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Make, model, and plate number cannot be empty",
		})
		return
	}

	var driver models.Driver
	if err := database.DB.Where("user_id = ?", userID).
		First(&driver).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You must register as a driver before adding a vehicle",
		})
		return
	}

	vehicle := models.Vehicle{
		DriverID:     driver.ID,
		Make:         input.Make,
		VehicleModel: input.VehicleModel,
		PlateNo:      input.PlateNo,
		Color:        input.Color,
	}

	if err := database.DB.Create(&vehicle).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Could not register vehicle. Check whether the plate number is already registered",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vehicle registered successfully",
		"vehicle": vehicle,
	})
}
