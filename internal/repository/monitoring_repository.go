package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/user/network-monitoring/internal/model"
)

type MonitoringRepository struct {
	db *gorm.DB
}

func NewMonitoringRepository(db *gorm.DB) *MonitoringRepository {
	return &MonitoringRepository{db: db}
}

func (r *MonitoringRepository) CreateResult(res *model.MonitoringResult) error {
	return r.db.Create(res).Error
}

func (r *MonitoringRepository) GetRecentResults(deviceID uuid.UUID, limit int) ([]model.MonitoringResult, error) {
	var results []model.MonitoringResult
	err := r.db.Where("device_id = ?", deviceID).Order("checked_at desc").Limit(limit).Find(&results).Error
	return results, err
}

func (r *MonitoringRepository) GetResultsHistory(deviceID uuid.UUID, since time.Time) ([]model.MonitoringResult, error) {
	var results []model.MonitoringResult
	err := r.db.Where("device_id = ? AND checked_at >= ?", deviceID, since).Order("checked_at asc").Find(&results).Error
	return results, err
}

func (r *MonitoringRepository) GetUptimeStats(deviceID uuid.UUID, since time.Time) (float64, error) {
	// Calculate uptime based on status column on device or based on monitoring results status code / packet loss.
	// Since MonitoringResult has no status text directly, we check DNS, HTTP status, and packet loss.
	// A device is considered online in a result if PacketLossPct < 100.
	var total int64
	var online int64

	err := r.db.Model(&model.MonitoringResult{}).
		Where("device_id = ? AND checked_at >= ?", deviceID, since).
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 100.0, nil // Default to 100 if no data
	}

	err = r.db.Model(&model.MonitoringResult{}).
		Where("device_id = ? AND checked_at >= ? AND packet_loss_pct < 100", deviceID, since).
		Count(&online).Error
	if err != nil {
		return 0, err
	}

	return (float64(online) / float64(total)) * 100.0, nil
}
