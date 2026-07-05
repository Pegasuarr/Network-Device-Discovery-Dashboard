package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/user/network-monitoring/internal/arp"
	"github.com/user/network-monitoring/internal/fingerprint"
	"github.com/user/network-monitoring/internal/ping"
	"github.com/user/network-monitoring/internal/portscan"
	"github.com/user/network-monitoring/internal/snmp"
	"github.com/user/network-monitoring/internal/vendor"
)

// ScanProfile constants
const (
	ProfileQuick    = "quick"
	ProfileDeep     = "deep"
	ProfilePortScan = "portscan"
	ProfilePingOnly = "ping"
)

// DiscoveredHost represents a device found during scanning.
type DiscoveredHost struct {
	IPAddress      string        `json:"ip_address"`
	Hostname       string        `json:"hostname"`
	MACAddress     string        `json:"mac_address"`
	MACVendor      string        `json:"mac_vendor"`
	Vendor         string        `json:"vendor"`
	OS             string        `json:"os"`
	DeviceType     string        `json:"device_type"`
	OpenPorts      []int         `json:"open_ports"`
	PingTimeMS     float64       `json:"ping_time_ms"`
	Status         string        `json:"status"` // online, offline
	SNMPEnabled    bool          `json:"snmp_enabled"`
	SNMPSysName    string        `json:"snmp_sys_name"`
	SNMPSysDescr   string        `json:"snmp_sys_descr"`
	SNMPSysUptime  uint32        `json:"snmp_sys_uptime"`
	SNMPCPUUsage   float64       `json:"snmp_cpu_usage"`
	SNMPRAMUsage   float64       `json:"snmp_ram_usage"`
	SNMPInterfaces []snmp.InterfaceInfo `json:"snmp_interfaces"`
}

// ProgressReport represents scan progress sent over websockets.
type ProgressReport struct {
	ScanID       uuid.UUID        `json:"scan_id"`
	TotalIPs     int              `json:"total_ips"`
	ScannedIPs   int              `json:"scanned_ips"`
	Percent      int              `json:"percent"`
	DevicesFound int              `json:"devices_found"`
	CurrentIP    string           `json:"current_ip"`
	Status       string           `json:"status"` // running, completed, cancelled
	LatestDevice *DiscoveredHost  `json:"latest_device,omitempty"`
}

// Registry to track running scans and their cancel functions.
var (
	ActiveScans   = make(map[uuid.UUID]context.CancelFunc)
	ActiveScansMu sync.Mutex
)

// RegisterScan registers a scan's cancel function.
func RegisterScan(id uuid.UUID, cancel context.CancelFunc) {
	ActiveScansMu.Lock()
	defer ActiveScansMu.Unlock()
	ActiveScans[id] = cancel
}

// CancelScan cancels a running scan by ID.
func CancelScan(id uuid.UUID) bool {
	ActiveScansMu.Lock()
	defer ActiveScansMu.Unlock()
	if cancel, found := ActiveScans[id]; found {
		cancel()
		delete(ActiveScans, id)
		return true
	}
	return false
}

// DeregisterScan removes a scan from the registry.
func DeregisterScan(id uuid.UUID) {
	ActiveScansMu.Lock()
	defer ActiveScansMu.Unlock()
	delete(ActiveScans, id)
}

// WorkerPoolScan coordinates concurrent scanning of target IPs using a pool of workers.
func WorkerPoolScan(
	ctx context.Context,
	scanID uuid.UUID,
	ips []string,
	profile string,
	progressChan chan<- ProgressReport,
) ([]DiscoveredHost, error) {
	totalIPs := len(ips)
	if totalIPs == 0 {
		return nil, fmt.Errorf("no IP addresses to scan")
	}

	var scannedCount int32
	var foundCount int32

	ipChan := make(chan string, totalIPs)
	resultsChan := make(chan DiscoveredHost, totalIPs)

	// Populate IP queue
	for _, ip := range ips {
		ipChan <- ip
	}
	close(ipChan)

	// Determine worker count based on size
	workerCount := 150
	if totalIPs < workerCount {
		workerCount = totalIPs
	}

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for targetIP := range ipChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				host, alive := ScanIP(ctx, targetIP, profile)
				atomic.AddInt32(&scannedCount, 1)

				currentScanned := int(atomic.LoadInt32(&scannedCount))
				percent := (currentScanned * 100) / totalIPs

				var latestDevice *DiscoveredHost
				if alive {
					resultsChan <- host
					atomic.AddInt32(&foundCount, 1)
					latestDevice = &host
				}

				// Report progress
				select {
				case progressChan <- ProgressReport{
					ScanID:       scanID,
					TotalIPs:     totalIPs,
					ScannedIPs:   currentScanned,
					Percent:      percent,
					DevicesFound: int(atomic.LoadInt32(&foundCount)),
					CurrentIP:    targetIP,
					Status:       "running",
					LatestDevice: latestDevice,
				}:
				default:
					// Don't block workers if queue is full
				}
			}
		}()
	}

	wg.Wait()
	close(resultsChan)

	var discovered []DiscoveredHost
	for res := range resultsChan {
		discovered = append(discovered, res)
	}

	return discovered, nil
}

