export interface Organization {
  id: string;
  name: string;
  created_at: string;
}

export interface Role {
  id: number;
  name: string;
  description: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  role_id: number;
  role: Role;
  created_at: string;
}

export interface DeviceGroup {
  id: string;
  name: string;
  description: string;
}

export interface Device {
  id: string;
  organization_id: string;
  name: string;
  hostname: string;
  ip_address: string;
  mac_address: string;
  device_type: "server" | "router" | "switch" | "printer" | "pc" | string;
  os: string;
  vendor: string;
  location: string;
  status: "online" | "offline" | "warning" | "unreachable" | "maintenance" | "unknown";
  monitoring_interval: number;
  tags: string;
  notes: string;
  enabled: boolean;
  group_id?: string;
  parent_id?: string;
  maintenance_start?: string;
  maintenance_end?: string;
  created_at: string;
  updated_at: string;
  group?: DeviceGroup;
}

export interface AlertRule {
  id: string;
  name: string;
  device_id?: string;
  metric: "latency_ms" | "packet_loss" | "response_time" | "status" | "cpu" | "ram" | "disk" | "ssl_days";
  operator: ">" | "<" | "==" | "!=";
  value: number;
  duration: number;
  level: "info" | "warning" | "critical";
  enabled: boolean;
}

export interface MonitoringResult {
  device_id: string;
  latency_ms: number;
  packet_loss_pct: number;
  response_time_ms: number;
  http_status: number;
  ssl_days: number;
  dns_resolved: boolean;
  cpu_usage: number;
  ram_usage: number;
  disk_usage: number;
  checked_at: string;
}

export interface Alert {
  id: string;
  device_id: string;
  rule_id?: string;
  type: string;
  message: string;
  level: "info" | "warning" | "critical";
  status: "active" | "resolved";
  created_at: string;
  resolved_at?: string;
  device: Device;
}

export interface AuditLog {
  id: string;
  username: string;
  action: string;
  resource_type: string;
  payload: string;
  ip_address: string;
  created_at: string;
}

export interface DashboardStats {
  total_devices: number;
  online_devices: number;
  offline_devices: number;
  warning_devices: number;
  unreachable_devices: number;
  active_alerts: number;
  avg_latency_ms: number;
  avg_packet_loss: number;
  avg_cpu: number;
  avg_ram: number;
  avg_disk: number;
  device_type_counts: Record<string, number>;
}
