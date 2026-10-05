package handlers

import (
	"net/http"
	"strings"
	"time"

	"shipshipship/database"
	"shipshipship/middleware"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users can sign in with their username or email address
	login := strings.TrimSpace(req.Username)
	var user models.User
	err := database.GetDB().
		Where("LOWER(username) = LOWER(?) OR (email <> '' AND LOWER(email) = LOWER(?))", login, login).
		First(&user).Error
	if err != nil || !user.Active || user.AuthProvider != models.AuthProviderLocal || !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	now := time.Now()
	database.GetDB().Model(&user).Update("last_login_at", now)

	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func ValidateToken(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user context"})
		return
	}

	response := gin.H{
		"valid":        true,
		"username":     username,
		"display_name": username,
		"role":         c.GetString("role"),
	}

	if userID := c.GetUint("user_id"); userID != 0 {
		var user models.User
		if err := database.GetDB().First(&user, userID).Error; err == nil {
			response["id"] = user.ID
			response["display_name"] = user.DisplayName
			response["email"] = user.Email
		}
	}

	c.JSON(http.StatusOK, response)
}

func CheckDemoMode(c *gin.Context) {
	isDemoMode := middleware.IsDemoMode()
	c.JSON(http.StatusOK, gin.H{
		"demo_mode": isDemoMode,
	})
}
