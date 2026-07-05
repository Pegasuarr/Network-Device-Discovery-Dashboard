import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { ScanHistory } from "../types";
import { History as HistoryIcon, Clock, Server, CheckCircle2, XCircle, Ban, Loader2 } from "lucide-react";

export const History: React.FC = () => {
  const [history, setHistory] = useState<ScanHistory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchHistory();
  }, []);

  const fetchHistory = async () => {
    try {
      const res = await api.get("/discovery/history");
      setHistory(res.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const classes = "inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold uppercase tracking-wider text-white";
    switch (status) {
      case "completed":
        return <span className={`${classes} bg-green-500`}><CheckCircle2 className="h-3 w-3 mr-1" /> Completed</span>;
      case "cancelled":
        return <span className={`${classes} bg-slate-400`}><Ban className="h-3 w-3 mr-1" /> Cancelled</span>;
      case "failed":
        return <span className={`${classes} bg-red-500`}><XCircle className="h-3 w-3 mr-1" /> Failed</span>;
      default:
        return <span className={`${classes} bg-yellow-500`}>Running</span>;
    }
  };

  const formatDuration = (ms: number) => {
    if (ms <= 0) return "N/A";
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  };

  return (
    <div className="space-y-6 pb-12">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100 flex items-center">
          <HistoryIcon className="h-7 w-7 mr-2 text-indigo-500" />
          Discovery Scan History
        </h1>
        <p className="text-sm text-slate-500 mt-1">
          Historical log of all subnet CIDR scans executed by the scanner engine.
        </p>
      </div>

      {/* History Log table */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm overflow-x-auto">
        {loading ? (
          <div className="flex items-center justify-center p-12">
            <Loader2 className="h-6 w-6 animate-spin text-indigo-500" />
            <span className="ml-3 text-slate-500 font-semibold text-sm">Querying scan logs...</span>
          </div>
        ) : history.length === 0 ? (
          <div className="text-center text-slate-500 p-12 text-sm flex flex-col items-center space-y-2">
            <HistoryIcon className="h-8 w-8 text-slate-350" />
            <p>No historical scans found. Trigger a scan from the discovery panel.</p>
          </div>
        ) : (
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-slate-50 dark:bg-slate-900/40 text-slate-400 text-xs font-bold uppercase tracking-wider border-b border-slate-200 dark:border-darkBorder">
                <th className="px-6 py-4">Target Range</th>
                <th className="px-6 py-4">Profile</th>
                <th className="px-6 py-4">Scan Type</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4">Devices Found</th>
                <th className="px-6 py-4">Duration</th>
                <th className="px-6 py-4">Timestamp</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-darkBorder text-slate-700 dark:text-slate-300 text-sm">
              {history.map((h) => (
                <tr key={h.id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/20 transition-all">
                  <td className="px-6 py-4 font-mono font-semibold text-slate-800 dark:text-slate-200">{h.target}</td>
                  <td className="px-6 py-4 capitalize">{h.scan_profile}</td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${h.scan_type === "scheduled" ? "bg-purple-100 text-purple-700 dark:bg-purple-950/30 dark:text-purple-400" : "bg-blue-100 text-blue-700 dark:bg-blue-950/30 dark:text-blue-400"}`}>
                      {h.scan_type}
                    </span>
                  </td>
                  <td className="px-6 py-4">{getStatusBadge(h.status)}</td>
                  <td className="px-6 py-4 font-mono text-xs">
                    <span className="inline-flex items-center text-green-600 dark:text-green-400 font-bold">
                      <Server className="h-3.5 w-3.5 mr-1" />
                      {h.devices_found}
                    </span>
                  </td>
                  <td className="px-6 py-4 font-mono text-xs text-slate-500 dark:text-slate-400">
                    <span className="inline-flex items-center">
                      <Clock className="h-3.5 w-3.5 mr-1" />
                      {formatDuration(h.duration_ms)}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-xs font-mono text-slate-400">
                    {new Date(h.started_at).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
