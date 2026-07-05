# AETHER NOC - Enterprise-Grade Network Monitoring Dashboard

AETHER NOC is a production-ready, enterprise-grade Network Monitoring Dashboard built with a Go backend (following Clean Architecture principles) and a React TypeScript frontend. It is designed to scale across multiple organizations (Multi-Tenant) while providing automated network discovery, hierarchical topology maps, and intelligent alert suppression.

---

## Technical Stack

* **Backend**: Golang (1.24+), Gin Web Framework, GORM ORM, PostgreSQL database, Redis session/sliding-window cache, Gorilla WebSocket, Prometheus Exporter.
* **Frontend**: React, TypeScript, Vite, Tailwind CSS, Recharts, Axios, Lucide Icons.
* **Deployment**: Docker, Docker Compose, Kubernetes StatefulSets/Deployments, Ingress Routing.

---

## Clean Architecture Directory Structure

```
cmd/
    server/             # Main executable bootstrapper
internal/
    api/                # HTTP REST Controllers & Gin routing
    service/            # Business logic orchestration (Auth, Devices, Alerts, Discovery, Dashboard)
    repository/         # GORM postgres queries & Redis handlers
    model/              # Database entity models
    middleware/         # JWT, Tenancy partitioning, Rate Limiting, Audit logs
    websocket/          # Partitioned WebSocket Hub rooms
    monitor/            # Background engine workers, Tickers, and parent-dependency analyzers
pkg/
    logger/             # Structured JSON log wrapper (log/slog)
configs/                # Config models & env loader
migrations/             # raw SQL migrations down/up
docker/                 # Dockerfiles for containers
k8s/                    # Kubernetes Deployments, Services, ConfigMaps, Ingress
```

---

## REST API Endpoints Reference

### Authentication (Public)
* `POST /api/v1/auth/register`: Create user account.
* `POST /api/v1/auth/login`: Authenticate and return Access & Refresh tokens.
* `POST /api/v1/auth/refresh`: Re-issue access tokens using a valid refresh token.
* `POST /api/v1/auth/logout`: Invalidate refresh token sessions.

### Devices (Protected)
* `GET /api/v1/devices`: Fetch all devices.
* `POST /api/v1/devices`: Create a new device.
* `PUT /api/v1/devices/:id`: Update device metadata/uplinks.
* `DELETE /api/v1/devices/:id`: Remove device.
* `GET /api/v1/devices/:id/history`: Fetch recent latency/packet-loss records.
* `POST /api/v1/devices/import`: Upload CSV device sheet.
* `GET /api/v1/devices/export`: Download current device inventory sheet.

### Incidents & Rules (Protected)
* `GET /api/v1/alerts`: Fetch active/resolved alert logs.
* `PUT /api/v1/alerts/:id/resolve`: Manually acknowledge/resolve an alert.
* `GET /api/v1/rules`: Fetch alert threshold rules.
* `POST /api/v1/rules`: Create custom alarm metric rules.
* `DELETE /api/v1/rules/:id`: Delete rule.

### Scanner & Network Discovery (Protected)
* `POST /api/v1/discovery/scan`: Sweep CIDR range (e.g. `192.168.1.0/24`) and return live hosts.

### Telemetry & Stats (Protected)
* `GET /api/v1/dashboard/stats`: Returns count of online/offline status ratios and telemetry averages.
* `GET /api/v1/dashboard/latency`: Returns aggregated latency historical lines.

### System Telemetry (Scraper)
* `GET /metrics`: Exposed Prometheus endpoint.

---

## Enterprise Core Features

### 1. Multi-Tenant Partitioning
All database entities partition under an `organization_id`. WebSocket connection client hubs isolate updates to respective organizational rooms. The authentication middleware validates API keys or JWT tokens to enforce strict tenant scoping.

### 2. Automated Network Discovery
Input any subnet range (e.g. `192.168.1.0/24`). The scanner initiates unprivileged ping checks and port-knocks common ports (SSH, HTTP, SMB, RDP) to detect active IPs, fingerprint Operating Systems, resolve hostnames, and allows clicking "Enroll" to auto-convert them into monitored devices.

### 3. Hierarchical Alert Suppression
Devices support a `parent_id` (parent switch/router). If a parent device is flagged `offline`, alert triggers for all child devices are suppressed, updating their status to `unreachable` instead of firing spam alerts.

### 4. Self-Monitoring & Telemetry
Integrates a Prometheus metrics scraper on `/metrics` to expose system metrics, active alerts, and connection stats.

---

## Running Locally

### Prerequisites
* Go 1.24+
* Node.js 18+ & npm
* PostgreSQL & Redis instances

### Step 1: Clone and Set Environment
Copy the `.env.example` in the root:
```bash
cp .env.example .env
```

### Step 2: Boot Backend
```bash
# Run database migrations and start server
go run cmd/server/main.go
```
*Port opens on http://localhost:8080. Seed admin user: `admin`/`admin123`.*

### Step 3: Boot Frontend
```bash
cd frontend
npm install
npm run dev
```
*Opens on http://localhost:5173 (redirects to http://localhost:3000 in Nginx).*

---

## Production Deployment (Docker Compose)

The easiest way to boot the full production stack is using Docker Compose:
```bash
docker-compose up --build
```
This launches:
1. `postgres`: Persistent database volume.
2. `redis`: Core key-value session cache.
3. `backend`: REST server running Go checks.
4. `frontend`: Compiled SPA React app hosted in Nginx.

---

## Kubernetes Orchestration

Deployments manifests reside in `/k8s`. To install:
```bash
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/backend.yaml
kubectl apply -f k8s/frontend.yaml
kubectl apply -f k8s/ingress.yaml
```
