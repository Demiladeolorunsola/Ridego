package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/Demiladeolorunsola/ridego/database"
	"github.com/Demiladeolorunsola/ridego/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide a valid name, email, and password of at least 8 characters",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	name := strings.TrimSpace(input.Name)

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	var existing models.User
	result := database.DB.Where("email = ?", email).First(&existing)

	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is already registered"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not secure password",
		})
		return
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Phone:    strings.TrimSpace(input.Phone),
		Role:     "passenger",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create account",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		},
	})
}

func Login(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "A valid email and password are required",
			})
			return
		}

		email := strings.ToLower(strings.TrimSpace(input.Email))

		var user models.User
		if err := database.DB.Where("email = ?", email).
			First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(input.Password),
		); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}

		claims := jwt.MapClaims{
			"user_id": user.ID,
			"role":    user.Role,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not create login token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   tokenString,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}
}

func Profile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		},
	})
}

func RegisterDriver(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input struct {
		LicenseNo string `json:"license_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "A valid license number is required",
		})
		return
	}

	input.LicenseNo = strings.TrimSpace(input.LicenseNo)
	if input.LicenseNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "License number cannot be empty",
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	if user.Role != "passenger" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Only passenger accounts can register as drivers",
		})
		return
	}

	var existingDriver models.Driver
	if err := database.DB.Where("user_id = ?", user.ID).
		First(&existingDriver).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Driver profile already exists",
		})
		return
	}

	driver := models.Driver{
		UserID:    user.ID,
		LicenseNo: input.LicenseNo,
		Available: false,
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not start driver registration",
		})
		return
	}

	if err := tx.Create(&driver).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create driver profile",
		})
		return
	}

	if err := tx.Model(&user).Update("role", "driver").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update account role",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not complete driver registration",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Driver profile created successfully",
		"driver": gin.H{
			"id":         driver.ID,
			"user_id":    driver.UserID,
			"license_no": driver.LicenseNo,
			"available":  driver.Available,
		},
	})
}
