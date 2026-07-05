package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/ping"
	"github.com/user/network-monitoring/internal/portscan"
	"github.com/user/network-monitoring/internal/repository"
	"github.com/user/network-monitoring/internal/service"
	"github.com/user/network-monitoring/internal/snmp"
	"github.com/user/network-monitoring/internal/websocket"
)

type Scheduler struct {
	alertService *service.AlertService
	hub          *websocket.Hub
	lastChecked  map[uuid.UUID]time.Time
	mu           sync.Mutex
}

func NewScheduler(alertService *service.AlertService, hub *websocket.Hub) *Scheduler {
	return &Scheduler{
		alertService: alertService,
		hub:          hub,
		lastChecked:  make(map[uuid.UUID]time.Time),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("Starting lightweight network monitoring background scheduler...")

	// Tick every 3 seconds to check for due devices
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping background monitoring scheduler...")
			return
		case <-ticker.C:
			s.pollAndExecute(ctx)
		}
	}
}

func (s *Scheduler) pollAndExecute(ctx context.Context) {
	var devices []model.Device
	err := repository.DB.Where("enabled = ?", true).Find(&devices).Error
	if err != nil {
		slog.Error("Scheduler failed to fetch devices", "error", err)
		return
	}

	now := time.Now()
	var wg sync.WaitGroup

	for _, device := range devices {
		s.mu.Lock()
		lastTime, exists := s.lastChecked[device.ID]
		s.mu.Unlock()

		interval := time.Duration(device.MonitoringInterval) * time.Second
		if interval <= 0 {
			interval = 60 * time.Second
		}

		if !exists || now.Sub(lastTime) >= interval {
			s.mu.Lock()
			s.lastChecked[device.ID] = now
			s.mu.Unlock()

			wg.Add(1)
			go func(d model.Device) {
				defer wg.Done()
				s.executeDeviceCheck(ctx, d)
			}(device)
		}
	}

	wg.Wait()
}

