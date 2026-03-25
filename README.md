# ⎈ Multi-Cluster Kubernetes Dashboard

A **zero-config**, real-time Kubernetes dashboard that automatically discovers and monitors all your local Kind/Minikube clusters. No Prometheus, no Helm, no YAML config — just run it and go.

> Built with Go · React · TailwindCSS · shadcn/ui · Recharts

---

## ✨ Key Features

| Feature | Description |
|---------|-------------|
| 🔍 **Zero-Config Auto-Discovery** | Watches `~/.kube/config` via `fsnotify` — new clusters appear automatically |
| 📦 **Auto-Install Metrics Server** | Detects missing `metrics-server` and deploys it to Kind/Minikube clusters automatically |
| 📊 **Live CPU & Memory Metrics** | Real-time utilization via Kubernetes Metrics API (`k8s.io/metrics`) |
| 🧊 **Premium React UI** | shadcn/ui components with TailwindCSS, dark theme, glassmorphism panels, Recharts |
| ⚡ **React Query Polling** | Dashboard auto-refreshes every 5 seconds via React Query |
| 🚫 **No Dead Clusters** | Only online, reachable clusters are shown — offline contexts are hidden automatically |
| 📈 **Historical Charts** | 24-hour CPU/Memory history stored in SQLite, rendered with Recharts |
| 🚨 **Alert System** | Automatic alerts for pod failures and high resource usage |

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│              React Frontend (Vite + shadcn/ui + Recharts)        │
│              Served on :3000, proxies /api → :8080               │
├──────────────────────────────────────────────────────────────────┤
│                    Go Backend Server (Gin)                        │
├──────────┬──────────────┬─────────────────┬─────────────────────┤
│ Handlers │ ClusterRegistry (fsnotify)     │  MetricsStore       │
│ (REST    │  ├─ Watches ~/.kube/config     │  (SQLite/GORM)      │
│  API)    │  ├─ Auto-discovers contexts    │                     │
│          │  └─ Auto-installs metrics-svr  │                     │
└──────────┴──────────────┴─────────────────┴─────────────────────┘
         │                  │                    │
    ┌────▼────┐       ┌────▼────┐          ┌────▼────┐
    │ Kind    │       │ Kind    │          │ Minikube│
    │ Cluster │       │ Cluster │          │ Cluster │
    └─────────┘       └─────────┘          └─────────┘
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **Node.js 18+** and **npm**
- **kubectl** configured with at least one cluster context
- **Kind** or **Minikube** (for local clusters)

### Run

```bash
git clone https://github.com/your-username/multi-cluster-dashboard.git
cd multi-cluster-dashboard

# Terminal 1: Start Go backend
go mod tidy
go run cmd/server/main.go

# Terminal 2: Start React frontend
cd frontend
npm install
npm run dev
```

Open **http://localhost:3000** in your browser.

---

## 📁 Project Structure

```
multi-cluster-dashboard/
├── cmd/server/main.go              # Go server entrypoint
├── internal/
│   ├── handlers/api.go             # REST API (clusters, nodes, pods, alerts)
│   ├── services/autodiscover.go    # Cluster registry, fsnotify, auto-install
│   ├── models/                     # Data structures
│   └── store/                      # SQLite metrics storage
├── frontend/                       # React app (Vite + shadcn/ui)
│   ├── src/
│   │   ├── lib/api.ts              # Go API fetch client
│   │   ├── lib/hooks.ts            # React Query hooks (polling)
│   │   ├── pages/Index.tsx         # Main dashboard page
│   │   └── components/dashboard/   # Sidebar, Overview, Nodes, Pods, Alerts
│   └── vite.config.ts              # Vite config with API proxy
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
| `/api/alerts` | GET | Active alerts |

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.21, Gin |
| **Frontend** | React 18, Vite, TailwindCSS, shadcn/ui, Recharts |
| **Kubernetes** | client-go, k8s.io/metrics |
| **Storage** | SQLite (GORM) |
| **File Watching** | fsnotify |

---

## 📄 License

MIT
