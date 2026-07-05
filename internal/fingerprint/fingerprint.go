package fingerprint

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Signature holds the fingerprinted OS details and device classification.
type Signature struct {
	OS         string `json:"os"`
	DeviceType string `json:"device_type"`
	Vendor     string `json:"vendor"`
}

// Fingerprint analyses open ports and performs banner grabs to identify the device.
func Fingerprint(ctx context.Context, ip string, openPorts []int) Signature {
	sig := Signature{
		OS:         "Unknown OS",
		DeviceType: "workstation", // default
		Vendor:     "Unknown Vendor",
	}

	hasSSH := false
	hasHTTP := false
	hasHTTPS := false
	hasRDP := false
	hasSMB := false
	hasTelnet := false
	hasPrinter := false // 9100, 515

	for _, p := range openPorts {
		switch p {
		case 22:
			hasSSH = true
		case 23:
			hasTelnet = true
		case 80:
			hasHTTP = true
		case 443:
			hasHTTPS = true
		case 3389:
			hasRDP = true
		case 135, 139, 445:
			hasSMB = true
		case 515, 9100:
			hasPrinter = true
		}
	}

	// 1. Check Printer port
	if hasPrinter {
		sig.DeviceType = "printer"
		sig.OS = "Embedded RTOS"
		sig.Vendor = "Canon/HP"
		return sig
	}

	// 2. Perform banner grabbing for SSH if open
	if hasSSH {
		sshBanner := grabSSHBanner(ctx, ip, 22)
		if sshBanner != "" {
			if strings.Contains(sshBanner, "Ubuntu") {
				sig.OS = "Ubuntu Linux"
				sig.DeviceType = "server"
				sig.Vendor = "Canonical"
				return sig
			}
			if strings.Contains(sshBanner, "Debian") {
				sig.OS = "Debian Linux"
				sig.DeviceType = "server"
				sig.Vendor = "Debian"
				return sig
			}
			if strings.Contains(sshBanner, "CentOS") || strings.Contains(sshBanner, "RedHat") {
				sig.OS = "Red Hat Enterprise Linux"
				sig.DeviceType = "server"
				sig.Vendor = "Red Hat"
				return sig
			}
			if strings.Contains(sshBanner, "MikroTik") || strings.Contains(sshBanner, "RouterOS") {
				sig.OS = "MikroTik RouterOS"
				sig.DeviceType = "router"
				sig.Vendor = "MikroTik"
				return sig
			}
			if strings.Contains(sshBanner, "pfSense") || strings.Contains(sshBanner, "FreeBSD") {
				sig.OS = "pfSense (FreeBSD)"
				sig.DeviceType = "router"
				sig.Vendor = "Netgate"
				return sig
			}
			if strings.Contains(sshBanner, "Cisco") {
				sig.OS = "Cisco IOS"
				sig.DeviceType = "switch"
				sig.Vendor = "Cisco"
				return sig
			}
		}
	}

	// 3. Perform HTTP server banner grabbing if open
	if hasHTTP || hasHTTPS {
		httpBanner := grabHTTPBanner(ctx, ip, hasHTTPS)
		if httpBanner != "" {
			if strings.Contains(httpBanner, "IIS") {
				sig.OS = "Windows Server 2022"
				sig.DeviceType = "server"
				sig.Vendor = "Microsoft"
				return sig
			}
			if strings.Contains(httpBanner, "pfSense") {
				sig.OS = "pfSense"
				sig.DeviceType = "router"
				sig.Vendor = "Netgate"
				return sig
			}
			if strings.Contains(httpBanner, "Synology") || strings.Contains(httpBanner, "DSM") {
				sig.OS = "Synology DSM"
				sig.DeviceType = "server"
				sig.Vendor = "Synology"
				return sig
			}
		}
	}

	// 4. Fallback checking based on ports
	if hasSMB || hasRDP {
		sig.OS = "Windows 11"
		sig.DeviceType = "workstation"
		sig.Vendor = "Microsoft"
		if hasSMB && !hasRDP && !hasHTTP {
			// Could be an AD domain controller or server
			sig.OS = "Windows Server"
			sig.DeviceType = "server"
		}
		return sig
	}

	if hasSSH {
		sig.OS = "Linux"
		sig.DeviceType = "server"
		sig.Vendor = "Generic Linux"
		return sig
	}

	if hasTelnet {
		sig.OS = "Embedded OS"
		sig.DeviceType = "switch"
		sig.Vendor = "Cisco/Aruba"
		return sig
	}

	if hasHTTP || hasHTTPS {
		// Web portal only
		sig.OS = "Embedded firmware"
		sig.DeviceType = "iot"
		sig.Vendor = "Generic IoT"
	}

	return sig
}

func grabSSHBanner(ctx context.Context, ip string, port int) string {
	dialer := net.Dialer{Timeout: 400 * time.Millisecond}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return ""
	}
	defer conn.Close()

	// Read SSH banner line
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	reader := bufio.NewReader(conn)
	banner, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return banner
}

func grabHTTPBanner(ctx context.Context, ip string, isHTTPS bool) string {
	schema := "http"
	port := 80
	if isHTTPS {
		schema = "https"
		port = 443
	}

	url := fmt.Sprintf("%s://%s:%d", schema, ip, port)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   400 * time.Millisecond,
		Transport: tr,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		// Try Server Header check in error if response is semi-sent
		return ""
	}
	defer resp.Body.Close()

	serverHeader := resp.Header.Get("Server")
	return serverHeader
}
