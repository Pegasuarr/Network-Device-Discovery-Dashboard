package vendor

import (
	"strings"
)

// OUIDatabase contains a list of common MAC OUI prefixes mapped to hardware vendors.
var OUIDatabase = map[string]string{
	// VMware (Virtual Machines)
	"00:05:69": "VMware",
	"00:0C:29": "VMware",
	"00:1C:14": "VMware",
	"00:50:56": "VMware",

	// Cisco Systems
	"00:00:0C": "Cisco",
	"00:11:92": "Cisco",
	"00:1D:70": "Cisco",
	"00:1E:4A": "Cisco",
	"00:21:A0": "Cisco",
	"00:24:97": "Cisco",
	"00:26:0B": "Cisco",
	"00:3A:9A": "Cisco",
	"00:50:73": "Cisco",
	"00:90:F2": "Cisco",
	"3C:52:82": "Cisco",
	"74:A2:E6": "Cisco",

	// Apple
	"00:03:93": "Apple",
	"00:17:F2": "Apple",
	"00:1C:42": "Apple",
	"00:1D:4F": "Apple",
	"00:1E:52": "Apple",
	"00:1F:5B": "Apple",
	"00:23:32": "Apple",
	"00:24:36": "Apple",
	"00:25:00": "Apple",
	"00:25:4B": "Apple",
	"00:26:08": "Apple",
	"00:26:BB": "Apple",
	"04:0C:CE": "Apple",
	"04:15:52": "Apple",
	"04:26:34": "Apple",
	"04:52:F3": "Apple",
	"BC:30:5B": "Apple",

	// Microsoft
	"00:03:FF": "Microsoft",
	"00:12:5A": "Microsoft",
	"00:15:5D": "Microsoft (Hyper-V)",
	"00:1D:D8": "Microsoft",
	"00:22:48": "Microsoft",
	"00:25:22": "Microsoft",
	"00:50:F2": "Microsoft",
	"28:18:78": "Microsoft",

	// Dell
	"00:06:5B": "Dell",
	"00:08:74": "Dell",
	"00:0A:F7": "Dell",
	"00:0B:DB": "Dell",
	"00:0F:1F": "Dell",
	"00:13:72": "Dell",
	"00:14:22": "Dell",
	"00:15:C5": "Dell",
	"00:16:EC": "Dell",
	"00:18:8B": "Dell",
	"00:19:B9": "Dell",
	"00:1A:80": "Dell",
	"00:1C:23": "Dell",
	"00:1D:09": "Dell",
	"00:1E:4F": "Dell",
	"00:21:9B": "Dell",
	"00:22:19": "Dell",
	"00:23:AE": "Dell",

	// Hewlett Packard (HP)
	"00:08:02": "HP",
	"00:0B:CD": "HP",
	"00:0C:00": "HP",
	"00:0D:9D": "HP",
	"00:0E:7F": "HP",
	"00:0F:20": "HP",
	"00:10:83": "HP",
	"00:11:0A": "HP",
	"00:11:85": "HP",
	"00:12:79": "HP",
	"00:13:21": "HP",
	"00:14:38": "HP",
	"00:15:60": "HP",
	"00:16:35": "HP",
	"00:17:08": "HP",
	"00:18:71": "HP",
	"00:19:BB": "HP",
	"00:1A:4B": "HP",
	"00:1B:78": "HP",
	"00:1C:25": "HP",

	// Intel
	"00:03:47": "Intel",
	"00:04:23": "Intel",
	"00:08:A1": "Intel",
	"00:0C:43": "Intel",
	"00:0E:0C": "Intel",
	"00:0E:35": "Intel",
	"00:13:02": "Intel",
	"00:13:E8": "Intel",
	"00:15:00": "Intel",
	"00:16:EA": "Intel",
	"00:18:DE": "Intel",
	"00:19:D1": "Intel",
	"00:1A:11": "Intel",
	"00:1B:21": "Intel",
	"00:1C:BF": "Intel",
	"00:1D:0F": "Intel",
	"00:1E:64": "Intel",
	"00:1F:3B": "Intel",
	"00:21:5A": "Intel",
	"00:21:6A": "Intel",
	"00:22:FA": "Intel",

	// Synology
	"00:11:32": "Synology",

	// Ubiquiti Networks
	"00:15:6D": "Ubiquiti",
	"00:27:22": "Ubiquiti",
	"04:18:D6": "Ubiquiti",
	"24:A4:3C": "Ubiquiti",
	"78:8A:20": "Ubiquiti",
	"80:2A:A8": "Ubiquiti",
	"B4:FB:E4": "Ubiquiti",
	"F0:9F:C2": "Ubiquiti",

	// Netgear
	"00:0F:B5": "Netgear",
	"00:14:6C": "Netgear",
	"00:18:4D": "Netgear",
	"00:1B:2F": "Netgear",
	"00:1F:33": "Netgear",
	"00:22:3F": "Netgear",
	"00:24:B2": "Netgear",
	"00:26:F2": "Netgear",
	"20:E5:2A": "Netgear",
	"28:80:88": "Netgear",

	// TP-Link
	"00:14:78": "TP-Link",
	"00:21:29": "TP-Link",
	"00:23:CD": "TP-Link",
	"00:25:86": "TP-Link",
	"00:27:19": "TP-Link",
	"10:7B:EF": "TP-Link",
	"14:CC:20": "TP-Link",
	"18:A6:C7": "TP-Link",
	"18:D6:C7": "TP-Link",
	"3C:84:6F": "TP-Link",
	"50:C7:BF": "TP-Link",
	"54:E6:FC": "TP-Link",
	"70:4F:57": "TP-Link",
	"74:DA:38": "TP-Link",

	// Samsung
	"00:00:F0": "Samsung",
	"00:02:78": "Samsung",
	"00:07:AB": "Samsung",
	"00:0D:E6": "Samsung",
	"00:12:47": "Samsung",
	"00:12:FC": "Samsung",
	"00:15:99": "Samsung",
	"00:16:6C": "Samsung",
	"00:17:C8": "Samsung",
	"00:18:AF": "Samsung",
	"00:1A:8A": "Samsung",
	"00:1B:98": "Samsung",
	"00:1D:25": "Samsung",
	"00:1E:7D": "Samsung",
	"00:1E:C2": "Samsung",
	"00:1F:DC": "Samsung",
	"00:21:1E": "Samsung",
	"00:21:D2": "Samsung",

	// Linksys
	"00:04:5A": "Linksys",
	"00:06:25": "Linksys",
	"00:0C:41": "Linksys",
	"00:0E:E8": "Linksys",
	"00:12:17": "Linksys",
	"00:13:10": "Linksys",
	"00:14:BF": "Linksys",
	"00:18:39": "Linksys",
	"00:1D:7E": "Linksys",
	"00:22:6B": "Linksys",
	"00:23:69": "Linksys",
	"00:25:9C": "Linksys",
	"00:26:5B": "Linksys",

	// Huawei
	"00:0B:C1": "Huawei",
	"00:16:3E": "Huawei",
	"00:18:82": "Huawei",
	"00:1E:10": "Huawei",
	"00:22:A1": "Huawei",
	"00:25:68": "Huawei",
	"00:25:9E": "Huawei",
	"08:19:A6": "Huawei",
	"0C:37:DC": "Huawei",
	"10:1B:54": "Huawei",
	"10:47:80": "Huawei",
	"10:C6:1F": "Huawei",
	"14:3E:BF": "Huawei",
	"18:84:B1": "Huawei",

	// Realtek (Network adapters)
	"00:E0:4C": "Realtek",
	"00:13:3B": "Realtek",
	"52:54:00": "QEMU Virtual Nic (Realtek)",
}

// Lookup queries the database for the vendor of the given MAC address.
func Lookup(mac string) string {
	if len(mac) < 8 {
		return "Unknown Vendor"
	}
	
	// Normalize to uppercase and extract OUI prefix (first 8 characters: "XX:XX:XX")
	macClean := strings.ToUpper(strings.TrimSpace(mac))
	macClean = strings.ReplaceAll(macClean, "-", ":")
	
	if len(macClean) >= 8 {
		oui := macClean[:8]
		if vendor, found := OUIDatabase[oui]; found {
			return vendor
		}
	}
	
	return "Unknown Vendor"
}
