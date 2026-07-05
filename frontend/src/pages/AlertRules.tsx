import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { Alert, AlertRule, Device } from "../types";
import { useSocket } from "../context/SocketContext";
import { AlertTriangle, Plus, Trash2, ShieldCheck } from "lucide-react";

export const AlertRules: React.FC = () => {
  const { lastMessage } = useSocket();
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);


  // Modal State
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    device_id: "",
    metric: "latency_ms",
    operator: ">",
    value: 100,
    duration: 30,
    level: "warning",
    enabled: true,
  });

  useEffect(() => {
    fetchAlerts();
    fetchRules();
    fetchDevices();
  }, []);

  useEffect(() => {
    if (!lastMessage) return;
    const { type } = lastMessage;
    if (type === "alert") {
      fetchAlerts();
    }
  }, [lastMessage]);

  const fetchAlerts = async () => {
    try {
      const res = await api.get("/alerts");
      setAlerts(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const fetchRules = async () => {
    try {
      const res = await api.get("/rules");
      setRules(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const fetchDevices = async () => {
    try {
      const res = await api.get("/devices");
      setDevices(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const handleResolveAlert = async (id: string) => {
    try {
      await api.put(`/alerts/${id}/resolve`);
      fetchAlerts();
    } catch (err) {
      console.error(err);
    }
  };

  const handleDeleteRule = async (id: string) => {
    if (!window.confirm("Delete this alert rule?")) return;
    try {
      await api.delete(`/rules/${id}`);
      fetchRules();
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateRule = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      ...formData,
      device_id: formData.device_id === "" ? null : formData.device_id,
    };

    try {
      await api.post("/rules", payload);
      setIsModalOpen(false);
      fetchRules();
      setFormData({
        name: "",
        device_id: "",
        metric: "latency_ms",
        operator: ">",
        value: 100,
        duration: 30,
        level: "warning",
        enabled: true,
      });
    } catch (err) {
      console.error(err);
    }
  };

  const getAlertBadge = (level: string) => {
    const classes = "inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider text-white";
    switch (level) {
      case "critical": return <span className={`${classes} bg-red-500`}>Critical</span>;
      case "warning": return <span className={`${classes} bg-amber-500`}>Warning</span>;
      default: return <span className={`${classes} bg-blue-500`}>Info</span>;
    }
  };

  const activeAlerts = alerts.filter((a) => a.status === "active");


  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
          Incident Center & Rules
        </h1>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-lg text-sm shadow-md"
        >
          <Plus className="h-4 w-4 mr-2" />
          Add Alarm Rule
        </button>
      </div>

      {/* Active Incidents list */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm overflow-hidden">
        <div className="p-4 bg-slate-50 dark:bg-slate-900/30 border-b border-slate-200 dark:border-darkBorder flex justify-between items-center">
          <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">Active Incidents ({activeAlerts.length})</h3>
        </div>

        {activeAlerts.length === 0 ? (
          <div className="p-8 text-center text-slate-400 text-sm">
            <ShieldCheck className="h-8 w-8 text-emerald-500 mx-auto mb-2" />
            No active incidents. Network health stable.
          </div>
        ) : (
          <div className="divide-y divide-slate-100 dark:divide-darkBorder">
            {activeAlerts.map((a) => (
              <div key={a.id} className="flex justify-between items-center p-4 hover:bg-slate-50/20 dark:hover:bg-slate-800/10">
                <div className="flex items-center space-x-3">
                  <div className={`p-2 rounded-full ${a.level === "critical" ? "bg-red-500/10 text-red-500" : "bg-amber-500/10 text-amber-500"}`}>
                    <AlertTriangle className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="text-sm font-bold text-slate-800 dark:text-slate-200">{a.device?.name} ({a.device?.ip_address})</h4>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{a.message}</p>
                  </div>
                </div>
                <div className="flex items-center space-x-4">
                  <div className="text-right">
                    {getAlertBadge(a.level)}
                    <span className="block text-[10px] text-slate-400 mt-1 font-mono">{new Date(a.created_at).toLocaleTimeString()}</span>
                  </div>
                  <button
                    onClick={() => handleResolveAlert(a.id)}
                    className="px-3 py-1.5 bg-emerald-55 hover:bg-emerald-600 dark:bg-emerald-950/20 dark:hover:bg-emerald-900/40 text-emerald-600 dark:text-emerald-400 font-bold rounded-lg text-xs border border-emerald-200 dark:border-emerald-900/30"
                  >
                    Acknowledge
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Rules list */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm overflow-hidden">
        <div className="p-4 bg-slate-55 dark:bg-slate-900/30 border-b border-slate-200 dark:border-darkBorder">
          <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400 font-bold">Configured Threshold Rules</h3>
        </div>

        {rules.length === 0 ? (
          <p className="text-center text-slate-500 p-8 text-sm">No alert rules configured.</p>
        ) : (
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-slate-50/50 dark:bg-slate-900/10 text-slate-400 text-xs font-bold border-b border-slate-200 dark:border-darkBorder uppercase">
                <th className="px-6 py-4">Rule Name</th>
                <th className="px-6 py-4">Metric/Threshold</th>
                <th className="px-6 py-4">Duration</th>
                <th className="px-6 py-4">Severity</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4 text-right font-bold">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-darkBorder text-slate-700 dark:text-slate-350 text-sm">
              {rules.map((r) => (
                <tr key={r.id} className="hover:bg-slate-50/10 dark:hover:bg-slate-800/10">
                  <td className="px-6 py-4 font-semibold text-slate-850 dark:text-slate-200">{r.name}</td>
                  <td className="px-6 py-4 font-mono text-xs text-indigo-500 dark:text-indigo-400">
                    {r.metric} {r.operator} {r.value}
                  </td>
                  <td className="px-6 py-4 font-mono text-xs">{r.duration}s</td>
                  <td className="px-6 py-4 capitalize">{getAlertBadge(r.level)}</td>
                  <td className="px-6 py-4">
                    {r.enabled ? (
                      <span className="text-emerald-500 text-xs font-bold">Active</span>
                    ) : (
                      <span className="text-slate-400 text-xs font-bold">Disabled</span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-right pr-8">
                    <button
                      onClick={() => handleDeleteRule(r.id)}
                      className="text-slate-400 hover:text-red-500"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Rules Builder Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 overflow-y-auto">
          <div className="w-full max-w-md bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-2xl shadow-2xl p-6">
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-6">Create Alert Rule</h3>

            <form onSubmit={handleCreateRule} className="space-y-4">
              <div>
                <label className="block text-xs font-bold text-slate-400 mb-1">Rule Name</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g. Server Disk Space Warning"
                  className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Target Metric</label>
                  <select
                    value={formData.metric}
                    onChange={(e) => setFormData({ ...formData, metric: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  >
                    <option value="latency_ms">Ping Latency (ms)</option>
                    <option value="packet_loss">Packet Loss (%)</option>
                    <option value="response_time">TCP Port Response (ms)</option>
                    <option value="cpu">CPU Usage (%)</option>
                    <option value="ram">RAM Usage (%)</option>
                    <option value="disk">Disk Usage (%)</option>
                    <option value="status">Device Offline Status</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Target Device</label>
                  <select
                    value={formData.device_id}
                    onChange={(e) => setFormData({ ...formData, device_id: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                  >
                    <option value="">Apply Globals (All Devices)</option>
                    {devices.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.name} ({d.ip_address})
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Comparison Operator</label>
                  <select
                    value={formData.operator}
                    onChange={(e) => setFormData({ ...formData, operator: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  >
                    <option value=">">&gt; (Greater than)</option>
                    <option value="<">&lt; (Less than)</option>
                    <option value="==">== (Equals)</option>
                    <option value="!=">!= (Not Equals)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Threshold Limit Value</label>
                  <input
                    type="number"
                    required
                    step="0.01"
                    value={formData.value}
                    onChange={(e) => setFormData({ ...formData, value: parseFloat(e.target.value) || 0 })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Duration Trigger (s)</label>
                  <input
                    type="number"
                    required
                    value={formData.duration}
                    onChange={(e) => setFormData({ ...formData, duration: parseInt(e.target.value) || 0 })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  />
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Severity Level</label>
                  <select
                    value={formData.level}
                    onChange={(e) => setFormData({ ...formData, level: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  >
                    <option value="info">Info</option>
                    <option value="warning">Warning</option>
                    <option value="critical">Critical</option>
                  </select>
                </div>
              </div>

              <div className="flex justify-end space-x-3 mt-8">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 font-semibold rounded-lg text-sm"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-lg text-sm shadow-md"
                >
                  Create Rule
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
