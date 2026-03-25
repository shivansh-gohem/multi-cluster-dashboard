// API client — fetches live data from the Go backend

const API_BASE = '/api';

// ===== Types (matching Go API responses) =====

export type ClusterStatus = 'healthy' | 'warning' | 'critical' | 'offline';
export type PodStatus = 'Running' | 'Pending' | 'Failed' | 'CrashLoopBackOff' | 'Succeeded';
export type NodeStatus = 'Ready' | 'NotReady';
export type AlertSeverity = 'warning' | 'critical';

export interface Cluster {
  name: string;
  displayName: string;
  status: ClusterStatus;
  nodeCount: number;
  podCount: number;
  cpuUsage: number;
  memoryUsage: number;
  reachable: boolean;
  type: 'kind' | 'minikube' | 'other';
  connected: boolean;
  runningPods?: number;
  failedPods?: number;
}

export interface Node {
  name: string;
  status: NodeStatus;
  roles: string[];
  cpuUsage: number;
  memoryUsage: number;
  age: string;
  version?: string;
}

export interface Pod {
  name: string;
  namespace: string;
  status: PodStatus;
  restarts: number;
  age: string;
  node?: string;
}

export interface Alert {
  id: string;
  severity: AlertSeverity;
  cluster: string;
  resource: string;
  message: string;
  timestamp: string;
}

export interface TimeSeriesPoint {
  time: string;
  value: number;
}

// ===== Fetch Functions =====

export async function fetchClusters(): Promise<Cluster[]> {
  const res = await fetch(`${API_BASE}/clusters`);
  if (!res.ok) throw new Error('Failed to fetch clusters');
  const data = await res.json();

  return (data.clusters || []).map((c: any) => ({
    name: c.name,
    displayName: c.displayName || c.name,
    status: (c.status || 'Unknown').toLowerCase() as ClusterStatus,
    nodeCount: c.nodeCount || 0,
    podCount: c.podCount || 0,
    cpuUsage: Math.round((c.cpuUsage || 0) * 10) / 10,
    memoryUsage: Math.round((c.memoryUsage || 0) * 10) / 10,
    reachable: c.reachable ?? false,
    type: c.name?.startsWith('kind-') ? 'kind' : c.name?.startsWith('minikube') ? 'minikube' : 'other',
    connected: c.reachable ?? false,
    runningPods: c.runningPods || 0,
    failedPods: c.failedPods || 0,
  }));
}

export async function fetchNodes(clusterName: string): Promise<Node[]> {
  const res = await fetch(`${API_BASE}/clusters/${clusterName}/nodes`);
  if (!res.ok) throw new Error('Failed to fetch nodes');
  const data = await res.json();

  return (data.nodes || []).map((n: any) => ({
    name: n.name,
    status: n.status as NodeStatus,
    roles: n.roles || ['<none>'],
    cpuUsage: 0, // Go API doesn't return per-node CPU yet
    memoryUsage: 0,
    age: n.age || '',
    version: n.version || '',
  }));
}

export async function fetchPods(clusterName: string): Promise<Pod[]> {
  const res = await fetch(`${API_BASE}/clusters/${clusterName}/pods`);
  if (!res.ok) throw new Error('Failed to fetch pods');
  const data = await res.json();

  return (data.pods || []).map((p: any) => ({
    name: p.name,
    namespace: p.namespace,
    status: p.status as PodStatus,
    restarts: p.restarts || 0,
    age: p.age || '',
    node: p.node || '',
  }));
}

export async function fetchAlerts(): Promise<Alert[]> {
  const res = await fetch(`${API_BASE}/alerts`);
  if (!res.ok) throw new Error('Failed to fetch alerts');
  const data = await res.json();

  return (data.alerts || []).map((a: any, index: number) => ({
    id: String(index + 1),
    severity: (a.severity || 'warning').toLowerCase() as AlertSeverity,
    cluster: a.cluster || '',
    resource: a.resource || a.cluster || '',
    message: a.message || '',
    timestamp: a.timestamp || 'just now',
  }));
}

export async function fetchHistory(clusterName: string): Promise<{ cpu: TimeSeriesPoint[]; memory: TimeSeriesPoint[] }> {
  const res = await fetch(`${API_BASE}/clusters/${clusterName}/history`);
  if (!res.ok) throw new Error('Failed to fetch history');
  const data = await res.json();

  const snapshots = data.snapshots || [];
  const cpu: TimeSeriesPoint[] = snapshots.map((s: any) => ({
    time: new Date(s.Timestamp || s.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false }),
    value: Math.round((s.CpuUsage || s.cpuUsage || 0) * 10) / 10,
  }));
  const memory: TimeSeriesPoint[] = snapshots.map((s: any) => ({
    time: new Date(s.Timestamp || s.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false }),
    value: Math.round((s.MemoryUsage || s.memoryUsage || 0) * 10) / 10,
  }));

  return { cpu, memory };
}
