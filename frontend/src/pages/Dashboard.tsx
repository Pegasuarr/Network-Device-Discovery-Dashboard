import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { DashboardStats, Device, DeviceTimeline } from "../types";
import { useSocket } from "../context/SocketContext";
import {
  Activity,
  Server,
  CheckCircle,
  XCircle,
  Clock,
  Globe,
  Radio,
  Bell,
  Database,
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
  const [timeline, setTimeline] = useState<DeviceTimeline[]>([]);
  const [loading, setLoading] = useState(true);

  // Load telemetry data
  const loadDashboardData = async () => {
    try {
      const [statsRes, trendRes, devRes, timelineRes] = await Promise.all([
        api.get("/dashboard/stats"),
        api.get("/dashboard/latency"),
        api.get("/devices"),
        api.get("/notifications"),
      ]);
      setStats(statsRes.data);
      setLatencyTrend(trendRes.data);
      setDevices(devRes.data);
      setTimeline(timelineRes.data);
    } catch (err) {
      console.error("Dashboard error loading data", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDashboardData();
  }, []);

  // Listen to live WebSocket events to update the dashboard dynamically
  useEffect(() => {
    if (!lastMessage) return;

    const { type } = lastMessage;

    if (type === "device_status" || type === "ping_result") {
      // Re-trigger statistical loads to update graphs and counters
      api.get("/dashboard/stats").then((res) => setStats(res.data));
    }

    if (type === "device_notification") {
      // Refresh notifications feed and stats
      loadDashboardData();
    }
  }, [lastMessage]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center h-[70vh] space-y-4">
        <Activity className="h-10 w-10 text-indigo-500 animate-spin" />
        <span className="text-slate-500 dark:text-slate-400 font-semibold tracking-wider">
          POLLING TELEMETRY GRAPHS...
        </span>
      </div>
    );
  }

  // Prepare device type distribution pie chart
  const deviceTypeData = devices.reduce((acc: any[], device) => {
    const type = device.device_type || "workstation";
    const existing = acc.find((item) => item.name === type);
    if (existing) {
      existing.value += 1;
    } else {
      acc.push({ name: type, value: 1 });
    }
    return acc;
  }, []);

  const COLORS = ["#4F46E5", "#10B981", "#F59E0B", "#EF4444", "#8B5CF6", "#EC4899", "#3B82F6"];

  return (
    <div className="space-y-6 pb-12">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
            NOC Operational Dashboard
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
            Real-time status of local network devices and discovery scanning.
          </p>
        </div>
        <span className="text-xs font-mono text-slate-500 bg-slate-100 dark:bg-slate-800 px-3 py-1.5 rounded-lg border border-slate-200 dark:border-darkBorder">
          System Time: {new Date().toLocaleTimeString()}
        </span>
      </div>

      {/* Network discovery status widgets */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-5">
        {/* Total Devices */}
        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-indigo-500/10 rounded-lg text-indigo-500 mr-4">
            <Server className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-slate-400 font-semibold uppercase tracking-wider">Total Devices</p>
            <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{stats?.total_devices}</h3>
          </div>
        </div>

        {/* Online Devices */}
        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-emerald-500/10 rounded-lg text-emerald-500 mr-4">
            <CheckCircle className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-slate-400 font-semibold uppercase tracking-wider">Online</p>
            <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{stats?.online_devices}</h3>
          </div>
        </div>

        {/* Offline Devices */}
        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-red-500/10 rounded-lg text-red-500 mr-4">
            <XCircle className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-slate-400 font-semibold uppercase tracking-wider">Offline</p>
            <h3 className="text-2xl font-bold text-slate-800 dark:text-slate-100">{stats?.offline_devices}</h3>
          </div>
        </div>

        {/* Current Subnet */}
        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-amber-500/10 rounded-lg text-amber-500 mr-4">
            <Globe className="h-6 w-6" />
          </div>
          <div className="truncate">
            <p className="text-xs text-slate-400 font-semibold uppercase tracking-wider">Active Network</p>
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 truncate">{stats?.current_network || "None"}</h3>
          </div>
        </div>

        {/* Last Scan Time */}
        <div className="flex items-center p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm">
          <div className="p-3 bg-purple-500/10 rounded-lg text-purple-500 mr-4">
            <Clock className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-slate-400 font-semibold uppercase tracking-wider">Last Scan Duration</p>
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100">
              {stats?.scan_duration ? `${stats.scan_duration.toFixed(1)}s` : "N/A"}
            </h3>
          </div>
        </div>
      </div>

      {/* Network statistics gauges */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Avg Ping Latency</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">
            {stats?.avg_latency_ms ? `${stats.avg_latency_ms.toFixed(2)} ms` : "0.00 ms"}
          </h2>
        </div>
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Avg Packet Loss</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">
            {stats?.avg_packet_loss ? `${stats.avg_packet_loss.toFixed(1)}%` : "0.0%"}
          </h2>
        </div>
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Active SNMP Devices</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">
            {devices.filter(d => d.snmp_enabled).length}
          </h2>
        </div>
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-4 shadow-sm text-center">
          <p className="text-xs text-slate-400 font-semibold uppercase">Total Alerts Registered</p>
          <h2 className="text-xl font-bold mt-1 text-slate-800 dark:text-slate-200">{stats?.active_alerts}</h2>
        </div>
      </div>

      {/* Charts section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Latency line chart */}
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm lg:col-span-2">
          <h2 className="text-sm font-bold text-slate-800 dark:text-slate-100 mb-4 flex items-center">
            <Radio className="h-4 w-4 mr-2 text-indigo-500 animate-pulse" />
            NOC Average Ping Latency Trend (Last 24h)
          </h2>
          <div className="h-64">
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
                  <Line type="monotone" dataKey="latency_ms" stroke="#4F46E5" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        {/* Pie Distribution of Device types */}
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
          <h2 className="text-sm font-bold text-slate-800 dark:text-slate-100 mb-4 flex items-center">
            <Database className="h-4 w-4 mr-2 text-emerald-500" />
            Device Classification Types
          </h2>
          <div className="h-64 flex flex-col items-center justify-center">
            {deviceTypeData.length === 0 ? (
              <div className="text-slate-400 text-sm">No device data available</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={deviceTypeData}
                    cx="50%"
                    cy="45%"
                    innerRadius={50}
                    outerRadius={75}
                    paddingAngle={4}
                    dataKey="value"
                  >
                    {deviceTypeData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
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

      {/* Notifications timeline */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm">
        <h2 className="text-sm font-bold text-slate-800 dark:text-slate-100 mb-4 flex items-center">
          <Bell className="h-4 w-4 mr-2 text-indigo-500 animate-bounce" />
          Live Network Activity Log
        </h2>
        {timeline.length === 0 ? (
          <p className="text-xs text-slate-500 text-center py-6">No network activity logs recorded yet.</p>
        ) : (
          <div className="space-y-3 max-h-72 overflow-y-auto pr-2">
            {timeline.slice(0, 10).map((t) => (
              <div
                key={t.id}
                className="flex justify-between items-center p-3 bg-slate-50 dark:bg-slate-900/40 border border-slate-150 dark:border-darkBorder rounded-lg"
              >
                <div className="flex items-center space-x-3">
                  <div
                    className={`p-1.5 rounded-full ${
                      t.event_type === "join"
                        ? "bg-indigo-500/10 text-indigo-500"
                        : t.event_type === "online"
                        ? "bg-green-500/10 text-green-500"
                        : "bg-red-500/10 text-red-500"
                    }`}
                  >
                    {t.event_type === "join" ? (
                      <Server className="h-4 w-4" />
                    ) : t.event_type === "online" ? (
                      <CheckCircle className="h-4 w-4" />
                    ) : (
                      <XCircle className="h-4 w-4" />
                    )}
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-slate-700 dark:text-slate-200 capitalize">
                      {t.event_type} Event
                    </h4>
                    <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">{t.message}</p>
                  </div>
                </div>
                <div className="text-right">
                  <span className="text-[10px] text-slate-400 font-mono">
                    {new Date(t.checked_at).toLocaleTimeString()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
