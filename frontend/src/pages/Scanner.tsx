import React, { useState, useEffect } from "react";
import api from "../services/api";
import { useSocket } from "../context/SocketContext";
import type { Device } from "../types";
import {
  Play,
  Square,
  Activity,
  Server,
  Network,
  Cpu,
  CheckCircle,
  Clock,
  Terminal,
} from "lucide-react";

interface DiscoveredHost {
  ip_address: string;
  hostname: string;
  mac_address: string;
  mac_vendor: string;
  vendor: string;
  os: string;
  device_type: string;
  open_ports: number[];
  ping_time_ms: number;
  status: string;
  snmp_enabled: boolean;
}

export const Scanner: React.FC = () => {
  const { lastMessage } = useSocket();
  const [target, setTarget] = useState("192.168.1.0/24");
  const [profile, setProfile] = useState("quick");
  const [isScanning, setIsScanning] = useState(false);
  const [scanID, setScanID] = useState<string | null>(null);
  
  // Progress states
  const [scannedCount, setScannedCount] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const [percent, setPercent] = useState(0);
  const [devicesFound, setDevicesFound] = useState(0);
  const [currentIP, setCurrentIP] = useState("");
  const [consoleLogs, setConsoleLogs] = useState<string[]>([]);
  const [discoveredDevices, setDiscoveredDevices] = useState<DiscoveredHost[]>([]);

  // Listen to WebSocket triggers
  useEffect(() => {
    if (!lastMessage) return;

    const { type, payload } = lastMessage;

    if (type === "scan_progress") {
      const {
        scan_id,
        total_ips,
        scanned_ips,
        percent: pct,
        devices_found,
        current_ip,
        status,
        latest_device,
      } = payload;

      // Ensure it matches current scan
      if (scanID && scan_id === scanID) {
        setScannedCount(scanned_ips);
        setTotalCount(total_ips);
        setPercent(pct);
        setDevicesFound(devices_found);
        setCurrentIP(current_ip);

        // Add to log lines
        const logLine = `[INFO] Probed ${current_ip} - Status: ${latest_device ? "ALIVE (" + latest_device.ping_time_ms + "ms)" : "DEAD"}`;
        setConsoleLogs((prev) => [logLine, ...prev].slice(0, 100));

        if (latest_device) {
          // Check if already in the list to avoid duplicate listings in table
          setDiscoveredDevices((prev) => {
            const exists = prev.some((d) => d.ip_address === latest_device.ip_address);
            if (!exists) {
              return [latest_device, ...prev];
            }
            return prev;
          });
        }

        if (status === "completed" || status === "cancelled" || status === "failed") {
          setIsScanning(false);
          const finalLog = `[SYSTEM] Scan finished with status: ${status.toUpperCase()}. Duration: ${total_ips} IPs checked, ${devices_found} alive hosts discovered.`;
          setConsoleLogs((prev) => [finalLog, ...prev]);
        }
      }
    }
  }, [lastMessage, scanID]);

  const handleStartScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (isScanning) return;

    setIsScanning(true);
    setScannedCount(0);
    setTotalCount(0);
    setPercent(0);
    setDevicesFound(0);
    setCurrentIP("");
    setDiscoveredDevices([]);
    setConsoleLogs([`[SYSTEM] Connecting to scanner, initializing CIDR scan on ${target}...`]);

    try {
      const res = await api.post("/discovery/scan", {
        target: target,
        profile: profile,
      });

      setScanID(res.data.scan_id);
      setConsoleLogs((prev) => [`[SYSTEM] Scan started successfully. Task UUID: ${res.data.scan_id}`, ...prev]);
    } catch (err: any) {
      console.error(err);
      setIsScanning(false);
      const errMsg = err.response?.data?.error || "Connection timeout";
      setConsoleLogs((prev) => [`[ERROR] Failed to start scan: ${errMsg}`, ...prev]);
    }
  };

  const handleCancelScan = async () => {
    if (!scanID) return;
    setConsoleLogs((prev) => [`[SYSTEM] Sending cancellation request for Scan ID ${scanID}...`, ...prev]);
    try {
      await api.post("/discovery/scan/cancel", {
        scan_id: scanID,
      });
    } catch (err) {
      console.error(err);
      setConsoleLogs((prev) => [`[ERROR] Cancel request failed.`, ...prev]);
    }
  };

  return (
    <div className="space-y-6 pb-12">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
          Network Discovery Scanner
        </h1>
        <p className="text-sm text-slate-500 mt-1">
          Perform unprivileged multi-profile subnet scans to identify connected hosts.
        </p>
      </div>

      {/* Control Panel */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
        <form onSubmit={handleStartScan} className="flex flex-col md:flex-row md:items-end gap-4">
          <div className="flex-1">
            <label className="block text-xs font-bold text-slate-400 uppercase tracking-wider mb-2">Scan Target Range</label>
            <input
              type="text"
              value={target}
              required
              disabled={isScanning}
              onChange={(e) => setTarget(e.target.value)}
              placeholder="e.g. 192.168.1.0/24 or 192.168.1.10"
              className="w-full px-4 py-2.5 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
            />
          </div>

          <div className="w-full md:w-48">
            <label className="block text-xs font-bold text-slate-400 uppercase tracking-wider mb-2">Scan Profile</label>
            <select
              value={profile}
              disabled={isScanning}
              onChange={(e) => setProfile(e.target.value)}
              className="w-full px-4 py-2.5 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 focus:outline-none text-sm"
            >
              <option value="quick">Quick Scan (Ping + Ports)</option>
              <option value="deep">Deep Scan (Port Scan + OS + SNMP)</option>
              <option value="portscan">Port Scan Only</option>
              <option value="ping">Ping Only</option>
            </select>
          </div>

          <div className="flex space-x-3">
            {!isScanning ? (
              <button
                type="submit"
                className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-bold shadow-md flex items-center transition-all"
              >
                <Play className="h-4 w-4 mr-2" />
                Start Scan
              </button>
            ) : (
              <button
                type="button"
                onClick={handleCancelScan}
                className="px-6 py-2.5 bg-red-600 hover:bg-red-500 text-white rounded-lg text-sm font-bold shadow-md flex items-center transition-all animate-pulse"
              >
                <Square className="h-4 w-4 mr-2" />
                Cancel Scan
              </button>
            )}
          </div>
        </form>
      </div>

      {/* Progress Cards */}
      {isScanning && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-5">
          <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
            <p className="text-xs text-slate-400 font-semibold uppercase">Scan Progress</p>
            <div className="flex items-center justify-between mt-2">
              <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{percent}%</h3>
              <span className="text-xs font-mono text-slate-500">{scannedCount}/{totalCount} IPs</span>
            </div>
            <div className="w-full bg-slate-100 dark:bg-slate-800 h-2 rounded-full mt-3 overflow-hidden">
              <div
                className="bg-indigo-600 h-full transition-all duration-300"
                style={{ width: `${percent}%` }}
              ></div>
            </div>
          </div>

          <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
            <p className="text-xs text-slate-400 font-semibold uppercase">Hosts Discovered</p>
            <h3 className="text-2xl font-bold text-green-500 mt-2 flex items-center">
              <Server className="h-6 w-6 mr-2" />
              {devicesFound}
            </h3>
          </div>

          <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
            <p className="text-xs text-slate-400 font-semibold uppercase">Current Target IP</p>
            <h3 className="text-xl font-mono font-bold text-slate-800 dark:text-slate-100 mt-2 truncate">
              {currentIP || "Waiting..."}
            </h3>
          </div>

          <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm flex items-center justify-center">
            <div className="flex items-center space-x-2 text-indigo-500">
              <Activity className="h-5 w-5 animate-spin" />
              <span className="text-xs font-bold uppercase tracking-wider animate-pulse">Running Scan...</span>
            </div>
          </div>
        </div>
      )}

      {/* Main Grid: Discovered Devices & Console logs */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Discovered Devices Table */}
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm lg:col-span-2 overflow-hidden flex flex-col">
          <h2 className="text-sm font-bold text-slate-800 dark:text-slate-100 mb-4 flex items-center">
            <Network className="h-4 w-4 mr-2 text-indigo-500" />
            Live Scanned Hosts ({discoveredDevices.length})
          </h2>
          <div className="overflow-x-auto flex-1 max-h-[450px]">
            {discoveredDevices.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20 text-slate-400 space-y-2">
                <Server className="h-8 w-8" />
                <p className="text-xs">No active devices discovered on this run.</p>
              </div>
            ) : (
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="bg-slate-50 dark:bg-slate-900/40 text-slate-400 text-[11px] font-bold uppercase tracking-wider border-b border-slate-200 dark:border-darkBorder">
                    <th className="px-4 py-3">IP Address</th>
                    <th className="px-4 py-3">Hostname</th>
                    <th className="px-4 py-3">MAC / Vendor</th>
                    <th className="px-4 py-3">OS</th>
                    <th className="px-4 py-3">Ping</th>
                    <th className="px-4 py-3">SNMP</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-150 dark:divide-darkBorder text-slate-700 dark:text-slate-300 text-xs">
                  {discoveredDevices.map((dev) => (
                    <tr key={dev.ip_address} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/10">
                      <td className="px-4 py-3 font-mono font-bold">{dev.ip_address}</td>
                      <td className="px-4 py-3 truncate max-w-[120px]">{dev.hostname}</td>
                      <td className="px-4 py-3">
                        <div className="font-mono text-[10px]">{dev.mac_address || "N/A"}</div>
                        <div className="text-[10px] text-slate-400">{dev.mac_vendor}</div>
                      </td>
                      <td className="px-4 py-3">{dev.os}</td>
                      <td className="px-4 py-3 font-mono text-green-500 font-bold">{dev.ping_time_ms.toFixed(1)}ms</td>
                      <td className="px-4 py-3">
                        {dev.snmp_enabled ? (
                          <span className="px-1.5 py-0.5 rounded bg-green-500/10 text-green-500 font-bold text-[10px] uppercase">
                            Active
                          </span>
                        ) : (
                          <span className="text-slate-400 text-[10px]">Disabled</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>

        {/* Real-time scanning Console logs */}
        <div className="bg-slate-900 border border-slate-850 rounded-xl p-5 shadow-sm text-slate-200 font-mono text-xs flex flex-col h-[525px]">
          <h2 className="text-xs font-bold text-slate-400 mb-3 uppercase tracking-wider flex items-center border-b border-slate-800 pb-2">
            <Terminal className="h-4 w-4 mr-2 text-indigo-400" />
            Scanner System logs
          </h2>
          <div className="flex-1 overflow-y-auto space-y-1.5 scrollbar-thin">
            {consoleLogs.length === 0 ? (
              <p className="text-slate-600 italic">Scanner idle. Start a scan to view diagnostic events.</p>
            ) : (
              consoleLogs.map((log, index) => {
                let color = "text-slate-300";
                if (log.includes("[ERROR]")) color = "text-red-400";
                if (log.includes("[SYSTEM]")) color = "text-indigo-400";
                if (log.includes("ALIVE")) color = "text-green-400";

                return (
                  <p key={index} className={`${color} leading-relaxed`}>
                    {log}
                  </p>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
