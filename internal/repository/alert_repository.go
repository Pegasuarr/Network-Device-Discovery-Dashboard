package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/user/network-monitoring/internal/model"
)

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(alert *model.Alert) error {
	return r.db.Create(alert).Error
}

func (r *AlertRepository) FindByID(id uuid.UUID, orgID uuid.UUID) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.Preload("Device").Where("id = ? AND organization_id = ?", id, orgID).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *AlertRepository) ListActive(orgID uuid.UUID) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.Preload("Device").Where("organization_id = ? AND status = ?", orgID, "active").Find(&alerts).Error
	return alerts, err
}

func (r *AlertRepository) ListAll(orgID uuid.UUID) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.Preload("Device").Where("organization_id = ?", orgID).Order("created_at desc").Find(&alerts).Error
	return alerts, err
}

func (r *AlertRepository) Resolve(id uuid.UUID, orgID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&model.Alert{}).
		Where("id = ? AND organization_id = ?", id, orgID).
		Updates(map[string]interface{}{"status": "resolved", "resolved_at": &now}).Error
}

func (r *AlertRepository) ResolveActiveByDeviceAndType(deviceID uuid.UUID, alertType string) error {
	now := time.Now()
	return r.db.Model(&model.Alert{}).
		Where("device_id = ? AND type = ? AND status = ?", deviceID, alertType, "active").
		Updates(map[string]interface{}{"status": "resolved", "resolved_at": &now}).Error
}

// Alert Rules
func (r *AlertRepository) CreateRule(rule *model.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *AlertRepository) ListRules(orgID uuid.UUID) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("organization_id = ? OR organization_id = ?", orgID, DefaultOrgID).Find(&rules).Error
	return rules, err
}

func (r *AlertRepository) FindRuleByID(id uuid.UUID, orgID uuid.UUID) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.Where("id = ? AND organization_id = ?", id, orgID).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *AlertRepository) UpdateRule(rule *model.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *AlertRepository) DeleteRule(id uuid.UUID, orgID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, orgID).Delete(&model.AlertRule{}).Error
}

// Notification Logs
func (r *AlertRepository) CreateNotificationLog(log *model.NotificationLog) error {
	return r.db.Create(log).Error
}
