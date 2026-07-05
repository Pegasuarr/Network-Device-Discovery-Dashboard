package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/user/network-monitoring/internal/middleware"
	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
	"github.com/user/network-monitoring/internal/scanner"
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

	search := c.Query("search")
	status := c.Query("status")
	deviceType := c.Query("device_type")
	vendor := c.Query("vendor")
	sortBy := c.Query("sort_by")
	sortOrder := c.DefaultQuery("sort_order", "asc")

	if search == "" && status == "" && deviceType == "" && vendor == "" && sortBy == "" && c.Query("page") == "" {
		devices, err := h.deviceService.ListDevices(orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, devices)
		return
	}

	query := repository.DB.Model(&model.Device{}).Where("organization_id = ?", orgID)

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("ip_address LIKE ? OR hostname LIKE ? OR mac_address LIKE ? OR mac_vendor LIKE ? OR vendor LIKE ? OR name LIKE ?", s, s, s, s, s, s)
	}

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	if deviceType != "" && deviceType != "all" {
		query = query.Where("device_type = ?", deviceType)
	}
	if vendor != "" && vendor != "all" {
		query = query.Where("mac_vendor = ? OR vendor = ?", vendor, vendor)
	}

	var total int64
	query.Count(&total)

	order := "ip_address asc"
	if sortBy != "" {
		col := ""
		switch sortBy {
		case "ip":
			col = "ip_address"
		case "hostname":
			col = "hostname"
		case "vendor":
			col = "mac_vendor"
		case "last_seen":
			col = "last_seen"
		case "ping", "latency":
			col = "availability_pct"
		default:
			col = "ip_address"
		}
		if col != "" {
			if sortOrder != "desc" {
				sortOrder = "asc"
			}
			order = fmt.Sprintf("%s %s", col, sortOrder)
		}
	}

	var devices []model.Device
	var err error

	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	if pageStr != "" || limitStr != "" {
		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)
		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 10
		}
		offset := (page - 1) * limit
		err = query.Order(order).Limit(limit).Offset(offset).Find(&devices).Error
	} else {
		err = query.Order(order).Find(&devices).Error
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pageStr != "" {
		c.JSON(http.StatusOK, gin.H{
			"data":  devices,
			"total": total,
		})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func (h *Handlers) GetDeviceDetail(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	device, err := h.deviceService.GetDevice(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	var timeline []model.DeviceTimeline
	repository.DB.Where("device_id = ?", id).Order("checked_at desc").Limit(30).Find(&timeline)

	var ports []int
	if device.OpenPorts != "" {
		_ = json.Unmarshal([]byte(device.OpenPorts), &ports)
	}

	var interfaces interface{}
	if device.SNMPInterfaces != "" {
		_ = json.Unmarshal([]byte(device.SNMPInterfaces), &interfaces)
	}

	c.JSON(http.StatusOK, gin.H{
		"device":     device,
		"timeline":   timeline,
		"ports":      ports,
		"interfaces": interfaces,
	})
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
	orgID, _ := middleware.GetTenantID(c)
	var req struct {
		Target  string `json:"target"`
		CIDR    string `json:"cidr"`
		Profile string `json:"profile"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	target := req.Target
	if target == "" {
		target = req.CIDR
	}
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target or CIDR is required"})
		return
	}

	profile := req.Profile
	if profile == "" {
		profile = "quick"
	}

	scanID, err := h.discoveryService.StartScan(orgID, target, profile, "manual")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"scan_id": scanID, "status": "running"})
}

func (h *Handlers) CancelScan(c *gin.Context) {
	var req struct {
		ScanID uuid.UUID `json:"scan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cancelled := scanner.CancelScan(req.ScanID)
	if !cancelled {
		c.JSON(http.StatusNotFound, gin.H{"error": "No running scan found matching that ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scan cancel request transmitted."})
}

func (h *Handlers) GetScanHistory(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var history []model.ScanHistory
	err := repository.DB.Where("organization_id = ?", orgID).Order("started_at desc").Find(&history).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *Handlers) ListScanSchedules(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var schedules []model.ScanSchedule
	err := repository.DB.Where("organization_id = ?", orgID).Find(&schedules).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

func (h *Handlers) CreateScanSchedule(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var schedule model.ScanSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule.ID = uuid.New()
	schedule.OrganizationID = orgID
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	if err := repository.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

func (h *Handlers) DeleteScanSchedule(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	if err := repository.DB.Where("id = ? AND organization_id = ?", id, orgID).Delete(&model.ScanSchedule{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule rule deleted"})
}

func (h *Handlers) GetStatistics(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)

	var totalDevices int64
	var onlineDevices int64
	var offlineDevices int64

	repository.DB.Model(&model.Device{}).Where("organization_id = ?", orgID).Count(&totalDevices)
	repository.DB.Model(&model.Device{}).Where("organization_id = ? AND status = 'online'", orgID).Count(&onlineDevices)
	repository.DB.Model(&model.Device{}).Where("organization_id = ? AND status = 'offline'", orgID).Count(&offlineDevices)

	// Fetch device type distribution
	type Row struct {
		DeviceType string `json:"device_type"`
		Count      int64  `json:"count"`
	}
	var types []Row
	repository.DB.Model(&model.Device{}).
		Select("device_type, count(*) as count").
		Where("organization_id = ?", orgID).
		Group("device_type").
		Scan(&types)

	c.JSON(http.StatusOK, gin.H{
		"total_devices":   totalDevices,
		"online_devices":  onlineDevices,
		"offline_devices": offlineDevices,
		"types":           types,
	})
}

func (h *Handlers) ListNotifications(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var timelines []model.DeviceTimeline
	err := repository.DB.
		Joins("JOIN devices ON device_timelines.device_id = devices.id").
		Where("devices.organization_id = ?", orgID).
		Order("checked_at desc").
		Limit(30).
		Find(&timelines).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timelines)
}

func (h *Handlers) ListDeviceTypes(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var types []string
	err := repository.DB.Model(&model.Device{}).
		Where("organization_id = ?", orgID).
		Distinct("device_type").
		Pluck("device_type", &types).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types)
}

func (h *Handlers) ListVendors(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var vendors []string
	err := repository.DB.Model(&model.Device{}).
		Where("organization_id = ? AND mac_vendor IS NOT NULL AND mac_vendor != ''", orgID).
		Distinct("mac_vendor").
		Pluck("mac_vendor", &vendors).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vendors)
}

func (h *Handlers) ExportDevicesPDF(c *gin.Context) {
	orgID, _ := middleware.GetTenantID(c)
	var devices []model.Device
	err := repository.DB.Where("organization_id = ?", orgID).Order("ip_address asc").Find(&devices).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate a beautiful, print-ready HTML page.
	// Users can easily save/print this to PDF using native browser shortcuts (Ctrl+P / Cmd+P).
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Network Device Inventory Report</title>
    <style>
        body { font-family: 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #1e293b; margin: 30px; }
        .header { display: flex; justify-content: space-between; align-items: center; border-bottom: 2px solid #e2e8f0; padding-bottom: 15px; margin-bottom: 30px; }
        .title h1 { margin: 0; font-size: 24px; color: #4f46e5; }
        .title p { margin: 5px 0 0 0; font-size: 13px; color: #64748b; }
        .meta { text-align: right; font-size: 13px; color: #64748b; }
        table { width: 100%; border-collapse: collapse; margin-top: 15px; }
        th { background-color: #f8fafc; color: #475569; font-weight: bold; text-align: left; font-size: 12px; text-transform: uppercase; border-bottom: 2px solid #cbd5e1; padding: 10px 8px; }
        td { border-bottom: 1px solid #e2e8f0; padding: 10px 8px; font-size: 13px; }
        tr:hover { background-color: #f8fafc; }
        .status-online { color: #16a34a; font-weight: bold; }
        .status-offline { color: #dc2626; font-weight: bold; }
        .badge { display: inline-block; padding: 3px 6px; border-radius: 4px; font-size: 11px; font-weight: bold; text-transform: uppercase; }
        .badge-type { background-color: #e0e7ff; color: #4338ca; }
        .ports { font-family: monospace; font-size: 11px; color: #475569; word-break: break-all; }
        @media print {
            body { margin: 0; }
            .no-print { display: none; }
            button { display: none; }
        }
        .btn-print { background-color: #4f46e5; color: white; border: none; padding: 8px 16px; border-radius: 6px; font-weight: bold; cursor: pointer; font-size: 13px; }
        .btn-print:hover { background-color: #4338ca; }
    </style>
</head>
<body>
    <div class="header">
        <div class="title">
            <h1>Network Discovery Inventory Report</h1>
            <p>Generated dynamically from Device Discovery Controller</p>
        </div>
        <div class="meta">
            <div><strong>Date:</strong> ` + time.Now().Format("2006-01-02 15:04:05") + `</div>
            <div style="margin-top: 5px;"><button class="btn-print" onclick="window.print()">Print / Save PDF</button></div>
        </div>
    </div>
    
    <table>
        <thead>
            <tr>
                <th>IP Address</th>
                <th>Hostname</th>
                <th>MAC Address</th>
                <th>Vendor</th>
                <th>Device Type</th>
                <th>OS</th>
                <th>Status</th>
                <th>Availability</th>
                <th>Open Ports</th>
            </tr>
        </thead>
        <tbody>`

	for _, d := range devices {
		statusClass := "status-offline"
		if d.Status == "online" {
			statusClass = "status-online"
		}
		
		portsStr := d.OpenPorts
		if portsStr == "" || portsStr == "[]" {
			portsStr = "None"
		}

		html += fmt.Sprintf(`
            <tr>
                <td style="font-family: monospace; font-weight: bold;">%s</td>
                <td>%s</td>
                <td style="font-family: monospace;">%s</td>
                <td>%s</td>
                <td><span class="badge badge-type">%s</span></td>
                <td>%s</td>
                <td class="%s">%s</td>
                <td style="font-family: monospace;">%.1f%%</td>
                <td class="ports">%s</td>
            </tr>`,
			d.IPAddress, d.Hostname, d.MACAddress, d.MACVendor, d.DeviceType, d.OS, statusClass, d.Status, d.AvailabilityPct, portsStr)
	}

	html += `
        </tbody>
    </table>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
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

	// Enrich with details on scans
	var lastScan model.ScanHistory
	var currentNetwork string = "N/A"
	var lastScanDuration float64 = 0.0
	var lastScanTime string = "N/A"

	err = repository.DB.Where("organization_id = ? AND status = 'completed'", orgID).Order("started_at desc").First(&lastScan).Error
	if err == nil {
		currentNetwork = lastScan.Target
		lastScanDuration = float64(lastScan.DurationMS) / 1000.0 // to seconds
		lastScanTime = lastScan.EndedAt.Format("15:04:05")
	}

	// We append scanner telemetry to the dashboard stats JSON
	c.JSON(http.StatusOK, gin.H{
		"total_devices":      stats.TotalDevices,
		"online_devices":     stats.OnlineDevices,
		"offline_devices":    stats.OfflineDevices,
		"warning_devices":    stats.WarningDevices,
		"unreachable_devices": stats.UnreachableDevices,
		"active_alerts":      stats.ActiveAlerts,
		"avg_latency_ms":     stats.AvgLatencyMS,
		"avg_packet_loss":    stats.AvgPacketLoss,
		"avg_cpu":            stats.AvgCPU,
		"avg_ram":            stats.AvgRAM,
		"avg_disk":           stats.AvgDisk,
		"current_network":    currentNetwork,
		"scan_duration":      lastScanDuration,
		"last_scan_time":     lastScanTime,
	})
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

