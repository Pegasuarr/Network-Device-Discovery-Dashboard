package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
)

type DashboardService struct {
	deviceRepo *repository.DeviceRepository
	alertRepo  *repository.AlertRepository
}

func NewDashboardService(deviceRepo *repository.DeviceRepository, alertRepo *repository.AlertRepository) *DashboardService {
	return &DashboardService{
		deviceRepo: deviceRepo,
		alertRepo:  alertRepo,
	}
}

type DashboardStats struct {
	TotalDevices       int            `json:"total_devices"`
	OnlineDevices      int            `json:"online_devices"`
	OfflineDevices     int            `json:"offline_devices"`
	WarningDevices     int            `json:"warning_devices"`
	UnreachableDevices int            `json:"unreachable_devices"`
	ActiveAlerts       int            `json:"active_alerts"`
	AvgLatencyMS       float64        `json:"avg_latency_ms"`
	AvgPacketLoss      float64        `json:"avg_packet_loss"`
	AvgCPU             float64        `json:"avg_cpu"`
	AvgRAM             float64        `json:"avg_ram"`
	AvgDisk            float64        `json:"avg_disk"`
	DeviceTypeCounts   map[string]int `json:"device_type_counts"`
}

type LatencyTrendPoint struct {
	Time      string  `json:"time"`
	LatencyMS float64 `json:"latency_ms"`
}

func (s *DashboardService) GetStats(orgID uuid.UUID) (*DashboardStats, error) {
	devices, err := s.deviceRepo.ListByOrg(orgID)
	if err != nil {
		return nil, err
	}

	stats := &DashboardStats{
		DeviceTypeCounts: make(map[string]int),
	}

	stats.TotalDevices = len(devices)
	deviceIDs := make([]uuid.UUID, 0, len(devices))

	for _, d := range devices {
		deviceIDs = append(deviceIDs, d.ID)
		switch d.Status {
		case "online":
			stats.OnlineDevices++
		case "offline":
			stats.OfflineDevices++
		case "warning":
			stats.WarningDevices++
		case "unreachable":
			stats.UnreachableDevices++
		}
		stats.DeviceTypeCounts[d.DeviceType]++
	}

	// Active Alerts
	activeAlerts, err := s.alertRepo.ListActive(orgID)
	if err == nil {
		stats.ActiveAlerts = len(activeAlerts)
	}

	// Compute averages from monitoring_results of the last 1 hour
	if len(deviceIDs) > 0 {
		var agg struct {
			AvgLatency    float64
			AvgPacketLoss float64
			AvgCPU        float64
			AvgRAM        float64
			AvgDisk       float64
		}
		oneHourAgo := time.Now().Add(-1 * time.Hour)
		err = repository.DB.Model(&model.MonitoringResult{}).
			Select("COALESCE(AVG(latency_ms), 0) as avg_latency, COALESCE(AVG(packet_loss_pct), 0) as avg_packet_loss, COALESCE(AVG(cpu_usage), 0) as avg_cpu, COALESCE(AVG(ram_usage), 0) as avg_ram, COALESCE(AVG(disk_usage), 0) as avg_disk").
			Where("device_id IN ? AND checked_at >= ?", deviceIDs, oneHourAgo).
			Scan(&agg).Error

		if err == nil {
			stats.AvgLatencyMS = agg.AvgLatency
			stats.AvgPacketLoss = agg.AvgPacketLoss
			stats.AvgCPU = agg.AvgCPU
			stats.AvgRAM = agg.AvgRAM
			stats.AvgDisk = agg.AvgDisk
		}
	}

	return stats, nil
}

func (s *DashboardService) GetGlobalLatencyTrend(orgID uuid.UUID) ([]LatencyTrendPoint, error) {
	devices, err := s.deviceRepo.ListByOrg(orgID)
	if err != nil || len(devices) == 0 {
		return []LatencyTrendPoint{}, nil
	}

	deviceIDs := make([]uuid.UUID, 0, len(devices))
	for _, d := range devices {
		deviceIDs = append(deviceIDs, d.ID)
	}

	// Fetch aggregated hourly latency averages for the last 24 hours
	type TrendRow struct {
		Hour      time.Time
		LatencyMS float64
	}

	var rows []TrendRow
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	// Group hourly using DATE_TRUNC (Postgres specific)
	err = repository.DB.Model(&model.MonitoringResult{}).
		Select("date_trunc('hour', checked_at) as hour, COALESCE(AVG(latency_ms), 0) as latency_ms").
		Where("device_id IN ? AND checked_at >= ?", deviceIDs, twentyFourHoursAgo).
		Group("hour").
		Order("hour asc").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	trendPoints := make([]LatencyTrendPoint, len(rows))
	for i, r := range rows {
		trendPoints[i] = LatencyTrendPoint{
			Time:      r.Hour.Format("15:04"),
			LatencyMS: r.LatencyMS,
		}
	}

	return trendPoints, nil
}
