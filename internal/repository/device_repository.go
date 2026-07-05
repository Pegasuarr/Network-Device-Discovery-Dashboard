package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/user/network-monitoring/internal/model"
)

type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(device *model.Device) error {
	return r.db.Create(device).Error
}

func (r *DeviceRepository) FindByID(id uuid.UUID, orgID uuid.UUID) (*model.Device, error) {
	var device model.Device
	err := r.db.Preload("Group").Where("id = ? AND organization_id = ?", id, orgID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepository) ListByOrg(orgID uuid.UUID) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Preload("Group").Where("organization_id = ?", orgID).Find(&devices).Error
	return devices, err
}

func (r *DeviceRepository) ListActive() ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Where("enabled = ?", true).Find(&devices).Error
	return devices, err
}

func (r *DeviceRepository) Update(device *model.Device) error {
	return r.db.Save(device).Error
}

func (r *DeviceRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&model.Device{}).Where("id = ?", id).Update("status", status).Error
}

func (r *DeviceRepository) Delete(id uuid.UUID, orgID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, orgID).Delete(&model.Device{}).Error
}

// Device Groups
func (r *DeviceRepository) CreateGroup(group *model.DeviceGroup) error {
	return r.db.Create(group).Error
}

func (r *DeviceRepository) ListGroups(orgID uuid.UUID) ([]model.DeviceGroup, error) {
	var groups []model.DeviceGroup
	err := r.db.Where("organization_id = ?", orgID).Find(&groups).Error
	return groups, err
}

func (r *DeviceRepository) DeleteGroup(id uuid.UUID, orgID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, orgID).Delete(&model.DeviceGroup{}).Error
}