// ScanIP evaluates a single IP based on the profile
func ScanIP(ctx context.Context, ip string, profile string) (DiscoveredHost, bool) {
	// 1. Check if host is alive via Ping
	pingTimeout := 800 * time.Millisecond
	if profile == ProfileDeep {
		pingTimeout = 1200 * time.Millisecond
	}

	latency, alive := ping.Ping(ctx, ip, pingTimeout)
	
	// Fallback to TCP checks for Quick and Deep profiles
	if !alive && (profile == ProfileQuick || profile == ProfileDeep) {
		// Try TCP connection to port 80 or 443
		for _, port := range []int{80, 443, 22, 3389} {
			address := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
			dialer := net.Dialer{Timeout: 200 * time.Millisecond}
			conn, err := dialer.DialContext(ctx, "tcp", address)
			if err == nil {
				conn.Close()
				alive = true
				latency = 10.0 // assume 10ms
				break
			}
		}
	}

	if !alive {
		return DiscoveredHost{}, false
	}

	// 2. Resolve DNS Hostname
	hostname := ""
	addrs, err := net.LookupAddr(ip)
	if err == nil && len(addrs) > 0 {
		hostname = strings.TrimSuffix(addrs[0], ".")
	} else {
		hostname = fmt.Sprintf("host-%s", strings.ReplaceAll(ip, ".", "-"))
	}

	// 3. Obtain MAC Address via local ARP table
	mac := ""
	macVendor := "Unknown Vendor"
	if localMAC, ok := arp.Lookup(ctx, ip); ok {
		mac = localMAC
		macVendor = vendor.Lookup(mac)
	}

	host := DiscoveredHost{
		IPAddress:  ip,
		Hostname:   hostname,
		MACAddress: mac,
		MACVendor:  macVendor,
		Vendor:     macVendor, // Default same as OUI vendor
		PingTimeMS: latency,
		Status:     "online",
		OS:         "Unknown OS",
		DeviceType: "workstation",
	}

	// 4. Port scan and fingerprinting depending on profile
	if profile == ProfileDeep || profile == ProfilePortScan {
		portTimeout := 300 * time.Millisecond
		if profile == ProfileDeep {
			portTimeout = 500 * time.Millisecond
		}

		// Perform port scan
		host.OpenPorts = portscan.Scan(ctx, ip, nil, portTimeout)

		// OS & Device Type fingerprinting
		sig := fingerprint.Fingerprint(ctx, ip, host.OpenPorts)
		host.OS = sig.OS
		host.DeviceType = sig.DeviceType
		if sig.Vendor != "Unknown Vendor" {
			host.Vendor = sig.Vendor
		}
	} else if profile == ProfileQuick {
		// Just run a quick check on basic ports to classify device
		quickPorts := []int{22, 80, 443, 3389, 445}
		host.OpenPorts = portscan.Scan(ctx, ip, quickPorts, 150*time.Millisecond)
		sig := fingerprint.Fingerprint(ctx, ip, host.OpenPorts)
		host.OS = sig.OS
		host.DeviceType = sig.DeviceType
		if sig.Vendor != "Unknown Vendor" {
			host.Vendor = sig.Vendor
		}
	}

	// 5. SNMP Probe for Router / Switch / Printer or if port 161 is open
	if profile == ProfileDeep || (profile == ProfileQuick && (host.DeviceType == "router" || host.DeviceType == "switch" || host.DeviceType == "printer")) {
		snmpData := snmp.Probe(ctx, ip, host.DeviceType, host.OS)
		if snmpData.Enabled {
			host.SNMPEnabled = true
			if snmpData.SysName != "" {
				host.SNMPSysName = snmpData.SysName
			}
			if snmpData.SysDescr != "" {
				host.SNMPSysDescr = snmpData.SysDescr
			}
			host.SNMPSysUptime = snmpData.SysUptime
			host.SNMPCPUUsage = snmpData.CPUUsage
			host.SNMPRAMUsage = snmpData.RAMUsage
			
			// Interfaces parse
			if snmpData.Interfaces != "" {
				var ifaces []snmp.InterfaceInfo
				if err := json.Unmarshal([]byte(snmpData.Interfaces), &ifaces); err == nil {
					host.SNMPInterfaces = ifaces
				}
			}
		}
	}

	return host, true
}
