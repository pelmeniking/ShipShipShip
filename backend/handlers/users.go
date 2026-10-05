package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"shipshipship/database"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password" binding:"required"`
	Role        string `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	Email       *string `json:"email"`
	DisplayName *string `json:"display_name"`
	Password    *string `json:"password"`
	Role        *string `json:"role"`
	Active      *bool   `json:"active"`
}

type ChangeOwnPasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// GetUsers returns all users
func GetUsers(c *gin.Context) {
	var users []models.User
	if err := database.GetDB().Order("username ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// CreateUser creates a new user with a local password
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}
	if !models.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	if err := checkUserUnique(req.Username, req.Email, 0); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = req.Username
	}

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		DisplayName:  displayName,
		Role:         req.Role,
		Active:       true,
		AuthProvider: models.AuthProviderLocal,
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser changes a user's details, role, status or password
func UpdateUser(c *gin.Context) {
	user, ok := findUserParam(c)
	if !ok {
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetUint("user_id")

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if err := checkUserUnique("", email, user.ID); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		user.Email = email
	}
	if req.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*req.DisplayName)
		if user.DisplayName == "" {
			user.DisplayName = user.Username
		}
	}
	if req.Role != nil {
		if !models.IsValidRole(*req.Role) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}
		if user.ID == currentUserID && *req.Role != models.RoleAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot remove your own admin role"})
			return
		}
		user.Role = *req.Role
	}
	if req.Active != nil {
		if user.ID == currentUserID && !*req.Active {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot deactivate your own account"})
			return
		}
		user.Active = *req.Active
	}
	if req.Password != nil && *req.Password != "" {
		if err := user.SetPassword(*req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		// Save writes zero values too, so deactivation (active=false) is persisted
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return ensureActiveAdminRemains(tx)
	})
	if err != nil {
		if errors.Is(err, errLastAdmin) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser removes a user
func DeleteUser(c *gin.Context) {
	user, ok := findUserParam(c)
	if !ok {
		return
	}

	if user.ID == c.GetUint("user_id") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot delete your own account"})
		return
	}

	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(user).Error; err != nil {
			return err
		}
		return ensureActiveAdminRemains(tx)
	})
	if err != nil {
		if errors.Is(err, errLastAdmin) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// ChangeOwnPassword lets any signed-in user change their own password
func ChangeOwnPassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not available in demo mode"})
		return
	}

	var req ChangeOwnPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if !user.CheckPassword(req.CurrentPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}
	if err := user.SetPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.GetDB().Model(&user).Update("password_hash", user.PasswordHash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed"})
}

var errLastAdmin = errors.New("at least one active admin is required")

func ensureActiveAdminRemains(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&models.User{}).
		Where("role = ? AND active = ?", models.RoleAdmin, true).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errLastAdmin
	}
	return nil
}

func findUserParam(c *gin.Context) (*models.User, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return nil, false
	}

	var user models.User
	if err := database.GetDB().First(&user, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return nil, false
	}
	return &user, true
}

// checkUserUnique makes sure no other user already has the username or email (case-insensitive)
func checkUserUnique(username, email string, excludeID uint) error {
	db := database.GetDB()
	if username != "" {
		var count int64
		db.Model(&models.User{}).Where("LOWER(username) = LOWER(?) AND id <> ?", username, excludeID).Count(&count)
		if count > 0 {
			return errors.New("username is already taken")
		}
	}
	if email != "" {
		var count int64
		db.Model(&models.User{}).Where("LOWER(email) = LOWER(?) AND id <> ?", email, excludeID).Count(&count)
		if count > 0 {
			return errors.New("email is already in use")
		}
	}
	return nil
}
