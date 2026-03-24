<![CDATA[# Project Synopsis

## Multi-Cluster Kubernetes Health Monitoring Dashboard

---

### 1. Introduction

As organizations increasingly adopt Kubernetes for container orchestration, the challenge of monitoring multiple clusters simultaneously has become a critical operational concern. Existing solutions like Prometheus + Grafana require significant infrastructure setup and are often too resource-intensive for local development environments.

This project presents a **lightweight, zero-configuration dashboard** that provides unified, real-time health monitoring across multiple Kubernetes clusters. It is specifically designed for developers and DevOps engineers who work with local cluster provisioners such as Kind and Minikube.

---

### 2. Problem Statement

- **Fragmented Visibility**: Developers running multiple local Kubernetes clusters lack a single pane of glass to monitor their health.
- **Manual Configuration Overhead**: Traditional monitoring stacks require manual installation of Prometheus, Grafana, and per-cluster configuration for each new cluster.
- **Resource Constraints**: Full observability stacks (Prometheus + Grafana) consume 500 MB+ RAM per cluster, which is impractical for local development.
- **Stale Dashboard Data**: Clusters created or destroyed during development sessions are not dynamically reflected in existing tools.

---

### 3. Proposed Solution

A Go-based web dashboard that:

1. **Automatically discovers** all Kubernetes clusters from the user's kubeconfig file using filesystem watching (`fsnotify`).
2. **Auto-provisions** the lightweight Kubernetes Metrics Server on clusters that lack it, eliminating manual setup.
3. **Streams real-time metrics** (CPU, Memory, Node count, Pod count) directly from the native Kubernetes Metrics API — no Prometheus required.
4. **Hides offline clusters** to maintain a clean, noise-free interface.
5. **Provides a premium UI** using modern web design patterns (Glassmorphism, micro-animations, responsive layouts).

---

### 4. System Architecture

```
                    ┌────────────────────────────┐
                    │      Web Browser (UI)       │
                    │  HTMX · Alpine.js · Chart.js│
                    └────────────┬───────────────┘
                                 │ HTTP (HTMX Polling)
                    ┌────────────▼───────────────┐
                    │     Go Backend (Gin)        │
                    │  ┌───────────────────────┐  │
                    │  │  Cluster Registry      │  │
                    │  │  (fsnotify watcher)    │  │
                    │  └──────────┬────────────┘  │
                    │  ┌──────────▼────────────┐  │
                    │  │  Metrics Collector     │  │
                    │  │  (k8s.io/metrics)      │  │
                    │  └──────────┬────────────┘  │
                    │  ┌──────────▼────────────┐  │
                    │  │  SQLite Store (GORM)   │  │
                    │  └───────────────────────┘  │
                    └──────┬─────┬─────┬─────────┘
                           │     │     │
                    ┌──────▼┐ ┌──▼──┐ ┌▼──────┐
                    │Cluster│ │Clstr│ │Cluster│
                    │  #1   │ │ #2  │ │  #N   │
                    └───────┘ └─────┘ └───────┘
```

---

### 5. Key Modules

| Module | Responsibility |
|--------|---------------|
| `ClusterRegistry` | Watches kubeconfig, discovers contexts, manages client connections |
| `ensureMetricsServer` | Auto-installs metrics-server on clusters missing it |
| `GetUtilization` | Queries Kubernetes Metrics API for real-time CPU/Memory |
| `MetricsStore` | Persists 24-hour snapshots in SQLite for historical charts |
| `APIHandler` | REST endpoints for clusters, nodes, pods, alerts |
| `PageHandler` | Server-side rendered HTML pages with HTMX interactivity |

---

### 6. Technologies Used

| Category | Technology |
|----------|------------|
| Language | Go 1.21 |
| Web Framework | Gin |
| Frontend | HTMX, Alpine.js, Chart.js |
| Kubernetes SDK | client-go, k8s.io/metrics |
| Database | SQLite (GORM ORM) |
| File Watching | fsnotify |
| Styling | CSS3 (Glassmorphism), Google Fonts |

---

### 7. Key Features

1. **Zero-Configuration Auto-Discovery** — No YAML or config files needed. The dashboard reads the kubeconfig and connects to all available clusters.
2. **Auto-Provisioning** — Automatically installs the Kubernetes Metrics Server on Kind/Minikube clusters, including TLS compatibility patches.
3. **Live Dashboard** — HTMX-powered polling refreshes cluster health every 5 seconds without full page reloads.
4. **Smart Filtering** — Only online, reachable clusters are displayed. Offline clusters are hidden to reduce noise.
5. **Historical Analytics** — 24-hour CPU and Memory trends stored in SQLite and rendered as interactive Chart.js graphs.
6. **Alert Engine** — Generates alerts for unreachable clusters, failed pods, and resource threshold breaches.
7. **Premium UI/UX** — Dark-mode glassmorphism design with animated gradients, neon glows, and the Outfit typeface.

---

### 8. Use Cases

- **Local Development**: Developers running multiple Kind clusters for microservice testing.
- **CI/CD Environments**: Quick health checks across ephemeral test clusters.
- **Learning & Education**: Students studying Kubernetes multi-cluster management.
- **DevOps Tooling**: Lightweight alternative to Prometheus + Grafana for non-production environments.

---

### 9. Future Scope

- **Namespace-level Filtering**: Allow monitoring specific namespaces within clusters.
- **Custom Alerts**: User-configurable alert thresholds and notification channels (Slack, Email).
- **Deployment Management**: Trigger rollouts and rollbacks directly from the dashboard.
- **Multi-User Auth**: Role-based access control for team environments.
- **Helm Chart**: Package the dashboard itself as a Helm chart for in-cluster deployment.
- **Cloud Cluster Support**: Extend auto-discovery to GKE, EKS, and AKS via cloud provider APIs.

---

### 10. Conclusion

This project demonstrates that effective Kubernetes multi-cluster monitoring does not require heavy infrastructure like Prometheus. By leveraging the native Kubernetes Metrics API, filesystem watching, and auto-provisioning, the dashboard achieves a truly zero-configuration experience. The modern, responsive UI ensures that developers can monitor their cluster fleet at a glance, while the auto-install mechanism eliminates the repetitive manual setup that plagues local Kubernetes development workflows.
]]>
