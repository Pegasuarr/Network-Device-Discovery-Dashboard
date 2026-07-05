package portscan

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// DefaultPorts contains the standard list of ports used for discovery scanning.
var DefaultPorts = []int{
	21,   // FTP
	22,   // SSH
	23,   // Telnet
	25,   // SMTP
	53,   // DNS
	80,   // HTTP
	135,  // MSRPC
	137,  // NetBIOS
	139,  // NetBIOS Session
	443,  // HTTPS
	445,  // SMB
	1433, // MSSQL
	3306, // MySQL
	3389, // RDP
	5432, // PostgreSQL
	8080, // HTTP Alternate
}

// ServiceNames maps standard ports to their common service names.
var ServiceNames = map[int]string{
	21:   "FTP",
	22:   "SSH",
	23:   "Telnet",
	25:   "SMTP",
	53:   "DNS",
	80:   "HTTP",
	135:  "MSRPC",
	137:  "NetBIOS",
	139:  "NetBIOS",
	443:  "HTTPS",
	445:  "SMB",
	1433: "MSSQL",
	3306: "MySQL",
	3389: "RDP",
	5432: "PostgreSQL",
	8080: "HTTP-ALT",
}

// Scan performs a concurrent scan of the specified ports on a target IP address.
func Scan(ctx context.Context, ip string, ports []int, timeout time.Duration) []int {
	if len(ports) == 0 {
		ports = DefaultPorts
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var openPorts []int

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			// Check context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			address := net.JoinHostPort(ip, fmt.Sprintf("%d", p))
			dialer := &net.Dialer{Timeout: timeout}
			conn, err := dialer.DialContext(ctx, "tcp", address)
			if err == nil {
				conn.Close()
				mu.Lock()
				openPorts = append(openPorts, p)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return openPorts
}
