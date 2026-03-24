<![CDATA[# ⎈ Multi-Cluster Kubernetes Dashboard

A **zero-config**, real-time Kubernetes dashboard that automatically discovers and monitors all your local Kind/Minikube clusters. No Prometheus, no Helm, no YAML config — just run it and go.

> Built with Go · HTMX · Alpine.js · Chart.js · Glassmorphism UI

---

## ✨ Key Features

| Feature | Description |
|---------|-------------|
| 🔍 **Zero-Config Auto-Discovery** | Watches `~/.kube/config` via `fsnotify` — new clusters appear automatically |
| 📦 **Auto-Install Metrics Server** | Detects missing `metrics-server` and deploys it to Kind/Minikube clusters automatically |
| 📊 **Live CPU & Memory Metrics** | Real-time utilization via Kubernetes Metrics API (`k8s.io/metrics`) |
| 🧊 **Glassmorphism UI** | Premium dark-theme dashboard with frosted glass panels, gradient accents, and micro-animations |
| ⚡ **HTMX Live Polling** | Dashboard auto-refreshes every 5 seconds — no page reloads |
| 🚫 **No Dead Clusters** | Only online, reachable clusters are shown — offline contexts are hidden automatically |
| 📈 **Historical Charts** | 24-hour CPU/Memory history stored in SQLite, rendered with Chart.js |
| 🚨 **Alert System** | Automatic alerts for pod failures and high resource usage |

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    Go Dashboard Server (Gin)                     │
├──────────┬──────────────┬─────────────────┬─────────────────────┤
│ Handlers │ ClusterRegistry (fsnotify)     │  MetricsStore       │
│ (API +   │  ├─ Watches ~/.kube/config     │  (SQLite/GORM)      │
│  Pages)  │  ├─ Auto-discovers contexts    │                     │
│          │  └─ Auto-installs metrics-svr  │                     │
├──────────┴──────────────┴─────────────────┴─────────────────────┤
│              HTMX + Alpine.js + Chart.js Frontend               │
└──────────────────────────────────────────────────────────────────┘
         │                  │                    │
    ┌────▼────┐       ┌────▼────┐          ┌────▼────┐
    │ Kind    │       │ Kind    │          │ Minikube│
    │ Cluster │       │ Cluster │          │ Cluster │
    │ (auto)  │       │ (auto)  │          │ (auto)  │
    └─────────┘       └─────────┘          └─────────┘
```

**How it works:**
1. `ClusterRegistry` reads all contexts from your kubeconfig
2. For each context, it pings the API server to check reachability
3. If a cluster is online but has no `metrics-server`, it auto-installs one
4. The dashboard only displays reachable clusters with live metrics
5. `fsnotify` watches for kubeconfig changes — spin up a new cluster and it appears instantly

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **kubectl** configured with at least one cluster context
- **Kind** or **Minikube** (for local clusters)

### Run

```bash
git clone https://github.com/shivansh-gohem/multi-cluster-dashboard.git
cd multi-cluster-dashboard
go mod tidy
go run cmd/server/main.go
```

Open **http://localhost:8080** in your browser.

That's it. The dashboard will automatically discover all your kubeconfig contexts, connect to the online ones, install metrics-server where needed, and start showing live data.

### Test Auto-Discovery

```bash
# Create a new cluster — it appears on the dashboard automatically!
kind create cluster --name demo

# Delete it — it disappears automatically!
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
│   │   └── autodiscover.go         # Cluster registry, fsnotify, auto-install logic
│   ├── models/                     # Data structures
│   └── store/                      # SQLite metrics storage (GORM)
├── templates/
│   ├── dashboard.html              # Main overview page
│   ├── cluster_detail.html         # Per-cluster detail page (nodes, pods, charts)
│   └── alerts.html                 # Alerts page
├── static/css/styles.css           # Glassmorphism dark theme
├── k8s-configs/clusters.yaml       # Optional: display name overrides
└── README.md
```

---

## 🔌 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/clusters` | GET | All online clusters with health, CPU, memory, node/pod counts |
| `/api/clusters/:name` | GET | Single cluster details |
| `/api/clusters/:name/nodes` | GET | Node list with status, roles, version, age |
| `/api/clusters/:name/pods` | GET | Pod list with namespace, status, restarts, node, age |
| `/api/clusters/:name/history` | GET | 24h CPU/Memory history snapshots |
| `/api/alerts` | GET | Active alerts (unreachable clusters, failed pods) |

---

## ⚙️ Configuration (Optional)

The dashboard works with **zero configuration**. However, you can optionally create `k8s-configs/clusters.yaml` to override display names:

```yaml
clusters:
  - context: "kind-production"
    displayName: "🏭 Production Cluster"
  - context: "kind-staging"
    displayName: "🧪 Staging Environment"
```

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `CLUSTER_CONFIG` | `k8s-configs/clusters.yaml` | Path to optional YAML config |
| `KUBECONFIG` | `~/.kube/config` | Path to kubeconfig file |
| `PORT` | `8080` | Dashboard server port |

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.21, Gin Web Framework |
| **Frontend** | HTMX 1.9, Alpine.js 3.x, Chart.js 4.4 |
| **Styling** | Vanilla CSS, Glassmorphism, Google Fonts (Outfit) |
| **Kubernetes** | client-go, k8s.io/metrics (Metrics API) |
| **Storage** | SQLite via GORM |
| **File Watching** | fsnotify |

---

## 📊 How Metrics Work

This dashboard does **NOT** require Prometheus. Instead, it uses the native **Kubernetes Metrics API**:

1. When a cluster is discovered, the backend checks for a `metrics-server` deployment in `kube-system`
2. If missing (common on fresh Kind clusters), it automatically runs:
   - `kubectl apply -f` the official metrics-server manifest
   - Patches the deployment with `--kubelet-insecure-tls` for Kind/Minikube compatibility
3. Once metrics-server is ready (~60s), the dashboard queries `NodeMetricses` to calculate:
   - **CPU %** = `(used millicores / allocatable millicores) × 100`
   - **Memory %** = `(used bytes / allocatable bytes) × 100`

---

## 🔨 Development

```bash
# Run tests
go test ./... -v

# Build binary
go build -o dashboard cmd/server/main.go

# Run binary
./dashboard
```

---

## 📄 License

MIT
]]>
