package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateAndValidateToken(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	role := "Operator"
	expiryMin := 5

	token, err := GenerateToken(userID, orgID, role, expiryMin)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, claims.UserID)
	}

	if claims.OrganizationID != orgID {
		t.Errorf("Expected OrganizationID %v, got %v", orgID, claims.OrganizationID)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
}

func TestExpiredToken(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	role := "Viewer"
	// Set to negative minutes to simulate expired token
	expiryMin := -5

	token, err := GenerateToken(userID, orgID, role, expiryMin)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, err = ValidateToken(token)
	if err == nil {
		t.Error("Expected error validating an expired token, but got nil")
	}
}
