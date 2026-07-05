package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/user/network-monitoring/internal/auth"
	"github.com/user/network-monitoring/internal/repository"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check API Key first (for automated scraping or CLI integrations)
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			var setting struct {
				OrganizationID uuid.UUID
				Value          string
			}
			err := repository.DB.Table("settings").
				Where("key = ? AND value = ?", "api_key", apiKey).
				First(&setting).Error
			if err == nil {
				// Authenticated via API Key as Operator
				c.Set("userID", uuid.Nil)
				c.Set("orgID", setting.OrganizationID)
				c.Set("role", "Operator")
				c.Next()
				return
			}
		}

		// 2. Fallback to JWT Bearer authentication
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("orgID", claims.OrganizationID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role information missing"})
			c.Abort()
			return
		}

		userRole := roleVal.(string)
		isAllowed := false

		// Admin has full control
		if userRole == "Admin" {
			isAllowed = true
		} else {
			for _, r := range allowedRoles {
				if r == userRole {
					isAllowed = true
					break
				}
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied for this role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetTenantID(c *gin.Context) (uuid.UUID, bool) {
	orgIDVal, exists := c.Get("orgID")
	if !exists {
		return uuid.Nil, false
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	return orgID, ok
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := userIDVal.(uuid.UUID)
	return userID, ok
}
