import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { Device } from "../types";
import { useSocket } from "../context/SocketContext";
import {
  Search,
  Plus,
  Trash2,
  Edit,
  Upload,
  Download,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  HelpCircle,
  Wrench,
  Loader2,
} from "lucide-react";

export const Devices: React.FC = () => {
  const { lastMessage } = useSocket();
  const [devices, setDevices] = useState<Device[]>([]);

  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [typeFilter, setTypeFilter] = useState("all");

  // Modal states
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingDevice, setEditingDevice] = useState<Device | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    hostname: "",
    ip_address: "",
    mac_address: "",
    device_type: "server",
    os: "",
    vendor: "",
    location: "",
    monitoring_interval: 60,
    parent_id: "",
    group_id: "",
    tags: "",
    notes: "",
    enabled: true,
  });

  const [csvFile, setCsvFile] = useState<File | null>(null);
  const [csvUploading, setCsvUploading] = useState(false);

  useEffect(() => {
    fetchDevices();
  }, []);

  useEffect(() => {
    if (!lastMessage) return;
    const { type, payload } = lastMessage;
    if (type === "device_status") {
      const { device_id, status } = payload;
      setDevices((prev) =>
        prev.map((d) => (d.id === device_id ? { ...d, status } : d))
      );
    }
  }, [lastMessage]);

  const fetchDevices = async () => {
    try {
      const res = await api.get("/devices");
      setDevices(res.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };



  const handleOpenAddModal = () => {
    setEditingDevice(null);
    setFormData({
      name: "",
      hostname: "",
      ip_address: "",
      mac_address: "",
      device_type: "server",
      os: "",
      vendor: "",
      location: "",
      monitoring_interval: 60,
      parent_id: "",
      group_id: "",
      tags: "",
      notes: "",
      enabled: true,
    });
    setIsModalOpen(true);
  };

  const handleOpenEditModal = (dev: Device) => {
    setEditingDevice(dev);
    setFormData({
      name: dev.name,
      hostname: dev.hostname,
      ip_address: dev.ip_address,
      mac_address: dev.mac_address || "",
      device_type: dev.device_type,
      os: dev.os || "",
      vendor: dev.vendor || "",
      location: dev.location || "",
      monitoring_interval: dev.monitoring_interval,
      parent_id: dev.parent_id || "",
      group_id: dev.group_id || "",
      tags: dev.tags || "",
      notes: dev.notes || "",
      enabled: dev.enabled,
    });
    setIsModalOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      ...formData,
      parent_id: formData.parent_id === "" ? null : formData.parent_id,
      group_id: formData.group_id === "" ? null : formData.group_id,
    };

    try {
      if (editingDevice) {
        await api.put(`/devices/${editingDevice.id}`, payload);
      } else {
        await api.post("/devices", payload);
      }
      setIsModalOpen(false);
      fetchDevices();
    } catch (err) {
      console.error(err);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to delete this device?")) return;
    try {
      await api.delete(`/devices/${id}`);
      fetchDevices();
    } catch (err) {
      console.error(err);
    }
  };

  const handleImportCsv = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!csvFile) return;

    const data = new FormData();
    data.append("file", csvFile);
    setCsvUploading(true);

    try {
      await api.post("/devices/import", data, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setCsvFile(null);
      fetchDevices();
      alert("Devices imported successfully!");
    } catch (err) {
      console.error(err);
      alert("Failed to import CSV.");
    } finally {
      setCsvUploading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const classes = "inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold uppercase tracking-wider text-white";
    switch (status) {
      case "online":
        return <span className={`${classes} bg-green-500`}><CheckCircle2 className="h-3 w-3 mr-1" /> Online</span>;
      case "offline":
        return <span className={`${classes} bg-red-500`}><XCircle className="h-3 w-3 mr-1" /> Offline</span>;
      case "warning":
        return <span className={`${classes} bg-amber-500`}><AlertTriangle className="h-3 w-3 mr-1" /> Warning</span>;
      case "unreachable":
        return <span className={`${classes} bg-slate-400`}><HelpCircle className="h-3 w-3 mr-1" /> Unreachable</span>;
      case "maintenance":
        return <span className={`${classes} bg-purple-500`}><Wrench className="h-3 w-3 mr-1" /> Maintenance</span>;
      default:
        return <span className={`${classes} bg-slate-500`}>Unknown</span>;
    }
  };

  // Filters
  const filteredDevices = devices.filter((d) => {
    const matchSearch =
      d.name.toLowerCase().includes(search.toLowerCase()) ||
      d.ip_address.includes(search) ||
      d.hostname.toLowerCase().includes(search.toLowerCase());
    const matchStatus = statusFilter === "all" || d.status === statusFilter;
    const matchType = typeFilter === "all" || d.device_type === typeFilter;
    return matchSearch && matchStatus && matchType;
  });

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-3 sm:space-y-0">
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
          Devices Inventory
        </h1>

        <div className="flex items-center space-x-3">
          {/* CSV Import */}
          <form onSubmit={handleImportCsv} className="flex items-center space-x-2">
            <label className="flex items-center px-3 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-lg text-sm font-semibold cursor-pointer border border-slate-200 dark:border-darkBorder">
              <Upload className="h-4 w-4 mr-2" />
              <span>{csvFile ? csvFile.name.substring(0, 10) + "..." : "Import CSV"}</span>
              <input
                type="file"
                accept=".csv"
                className="hidden"
                onChange={(e) => setCsvFile(e.target.files?.[0] || null)}
              />
            </label>
            {csvFile && (
              <button
                type="submit"
                disabled={csvUploading}
                className="px-3 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-semibold flex items-center"
              >
                {csvUploading ? <Loader2 className="h-4 w-4 animate-spin" /> : "Upload"}
              </button>
            )}
          </form>

          {/* CSV Export */}
          <a
            href="http://localhost:8080/api/v1/devices/export"
            download
            className="flex items-center px-3 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-lg text-sm font-semibold border border-slate-200 dark:border-darkBorder"
          >
            <Download className="h-4 w-4 mr-2" />
            Export CSV
          </a>

          {/* Add Device */}
          <button
            onClick={handleOpenAddModal}
            className="flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-bold shadow-md transition-all"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Device
          </button>
        </div>
      </div>

      {/* Filters row */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder p-4 rounded-xl shadow-sm">
        <div className="relative">
          <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-slate-400">
            <Search className="h-4 w-4" />
          </span>
          <input
            type="text"
            placeholder="Search by name, IP, hostname..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 placeholder-slate-400 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>

        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          <option value="all">All Statuses</option>
          <option value="online">Online</option>
          <option value="offline">Offline</option>
          <option value="warning">Warning</option>
          <option value="unreachable">Unreachable</option>
          <option value="maintenance">Maintenance</option>
        </select>

        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          <option value="all">All Device Types</option>
          <option value="server">Server</option>
          <option value="router">Router</option>
          <option value="switch">Switch</option>
          <option value="printer">Printer</option>
          <option value="pc">Workstation</option>
        </select>
      </div>

      {/* Inventory Table */}
      <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm overflow-x-auto">
        {loading ? (
          <div className="flex items-center justify-center p-12">
            <Loader2 className="h-6 w-6 animate-spin text-indigo-500" />
            <span className="ml-3 text-slate-500 font-semibold text-sm">Querying active list...</span>
          </div>
        ) : filteredDevices.length === 0 ? (
          <p className="text-center text-slate-500 p-12 text-sm">No devices match criteria.</p>
        ) : (
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-slate-50 dark:bg-slate-900/40 text-slate-400 text-xs font-bold uppercase tracking-wider border-b border-slate-200 dark:border-darkBorder">
                <th className="px-6 py-4">Name</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4">IP Address</th>
                <th className="px-6 py-4">Device Type</th>
                <th className="px-6 py-4">Interval</th>
                <th className="px-6 py-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-darkBorder text-slate-700 dark:text-slate-300 text-sm">
              {filteredDevices.map((d) => (
                <tr key={d.id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/20 transition-all">
                  <td className="px-6 py-4 font-semibold text-slate-800 dark:text-slate-200">{d.name}</td>
                  <td className="px-6 py-4">{getStatusBadge(d.status)}</td>
                  <td className="px-6 py-4 font-mono text-xs">{d.ip_address}</td>
                  <td className="px-6 py-4 capitalize">{d.device_type}</td>
                  <td className="px-6 py-4 font-mono text-xs">{d.monitoring_interval}s</td>
                  <td className="px-6 py-4 text-right space-x-3">
                    <button
                      onClick={() => handleOpenEditModal(d)}
                      className="p-1 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400"
                    >
                      <Edit className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => handleDelete(d.id)}
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

      {/* Add / Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 overflow-y-auto">
          <div className="w-full max-w-lg bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-2xl shadow-2xl p-6">
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-6">
              {editingDevice ? "Edit System Device" : "Register New Device"}
            </h3>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Device Name</label>
                  <input
                    type="text"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Hostname / Domain</label>
                  <input
                    type="text"
                    required
                    value={formData.hostname}
                    onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">IP Address</label>
                  <input
                    type="text"
                    required
                    value={formData.ip_address}
                    onChange={(e) => setFormData({ ...formData, ip_address: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono text-slate-800 dark:text-slate-100 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Device Type</label>
                  <select
                    value={formData.device_type}
                    onChange={(e) => setFormData({ ...formData, device_type: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                  >
                    <option value="server">Server</option>
                    <option value="router">Router</option>
                    <option value="switch">Switch</option>
                    <option value="printer">Printer</option>
                    <option value="pc">Workstation</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Operating System</label>
                  <input
                    type="text"
                    value={formData.os}
                    onChange={(e) => setFormData({ ...formData, os: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Vendor</label>
                  <input
                    type="text"
                    value={formData.vendor}
                    onChange={(e) => setFormData({ ...formData, vendor: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Monitoring Interval (s)</label>
                  <input
                    type="number"
                    required
                    value={formData.monitoring_interval}
                    onChange={(e) => setFormData({ ...formData, monitoring_interval: parseInt(e.target.value) || 60 })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Parent Uplink Device</label>
                  <select
                    value={formData.parent_id}
                    onChange={(e) => setFormData({ ...formData, parent_id: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                  >
                    <option value="">No Parent (Root)</option>
                    {devices
                      .filter((d) => d.id !== editingDevice?.id) // avoid self-loop
                      .map((d) => (
                        <option key={d.id} value={d.id}>
                          {d.name} ({d.ip_address})
                        </option>
                      ))}
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-400 mb-1">Location / Rack</label>
                <input
                  type="text"
                  value={formData.location}
                  onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                  className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100"
                />
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
                  Save Device
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
