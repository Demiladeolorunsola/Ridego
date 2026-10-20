package main

import (
	"log"
	"os"

	"github.com/Demiladeolorunsola/ridego/database"
	"github.com/Demiladeolorunsola/ridego/handlers"
	"github.com/Demiladeolorunsola/ridego/middleware"
	"github.com/Demiladeolorunsola/ridego/models"
	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDatabase()

	err := database.DB.AutoMigrate(
		&models.User{},
		&models.Driver{},
		&models.Vehicle{},
		&models.Ride{},
	)
	if err != nil {
		log.Fatal("Database migration failed:", err)
	}

	log.Println("Database tables migrated successfully")

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "development-only-change-this-secret-before-deployment"
		log.Println("WARNING: Using development JWT secret")
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to RideGo API",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Public authentication routes.
	router.POST("/api/register", handlers.Register)
	router.POST("/api/login", handlers.Login(secret))

	// Protected routes.
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(secret))
	{
		protected.GET("/profile", handlers.Profile)

		// Ride routes.
		protected.POST("/rides", handlers.CreateRide)
		protected.GET("/rides", handlers.GetMyRides)
		protected.PATCH("/rides/:id/cancel", handlers.CancelRide)
		protected.POST("/drivers/register", handlers.RegisterDriver)
		protected.POST("/vehicles", handlers.RegisterVehicle)
		protected.PATCH("/drivers/availability", handlers.UpdateDriverAvailability)
		protected.PATCH("/rides/:id/accept", handlers.AcceptRide)
		protected.PATCH("/rides/:id/start", handlers.StartRide)
		protected.PATCH("/rides/:id/complete", handlers.CompleteRide)
	}

	// Start the RideGo API server.
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed:", err)
	}

}
