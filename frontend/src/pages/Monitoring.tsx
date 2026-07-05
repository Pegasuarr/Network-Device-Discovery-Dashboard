import React, { useState } from "react";
import api from "../services/api";
import { Play, Plus, Loader2, Network, ShieldCheck, Terminal } from "lucide-react";

interface DiscoveredHost {
  ip_address: string;
  hostname: string;
  vendor: string;
  os: string;
  device_type: string;
  open_ports: number[];
  ping_time_ms: number;
}

export const Monitoring: React.FC = () => {
  const [cidr, setCidr] = useState("127.0.0.1/32");
  const [scanning, setScanning] = useState(false);
  const [results, setResults] = useState<DiscoveredHost[]>([]);
  const [addedIps, setAddedIps] = useState<Record<string, boolean>>({});

  const handleScan = async (e: React.FormEvent) => {
    e.preventDefault();
    setScanning(true);
    setResults([]);

    try {
      const res = await api.post("/discovery/scan", { cidr });
      setResults(res.data || []);
    } catch (err) {
      console.error(err);
      alert("Scan failed. Verify subnet range format (e.g. 192.168.1.0/24)");
    } finally {
      setScanning(false);
    }
  };

  const handleAutoAdd = async (host: DiscoveredHost) => {
    try {
      const payload = {
        name: host.hostname || `Host ${host.ip_address}`,
        hostname: host.ip_address,
        ip_address: host.ip_address,
        device_type: host.device_type,
        os: host.os,
        vendor: host.vendor,
        monitoring_interval: 60,
        enabled: true,
      };

      await api.post("/devices", payload);
      setAddedIps((prev) => ({ ...prev, [host.ip_address]: true }));
    } catch (err) {
      console.error("Auto add failed", err);
      alert("Failed to enroll device.");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
          Network Discovery Scanner
        </h1>
        <span className="text-xs text-slate-500 font-semibold bg-slate-100 dark:bg-slate-800 px-3 py-1 border border-slate-200 dark:border-darkBorder rounded-lg">
          Active Sweeper: IP/TCP Sweep
        </span>
      </div>

      {/* Discovery Input Form */}
      <div className="p-6 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm space-y-4">
        <div>
          <h3 className="text-sm font-bold text-slate-700 dark:text-slate-200">CIDR Subnet Scanning</h3>
          <p className="text-xs text-slate-400 mt-1">
            Specify an IP address block in CIDR notation. The scan sweeps the range to identify active listeners and fingerprints device signatures.
          </p>
        </div>

        <form onSubmit={handleScan} className="flex flex-col sm:flex-row items-end sm:items-center space-y-3 sm:space-y-0 sm:space-x-4">
          <div className="flex-1 w-full">
            <label className="block text-[10px] uppercase font-bold text-slate-400 mb-1.5">Target CIDR Block</label>
            <div className="relative">
              <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-slate-500">
                <Network className="h-5 w-5" />
              </span>
              <input
                type="text"
                required
                value={cidr}
                onChange={(e) => setCidr(e.target.value)}
                placeholder="e.g. 192.168.1.0/24"
                className="w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 placeholder-slate-400 focus:outline-none"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={scanning}
            className="w-full sm:w-auto px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-bold shadow-md flex items-center justify-center disabled:opacity-50"
          >
            {scanning ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Scanning Subnet...
              </>
            ) : (
              <>
                <Play className="h-4 w-4 mr-2" />
                Run Discovery Scan
              </>
            )}
          </button>
        </form>
      </div>

      {/* Discovered results table */}
      {scanning && (
        <div className="flex flex-col items-center justify-center p-12 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <Terminal className="h-10 w-10 text-indigo-500 animate-bounce mb-3" />
          <h4 className="text-sm font-bold text-slate-700 dark:text-slate-300">Sweep active. Port knocking IP ranges...</h4>
          <p className="text-xs text-slate-500 mt-1">This may take up to 15 seconds depending on subnet sizing.</p>
        </div>
      )}

      {!scanning && results.length > 0 && (
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm overflow-hidden">
          <div className="p-4 bg-slate-55 dark:bg-slate-900/30 border-b border-slate-200 dark:border-darkBorder">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">Scan Results ({results.length} Hosts Active)</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-slate-50/50 dark:bg-slate-900/10 text-slate-400 text-xs font-bold uppercase border-b border-slate-200 dark:border-darkBorder">
                  <th className="px-6 py-4">IP Address</th>
                  <th className="px-6 py-4">Resolved Host</th>
                  <th className="px-6 py-4">Signature / OS</th>
                  <th className="px-6 py-4">Open Ports</th>
                  <th className="px-6 py-4">RTT</th>
                  <th className="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-darkBorder text-slate-700 dark:text-slate-300 text-sm">
                {results.map((r, idx) => (
                  <tr key={idx} className="hover:bg-slate-50/20 dark:hover:bg-slate-800/10">
                    <td className="px-6 py-4 font-mono font-bold text-xs text-indigo-500">{r.ip_address}</td>
                    <td className="px-6 py-4 font-semibold text-slate-850 dark:text-slate-200">{r.hostname}</td>
                    <td className="px-6 py-4">
                      <span className="block text-xs font-semibold">{r.vendor}</span>
                      <span className="block text-[10px] text-slate-400 mt-0.5">{r.os} ({r.device_type})</span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {r.open_ports.map((port) => (
                          <span key={port} className="px-1.5 py-0.5 bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-darkBorder rounded text-[10px] font-mono">
                            {port}
                          </span>
                        ))}
                        {r.open_ports.length === 0 && <span className="text-xs text-slate-500 font-semibold">None detected</span>}
                      </div>
                    </td>
                    <td className="px-6 py-4 font-mono text-xs">{r.ping_time_ms} ms</td>
                    <td className="px-6 py-4 text-right">
                      {addedIps[r.ip_address] ? (
                        <span className="inline-flex items-center text-xs font-bold text-emerald-500 pr-4">
                          <ShieldCheck className="h-4 w-4 mr-1" /> Monitored
                        </span>
                      ) : (
                        <button
                          onClick={() => handleAutoAdd(r)}
                          className="px-3 py-1.5 bg-slate-100 hover:bg-indigo-500 hover:text-white dark:bg-slate-800 text-slate-700 dark:text-slate-350 dark:hover:text-white font-bold rounded-lg text-xs border border-slate-200 dark:border-darkBorder transition-all"
                        >
                          <Plus className="h-3.5 w-3.5 mr-1 inline" /> Enroll
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};
