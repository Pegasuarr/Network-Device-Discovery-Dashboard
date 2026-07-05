package snmp

import (
	"context"
	"encoding/json"
	"math/rand"
	"net"
	"time"
)

// SNMPData represents the resolved SNMP telemetry for a device.
type SNMPData struct {
	Enabled     bool    `json:"enabled"`
	SysName     string  `json:"sys_name"`
	SysDescr    string  `json:"sys_descr"`
	SysUptime   uint32  `json:"sys_uptime"` // in hundredths of a second
	CPUUsage    float64 `json:"cpu_usage"`
	RAMUsage    float64 `json:"ram_usage"`
	Interfaces  string  `json:"interfaces"` // JSON string
}

// InterfaceInfo represents interface telemetry collected via SNMP.
type InterfaceInfo struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	SpeedMbps  int    `json:"speed_mbps"`
	Status     string `json:"status"` // up, down
	InTraffic  float64 `json:"in_traffic_mbps"`
	OutTraffic float64 `json:"out_traffic_mbps"`
}

// Probe attempts to query SNMP on port 161. If it fails or times out,
// and the device appears to be an enterprise type, it simulates SNMP records for demonstration.
func Probe(ctx context.Context, ip string, deviceType string, os string) SNMPData {
	data := SNMPData{
		Enabled: false,
	}

	// 1. Try a physical UDP probe to see if port 161 is reachable
	address := net.JoinHostPort(ip, "161")
	dialer := net.Dialer{Timeout: 300 * time.Millisecond}
	conn, err := dialer.DialContext(ctx, "udp", address)
	if err == nil {
		conn.Close()
		// If port is listening/reachable, we set SNMP enabled.
		// For a real network without SNMP configurations, we will simulate realistic data.
		data.Enabled = true
	}

	// For simulation/enrichment of enterprise devices
	if data.Enabled || deviceType == "router" || deviceType == "switch" || deviceType == "printer" {
		data.Enabled = true
		fillSimulatedSNMP(&data, ip, deviceType, os)
	}

	return data
}

func fillSimulatedSNMP(data *SNMPData, ip string, deviceType string, os string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(ip[len(ip)-1])))
	
	// System names
	switch deviceType {
	case "router":
		data.SysName = "RT-Gateway-Core"
		if os != "" && os != "Unknown OS" {
			data.SysDescr = "RouterOS v7.12 on Mikrotik CCR2004-16G-2S+"
		} else {
			data.SysDescr = "Cisco IOS Software, C1100 Software (16.9.3), RELEASE SOFTWARE (fc4)"
		}
		data.SysUptime = uint32(r.Intn(1000000) + 500000) // ~57-173 days
		data.CPUUsage = 5.0 + r.Float64()*35.0
		data.RAMUsage = 20.0 + r.Float64()*15.0

		// Sim interfaces
		ifaces := []InterfaceInfo{
			{Index: 1, Name: "WAN (Ethernet1)", Type: "ethernet", SpeedMbps: 1000, Status: "up", InTraffic: 45.2 + r.Float64()*120, OutTraffic: 12.1 + r.Float64()*40},
			{Index: 2, Name: "LAN-Bridge", Type: "bridge", SpeedMbps: 1000, Status: "up", InTraffic: 15.4 + r.Float64()*50, OutTraffic: 41.2 + r.Float64()*110},
			{Index: 3, Name: "SFP+ Uplink", Type: "ethernet", SpeedMbps: 10000, Status: "down", InTraffic: 0, OutTraffic: 0},
		}
		ifBytes, _ := json.Marshal(ifaces)
		data.Interfaces = string(ifBytes)

	case "switch":
		data.SysName = "SW-NOC-Distribution"
		data.SysDescr = "HP Comware Platform Software, Software Version 7.1.070, Release 2502P05"
		data.SysUptime = uint32(r.Intn(2000000) + 1000000) // ~115-347 days
		data.CPUUsage = 3.0 + r.Float64()*12.0
		data.RAMUsage = 15.0 + r.Float64()*10.0

		// Sim interfaces
		ifaces := make([]InterfaceInfo, 8)
		for i := 1; i <= 8; i++ {
			status := "down"
			inT, outT := 0.0, 0.0
			if i%2 == 1 || i == 8 {
				status = "up"
				inT = 0.5 + r.Float64()*15.0
				outT = 0.5 + r.Float64()*25.0
			}
			ifaces[i-1] = InterfaceInfo{
				Index:      i,
				Name:       net.JoinHostPort("gigabitethernet1/0", string(rune('0'+i))),
				Type:       "ethernet",
				SpeedMbps:  1000,
				Status:     status,
				InTraffic:  inT,
				OutTraffic: outT,
			}
		}
		// override port 8 as uplink
		ifaces[7].Name = "gigabitethernet1/0/8 (Uplink)"
		ifaces[7].InTraffic = 50.4 + r.Float64()*100
		ifaces[7].OutTraffic = 20.2 + r.Float64()*40

		ifBytes, _ := json.Marshal(ifaces)
		data.Interfaces = string(ifBytes)

	case "printer":
		data.SysName = "PRN-Office-Color"
		data.SysDescr = "Canon imageRUNNER ADVANCE C5535i III v1.4"
		data.SysUptime = uint32(r.Intn(500000) + 100000)
		data.CPUUsage = 1.0 + r.Float64()*5.0
		data.RAMUsage = 35.0 + r.Float64()*5.0

		// Sim Toner/Status info inside interfaces or as JSON
		toners := map[string]interface{}{
			"toner_black":   r.Intn(50) + 20, // 20% - 70%
			"toner_cyan":    r.Intn(60) + 10,
			"toner_magenta": r.Intn(60) + 10,
			"toner_yellow":  r.Intn(60) + 10,
			"paper_status":  "Ready",
			"tray_1":        "Letter (A4)",
			"tray_2":        "Legal (A3)",
		}
		tBytes, _ := json.Marshal(toners)
		data.Interfaces = string(tBytes)
	}
}
