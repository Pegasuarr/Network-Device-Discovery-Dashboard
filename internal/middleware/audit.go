package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// Only log mutating requests
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete && method != http.MethodPatch {
			c.Next()
			return
		}

		// Read request body for audit trail
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// Get auth context
		userIDVal, userExists := c.Get("userID")
		orgIDVal, orgExists := c.Get("orgID")
		roleVal, _ := c.Get("role")

		if !userExists || !orgExists {
			return // Skip logging unauthenticated mutations (like login/register errors)
		}

		userID := userIDVal.(uuid.UUID)
		orgID := orgIDVal.(uuid.UUID)
		role := roleVal.(string)

		// Record the path
		path := c.Request.URL.Path
		ip := c.ClientIP()

		// Async write to db
		go func(uid, oid uuid.UUID, username, action, path, payload, ip string) {
			audit := model.AuditLog{
				ID:             uuid.New(),
				OrganizationID: oid,
				UserID:         uid,
				Username:       username,
				Action:         action,
				ResourceType:   path,
				ResourceID:     "", // can be parsed from route params if needed
				Payload:        payload,
				IPAddress:      ip,
				CreatedAt:      time.Now(),
			}
			repository.DB.Create(&audit)
		}(userID, orgID, role, method+" "+path, path, string(bodyBytes), ip)
	}
}
