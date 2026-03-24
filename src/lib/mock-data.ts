export type ClusterStatus = 'healthy' | 'warning' | 'critical' | 'offline';
export type PodStatus = 'Running' | 'Pending' | 'Failed' | 'CrashLoopBackOff' | 'Succeeded';
export type NodeStatus = 'Ready' | 'NotReady';
export type AlertSeverity = 'warning' | 'critical';

export interface Cluster {
  name: string;
  status: ClusterStatus;
  nodeCount: number;
  podCount: number;
  cpuUsage: number;
  memoryUsage: number;
  type: 'kind' | 'minikube';
  connected: boolean;
}

export interface Node {
  name: string;
  status: NodeStatus;
  roles: string[];
  cpuUsage: number;
  memoryUsage: number;
  age: string;
}

export interface Pod {
  name: string;
  namespace: string;
  status: PodStatus;
  restarts: number;
  age: string;
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

const generateTimeSeries = (baseValue: number, variance: number): TimeSeriesPoint[] =>
  Array.from({ length: 24 }, (_, i) => ({
    time: `${String(i).padStart(2, '0')}:00`,
    value: Math.max(0, Math.min(100, baseValue + (Math.random() - 0.5) * variance)),
  }));

export const clusters: Cluster[] = [
  { name: 'kind-production', status: 'healthy', nodeCount: 5, podCount: 47, cpuUsage: 62, memoryUsage: 71, type: 'kind', connected: true },
  { name: 'kind-staging', status: 'warning', nodeCount: 3, podCount: 28, cpuUsage: 78, memoryUsage: 85, type: 'kind', connected: true },
  { name: 'minikube-dev', status: 'healthy', nodeCount: 1, podCount: 12, cpuUsage: 34, memoryUsage: 45, type: 'minikube', connected: true },
  { name: 'kind-testing', status: 'critical', nodeCount: 2, podCount: 8, cpuUsage: 92, memoryUsage: 94, type: 'kind', connected: true },
  { name: 'minikube-local', status: 'offline', nodeCount: 1, podCount: 0, cpuUsage: 0, memoryUsage: 0, type: 'minikube', connected: false },
];

export const getNodesForCluster = (clusterName: string): Node[] => {
  const cluster = clusters.find(c => c.name === clusterName);
  if (!cluster) return [];
  return Array.from({ length: cluster.nodeCount }, (_, i) => ({
    name: `${clusterName}-node-${i + 1}`,
    status: (i === 0 && cluster.status === 'critical' ? 'NotReady' : 'Ready') as NodeStatus,
    roles: i === 0 ? ['control-plane', 'master'] : ['worker'],
    cpuUsage: Math.round(Math.random() * 40 + 30),
    memoryUsage: Math.round(Math.random() * 40 + 35),
    age: `${Math.floor(Math.random() * 30 + 1)}d`,
  }));
};

const namespaces = ['default', 'kube-system', 'monitoring', 'ingress-nginx', 'app'];
const podPrefixes = ['api-server', 'web-frontend', 'redis-cache', 'postgres-db', 'nginx-proxy', 'worker', 'scheduler', 'coredns', 'etcd', 'kube-proxy'];

export const getPodsForCluster = (clusterName: string): Pod[] => {
  const cluster = clusters.find(c => c.name === clusterName);
  if (!cluster) return [];
  return Array.from({ length: cluster.podCount }, (_, i) => {
    const statuses: PodStatus[] = ['Running', 'Running', 'Running', 'Running', 'Running', 'Pending', 'Failed', 'CrashLoopBackOff'];
    let status: PodStatus = cluster.status === 'healthy' ? 'Running' : statuses[Math.floor(Math.random() * statuses.length)];
    if (cluster.status === 'critical' && i < 3) status = 'Failed';
    return {
      name: `${podPrefixes[i % podPrefixes.length]}-${Math.random().toString(36).substring(2, 8)}`,
      namespace: namespaces[Math.floor(Math.random() * namespaces.length)],
      status,
      restarts: status === 'CrashLoopBackOff' ? Math.floor(Math.random() * 50 + 5) : Math.floor(Math.random() * 3),
      age: `${Math.floor(Math.random() * 48 + 1)}h`,
    };
  });
};

export const alerts: Alert[] = [
  { id: '1', severity: 'critical', cluster: 'kind-testing', resource: 'node/kind-testing-node-1', message: 'Node is NotReady — kubelet stopped responding', timestamp: '2 min ago' },
  { id: '2', severity: 'critical', cluster: 'kind-testing', resource: 'pod/api-server-x9f2k1', message: 'Pod in CrashLoopBackOff — OOMKilled', timestamp: '5 min ago' },
  { id: '3', severity: 'warning', cluster: 'kind-staging', resource: 'node/kind-staging-node-2', message: 'High memory pressure — 85% utilization', timestamp: '12 min ago' },
  { id: '4', severity: 'warning', cluster: 'kind-staging', resource: 'pod/redis-cache-m8k3j2', message: 'Pod restarted 3 times in the last hour', timestamp: '18 min ago' },
  { id: '5', severity: 'warning', cluster: 'kind-production', resource: 'deployment/web-frontend', message: 'Deployment rollout taking longer than expected', timestamp: '25 min ago' },
  { id: '6', severity: 'critical', cluster: 'kind-testing', resource: 'pod/postgres-db-q2w1e3', message: 'Persistent volume claim stuck in Pending state', timestamp: '32 min ago' },
];

export const getCpuTimeSeries = (clusterName: string): TimeSeriesPoint[] => {
  const cluster = clusters.find(c => c.name === clusterName);
  return generateTimeSeries(cluster?.cpuUsage ?? 50, 20);
};

export const getMemoryTimeSeries = (clusterName: string): TimeSeriesPoint[] => {
  const cluster = clusters.find(c => c.name === clusterName);
  return generateTimeSeries(cluster?.memoryUsage ?? 60, 15);
};
