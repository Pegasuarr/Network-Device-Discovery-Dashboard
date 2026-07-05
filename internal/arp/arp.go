package arp

import (
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Lookup retrieves the MAC address associated with the given IP from the system's ARP cache.
func Lookup(ctx context.Context, ip string) (string, bool) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "arp", "-a", ip)
	} else {
		cmd = exec.CommandContext(ctx, "arp", "-n", ip)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", false
	}

	output := string(out)
	mac := parseMACAddress(output)
	if mac == "" {
		// Fallback without ip specifier
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "arp", "-a")
		} else {
			cmd = exec.CommandContext(ctx, "arp", "-n")
		}
		out, err = cmd.CombinedOutput()
		if err == nil {
			mac = parseMACFromFullTable(string(out), ip)
		}
	}

	if mac != "" {
		// Standardize MAC address to AA:BB:CC:DD:EE:FF format (uppercase, colon separated)
		mac = strings.ReplaceAll(mac, "-", ":")
		mac = strings.ToUpper(mac)
		return mac, true
	}

	return "", false
}

// Regex to match MAC address formats: e.g. 00-11-22-33-44-55 or 00:11:22:33:44:55
var macRegex = regexp.MustCompile(`([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`)

func parseMACAddress(output string) string {
	matches := macRegex.FindString(output)
	return matches
}

func parseMACFromFullTable(output string, ip string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ip) {
			matches := macRegex.FindString(line)
			if matches != "" {
				return matches
			}
		}
	}
	return ""
}
