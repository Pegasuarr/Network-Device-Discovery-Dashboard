package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
	"github.com/user/network-monitoring/internal/scanner"
	"github.com/user/network-monitoring/internal/websocket"
)

type DiscoveryService struct {
	hub *websocket.Hub
}

func NewDiscoveryService(hub *websocket.Hub) *DiscoveryService {
	return &DiscoveryService{
		hub: hub,
	}
}

// StartScan initiates a concurrent network discovery scan for a CIDR or IP address.
func (s *DiscoveryService) StartScan(orgID uuid.UUID, target string, profile string, scanType string) (uuid.UUID, error) {
	// 1. Parse target to get a list of IP addresses
	ips, err := parseTargetIPs(target)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse scan target: %v", err)
	}

	scanID := uuid.New()

	// 2. Create ScanHistory entry
	history := &model.ScanHistory{
		ID:             scanID,
		OrganizationID: orgID,
		Target:         target,
		ScanProfile:    profile,
		StartedAt:      time.Now(),
		Status:         "running",
		ScanType:       scanType,
	}

	if err := repository.DB.Create(history).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to log scan history: %v", err)
	}

	// 3. Create scan context
	ctx, cancel := context.WithCancel(context.Background())
	scanner.RegisterScan(scanID, cancel)

	// 4. Execute scan in background
	go func() {
		defer scanner.DeregisterScan(scanID)
		startTime := time.Now()

		progressChan := make(chan scanner.ProgressReport, 50)
		
		// Run progress listener in parallel to stream progress over WebSocket
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for report := range progressChan {
				s.hub.Broadcast <- websocket.BroadcastEvent{
					OrgID: orgID,
					Type:  "scan_progress",
					Payload: report,
				}
			}
		}()

		// Run the worker pool scan
		discovered, err := scanner.WorkerPoolScan(ctx, scanID, ips, profile, progressChan)
		close(progressChan)
		wg.Wait()

		duration := time.Since(startTime).Milliseconds()
		endedAt := time.Now()

		historyUpdate := map[string]interface{}{
			"ended_at":      &endedAt,
			"duration_ms":   duration,
			"devices_found": len(discovered),
		}

		if err != nil {
			if ctx.Err() == context.Canceled {
				historyUpdate["status"] = "cancelled"
				slog.Info("Scan cancelled by user", "scan_id", scanID)
			} else {
				historyUpdate["status"] = "failed"
				slog.Error("Scan failed during concurrent probe", "scan_id", scanID, "error", err)
			}
		} else {
			historyUpdate["status"] = "completed"
			slog.Info("Scan completed successfully", "scan_id", scanID, "devices_found", len(discovered))

			// Ingest discovered devices into database
			s.ingestDiscoveredDevices(orgID, discovered)
		}

		// Update database history
		repository.DB.Model(&model.ScanHistory{}).Where("id = ?", scanID).Updates(historyUpdate)

		// Broadcast final scan status
		finalStatus := "completed"
		if historyUpdate["status"] == "cancelled" {
			finalStatus = "cancelled"
		} else if historyUpdate["status"] == "failed" {
			finalStatus = "failed"
		}

		s.hub.Broadcast <- websocket.BroadcastEvent{
			OrgID: orgID,
			Type:  "scan_progress",
			Payload: scanner.ProgressReport{
				ScanID:       scanID,
				TotalIPs:     len(ips),
				ScannedIPs:   len(ips),
				Percent:      100,
				DevicesFound: len(discovered),
				Status:       finalStatus,
			},
		}
	}()

	return scanID, nil
}