func (s *Scheduler) executeDeviceCheck(ctx context.Context, d model.Device) {
	// 1. Check maintenance mode
	now := time.Now()
	if d.MaintenanceStart != nil && d.MaintenanceEnd != nil {
		if now.After(*d.MaintenanceStart) && now.Before(*d.MaintenanceEnd) {
			if d.Status != "maintenance" {
				repository.DB.Model(&model.Device{}).Where("id = ?", d.ID).Update("status", "maintenance")
				s.broadcastStatus(d.OrganizationID, d.ID, "maintenance")
			}
			return
		}
	}

	// 2. Perform lightweight check (Ping or port 80/443 connection)
	latency, alive := ping.Ping(ctx, d.IPAddress, 600*time.Millisecond)
	if !alive && d.OpenPorts != "" {
		// Fallback to testing their open ports
		var ports []int
		if err := json.Unmarshal([]byte(d.OpenPorts), &ports); err == nil && len(ports) > 0 {
			// Scan a single open port
			open := portscan.Scan(ctx, d.IPAddress, []int{ports[0]}, 200*time.Millisecond)
			if len(open) > 0 {
				alive = true
				latency = 8.0
			}
		}
	}

	prevStatus := d.Status
	newStatus := "offline"
	if alive {
		newStatus = "online"
	}

	// 3. Dependency suppression
	if newStatus == "offline" && d.ParentID != nil {
		var parent model.Device
		if err := repository.DB.Where("id = ?", *d.ParentID).First(&parent).Error; err == nil {
			if parent.Status == "offline" {
				newStatus = "unreachable"
			}
		}
	}

	// 4. Update Device in DB
	updates := map[string]interface{}{
		"status":    newStatus,
		"last_seen": now,
	}

	if alive {
		// Increment online stats
		updates["total_online_time"] = d.TotalOnlineTime + int64(d.MonitoringInterval)
	}
	updates["number_of_scans"] = d.NumberOfScans + 1

	// Calculate availability percentage
	totalExpectedTime := (float64(d.NumberOfScans+1) * float64(d.MonitoringInterval))
	var onlineTime float64
	if alive {
		onlineTime = float64(d.TotalOnlineTime + int64(d.MonitoringInterval))
	} else {
		onlineTime = float64(d.TotalOnlineTime)
	}
	updates["availability_pct"] = (onlineTime / totalExpectedTime) * 100.0

	// 5. SNMP telemetry check if enabled
	var snmpCPU, snmpRAM float64
	var snmpUptime uint32
	if alive && d.SNMPEnabled {
		snmpData := snmp.Probe(ctx, d.IPAddress, d.DeviceType, d.OS)
		if snmpData.Enabled {
			updates["snmp_sys_uptime"] = snmpData.SysUptime
			updates["snmp_cpu_usage"] = snmpData.CPUUsage
			updates["snmp_ram_usage"] = snmpData.RAMUsage
			updates["snmp_interfaces"] = snmpData.Interfaces

			snmpCPU = snmpData.CPUUsage
			snmpRAM = snmpData.RAMUsage
			snmpUptime = snmpData.SysUptime
		}
	}

	repository.DB.Model(&model.Device{}).Where("id = ?", d.ID).Updates(updates)

	// Create a monitoring result record
	result := &model.MonitoringResult{
		ID:             uuid.New(),
		DeviceID:       d.ID,
		LatencyMS:      latency,
		PacketLossPct:  0.0,
		ResponseTimeMS: latency,
		DNSResolved:    true,
		CPUUsage:       snmpCPU,
		RAMUsage:       snmpRAM,
		CheckedAt:      now,
	}
	if !alive {
		result.PacketLossPct = 100.0
		result.LatencyMS = 0
		result.ResponseTimeMS = 0
	}
	repository.DB.Create(result)

	// Evaluate alert rules
	s.alertService.EvaluateRules(&d, result)

	// 6. Handle State Changes (log timelines & broadcast toast)
	if prevStatus != newStatus {
		timeline := &model.DeviceTimeline{
			ID:        uuid.New(),
			DeviceID:  d.ID,
			EventType: newStatus,
			Message:   fmt.Sprintf("Device status transitioned from %s to %s.", prevStatus, newStatus),
			CheckedAt: now,
		}
		repository.DB.Create(timeline)

		// Broadcast toast alert and status change
		s.hub.Broadcast <- websocket.BroadcastEvent{
			OrgID: d.OrganizationID,
			Type:  "device_notification",
			Payload: map[string]interface{}{
				"type":       newStatus, // "online" or "offline"
				"device_id":  d.ID,
				"ip_address": d.IPAddress,
				"name":       d.Name,
				"vendor":     d.MACVendor,
			},
		}

		s.broadcastStatus(d.OrganizationID, d.ID, newStatus)
	}

	// 7. Broadcast telemetry details
	s.broadcastResult(d.OrganizationID, d.ID, result, snmpUptime)
}

func (s *Scheduler) broadcastResult(orgID uuid.UUID, deviceID uuid.UUID, res *model.MonitoringResult, uptime uint32) {
	s.hub.Broadcast <- websocket.BroadcastEvent{
		OrgID: orgID,
		Type:  "ping_result",
		Payload: map[string]interface{}{
			"device_id":        deviceID,
			"latency_ms":       res.LatencyMS,
			"packet_loss_pct":  res.PacketLossPct,
			"response_time_ms": res.ResponseTimeMS,
			"cpu_usage":        res.CPUUsage,
			"ram_usage":        res.RAMUsage,
			"checked_at":       res.CheckedAt,
			"snmp_sys_uptime":  uptime,
		},
	}
}

func (s *Scheduler) broadcastStatus(orgID uuid.UUID, deviceID uuid.UUID, status string) {
	s.hub.Broadcast <- websocket.BroadcastEvent{
		OrgID: orgID,
		Type:  "device_status",
		Payload: map[string]interface{}{
			"device_id": deviceID,
			"status":    status,
		},
	}
}
