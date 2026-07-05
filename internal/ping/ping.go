package ping

import (
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Ping executes an unprivileged ping command using the host OS utility.
func Ping(ctx context.Context, ip string, timeout time.Duration) (float64, bool) {
	var cmd *exec.Cmd
	timeoutSec := int(timeout.Seconds())
	if timeoutSec <= 0 {
		timeoutSec = 1
	}

	if runtime.GOOS == "windows" {
		// -n 1: 1 packet, -w 1000: timeout in milliseconds
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", strconv.Itoa(timeoutSec*1000), ip)
	} else {
		// -c 1: 1 packet, -W 1: timeout in seconds
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", strconv.Itoa(timeoutSec), ip)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, false
	}

	output := string(out)
	latency := parsePingLatency(output)
	return latency, true
}

func parsePingLatency(output string) float64 {
	// Look for time=XXms or time=XX.X ms
	reTime := regexp.MustCompile(`time[<=](\d+(?:\.\d+)?)`)
	matches := reTime.FindStringSubmatch(output)
	if len(matches) > 1 {
		l, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return l
		}
	}

	// Windows fallback: Average = XXms
	reAvg := regexp.MustCompile(`Average = (\d+)ms`)
	matchesAvg := reAvg.FindStringSubmatch(output)
	if len(matchesAvg) > 1 {
		l, err := strconv.ParseFloat(matchesAvg[1], 64)
		if err == nil {
			return l
		}
	}

	// Linux fallback: rtt min/avg/max/mdev
	if strings.Contains(output, "rtt") || strings.Contains(output, "min/avg/max") {
		parts := strings.Split(output, "\n")
		for _, part := range parts {
			if strings.HasPrefix(part, "rtt") || strings.Contains(part, "min/avg/max") {
				fields := strings.Fields(part)
				if len(fields) > 3 {
					stats := strings.Split(fields[3], "/")
					if len(stats) > 1 {
						l, err := strconv.ParseFloat(stats[1], 64) // average is index 1
						if err == nil {
							return l
						}
					}
				}
			}
		}
	}

	return 1.0 // minimal fallback
}
