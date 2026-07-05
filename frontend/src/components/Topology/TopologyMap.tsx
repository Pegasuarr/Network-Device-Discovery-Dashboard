import React, { useState, useEffect, useRef } from "react";
import type { Device } from "../../types";
import { ZoomIn, ZoomOut, RotateCcw, Server, Router, Network, Laptop } from "lucide-react";

interface TopologyMapProps {
  devices: Device[];
}

interface PosNode {
  device: Device;
  x: number;
  y: number;
}

interface PosLink {
  fromX: number;
  fromY: number;
  toX: number;
  toY: number;
  status: string;
}

export const TopologyMap: React.FC<TopologyMapProps> = ({ devices }) => {
  const [nodes, setNodes] = useState<PosNode[]>([]);
  const [links, setLinks] = useState<PosLink[]>([]);
  const [transform, setTransform] = useState({ x: 0, y: 0, scale: 1 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [selectedNode, setSelectedNode] = useState<Device | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);

  // Compute Layered Tree Layout positions
  useEffect(() => {
    if (devices.length === 0) return;

    const width = 800;
    const computedNodes: PosNode[] = [];
    const computedLinks: PosLink[] = [];

    // Group devices by parents
    const roots = devices.filter(d => !d.parent_id);
    const childrenMap: Record<string, Device[]> = {};
    devices.forEach(d => {
      if (d.parent_id) {
        if (!childrenMap[d.parent_id]) childrenMap[d.parent_id] = [];
        childrenMap[d.parent_id].push(d);
      }
    });

    // Layer 0: Root nodes
    const rootCount = roots.length;
    roots.forEach((root, idx) => {
      const rootX = (width / (rootCount + 1)) * (idx + 1);
      const rootY = 80;
      computedNodes.push({ device: root, x: rootX, y: rootY });

      // Layer 1: Children of Root
      const level1 = childrenMap[root.id] || [];
      const l1Count = level1.length;
      level1.forEach((child1, idx1) => {
        const c1X = rootX - 100 + (200 / (l1Count > 1 ? l1Count - 1 : 1)) * idx1;
        const c1Y = 220;
        computedNodes.push({ device: child1, x: c1X, y: c1Y });
        computedLinks.push({
          fromX: rootX,
          fromY: rootY,
          toX: c1X,
          toY: c1Y,
          status: child1.status,
        });

        // Layer 2: Children of Level 1
        const level2 = childrenMap[child1.id] || [];
        const l2Count = level2.length;
        level2.forEach((child2, idx2) => {
          const c2X = c1X - 50 + (100 / (l2Count > 1 ? l2Count - 1 : 1)) * idx2;
          const c2Y = 360;
          computedNodes.push({ device: child2, x: c2X, y: c2Y });
          computedLinks.push({
            fromX: c1X,
            fromY: c1Y,
            toX: c2X,
            toY: c2Y,
            status: child2.status,
          });
        });
      });
    });

    // Capture orphans (devices that have parent_id pointing to missing devices)
    devices.forEach(d => {
      if (d.parent_id && !computedNodes.some(n => n.device.id === d.id)) {
        // Render on root layer
        const index = computedNodes.length;
        computedNodes.push({ device: d, x: 50 + index * 120, y: 80 });
      }
    });

    setNodes(computedNodes);
    setLinks(computedLinks);
  }, [devices]);

  // Handle Dragging / Panning
  const handleMouseDown = (e: React.MouseEvent) => {
    if ((e.target as HTMLElement).tagName === "circle" || (e.target as HTMLElement).tagName === "text") {
      return; // click on node, handled elsewhere
    }
    setIsDragging(true);
    setDragStart({ x: e.clientX - transform.x, y: e.clientY - transform.y });
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging) return;
    setTransform(prev => ({
      ...prev,
      x: e.clientX - dragStart.x,
      y: e.clientY - dragStart.y,
    }));
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleZoom = (factor: number) => {
    setTransform(prev => ({
      ...prev,
      scale: Math.max(0.5, Math.min(2.5, prev.scale * factor)),
    }));
  };

  const handleReset = () => {
    setTransform({ x: 0, y: 0, scale: 1 });
    setSelectedNode(null);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "online": return "#10B981"; // green
      case "offline": return "#EF4444"; // red
      case "warning": return "#F59E0B"; // yellow
      case "unreachable": return "#94A3B8"; // grey
      case "maintenance": return "#8B5CF6"; // purple
      default: return "#64748B"; // slate
    }
  };

  const getNodeIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case "router": return <Router className="h-5 w-5 text-white" />;
      case "switch": return <Network className="h-5 w-5 text-white" />;
      case "server": return <Server className="h-5 w-5 text-white" />;
      default: return <Laptop className="h-5 w-5 text-white" />;
    }
  };

  return (
    <div className="relative border border-slate-200 dark:border-darkBorder bg-slate-100 dark:bg-slate-900/60 rounded-xl overflow-hidden shadow-inner h-[500px]">
      {/* Controls Overlay */}
      <div className="absolute top-4 left-4 z-10 flex space-x-2">
        <button onClick={() => handleZoom(1.1)} className="p-2 bg-white dark:bg-darkCard hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-darkBorder rounded-lg shadow-sm">
          <ZoomIn className="h-4 w-4" />
        </button>
        <button onClick={() => handleZoom(0.9)} className="p-2 bg-white dark:bg-darkCard hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-darkBorder rounded-lg shadow-sm">
          <ZoomOut className="h-4 w-4" />
        </button>
        <button onClick={handleReset} className="p-2 bg-white dark:bg-darkCard hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-darkBorder rounded-lg shadow-sm">
          <RotateCcw className="h-4 w-4" />
        </button>
      </div>

      {/* SVG Canvas */}
      <div
        ref={containerRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        className={`w-full h-full cursor-${isDragging ? "grabbing" : "grab"}`}
      >
        <svg className="w-full h-full">
          <g transform={`translate(${transform.x}, ${transform.y}) scale(${transform.scale})`}>
            {/* Draw Links */}
            {links.map((link, idx) => (
              <line
                key={`link-${idx}`}
                x1={link.fromX}
                y1={link.fromY}
                x2={link.toX}
                y2={link.toY}
                stroke={link.status === "offline" ? "#EF4444" : "#22D3EE"}
                strokeWidth={2}
                strokeDasharray={link.status === "unreachable" ? "4,4" : "0"}
                className="transition-all duration-300 opacity-60"
              />
            ))}

            {/* Draw Nodes */}
            {nodes.map((node) => {
              const color = getStatusColor(node.device.status);
              const isSelected = selectedNode?.id === node.device.id;
              return (
                <g
                  key={node.device.id}
                  transform={`translate(${node.x}, ${node.y})`}
                  onClick={() => setSelectedNode(node.device)}
                  className="cursor-pointer group"
                >
                  <circle
                    r={26}
                    fill={color}
                    stroke={isSelected ? "#FFF" : "transparent"}
                    strokeWidth={3}
                    className="transition-all duration-300 shadow-md group-hover:scale-110"
                  />
                  {/* Centralized Icon */}
                  <foreignObject x={-10} y={-10} width={20} height={20} className="pointer-events-none">
                    <div className="flex items-center justify-center">
                      {getNodeIcon(node.device.device_type)}
                    </div>
                  </foreignObject>
                  {/* Label */}
                  <text
                    y={42}
                    textAnchor="middle"
                    fill="currentColor"
                    className="text-xs font-semibold select-none text-slate-800 dark:text-slate-200"
                  >
                    {node.device.name}
                  </text>
                  <text
                    y={56}
                    textAnchor="middle"
                    fill="currentColor"
                    className="text-[10px] font-mono select-none text-slate-500 dark:text-slate-400"
                  >
                    {node.device.ip_address}
                  </text>
                </g>
              );
            })}
          </g>
        </svg>
      </div>

      {/* Info Card Drawer */}
      {selectedNode && (
        <div className="absolute right-4 bottom-4 w-72 p-5 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-xl z-10">
          <div className="flex justify-between items-start mb-3">
            <div>
              <h4 className="text-sm font-bold text-slate-800 dark:text-slate-200">{selectedNode.name}</h4>
              <span className="text-xs text-slate-500 font-mono">{selectedNode.ip_address}</span>
            </div>
            <span
              className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider text-white"
              style={{ backgroundColor: getStatusColor(selectedNode.status) }}
            >
              {selectedNode.status}
            </span>
          </div>

          <div className="space-y-1.5 text-xs text-slate-600 dark:text-slate-300 border-t border-slate-100 dark:border-darkBorder pt-3">
            <div><span className="font-semibold text-slate-400">Host:</span> {selectedNode.hostname}</div>
            <div><span className="font-semibold text-slate-400">Type:</span> <span className="capitalize">{selectedNode.device_type}</span></div>
            <div><span className="font-semibold text-slate-400">OS:</span> {selectedNode.os}</div>
            <div><span className="font-semibold text-slate-400">Location:</span> {selectedNode.location}</div>
            {selectedNode.parent_id && (
              <div><span className="font-semibold text-slate-400">Uplink parent:</span> Uplinked</div>
            )}
          </div>
          <button
            onClick={() => setSelectedNode(null)}
            className="w-full mt-4 py-1.5 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 font-semibold rounded-lg text-xs transition-all"
          >
            Close Details
          </button>
        </div>
      )}
    </div>
  );
};
