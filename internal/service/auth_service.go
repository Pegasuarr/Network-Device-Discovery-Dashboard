package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/user/network-monitoring/configs"
	"github.com/user/network-monitoring/internal/auth"
	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *configs.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *configs.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Register(username, email, password string, orgID uuid.UUID, roleID uint) (*model.User, error) {
	// Check if username/email already exists
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, errors.New("username is already taken")
	}
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, errors.New("email is already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &model.User{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Username:       username,
		Email:          email,
		PasswordHash:   string(hashedPassword),
		RoleID:         roleID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, err
	}

	// Fetch with role preloaded
	return s.userRepo.FindByID(newUser.ID, orgID)
}

type LoginResult struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

func (s *AuthService) Login(username, password, ipAddress, userAgent string) (*LoginResult, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		s.logLogin(uuid.Nil, username, ipAddress, userAgent, "failed")
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.logLogin(user.ID, username, ipAddress, userAgent, "failed")
		return nil, errors.New("invalid username or password")
	}

	// Determine Role string representation
	roleStr := "Viewer"
	if user.RoleID == 1 {
		roleStr = "Admin"
	} else if user.RoleID == 2 {
		roleStr = "Operator"
	}

	// Generate JWT Access Token
	accToken, err := auth.GenerateToken(user.ID, user.OrganizationID, roleStr, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token
	refTokenVal, err := auth.GenerateToken(user.ID, user.OrganizationID, "Refresh", s.cfg.JWTRefreshTTL*24*60)
	if err != nil {
		return nil, err
	}

	// Save Refresh Token to Database
	refToken := &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refTokenVal,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.JWTRefreshTTL) * 24 * time.Hour),
	}
	if err := repository.DB.Create(refToken).Error; err != nil {
		return nil, err
	}

	s.logLogin(user.ID, username, ipAddress, userAgent, "success")

	return &LoginResult{
		User:         user,
		AccessToken:  accToken,
		RefreshToken: refTokenVal,
	}, nil
}

func (s *AuthService) Refresh(tokenStr string) (string, error) {
	claims, err := auth.ValidateToken(tokenStr)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Verify in DB and make sure it is not revoked
	var refToken model.RefreshToken
	err = repository.DB.Where("token = ? AND revoked_at IS NULL AND expires_at > ?", tokenStr, time.Now()).First(&refToken).Error
	if err != nil {
		return "", errors.New("refresh token expired or revoked")
	}

	// Fetch user details for new token
	var user model.User
	if err := repository.DB.Preload("Role").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return "", err
	}

	roleStr := "Viewer"
	if user.RoleID == 1 {
		roleStr = "Admin"
	} else if user.RoleID == 2 {
		roleStr = "Operator"
	}

	return auth.GenerateToken(user.ID, user.OrganizationID, roleStr, s.cfg.JWTAccessTTL)
}

func (s *AuthService) Logout(tokenStr string) error {
	now := time.Now()
	return repository.DB.Model(&model.RefreshToken{}).
		Where("token = ?", tokenStr).
		Update("revoked_at", &now).Error
}

func (s *AuthService) logLogin(userID uuid.UUID, username, ipAddress, userAgent, status string) {
	log := model.LoginLog{
		ID:        uuid.New(),
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Status:    status,
		CreatedAt: time.Now(),
	}
	repository.DB.Create(&log)
}
