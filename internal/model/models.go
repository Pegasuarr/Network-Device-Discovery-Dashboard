package model

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Role struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
}

type User struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Username       string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Email          string       `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash   string       `gorm:"type:varchar(255);not null" json:"-"`
	RoleID         uint         `gorm:"not null" json:"role_id"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"-"`
	Role           Role         `gorm:"foreignKey:RoleID" json:"role"`
}

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string     `gorm:"type:varchar(500);uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

type DeviceGroup struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
}

type Device struct {
	ID                 uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID     uuid.UUID    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name               string       `gorm:"type:varchar(100);not null" json:"name"`
	Hostname           string       `gorm:"type:varchar(255);not null" json:"hostname"`
	IPAddress          string       `gorm:"type:varchar(45);not null" json:"ip_address"`
	MACAddress         string       `gorm:"type:varchar(17)" json:"mac_address"`
	MACVendor          string       `gorm:"type:varchar(100)" json:"mac_vendor"`
	DeviceType         string       `gorm:"type:varchar(50);not null" json:"device_type"` // e.g. server, router, switch, printer, pc, phone, iot
	OS                 string       `gorm:"type:varchar(100)" json:"os"`
	Vendor             string       `gorm:"type:varchar(100)" json:"vendor"` // Generic vendor (or from SNMP)
	Location           string       `gorm:"type:varchar(255)" json:"location"`
	Status             string       `gorm:"type:varchar(20);default:'offline'" json:"status"` // online, offline
	MonitoringInterval int          `gorm:"default:60" json:"monitoring_interval"`            // in seconds
	Tags               string       `gorm:"type:text" json:"tags"`                            // comma separated or JSON string
	Notes              string       `gorm:"type:text" json:"notes"`
	Enabled            bool         `gorm:"default:true" json:"enabled"`
	GroupID            *uuid.UUID   `gorm:"type:uuid;index" json:"group_id"`
	ParentID           *uuid.UUID   `gorm:"type:uuid;index" json:"parent_id"` // For dependency suppression/topology mapping
	MaintenanceStart   *time.Time   `json:"maintenance_start"`
	MaintenanceEnd     *time.Time   `json:"maintenance_end"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`

	// Scanner & History fields
	FirstSeen          time.Time    `json:"first_seen"`
	LastSeen           time.Time    `json:"last_seen"`
	TotalOnlineTime    int64        `json:"total_online_time"` // in seconds
	NumberOfScans      int          `json:"number_of_scans"`
	AvailabilityPct    float64      `gorm:"type:decimal(5,2)" json:"availability_pct"`
	OpenPorts          string       `gorm:"type:text" json:"open_ports"` // JSON array of ports (e.g. [22, 80, 443])

	// SNMP Telemetry
	SNMPEnabled        bool         `gorm:"default:false" json:"snmp_enabled"`
	SNMPSysName        string       `gorm:"type:varchar(255)" json:"snmp_sys_name"`
	SNMPSysDescr       string       `gorm:"type:text" json:"snmp_sys_descr"`
	SNMPSysUptime      uint32       `json:"snmp_sys_uptime"`
	SNMPCPUUsage       float64      `gorm:"type:decimal(5,2)" json:"snmp_cpu_usage"`
	SNMPRAMUsage       float64      `gorm:"type:decimal(5,2)" json:"snmp_ram_usage"`
	SNMPInterfaces     string       `gorm:"type:text" json:"snmp_interfaces"` // JSON array of interface telemetry

	Group              *DeviceGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:SET NULL" json:"group,omitempty"`
}

