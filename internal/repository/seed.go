package repository

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/user/network-monitoring/internal/model"
)

var DefaultOrgID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func Seed(db *gorm.DB) error {
	// 1. Seed Roles
	roles := []model.Role{
		{ID: 1, Name: "Admin", Description: "Administrator with full dashboard control"},
		{ID: 2, Name: "Operator", Description: "Operator with read/write access to devices and alerts"},
		{ID: 3, Name: "Viewer", Description: "Viewer with read-only access"},
	}
	for _, role := range roles {
		if err := db.FirstOrCreate(&role, model.Role{ID: role.ID}).Error; err != nil {
			return err
		}
	}

	// 2. Seed Default Organization
	org := model.Organization{
		ID:   DefaultOrgID,
		Name: "Enterprise Global NOC",
	}
	if err := db.FirstOrCreate(&org, model.Organization{ID: org.ID}).Error; err != nil {
		return err
	}

	// 3. Seed Admin User
	var adminUser model.User
	err := db.Where("username = ?", "admin").First(&adminUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		adminUser = model.User{
			ID:             uuid.New(),
			OrganizationID: DefaultOrgID,
			Username:       "admin",
			Email:          "admin@enterprise.local",
			PasswordHash:   string(hashedPassword),
			RoleID:         1, // Admin
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}
	}

	// 4. Seed Default Device Groups
	groups := []model.DeviceGroup{
		{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), OrganizationID: DefaultOrgID, Name: "Core Infrastructure", Description: "Routers and Firewalls"},
		{ID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), OrganizationID: DefaultOrgID, Name: "Production Servers", Description: "Critical application servers"},
	}
	for _, group := range groups {
		if err := db.FirstOrCreate(&group, model.DeviceGroup{ID: group.ID}).Error; err != nil {
			return err
		}
	}

	// 5. Seed Core Devices with Dependencies
	gatewayID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	switchID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	webServerID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	coreGroup := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	srvGroup := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	devices := []model.Device{
		{
			ID:                 gatewayID,
			OrganizationID:     DefaultOrgID,
			Name:               "Core Gateway",
			Hostname:           "gateway.noc",
			IPAddress:          "127.0.0.1",
			DeviceType:         "router",
			OS:                 "Cisco IOS",
			Vendor:             "Cisco",
			Location:           "Rack A-1",
			Status:             "online",
			MonitoringInterval: 30,
			Enabled:            true,
			GroupID:            &coreGroup,
		},
		{
			ID:                 switchID,
			OrganizationID:     DefaultOrgID,
			Name:               "Switch NOC 01",
			Hostname:           "switch01.noc",
			IPAddress:          "127.0.0.1",
			DeviceType:         "switch",
			OS:                 "Arista EOS",
			Vendor:             "Arista",
			Location:           "Rack A-2",
			Status:             "online",
			MonitoringInterval: 30,
			Enabled:            true,
			GroupID:            &coreGroup,
			ParentID:           &gatewayID, // child of gateway
		},
		{
			ID:                 webServerID,
			OrganizationID:     DefaultOrgID,
			Name:               "Enterprise Web Server",
			Hostname:           "google.com", // monitoring google.com for test
			IPAddress:          "8.8.8.8",
			DeviceType:         "server",
			OS:                 "Ubuntu 24.04 LTS",
			Vendor:             "Dell",
			Location:           "Rack B-3",
			Status:             "online",
			MonitoringInterval: 60,
			Enabled:            true,
			GroupID:            &srvGroup,
			ParentID:           &switchID, // child of switch
		},
	}

	for _, d := range devices {
		if err := db.FirstOrCreate(&d, model.Device{ID: d.ID}).Error; err != nil {
			return err
		}
	}

	// 6. Seed Default Alert Rules
	rules := []model.AlertRule{
		{
			ID:             uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			OrganizationID: DefaultOrgID,
			Name:           "Device Offline Threshold",
			Metric:         "status",
			Operator:       "==",
			Value:          0, // 0 = offline
			Duration:       30,
			Level:          "critical",
			Enabled:        true,
		},
		{
			ID:             uuid.MustParse("88888888-8888-8888-8888-888888888888"),
			OrganizationID: DefaultOrgID,
			Name:           "High Latency Threshold",
			Metric:         "latency_ms",
			Operator:       ">",
			Value:          150,
			Duration:       60,
			Level:          "warning",
			Enabled:        true,
		},
	}
	for _, rule := range rules {
		if err := db.FirstOrCreate(&rule, model.AlertRule{ID: rule.ID}).Error; err != nil {
			return err
		}
	}

	// 7. Seed Settings
	settings := []model.Setting{
		{OrganizationID: DefaultOrgID, Key: "smtp_host", Value: "smtp.mailtrap.io", Group: "smtp"},
		{OrganizationID: DefaultOrgID, Key: "smtp_port", Value: "2525", Group: "smtp"},
		{OrganizationID: DefaultOrgID, Key: "smtp_username", Value: "", Group: "smtp"},
		{OrganizationID: DefaultOrgID, Key: "smtp_password", Value: "", Group: "smtp"},
		{OrganizationID: DefaultOrgID, Key: "slack_webhook", Value: "", Group: "notification"},
		{OrganizationID: DefaultOrgID, Key: "telegram_token", Value: "", Group: "notification"},
		{OrganizationID: DefaultOrgID, Key: "telegram_chat_id", Value: "", Group: "notification"},
		{OrganizationID: DefaultOrgID, Key: "discord_webhook", Value: "", Group: "notification"},
	}
	for _, setting := range settings {
		if err := db.FirstOrCreate(&setting, model.Setting{OrganizationID: setting.OrganizationID, Key: setting.Key}).Error; err != nil {
			return err
		}
	}

	return nil
}
