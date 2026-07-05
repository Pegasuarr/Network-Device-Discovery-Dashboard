package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/user/network-monitoring/internal/middleware"
	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
	"github.com/user/network-monitoring/internal/service"
)

type Handlers struct {
	authService      *service.AuthService
	deviceService    *service.DeviceService
	alertService     *service.AlertService
	discoveryService *service.DiscoveryService
	dashboardService *service.DashboardService
	userRepo         *repository.UserRepository
	isMonPaused      *bool
}

func NewHandlers(
	authSvc *service.AuthService,
	devSvc *service.DeviceService,
	alSvc *service.AlertService,
	discSvc *service.DiscoveryService,
	dashSvc *service.DashboardService,
	userRepo *repository.UserRepository,
	isMonPaused *bool,
) *Handlers {
	return &Handlers{
		authService:      authSvc,
		deviceService:    devSvc,
		alertService:     alSvc,
		discoveryService: discSvc,
		dashboardService: dashSvc,
		userRepo:         userRepo,
		isMonPaused:      isMonPaused,
	}
}

// --- AUTH HANDLERS ---

func (h *Handlers) Register(c *gin.Context) {
	var req struct {
		Username string    `json:"username" binding:"required"`
		Email    string    `json:"email" binding:"required,email"`
		Password string    `json:"password" binding:"required,min=6"`
		OrgID    uuid.UUID `json:"organization_id"`
		RoleID   uint      `json:"role_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use default org if not specified
	orgID := req.OrgID
	if orgID == uuid.Nil {
		orgID = repository.DefaultOrgID
	}

	roleID := req.RoleID
	if roleID == 0 {
		roleID = 3 // default to Viewer
	}

	user, err := h.authService.Register(req.Username, req.Email, req.Password, orgID, roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handlers) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	result, err := h.authService.Login(req.Username, req.Password, ip, ua)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handlers) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accToken, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accToken})
}

func (h *Handlers) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// --- USER MANAGEMENT HANDLERS ---

func (h *Handlers) ListUsers(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	users, err := h.userRepo.ListByOrg(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handlers) CreateUser(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		RoleID   uint   `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(req.Username, req.Email, req.Password, orgID, req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handlers) UpdateUser(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.FindByID(userID, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		RoleID   uint   `json:"role_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RoleID != 0 {
		user.RoleID = req.RoleID
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.PasswordHash = string(hashed)
	}

	user.UpdatedAt = time.Now()
	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handlers) DeleteUser(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.userRepo.Delete(userID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// --- DEVICES HANDLERS ---

func (h *Handlers) ListDevices(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	devices, err := h.deviceService.ListDevices(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, devices)
}

func (h *Handlers) CreateDevice(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var device model.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device.ID = uuid.New()
	device.OrganizationID = orgID

	if err := h.deviceService.CreateDevice(&device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func (h *Handlers) UpdateDevice(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	devIDStr := c.Param("id")
	devID, err := uuid.Parse(devIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	existing, err := h.deviceService.GetDevice(devID, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.OrganizationID = orgID // Enforce tenancy safety
	existing.ID = devID

	if err := h.deviceService.UpdateDevice(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}

func (h *Handlers) DeleteDevice(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	devIDStr := c.Param("id")
	devID, err := uuid.Parse(devIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	if err := h.deviceService.DeleteDevice(devID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}

func (h *Handlers) GetDeviceHistory(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	devIDStr := c.Param("id")
	devID, err := uuid.Parse(devIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	// Verify device exists and belongs to tenant
	_, err = h.deviceService.GetDevice(devID, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	// Get history for the last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	results := []model.MonitoringResult{}
	err = repository.DB.Where("device_id = ? AND checked_at >= ?", devID, since).Order("checked_at asc").Find(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *Handlers) ImportDevices(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	count, err := h.deviceService.ImportCSV(orgID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Import successful", "count": count})
}

func (h *Handlers) ExportDevices(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	c.Header("Content-Disposition", "attachment; filename=devices.csv")
	c.Header("Content-Type", "text/csv")
	if err := h.deviceService.ExportCSV(orgID, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// --- MONITORING CONTROLS HANDLERS ---

func (h *Handlers) GetMonitoringStatus(c *gin.Context) {
	status := "running"
	if *h.isMonPaused {
		status = "paused"
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func (h *Handlers) StartMonitoring(c *gin.Context) {
	*h.isMonPaused = false
	c.JSON(http.StatusOK, gin.H{"message": "Monitoring service resumed"})
}

func (h *Handlers) StopMonitoring(c *gin.Context) {
	*h.isMonPaused = true
	c.JSON(http.StatusOK, gin.H{"message": "Monitoring service paused"})
}

func (h *Handlers) TriggerDiscovery(c *gin.Context) {
	var req struct {
		CIDR string `json:"cidr" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	devices, err := h.discoveryService.ScanCIDR(ctx, req.CIDR)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// --- ALERT HANDLERS ---

func (h *Handlers) ListAlerts(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	alerts, err := h.alertService.AlertRepo.ListAll(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *Handlers) CreateAlertRule(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule.ID = uuid.New()
	rule.OrganizationID = orgID

	if err := h.alertService.AlertRepo.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *Handlers) ListAlertRules(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	rules, err := h.alertService.AlertRepo.ListRules(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *Handlers) ResolveAlert(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	alertIDStr := c.Param("id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.alertService.ResolveAlert(alertID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert marked resolved"})
}

func (h *Handlers) DeleteAlertRule(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	ruleIDStr := c.Param("id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	if err := h.alertService.AlertRepo.DeleteRule(ruleID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert rule deleted"})
}

// --- DASHBOARD HANDLERS ---

func (h *Handlers) GetDashboardStats(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	stats, err := h.dashboardService.GetStats(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handlers) GetDashboardLatency(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	trend, err := h.dashboardService.GetGlobalLatencyTrend(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}

// --- SETTINGS HANDLERS ---

func (h *Handlers) GetSettings(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var settings []model.Setting
	err := repository.DB.Where("organization_id = ?", orgID).Find(&settings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	setMap := make(map[string]string)
	for _, s := range settings {
		setMap[s.Key] = s.Value
	}
	c.JSON(http.StatusOK, setMap)
}

func (h *Handlers) SaveSettings(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for k, v := range req {
		setting := model.Setting{
			OrganizationID: orgID,
			Key:            k,
			Value:          v,
			Group:          "general",
		}
		if k == "smtp_host" || k == "smtp_port" || k == "smtp_username" || k == "smtp_password" {
			setting.Group = "smtp"
		} else if k == "slack_webhook" || k == "telegram_token" || k == "telegram_chat_id" || k == "discord_webhook" || k == "api_key" {
			setting.Group = "notification"
		}
		err := repository.DB.Save(&setting).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings saved successfully"})
}

