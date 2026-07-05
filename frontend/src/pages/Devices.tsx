import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { Device, DeviceTimeline, SNMPInterface } from "../types";
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
  Eye,
  Info,
  FileText,
} from "lucide-react";

export const Devices: React.FC = () => {
  const { lastMessage } = useSocket();
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  
  // Search, Filters & Sort
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [typeFilter, setTypeFilter] = useState("all");
  const [sortBy, setSortBy] = useState("ip");
  const [sortOrder, setSortOrder] = useState("asc");

  // Detail Drawer state
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [deviceTimeline, setDeviceTimeline] = useState<DeviceTimeline[]>([]);
  const [devicePorts, setDevicePorts] = useState<number[]>([]);
  const [deviceInterfaces, setDeviceInterfaces] = useState<SNMPInterface[]>([]);
  const [detailTab, setDetailTab] = useState<"general" | "network" | "stats">("general");

  // Modal states (Add / Edit)
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingDevice, setEditingDevice] = useState<Device | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    hostname: "",
    ip_address: "",
    mac_address: "",
    device_type: "pc",
    os: "",
    vendor: "",
    location: "",
    monitoring_interval: 60,
    parent_id: "",
    tags: "",
    notes: "",
    enabled: true,
  });

  const [csvFile, setCsvFile] = useState<File | null>(null);
  const [csvUploading, setCsvUploading] = useState(false);

  useEffect(() => {
    fetchDevices();
  }, [search, statusFilter, typeFilter, sortBy, sortOrder]);

  useEffect(() => {
    if (!lastMessage) return;
    const { type, payload } = lastMessage;
    
    if (type === "device_status") {
      const { device_id, status } = payload;
      setDevices((prev) =>
        prev.map((d) => (d.id === device_id ? { ...d, status } : d))
      );
      if (selectedDevice?.id === device_id) {
        setSelectedDevice((prev) => prev ? { ...prev, status } : null);
      }
    }
  }, [lastMessage, selectedDevice]);

  const fetchDevices = async () => {
    setLoading(true);
    try {
      const params: any = {
        search,
        status: statusFilter,
        device_type: typeFilter,
        sort_by: sortBy,
        sort_order: sortOrder,
      };
      const res = await api.get("/devices", { params });
      setDevices(res.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDeviceDetail = async (dev: Device) => {
    setDetailLoading(true);
    setSelectedDevice(dev);
    setDetailTab("general");
    setDeviceTimeline([]);
    setDevicePorts([]);
    setDeviceInterfaces([]);
    
    try {
      const res = await api.get(`/devices/${dev.id}`);
      setSelectedDevice(res.data.device);
      setDeviceTimeline(res.data.timeline || []);
      setDevicePorts(res.data.ports || []);
      setDeviceInterfaces(res.data.interfaces || []);
    } catch (err) {
      console.error("Failed to load device details", err);
    } finally {
      setDetailLoading(false);
    }
  };

  const handleOpenAddModal = () => {
    setEditingDevice(null);
    setFormData({
      name: "",
      hostname: "",
      ip_address: "",
      mac_address: "",
      device_type: "pc",
      os: "",
      vendor: "",
      location: "",
      monitoring_interval: 60,
      parent_id: "",
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
      setSelectedDevice(null);
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
    const classes = "inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider text-white";
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

  return (
    <div className="space-y-6 pb-12">
      {/* Top action bar */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-3 sm:space-y-0">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
            Devices Inventory
          </h1>
          <p className="text-sm text-slate-500 mt-1">Manage network configuration, uplinks, and details.</p>
        </div>

        <div className="flex items-center space-x-3">
          {/* CSV Import */}
          <form onSubmit={handleImportCsv} className="flex items-center space-x-2">
            <label className="flex items-center px-3 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-bold cursor-pointer border border-slate-200 dark:border-darkBorder">
              <Upload className="h-4 w-4 mr-2" />
              <span>{csvFile ? csvFile.name.substring(0, 8) + "..." : "Import CSV"}</span>
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
                className="px-2.5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-bold"
              >
                {csvUploading ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : "Upload"}
              </button>
            )}
          </form>

          {/* CSV Export */}
          <a
            href="http://localhost:8080/api/v1/devices/export"
            download
            className="flex items-center px-3 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-bold border border-slate-200 dark:border-darkBorder"
          >
            <Download className="h-4 w-4 mr-2" />
            CSV
          </a>

          {/* PDF Report Export */}
          <a
            href="http://localhost:8080/api/v1/devices/export/pdf"
            target="_blank"
            rel="noreferrer"
            className="flex items-center px-3 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-bold border border-slate-200 dark:border-darkBorder"
          >
            <FileText className="h-4 w-4 mr-2" />
            PDF Report
          </a>

          {/* Add Device */}
          <button
            onClick={handleOpenAddModal}
            className="flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-bold shadow-md transition-all"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Device
          </button>
        </div>
      </div>

      {/* Query filters */}
      <div className="grid grid-cols-1 sm:grid-cols-5 gap-3 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder p-4 rounded-xl shadow-sm">
        <div className="relative col-span-2">
          <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-slate-400">
            <Search className="h-4 w-4" />
          </span>
          <input
            type="text"
            placeholder="Search IP, hostname, MAC, vendor..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 placeholder-slate-400 text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>

        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 text-xs focus:outline-none"
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
          className="px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 text-xs focus:outline-none"
        >
          <option value="all">All Types</option>
          <option value="router">Router</option>
          <option value="switch">Switch</option>
          <option value="server">Server</option>
          <option value="pc">Workstation</option>
          <option value="phone">Phone</option>
          <option value="printer">Printer</option>
          <option value="iot">IoT Device</option>
        </select>

        <div className="flex space-x-2">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            className="flex-1 px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-slate-800 dark:text-slate-100 text-xs focus:outline-none"
          >
            <option value="ip">Sort: IP Address</option>
            <option value="hostname">Sort: Hostname</option>
            <option value="vendor">Sort: Vendor</option>
            <option value="last_seen">Sort: Last Seen</option>
          </select>
          <button
            onClick={() => setSortOrder(prev => prev === "asc" ? "desc" : "asc")}
            className="px-2.5 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-xs font-bold text-slate-700 dark:text-slate-300 hover:bg-slate-100"
          >
            {sortOrder.toUpperCase()}
          </button>
        </div>
      </div>

      {/* Main split display: Table + Detail drawer */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
        {/* Table list */}
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm overflow-hidden xl:col-span-2">
          {loading ? (
            <div className="flex items-center justify-center p-12">
              <Loader2 className="h-6 w-6 animate-spin text-indigo-500" />
              <span className="ml-3 text-slate-500 font-semibold text-xs">Querying devices...</span>
            </div>
          ) : devices.length === 0 ? (
            <p className="text-center text-slate-500 p-12 text-xs">No devices found matching parameters.</p>
          ) : (
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-slate-50 dark:bg-slate-900/40 text-slate-400 text-[10px] font-bold uppercase tracking-wider border-b border-slate-200 dark:border-darkBorder">
                  <th className="px-5 py-3.5">Name</th>
                  <th className="px-5 py-3.5">Status</th>
                  <th className="px-5 py-3.5">IP Address</th>
                  <th className="px-5 py-3.5">Vendor / OUI</th>
                  <th className="px-5 py-3.5">Type</th>
                  <th className="px-5 py-3.5 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-darkBorder text-slate-700 dark:text-slate-300 text-xs">
                {devices.map((d) => (
                  <tr
                    key={d.id}
                    onClick={() => handleOpenDeviceDetail(d)}
                    className={`hover:bg-slate-50/50 dark:hover:bg-slate-800/10 transition-all cursor-pointer ${
                      selectedDevice?.id === d.id ? "bg-indigo-50/30 dark:bg-indigo-950/10 border-l-4 border-l-indigo-600" : ""
                    }`}
                  >
                    <td className="px-5 py-3.5 font-semibold text-slate-800 dark:text-slate-200">{d.name}</td>
                    <td className="px-5 py-3.5">{getStatusBadge(d.status)}</td>
                    <td className="px-5 py-3.5 font-mono text-slate-500 dark:text-slate-400">{d.ip_address}</td>
                    <td className="px-5 py-3.5">
                      <div className="font-semibold text-slate-700 dark:text-slate-300">{d.mac_vendor || d.vendor || "Unknown"}</div>
                      <div className="text-[10px] text-slate-400 font-mono">{d.mac_address || "N/A"}</div>
                    </td>
                    <td className="px-5 py-3.5 capitalize">
                      <span className="px-2 py-0.5 bg-slate-100 dark:bg-slate-900 rounded text-[10px]">
                        {d.device_type}
                      </span>
                    </td>
                    <td className="px-5 py-3.5 text-right space-x-2" onClick={(e) => e.stopPropagation()}>
                      <button
                        onClick={() => handleOpenDeviceDetail(d)}
                        className="p-1 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400"
                        title="Quick View"
                      >
                        <Eye className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => handleOpenEditModal(d)}
                        className="p-1 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400"
                        title="Edit Settings"
                      >
                        <Edit className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => handleDelete(d.id)}
                        className="p-1 text-slate-400 hover:text-red-600 dark:hover:text-red-400"
                        title="Delete Device"
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

        {/* Detailed sidebar panel */}
        <div className="bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl p-5 shadow-sm min-h-[450px] flex flex-col">
          {!selectedDevice ? (
            <div className="flex flex-col items-center justify-center flex-1 text-slate-400 py-16 space-y-2">
              <Info className="h-8 w-8 text-slate-300" />
              <p className="text-xs">Select a device from the list to display hardware logs and live SNMP metrics.</p>
            </div>
          ) : detailLoading ? (
            <div className="flex flex-col items-center justify-center flex-1 py-16 space-y-3">
              <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
              <span className="text-slate-500 font-semibold text-xs">Querying telemetry details...</span>
            </div>
          ) : (
            <div className="space-y-5 flex-1 flex flex-col">
              {/* Device Header */}
              <div className="flex justify-between items-start border-b border-slate-100 dark:border-darkBorder pb-4">
                <div>
                  <h3 className="text-base font-bold text-slate-800 dark:text-slate-100">{selectedDevice.name}</h3>
                  <span className="text-xs font-mono text-slate-500">{selectedDevice.ip_address}</span>
                </div>
                <div className="flex flex-col items-end space-y-1">
                  {getStatusBadge(selectedDevice.status)}
                  <span className="text-[10px] text-slate-400">Available: {selectedDevice.availability_pct?.toFixed(1)}%</span>
                </div>
              </div>

              {/* Sub-tabs */}
              <div className="flex border-b border-slate-100 dark:border-darkBorder text-xs">
                <button
                  onClick={() => setDetailTab("general")}
                  className={`flex-1 pb-2 font-bold uppercase tracking-wider text-center ${
                    detailTab === "general" ? "border-b-2 border-indigo-600 text-indigo-600" : "text-slate-400"
                  }`}
                >
                  General
                </button>
                <button
                  onClick={() => setDetailTab("network")}
                  className={`flex-1 pb-2 font-bold uppercase tracking-wider text-center ${
                    detailTab === "network" ? "border-b-2 border-indigo-600 text-indigo-600" : "text-slate-400"
                  }`}
                >
                  Network & SNMP
                </button>
                <button
                  onClick={() => setDetailTab("stats")}
                  className={`flex-1 pb-2 font-bold uppercase tracking-wider text-center ${
                    detailTab === "stats" ? "border-b-2 border-indigo-600 text-indigo-600" : "text-slate-400"
                  }`}
                >
                  Activity
                </button>
              </div>

              {/* Tab Contents */}
              <div className="flex-1 text-xs text-slate-600 dark:text-slate-300 space-y-4">
                {detailTab === "general" && (
                  <div className="space-y-2.5">
                    <div><span className="font-semibold text-slate-400 block mb-0.5">Hostname:</span> {selectedDevice.hostname || "N/A"}</div>
                    <div><span className="font-semibold text-slate-400 block mb-0.5">MAC Address:</span> <code className="font-mono bg-slate-50 dark:bg-slate-900 px-1 py-0.5 rounded text-[11px]">{selectedDevice.mac_address || "N/A"}</code></div>
                    <div><span className="font-semibold text-slate-400 block mb-0.5">Hardware Manufacturer:</span> {selectedDevice.mac_vendor || selectedDevice.vendor || "Unknown"}</div>
                    <div><span className="font-semibold text-slate-400 block mb-0.5">Operating System:</span> {selectedDevice.os || "Unknown"}</div>
                    <div><span className="font-semibold text-slate-400 block mb-0.5">Physical Location:</span> {selectedDevice.location || "Default Rack / VLAN"}</div>
                    {selectedDevice.notes && (
                      <div className="bg-slate-50 dark:bg-slate-900 p-2 rounded-lg border border-slate-100 dark:border-darkBorder">
                        <span className="font-semibold text-slate-400 block mb-1">Administrative Notes:</span>
                        <p className="italic text-slate-500 leading-relaxed">{selectedDevice.notes}</p>
                      </div>
                    )}
                  </div>
                )}

                {detailTab === "network" && (
                  <div className="space-y-4">
                    {/* Open Ports */}
                    <div>
                      <span className="font-semibold text-slate-400 block mb-1.5">Open Services (TCP Ports)</span>
                      {devicePorts.length === 0 ? (
                        <span className="text-slate-400 italic">No open ports detected.</span>
                      ) : (
                        <div className="flex flex-wrap gap-1.5">
                          {devicePorts.map(p => (
                            <span key={p} className="px-2 py-0.5 font-mono text-[10px] bg-indigo-50 dark:bg-indigo-950/30 text-indigo-600 dark:text-indigo-400 border border-indigo-100 dark:border-indigo-950/60 rounded">
                              Port {p}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>

                    {/* SNMP info */}
                    {selectedDevice.snmp_enabled ? (
                      <div className="border-t border-slate-100 dark:border-darkBorder pt-3 space-y-2.5">
                        <span className="font-bold text-slate-800 dark:text-slate-200 block">SNMP Telemetry Data</span>
                        <div><span className="text-slate-400">SysName:</span> {selectedDevice.snmp_sys_name}</div>
                        <div className="max-h-12 overflow-y-auto"><span className="text-slate-400">Description:</span> {selectedDevice.snmp_sys_descr}</div>
                        <div><span className="text-slate-400">Uptime:</span> {Math.floor(selectedDevice.snmp_sys_uptime / 8640000)} days ({selectedDevice.snmp_sys_uptime} ticks)</div>

                        {/* CPU / RAM Usage */}
                        <div className="grid grid-cols-2 gap-3 pt-2">
                          <div className="bg-slate-50 dark:bg-slate-900/60 p-2.5 rounded border border-slate-150 dark:border-darkBorder">
                            <span className="text-[10px] text-slate-400 font-bold block">SNMP CPU Usage</span>
                            <span className="text-base font-bold text-slate-800 dark:text-slate-100">{selectedDevice.snmp_cpu_usage?.toFixed(1)}%</span>
                          </div>
                          <div className="bg-slate-50 dark:bg-slate-900/60 p-2.5 rounded border border-slate-150 dark:border-darkBorder">
                            <span className="text-[10px] text-slate-400 font-bold block">SNMP RAM Usage</span>
                            <span className="text-base font-bold text-slate-800 dark:text-slate-100">{selectedDevice.snmp_ram_usage?.toFixed(1)}%</span>
                          </div>
                        </div>

                        {/* SNMP Interfaces list */}
                        {deviceInterfaces.length > 0 && (
                          <div className="pt-2">
                            <span className="font-semibold text-slate-400 block mb-1">SNMP Interface Traffic Logs</span>
                            <div className="space-y-1.5 max-h-40 overflow-y-auto">
                              {deviceInterfaces.map((iface, index) => (
                                <div key={index} className="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-900/40 rounded border border-slate-100 dark:border-darkBorder font-mono text-[10px]">
                                  <div>
                                    <div className="font-bold text-slate-700 dark:text-slate-300">{iface.name}</div>
                                    <div className="text-[9px] text-slate-400">Speed: {iface.speed_mbps} Mbps</div>
                                  </div>
                                  <div className="text-right">
                                    <div className="text-green-600">In: {iface.in_traffic_mbps?.toFixed(1)}M</div>
                                    <div className="text-indigo-600">Out: {iface.out_traffic_mbps?.toFixed(1)}M</div>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        )}
                      </div>
                    ) : (
                      <div className="border-t border-slate-100 dark:border-darkBorder pt-3 text-slate-400 text-center py-4">
                        SNMP Daemon port 161 is inactive or unreachable on this device.
                      </div>
                    )}
                  </div>
                )}

                {detailTab === "stats" && (
                  <div className="space-y-4 flex flex-col h-full">
                    <div className="grid grid-cols-2 gap-3">
                      <div className="bg-slate-50 dark:bg-slate-900/60 p-2.5 rounded border border-slate-150 dark:border-darkBorder">
                        <span className="text-[10px] text-slate-400 font-bold block">First Discovered</span>
                        <span className="text-[10px] font-mono text-slate-800 dark:text-slate-200">{new Date(selectedDevice.first_seen).toLocaleDateString()}</span>
                      </div>
                      <div className="bg-slate-50 dark:bg-slate-900/60 p-2.5 rounded border border-slate-150 dark:border-darkBorder">
                        <span className="text-[10px] text-slate-400 font-bold block">Uptime count</span>
                        <span className="text-[10px] font-mono text-slate-800 dark:text-slate-200">{selectedDevice.number_of_scans} scans</span>
                      </div>
                    </div>

                    {/* Timeline feed */}
                    <div className="flex-1 flex flex-col">
                      <span className="font-semibold text-slate-400 block mb-2">Device Event Log Timeline</span>
                      {deviceTimeline.length === 0 ? (
                        <span className="text-slate-450 italic">No events logged for this device.</span>
                      ) : (
                        <div className="space-y-2 overflow-y-auto max-h-72 flex-1 pr-1">
                          {deviceTimeline.map((log) => (
                            <div key={log.id} className="p-2.5 bg-slate-50 dark:bg-slate-900/40 rounded border border-slate-100 dark:border-darkBorder relative pl-4">
                              <span className={`absolute left-1.5 top-3.5 w-1.5 h-1.5 rounded-full ${log.event_type === "online" ? "bg-green-500" : log.event_type === "join" ? "bg-indigo-500" : "bg-red-500"}`}></span>
                              <div className="flex justify-between items-center text-[10px] text-slate-400 border-b border-slate-100 dark:border-darkBorder/40 pb-1 mb-1 font-mono">
                                <span className="uppercase font-bold">{log.event_type}</span>
                                <span>{new Date(log.checked_at).toLocaleTimeString()}</span>
                              </div>
                              <p className="text-[11px] text-slate-600 dark:text-slate-300 font-sans leading-relaxed">{log.message}</p>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
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
                    <option value="pc">Workstation</option>
                    <option value="server">Server</option>
                    <option value="router">Router</option>
                    <option value="switch">Switch</option>
                    <option value="printer">Printer</option>
                    <option value="phone">Phone</option>
                    <option value="iot">IoT Device</option>
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
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-400 mb-1">Vendor</label>
                  <input
                    type="text"
                    value={formData.vendor}
                    onChange={(e) => setFormData({ ...formData, vendor: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
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
                      .filter((d) => d.id !== editingDevice?.id)
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
                  className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm text-slate-800 dark:text-slate-100 focus:outline-none"
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