// Ingest discovered devices, updates statistics and log device timelines.
func (s *DiscoveryService) ingestDiscoveredDevices(orgID uuid.UUID, hosts []scanner.DiscoveredHost) {
	for _, host := range hosts {
		var dev model.Device
		// Look up if device already exists by IP & Organization
		err := repository.DB.Where("organization_id = ? AND ip_address = ?", orgID, host.IPAddress).First(&dev).Error

		openPortsJSON, _ := json.Marshal(host.OpenPorts)
		snmpInterfacesJSON, _ := json.Marshal(host.SNMPInterfaces)

		if err == nil {
			// Device already exists, update stats and details
			prevStatus := dev.Status
			wasOffline := prevStatus == "offline"

			dev.Hostname = host.Hostname
			dev.MACAddress = host.MACAddress
			dev.MACVendor = host.MACVendor
			dev.Vendor = host.Vendor
			dev.OS = host.OS
			dev.DeviceType = host.DeviceType
			dev.OpenPorts = string(openPortsJSON)
			dev.Status = "online"
			dev.LastSeen = time.Now()
			dev.NumberOfScans++

			// Simple availability calculation:
			// Let's assume it was online for this scan. We increase its uptime stats.
			// availability = (online scans / total scans) * 100
			// (We count its historical availability)
			dev.TotalOnlineTime += int64(dev.MonitoringInterval)
			dev.AvailabilityPct = (float64(dev.TotalOnlineTime) / float64(int64(dev.NumberOfScans)*int64(dev.MonitoringInterval))) * 100.0
			if dev.AvailabilityPct > 100.0 {
				dev.AvailabilityPct = 100.0
			}

			// SNMP details
			dev.SNMPEnabled = host.SNMPEnabled
			dev.SNMPSysName = host.SNMPSysName
			dev.SNMPSysDescr = host.SNMPSysDescr
			dev.SNMPSysUptime = host.SNMPSysUptime
			dev.SNMPCPUUsage = host.SNMPCPUUsage
			dev.SNMPRAMUsage = host.SNMPRAMUsage
			dev.SNMPInterfaces = string(snmpInterfacesJSON)

			repository.DB.Save(&dev)

			// Record timeline event if status changed
			if wasOffline {
				timeline := &model.DeviceTimeline{
					ID:        uuid.New(),
					DeviceID:  dev.ID,
					EventType: "online",
					Message:   fmt.Sprintf("Device came online. IP: %s, MAC: %s", dev.IPAddress, dev.MACAddress),
					CheckedAt: time.Now(),
				}
				repository.DB.Create(timeline)

				// Broadcast toast alert
				s.hub.Broadcast <- websocket.BroadcastEvent{
					OrgID: orgID,
					Type:  "device_notification",
					Payload: map[string]interface{}{
						"type":       "online",
						"device_id":  dev.ID,
						"ip_address": dev.IPAddress,
						"name":       dev.Name,
						"vendor":     dev.MACVendor,
					},
				}
			}
		} else {
			// Brand new device discovered
			newID := uuid.New()
			dev = model.Device{
				ID:                 newID,
				OrganizationID:     orgID,
				Name:               host.Hostname,
				Hostname:           host.Hostname,
				IPAddress:          host.IPAddress,
				MACAddress:         host.MACAddress,
				MACVendor:          host.MACVendor,
				Vendor:             host.Vendor,
				OS:                 host.OS,
				DeviceType:         host.DeviceType,
				Status:             "online",
				MonitoringInterval: 60,
				Enabled:            true,
				FirstSeen:          time.Now(),
				LastSeen:           time.Now(),
				TotalOnlineTime:    60,
				NumberOfScans:      1,
				AvailabilityPct:    100.0,
				OpenPorts:          string(openPortsJSON),
				SNMPEnabled:        host.SNMPEnabled,
				SNMPSysName:        host.SNMPSysName,
				SNMPSysDescr:       host.SNMPSysDescr,
				SNMPSysUptime:      host.SNMPSysUptime,
				SNMPCPUUsage:       host.SNMPCPUUsage,
				SNMPRAMUsage:       host.SNMPRAMUsage,
				SNMPInterfaces:     string(snmpInterfacesJSON),
			}

			// Save to database
			repository.DB.Create(&dev)

			// Log timeline event
			timeline := &model.DeviceTimeline{
				ID:        uuid.New(),
				DeviceID:  dev.ID,
				EventType: "join",
				Message:   fmt.Sprintf("New device joined the network. IP: %s, MAC: %s, Vendor: %s", dev.IPAddress, dev.MACAddress, dev.MACVendor),
				CheckedAt: time.Now(),
			}
			repository.DB.Create(timeline)

			// Broadcast join toast
			s.hub.Broadcast <- websocket.BroadcastEvent{
				OrgID: orgID,
				Type:  "device_notification",
				Payload: map[string]interface{}{
					"type":       "join",
					"device_id":  dev.ID,
					"ip_address": dev.IPAddress,
					"name":       dev.Hostname,
					"vendor":     dev.MACVendor,
				},
			}
		}
	}
}

// Helper to expand targets into lists of discrete IP addresses.
func parseTargetIPs(target string) ([]string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, fmt.Errorf("empty target")
	}

	// 1. Check if target contains CIDR notation (e.g. 192.168.1.0/24)
	if strings.Contains(target, "/") {
		ip, ipnet, err := net.ParseCIDR(target)
		if err != nil {
			return nil, err
		}

		var ips []string
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
			ips = append(ips, ip.String())
		}

		// Omit subnet and broadcast addresses for standard /24 and smaller subnets
		if len(ips) > 2 {
			return ips[1 : len(ips)-1], nil
		}
		return ips, nil
	}

	// 2. Check if target is a simple single IP
	parsedIP := net.ParseIP(target)
	if parsedIP != nil {
		return []string{target}, nil
	}

	// 3. Fallback check for hyphenated IP ranges (e.g. 192.168.1.1-192.168.1.50)
	if strings.Contains(target, "-") {
		parts := strings.Split(target, "-")
		if len(parts) == 2 {
			startIP := net.ParseIP(strings.TrimSpace(parts[0]))
			endIP := net.ParseIP(strings.TrimSpace(parts[1]))
			if startIP != nil && endIP != nil {
				return generateIPRange(startIP, endIP), nil
			}
		}
	}

	return nil, fmt.Errorf("unsupported target format: %s", target)
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func generateIPRange(start, end net.IP) []string {
	var ips []string
	curr := make(net.IP, len(start))
	copy(curr, start)

	for {
		ips = append(ips, curr.String())
		if curr.Equal(end) {
			break
		}
		incIP(curr)
		// prevent runaway loops
		if len(ips) > 1000 {
			break
		}
	}
	return ips
}
