// React Query hooks with auto-polling for live data

import { useQuery } from '@tanstack/react-query';
import { fetchClusters, fetchNodes, fetchPods, fetchAlerts, fetchHistory } from './api';
import type { Cluster, Node, Pod, Alert, TimeSeriesPoint } from './api';

// Re-export types so components can import from hooks
export type { Cluster, Node, Pod, Alert, TimeSeriesPoint };
export type ClusterStatus = Cluster['status'];
export type PodStatus = Pod['status'];
export type NodeStatus = Node['status'];
export type AlertSeverity = Alert['severity'];

const POLL_INTERVAL = 5000; // 5 seconds

export function useClusters() {
  return useQuery<Cluster[]>({
    queryKey: ['clusters'],
    queryFn: fetchClusters,
    refetchInterval: POLL_INTERVAL,
    initialData: [],
  });
}

export function useNodes(clusterName: string) {
  return useQuery<Node[]>({
    queryKey: ['nodes', clusterName],
    queryFn: () => fetchNodes(clusterName),
    refetchInterval: POLL_INTERVAL,
    enabled: !!clusterName,
    initialData: [],
  });
}

export function usePods(clusterName: string) {
  return useQuery<Pod[]>({
    queryKey: ['pods', clusterName],
    queryFn: () => fetchPods(clusterName),
    refetchInterval: POLL_INTERVAL,
    enabled: !!clusterName,
    initialData: [],
  });
}

export function useAlerts() {
  return useQuery<Alert[]>({
    queryKey: ['alerts'],
    queryFn: fetchAlerts,
    refetchInterval: POLL_INTERVAL,
    initialData: [],
  });
}

export function useHistory(clusterName: string) {
  return useQuery<{ cpu: TimeSeriesPoint[]; memory: TimeSeriesPoint[] }>({
    queryKey: ['history', clusterName],
    queryFn: () => fetchHistory(clusterName),
    refetchInterval: 30000, // 30 seconds for history
    enabled: !!clusterName,
    initialData: { cpu: [], memory: [] },
  });
}
