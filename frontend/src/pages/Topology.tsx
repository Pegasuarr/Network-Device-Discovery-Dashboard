import React, { useEffect, useState } from "react";
import api from "../services/api";
import type { Device } from "../types";
import { TopologyMap } from "../components/Topology/TopologyMap";
import { Network, Loader2, Info } from "lucide-react";

export const Topology: React.FC = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDevices();
  }, []);

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

  return (
    <div className="space-y-6 pb-12">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100 flex items-center">
          <Network className="h-7 w-7 mr-2 text-indigo-500" />
          Network Topology Map
        </h1>
        <p className="text-sm text-slate-500 mt-1">
          Visual representation of local network device hierarchies, uplinks, and connection vectors.
        </p>
      </div>

      {/* Info Banner */}
      <div className="flex items-start p-4 bg-indigo-500/10 border border-indigo-500/20 text-indigo-700 dark:text-indigo-400 rounded-xl text-xs space-x-3">
        <Info className="h-5 w-5 flex-shrink-0" />
        <div>
          <p className="font-bold">Understanding the Network Map:</p>
          <p className="mt-1 leading-relaxed">
            The hierarchy is generated automatically based on parent uplink configurations. You can configure a device's parent device in the <strong>Devices</strong> section. Nodes represent discovered hosts colored by their active status (Green = Online, Red = Offline, Yellow = Warning, Purple = Maintenance). You can click on any node to view details, or drag and zoom the map canvas.
          </p>
        </div>
      </div>

      {/* Topology Canvas */}
      {loading ? (
        <div className="flex flex-col items-center justify-center border border-slate-200 dark:border-darkBorder bg-white dark:bg-darkCard rounded-xl shadow-sm h-[500px] space-y-3">
          <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
          <span className="text-slate-500 text-sm font-semibold">Computing topology tree...</span>
        </div>
      ) : devices.length === 0 ? (
        <div className="flex flex-col items-center justify-center border border-slate-200 dark:border-darkBorder bg-white dark:bg-darkCard rounded-xl shadow-sm h-[500px] text-slate-400 space-y-2">
          <Network className="h-10 w-10" />
          <p className="text-xs">No active devices registered in inventory to build topology.</p>
        </div>
      ) : (
        <TopologyMap devices={devices} />
      )}
    </div>
  );
};
