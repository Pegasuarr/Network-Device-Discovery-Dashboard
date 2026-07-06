#  AETHER NOC - Network Device Discovery Dashboard

AETHER NOC is a production-ready, enterprise-grade **Network Device Discovery Dashboard** built with a **Go backend** (following Clean Architecture principles) and a **React TypeScript frontend**.

---

###  Purpose & Core Capabilities

The primary purpose of this application is to map and monitor local and enterprise network infrastructures in real-time:

1. **Automated Network Device Discovery**:
   Allows you to sweep custom CIDR ranges (such as `17.172.224.47/30`) to scan your home Wi-Fi or office network. The dashboard detects all active devices (smartphones, TVs, laptops, routers), resolves their hostnames, fingerprints their operating systems, and identifies open TCP ports.

2. **Real-Time Availability & Latency Monitoring**:
   Periodically queries all enrolled devices using ICMP pings and SNMP checks to record status history, average latency, and packet loss, streaming updates instantly to the frontend using WebSockets.

3. **Hierarchical Alert Suppression**:
   Maps parent-child relationships (e.g., switches behind a router). If a parent device drops offline, alarms for all downstream child devices are automatically suppressed (marking them as `unreachable` instead of `offline`) to prevent alert storms and spam.

4. **Multi-Tenant Scoping**:
   Safely partitions devices, groups, alert rules, and audit logs by organization so that multiple tenants can use a single hosted dashboard system securely.

5. **Threshold Alarms & Notification Triggers**:
   Monitors system metrics against configurable alert rules (e.g., latency > 150ms) and dispatches instant alert notifications to **Slack, Discord, Telegram, or Email** channels.
