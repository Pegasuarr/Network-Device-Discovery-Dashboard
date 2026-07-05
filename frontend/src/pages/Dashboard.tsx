import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { DashboardStats, Device, Alert } from "../types";
import { useSocket } from "../context/SocketContext";
import { TopologyMap } from "../components/Topology/TopologyMap";
import {
  Activity,
  Server,
  AlertTriangle,
  CheckCircle,
  XCircle,
} from "lucide-react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from "recharts";

export const Dashboard: React.FC = () => {
  const { lastMessage } = useSocket();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [latencyTrend, setLatencyTrend] = useState<any[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [activeAlerts, setActiveAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);

  // Load data
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statsRes, trendRes, devRes, alertRes] = await Promise.all([
          api.get("/dashboard/stats"),
          api.get("/dashboard/latency"),
          api.get("/devices"),
          api.get("/alerts"),
        ]);
        setStats(statsRes.data);
        setLatencyTrend(trendRes.data);
        setDevices(devRes.data);
        setActiveAlerts(alertRes.data.filter((a: Alert) => a.status === "active"));
      } catch (err) {
        console.error("Dashboard error fetching metrics", err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  // Listen to WebSocket broadcasts
  useEffect(() => {
    if (!lastMessage) return;

    const { type, payload } = lastMessage;

    if (type === "device_status") {
      const { device_id, status } = payload;
      setDevices((prev) =>
        prev.map((d) => (d.id === device_id ? { ...d, status } : d))
      );

      // Re-trigger stats fetch to update dashboard card numbers
      api.get("/dashboard/stats").then((res) => setStats(res.data));
    }

    if (type === "ping_result") {
      // Re-trigger stats to update averages
      api.get("/dashboard/stats").then((res) => setStats(res.data));
    }
  }, [lastMessage]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Activity className="h-8 w-8 text-indigo-500 animate-spin" />
        <span className="ml-3 text-slate-500 font-semibold">Loading telemetry data...</span>
      </div>
    );
  }

  // Prep Status Pie chart data
  const pieData = [
    { name: "Online", value: stats?.online_devices || 0, color: "#10B981" },
    { name: "Offline", value: stats?.offline_devices || 0, color: "#EF4444" },
    { name: "Warning", value: stats?.warning_devices || 0, color: "#F59E0B" },
    { name: "Unreachable", value: stats?.unreachable_devices || 0, color: "#94A3B8" },
    { name: "Maintenance", value: stats?.warning_devices ? 0 : 0, color: "#8B5CF6" }, // dynamic placeholder
  ].filter((slice) => slice.value > 0);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
          NOC Operational Dashboard
        </h1>
        <span className="text-xs font-mono text-slate-500 bg-slate-100 dark:bg-slate-800 px-3 py-1.5 rounded-lg border border-slate-200 dark:border-darkBorder">
          Last updated: {new Date().toLocaleTimeString()}
        </span>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-indigo-500/10 rounded-lg text-indigo-500 mr-4">
            <Server className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">Total Devices</p>
            <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{stats?.total_devices}</h3>
          </div>
        </div>

        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-emerald-500/10 rounded-lg text-emerald-500 mr-4">
            <CheckCircle className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">Online</p>
            <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{stats?.online_devices}</h3>
          </div>
        </div>

        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-red-500/10 rounded-lg text-red-500 mr-4">
            <XCircle className="h-6 w-6 animate-pulse" />
          </div>
          <div>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">Offline</p>
            <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{stats?.offline_devices}</h3>
          </div>
        </div>

        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-amber-500/10 rounded-lg text-amber-500 mr-4">
            <AlertTriangle className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">Active Alerts</p>
            <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{stats?.active_alerts}</h3>
          </div>
        </div>
      </div>

      {/* Network Metrics Gauge Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-5">
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Avg Latency</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">{stats?.avg_latency_ms?.toFixed(2)} ms</h2>
        </div>
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Packet Loss</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">{stats?.avg_packet_loss?.toFixed(1)}%</h2>
        </div>
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Avg CPU</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">{stats?.avg_cpu?.toFixed(1)}%</h2>
        </div>
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Avg RAM</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">{stats?.avg_ram?.toFixed(1)}%</h2>
        </div>
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Avg Disk</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">{stats?.avg_disk?.toFixed(1)}%</h2>
        </div>
      </div>

      {/* Network Topology Map Section */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-6 shadow-sm">
        <h2 className="text-base font-bold text-slate-800 dark:text-slate-100 mb-4">Network Topology Map</h2>
        <TopologyMap devices={devices} />
      </div>

      {/* Graphs row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Latency line chart */}
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm lg:col-span-2">
          <h2 className="text-sm font-bold text-slate-800 dark:text-slate-100 mb-4">NOC Average Latency Trend (Last 24h)</h2>
          <div className="h-60">
            {latencyTrend.length === 0 ? (
              <div className="flex items-center justify-center h-full text-slate-400 text-sm">
                No latency records logged yet
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={latencyTrend}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#222D44" />
                  <XAxis dataKey="time" stroke="#64748B" fontSize={11} />
                  <YAxis stroke="#64748B" fontSize={11} unit="ms" />
                  <Tooltip
                    contentStyle={{ backgroundColor: "#151B2C", borderColor: "#222D44" }}
                    labelStyle={{ color: "#94A3B8" }}
                  />
                  <Line type="monotone" dataKey="latency_ms" stroke="#3B82F6" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        {/* Pie distribution */}
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
          <h2 className="text-sm font-bold text-slate-800 dark:text-slate-100 mb-4">Device Status Ratios</h2>
          <div className="h-60 flex items-center justify-center">
            {pieData.length === 0 ? (
              <div className="text-slate-400 text-sm">No device data available</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={pieData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {pieData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend verticalAlign="bottom" height={36} />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>
      </div>

      {/* Recent Alerts Feed */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
        <h2 className="text-sm font-bold text-slate-800 dark:text-slate-100 mb-4">Active Alarm Timeline</h2>
        {activeAlerts.length === 0 ? (
          <p className="text-xs text-slate-500 text-center py-4">System cleared. No active critical alarms.</p>
        ) : (
          <div className="space-y-3">
            {activeAlerts.slice(0, 5).map((a) => (
              <div key={a.id} className="flex justify-between items-center p-3 bg-slate-50 dark:bg-slate-900/40 border border-slate-150 dark:border-darkBorder rounded-lg">
                <div className="flex items-center space-x-3">
                  <div className={`p-1.5 rounded-full ${a.level === "critical" ? "bg-red-500/10 text-red-500" : "bg-amber-500/10 text-amber-500"}`}>
                    <AlertTriangle className="h-4 w-4" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-slate-700 dark:text-slate-200">{a.device?.name}</h4>
                    <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">{a.message}</p>
                  </div>
                </div>
                <div className="text-right">
                  <span className={`px-2 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-white ${a.level === "critical" ? "bg-red-500" : "bg-amber-500"}`}>
                    {a.level}
                  </span>
                  <p className="text-[10px] text-slate-400 font-mono mt-1">{new Date(a.created_at).toLocaleTimeString()}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
