import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { ScanSchedule } from "../types";
import {
  Calendar,
  Plus,
  Trash2,
  CheckCircle,
  XCircle,
  Clock,
  Globe,
  Loader2,
} from "lucide-react";

export const Schedules: React.FC = () => {
  const [schedules, setSchedules] = useState<ScanSchedule[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  
  const [formData, setFormData] = useState({
    name: "",
    target: "192.168.1.0/24",
    cron_expression: "*/30 * * * *",
    scan_profile: "quick",
    enabled: true,
  });

  useEffect(() => {
    fetchSchedules();
  }, []);

  const fetchSchedules = async () => {
    try {
      const res = await api.get("/discovery/schedules");
      setSchedules(res.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenAddModal = () => {
    setFormData({
      name: "",
      target: "192.168.1.0/24",
      cron_expression: "*/30 * * * *",
      scan_profile: "quick",
      enabled: true,
    });
    setIsModalOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post("/discovery/schedules", formData);
      setIsModalOpen(false);
      fetchSchedules();
    } catch (err) {
      console.error(err);
      alert("Failed to save schedule.");
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to delete this scan schedule?")) return;
    try {
      await api.delete(`/discovery/schedules/${id}`);
      fetchSchedules();
    } catch (err) {
      console.error(err);
      alert("Failed to delete schedule.");
    }
  };

  return (
    <div className="space-y-6 pb-12">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100 flex items-center">
            <Calendar className="h-7 w-7 mr-2 text-indigo-500" />
            Scheduled Scans Configuration
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Automate discovery scans at periodic intervals using standard cron scheduling expressions.
          </p>
        </div>

        <button
          onClick={handleOpenAddModal}
          className="flex items-center px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-bold shadow-md transition-all"
        >
          <Plus className="h-4 w-4 mr-2" />
          Add Schedule Rule
        </button>
      </div>

      {/* Schedules list */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm overflow-x-auto">
        {loading ? (
          <div className="flex items-center justify-center p-12">
            <Loader2 className="h-6 w-6 animate-spin text-indigo-500" />
            <span className="ml-3 text-slate-500 font-semibold text-sm">Querying cron configurations...</span>
          </div>
        ) : schedules.length === 0 ? (
          <div className="text-center text-slate-500 p-12 text-sm flex flex-col items-center space-y-2">
            <Calendar className="h-8 w-8 text-slate-350" />
            <p>No scheduled tasks mapped. Click 'Add Schedule Rule' to automate discovery sweeps.</p>
          </div>
        ) : (
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-slate-50 dark:bg-slate-900/40 text-slate-400 text-xs font-bold uppercase tracking-wider border-b border-slate-200 dark:border-darkBorder">
                <th className="px-6 py-4">Rule Name</th>
                <th className="px-6 py-4">Target Range</th>
                <th className="px-6 py-4">Cron Expression</th>
                <th className="px-6 py-4">Profile</th>
                <th className="px-6 py-4">State</th>
                <th className="px-6 py-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-darkBorder text-slate-700 dark:text-slate-300 text-sm">
              {schedules.map((s) => (
                <tr key={s.id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/20 transition-all">
                  <td className="px-6 py-4 font-semibold text-slate-800 dark:text-slate-200">{s.name}</td>
                  <td className="px-6 py-4 font-mono text-xs">
                    <span className="inline-flex items-center">
                      <Globe className="h-3.5 w-3.5 mr-1.5 text-slate-400" />
                      {s.target}
                    </span>
                  </td>
                  <td className="px-6 py-4 font-mono text-xs">
                    <span className="inline-flex items-center">
                      <Clock className="h-3.5 w-3.5 mr-1.5 text-slate-400" />
                      {s.cron_expression}
                    </span>
                  </td>
                  <td className="px-6 py-4 capitalize">{s.scan_profile}</td>
                  <td className="px-6 py-4">
                    {s.enabled ? (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-green-500/10 text-green-500">
                        <CheckCircle className="h-3 w-3 mr-1" /> Active
                      </span>
                    ) : (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-slate-100 text-slate-500 dark:bg-slate-900 dark:text-slate-400">
                        <XCircle className="h-3 w-3 mr-1" /> Paused
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => handleDelete(s.id)}
                      className="p-1 text-slate-400 hover:text-red-600 dark:hover:text-red-400"
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

      {/* Add Schedule Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 overflow-y-auto">
          <div className="w-full max-w-md bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-2xl shadow-2xl p-6">
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-6">
              Create Automatic Scan Schedule
            </h3>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-bold text-slate-400 mb-1 uppercase tracking-wider">Rule Name</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g. Daily DMZ Sweep"
                  className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-400 mb-1 uppercase tracking-wider">Scan Target</label>
                <input
                  type="text"
                  required
                  value={formData.target}
                  onChange={(e) => setFormData({ ...formData, target: e.target.value })}
                  placeholder="e.g. 192.168.1.0/24"
                  className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono text-slate-800 dark:text-slate-100 focus:outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1 uppercase tracking-wider">Cron Expression</label>
                  <input
                    type="text"
                    required
                    value={formData.cron_expression}
                    onChange={(e) => setFormData({ ...formData, cron_expression: e.target.value })}
                    placeholder="e.g. */30 * * * *"
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono text-slate-800 dark:text-slate-100 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1 uppercase tracking-wider">Scan Profile</label>
                  <select
                    value={formData.scan_profile}
                    onChange={(e) => setFormData({ ...formData, scan_profile: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                  >
                    <option value="quick">Quick Scan</option>
                    <option value="deep">Deep Scan</option>
                    <option value="portscan">Port Scan</option>
                    <option value="ping">Ping Only</option>
                  </select>
                </div>
              </div>

              <div className="flex items-center space-x-2 pt-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                  className="rounded text-indigo-600 focus:ring-indigo-500 h-4 w-4"
                />
                <label htmlFor="enabled" className="text-sm font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  Enable Scheduled Sweeps
                </label>
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
