# ⎈ Multi-Cluster Kubernetes Dashboard

<div align="center">

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-client--go-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![HTMX](https://img.shields.io/badge/HTMX-1.9-36C?style=for-the-badge&logo=html5&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-GORM-003B57?style=for-the-badge&logo=sqlite&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**A zero-config, real-time Kubernetes dashboard that automatically discovers and monitors all your local Kind/Minikube clusters.**

No Prometheus. No Helm. No YAML config. Just run it and go.

[Quick Start](#-quick-start) · [Features](#-features) · [Architecture](#-architecture) · [API Docs](#-api-endpoints) · [Tech Stack](#-tech-stack)

</div>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🔍 **Zero-Config Auto-Discovery** | Watches `~/.kube/config` via `fsnotify` — new clusters appear the moment you create them |
| 📦 **Auto-Install Metrics Server** | Detects missing `metrics-server` and deploys it to Kind/Minikube clusters automatically |
| 📊 **Live CPU & Memory Metrics** | Real-time utilization via Kubernetes Metrics API — no Prometheus required |
| ⚡ **HTMX Live Polling** | Dashboard auto-refreshes every 5 seconds without page reloads |
| 🚫 **No Dead Clusters** | Only online, reachable clusters are shown — offline contexts are silently hidden |
| 📈 **Historical Charts** | 24-hour CPU/Memory history stored in SQLite and rendered with Chart.js |
| 🚨 **Alert System** | Automatic alerts for pod failures and high resource usage |
| 🎨 **Glassmorphism UI** | Premium dark-theme dashboard with frosted glass panels and micro-animations |

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    Go Dashboard Server (Gin)                     │
├──────────┬─────────────────────────────┬────────────────────────┤
│ Handlers │  ClusterRegistry (fsnotify) │  MetricsStore          │
│ (API +   │  ├─ Watches ~/.kube/config  │  (SQLite / GORM)       │
│  Pages)  │  ├─ Auto-discovers contexts │                        │
│          │  └─ Auto-installs metrics   │                        │
├──────────┴─────────────────────────────┴────────────────────────┤
│              HTMX + Alpine.js + Chart.js Frontend               │
└──────────────────────────────────────────────────────────────────┘
         │                  │                    │
    ┌────▼────┐       ┌────▼────┐          ┌────▼────┐
    │  Kind   │       │  Kind   │          │Minikube │
    │ Cluster │       │ Cluster │          │ Cluster │
    │  (auto) │       │  (auto) │          │  (auto) │
    └─────────┘       └─────────┘          └─────────┘
```

**How it works:**
1. `ClusterRegistry` reads all contexts from `~/.kube/config` on startup
2. Each context is pinged to check if the cluster is actually reachable
3. If a reachable cluster has no `metrics-server`, the dashboard installs it automatically
4. `fsnotify` watches for kubeconfig changes — spin up a new cluster and it appears in ~2 seconds
5. Only healthy, reachable clusters are shown in the UI

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **kubectl** configured with at least one cluster context
- **Kind** or **Minikube** installed locally

### Run

```bash
git clone https://github.com/shivansh-gohem/multi-cluster-dashboard.git
cd multi-cluster-dashboard
go mod tidy
go run cmd/server/main.go
```

Open **http://localhost:8080** in your browser.

That's it. No config files to edit.

---

### 🧪 Test Auto-Discovery

```bash
# Open the dashboard at http://localhost:8080, then in a new terminal:

# Create a new cluster — it appears on the dashboard automatically within 2 seconds
kind create cluster --name demo

# Delete it — it disappears automatically
kind delete cluster --name demo
```

---

## 📁 Project Structure

```
multi-cluster-dashboard/
├── cmd/server/main.go              # Server entrypoint & metrics collector
├── internal/
│   ├── handlers/
│   │   ├── api.go                  # REST API handlers (clusters, nodes, pods, alerts)
│   │   └── pages.go                # HTML page handlers
│   ├── services/
│   │   └── autodiscover.go         # ClusterRegistry, fsnotify watcher, auto-install logic
│   ├── models/                     # Data structures
│   └── store/                      # SQLite metrics storage via GORM
├── templates/
│   ├── dashboard.html              # Main overview page
│   ├── cluster_detail.html         # Per-cluster detail (nodes, pods, charts)
│   └── alerts.html                 # Alerts page
├── static/css/styles.css           # Glassmorphism dark theme
├── k8s-configs/clusters.yaml       # Optional: display name overrides only
└── README.md
```

---

## 🔌 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/clusters` | GET | All online clusters with health status, CPU %, memory %, node and pod counts |
| `/api/clusters/:name` | GET | Single cluster full details |
| `/api/clusters/:name/nodes` | GET | Node list with status, roles, version, and age |
| `/api/clusters/:name/pods` | GET | Pod list with namespace, status, restarts, node, and age |
| `/api/clusters/:name/history` | GET | 24-hour CPU/Memory snapshots for charts |
| `/api/alerts` | GET | Active alerts — unreachable clusters, failed pods, high resource usage |

---

## 📊 How Metrics Work (No Prometheus Needed)

This dashboard uses the native **Kubernetes Metrics API** instead of Prometheus:

1. When a cluster is discovered, the backend checks for a `metrics-server` deployment in `kube-system`
2. If missing (common on fresh Kind clusters), it automatically applies the official manifest and patches it with `--kubelet-insecure-tls` for Kind/Minikube TLS compatibility
3. Once `metrics-server` is ready (~60 seconds), the dashboard queries `NodeMetricses` to calculate:

```
CPU %    = (used millicores    / allocatable millicores)    × 100
Memory % = (used bytes         / allocatable bytes)         × 100
```

> **No Prometheus, no Helm charts, no port-forwarding required.**

---

## ⚙️ Configuration (Optional)

The dashboard works out of the box with **zero configuration**. Optionally, create `k8s-configs/clusters.yaml` to override display names:

```yaml
clusters:
  - context: "kind-production"
    displayName: "🏭 Production Cluster"
  - context: "kind-staging"
    displayName: "🧪 Staging Environment"
```

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `KUBECONFIG` | `~/.kube/config` | Path to kubeconfig file |
| `CLUSTER_CONFIG` | `k8s-configs/clusters.yaml` | Path to optional display name overrides |
| `PORT` | `8080` | Dashboard server port |

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.21, Gin Web Framework |
| **Frontend** | HTMX 1.9, Alpine.js 3.x, Chart.js 4.4 |
| **Styling** | Vanilla CSS, Glassmorphism, Google Fonts (Outfit) |
| **Kubernetes** | `client-go`, `k8s.io/metrics` (Metrics API) |
| **Storage** | SQLite via GORM |
| **File Watching** | `fsnotify` |

---

## 🔨 Development

```bash
# Run all tests
go test ./... -v

# Build binary
go build -o dashboard cmd/server/main.go

# Run binary
./dashboard
```

---

## 🚨 Alert Thresholds

| Condition | Severity |
|-----------|----------|
| CPU usage > 80% | ⚠️ Warning |
| CPU usage > 95% | 🔴 Critical |
| Memory usage > 80% | ⚠️ Warning |
| Memory usage > 95% | 🔴 Critical |
| Any pod in Failed state | ⚠️ Warning |
| Cluster unreachable | 🔴 Critical |

---

## 🤝 Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you'd like to change.

---

## 📄 License

[MIT](LICENSE)

---

<div align="center">
Built with ❤️ using Go · HTMX · client-go · Chart.js
</div>
