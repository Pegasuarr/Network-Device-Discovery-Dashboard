package monitor

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/user/network-monitoring/internal/model"
)

type Engine struct {
	httpClient *http.Client
}

func NewEngine() *Engine {
	return &Engine{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (e *Engine) RunAllChecks(device *model.Device) *model.MonitoringResult {
	res := &model.MonitoringResult{
		DeviceID:    device.ID,
		CheckedAt:   time.Now(),
		DNSResolved: true,
	}

	// 1. DNS check (Lookup IP if hostname is set)
	var resolvedIP string
	if device.Hostname != "" {
		ips, err := net.LookupIP(device.Hostname)
		if err != nil || len(ips) == 0 {
			res.DNSResolved = false
			resolvedIP = device.IPAddress
		} else {
			resolvedIP = ips[0].String()
		}
	} else {
		resolvedIP = device.IPAddress
	}

	// 2. Ping Check
	latency, loss := e.Ping(resolvedIP)
	res.LatencyMS = latency
	res.PacketLossPct = loss

	// 3. TCP Check (if device is a switch, router, or has port checking, check port 80/443/22)
	tcpPort := "80"
	if device.DeviceType == "router" || device.DeviceType == "switch" {
		tcpPort = "23" // Telnet / SSH port equivalent
	}
	startTCP := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(resolvedIP, tcpPort), 2*time.Second)
	if err == nil {
		conn.Close()
		res.ResponseTimeMS = float64(time.Since(startTCP).Milliseconds())
	} else {
		res.ResponseTimeMS = 0
	}

	// 4. HTTP Check (if server or specifically requested)
	if device.DeviceType == "server" || strings.HasPrefix(device.Hostname, "http") {
		url := fmt.Sprintf("http://%s", device.Hostname)
		if strings.HasPrefix(device.Hostname, "http") {
			url = device.Hostname
		}

		startHTTP := time.Now()
		resp, err := e.httpClient.Get(url)
		if err == nil {
			res.HTTPStatus = resp.StatusCode
			res.ResponseTimeMS = float64(time.Since(startHTTP).Milliseconds())
			resp.Body.Close()
		} else {
			res.HTTPStatus = 0
		}
	}

	// 5. SSL Check (if HTTPS)
	if strings.Contains(device.Hostname, "https") || strings.Contains(device.Hostname, "443") {
		hostOnly := device.Hostname
		if strings.HasPrefix(hostOnly, "https://") {
			hostOnly = strings.TrimPrefix(hostOnly, "https://")
		}
		if !strings.Contains(hostOnly, ":") {
			hostOnly = hostOnly + ":443"
		}

		days := e.CheckSSLExpiration(hostOnly)
		res.SSLDaysRemaining = days
	}

	// 6. Simulate Server System Telemetry
	// For demo and testing, servers report live CPU, RAM, Disk utilization
	if device.DeviceType == "server" {
		res.CPUUsage = 15.0 + rand.Float64()*45.0 // 15% - 60%
		res.RAMUsage = 40.0 + rand.Float64()*30.0 // 40% - 70%
		res.DiskUsage = 52.3                      // static average
	} else {
		res.CPUUsage = 5.0 + rand.Float64()*10.0
		res.RAMUsage = 12.0 + rand.Float64()*5.0
		res.DiskUsage = 0
	}

	return res
}

func (e *Engine) Ping(ip string) (float64, float64) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", ip)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 100.0 // 100% packet loss
	}

	output := string(out)
	latency := e.parsePingLatency(output)

	return latency, 0.0 // 0% packet loss, successfully pinged
}

func (e *Engine) parsePingLatency(output string) float64 {
	// Parse latency from ping output
	// Windows format: Minimum = 1ms, Maximum = 1ms, Average = 1ms or time=12ms
	// Linux format: rtt min/avg/max/mdev = 0.052/0.052/0.052/0.000 ms or time=12.3 ms
	reTime := regexp.MustCompile(`time[<=](\d+(?:\.\d+)?)`)
	matches := reTime.FindStringSubmatch(output)
	if len(matches) > 1 {
		l, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return l
		}
	}

	// Try average fallback
	reAvg := regexp.MustCompile(`Average = (\d+)ms`)
	matchesAvg := reAvg.FindStringSubmatch(output)
	if len(matchesAvg) > 1 {
		l, err := strconv.ParseFloat(matchesAvg[1], 64)
		if err == nil {
			return l
		}
	}

	return 1.0 // default minimum fallback
}

func (e *Engine) CheckSSLExpiration(host string) int {
	dialer := &net.Dialer{
		Timeout: 3 * time.Second,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", host, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return 0
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return 0
	}

	expiry := certs[0].NotAfter
	days := int(time.Until(expiry).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