type AlertRule struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`
	DeviceID       *uuid.UUID `gorm:"type:uuid;index" json:"device_id"` // Null means global rule
	Metric         string    `gorm:"type:varchar(50);not null" json:"metric"` // latency_ms, packet_loss, response_time, status, cpu, ram, disk, ssl_days
	Operator       string    `gorm:"type:varchar(5);not null" json:"operator"` // >, <, ==, !=
	Value          float64   `gorm:"type:decimal(10,2);not null" json:"value"`
	Duration       int       `gorm:"default:0" json:"duration"` // threshold breached duration in seconds
	Level          string    `gorm:"type:varchar(20);not null" json:"level"` // info, warning, critical
	Enabled        bool      `gorm:"default:true" json:"enabled"`
}

type MonitoringResult struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeviceID         uuid.UUID `gorm:"type:uuid;not null;index" json:"device_id"`
	LatencyMS        float64   `gorm:"type:decimal(10,2)" json:"latency_ms"`
	PacketLossPct    float64   `gorm:"type:decimal(5,2)" json:"packet_loss_pct"`
	ResponseTimeMS   float64   `gorm:"type:decimal(10,2)" json:"response_time_ms"`
	HTTPStatus       int       `json:"http_status"`
	SSLDaysRemaining int       `json:"ssl_days_remaining"`
	DNSResolved      bool      `json:"dns_resolved"`
	CPUUsage         float64   `gorm:"type:decimal(5,2)" json:"cpu_usage"`
	RAMUsage         float64   `gorm:"type:decimal(5,2)" json:"ram_usage"`
	DiskUsage        float64   `gorm:"type:decimal(5,2)" json:"disk_usage"`
	CheckedAt        time.Time `gorm:"index" json:"checked_at"`
}

type Alert struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	DeviceID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"device_id"`
	RuleID         *uuid.UUID `gorm:"type:uuid;index" json:"rule_id"`
	Type           string     `gorm:"type:varchar(50);not null" json:"type"`
	Message        string     `gorm:"type:text;not null" json:"message"`
	Level          string     `gorm:"type:varchar(20);not null" json:"level"` // info, warning, critical
	Status         string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, resolved
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	Device         Device     `gorm:"foreignKey:DeviceID" json:"device"`
}

type NotificationLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AlertID      uuid.UUID `gorm:"type:uuid;not null;index" json:"alert_id"`
	Channel      string    `gorm:"type:varchar(20);not null" json:"channel"` // email, telegram, slack, discord
	Status       string    `gorm:"type:varchar(20);not null" json:"status"` // sent, failed
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	SentAt       time.Time `json:"sent_at"`
}

type AuditLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Username       string    `gorm:"type:varchar(100);not null" json:"username"`
	Action         string    `gorm:"type:varchar(100);not null" json:"action"`
	ResourceType   string    `gorm:"type:varchar(50);not null" json:"resource_type"`
	ResourceID     string    `gorm:"type:varchar(100)" json:"resource_id"`
	Payload        string    `gorm:"type:text" json:"payload"`
	IPAddress      string    `gorm:"type:varchar(45)" json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

type LoginLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Username  string    `gorm:"type:varchar(100);not null" json:"username"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	Status    string    `gorm:"type:varchar(20);not null" json:"status"` // success, failed
	CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
	OrganizationID uuid.UUID `gorm:"type:uuid;primaryKey" json:"organization_id"`
	Key            string    `gorm:"type:varchar(100);primaryKey" json:"key"`
	Value          string    `gorm:"type:text" json:"value"`
	Group          string    `gorm:"type:varchar(50)" json:"group"` // smtp, telegram, general, custom
}

type DeviceTimeline struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeviceID  uuid.UUID `gorm:"type:uuid;not null;index" json:"device_id"`
	EventType string    `gorm:"type:varchar(50);not null" json:"event_type"` // online, offline, ip_change, hostname_change
	Message   string    `gorm:"type:text;not null" json:"message"`
	CheckedAt time.Time `gorm:"index" json:"checked_at"`
}

type ScanHistory struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	Target         string     `gorm:"type:varchar(255);not null" json:"target"`
	ScanProfile    string     `gorm:"type:varchar(50);not null" json:"scan_profile"`
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at"`
	Status         string     `gorm:"type:varchar(20);default:'running'" json:"status"` // running, completed, cancelled, failed
	DevicesFound   int        `json:"devices_found"`
	DurationMS     int64      `json:"duration_ms"`
	ScanType       string     `gorm:"type:varchar(20);default:'manual'" json:"scan_type"` // manual, scheduled
}

type ScanSchedule struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`
	Target         string    `gorm:"type:varchar(255);not null" json:"target"`
	CronExpression string    `gorm:"type:varchar(100);not null" json:"cron_expression"`
	ScanProfile    string    `gorm:"type:varchar(50);not null" json:"scan_profile"` // quick, deep, portscan, ping, custom
	Enabled        bool      `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
