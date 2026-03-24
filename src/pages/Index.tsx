import { useState, useEffect, useMemo } from 'react';
import { Navbar } from '@/components/dashboard/Navbar';
import { ClusterSidebar } from '@/components/dashboard/ClusterSidebar';
import { MobileClusterSelect } from '@/components/dashboard/MobileClusterSelect';
import { OverviewTab } from '@/components/dashboard/OverviewTab';
import { NodesTab } from '@/components/dashboard/NodesTab';
import { PodsTab } from '@/components/dashboard/PodsTab';
import { AlertsTab } from '@/components/dashboard/AlertsTab';
import { clusters, alerts, getNodesForCluster, getPodsForCluster } from '@/lib/mock-data';
import { cn } from '@/lib/utils';

type Tab = 'overview' | 'nodes' | 'pods' | 'alerts';
const tabs: { id: Tab; label: string }[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'nodes', label: 'Nodes' },
  { id: 'pods', label: 'Pods' },
  { id: 'alerts', label: 'Alerts' },
];

export default function Index() {
  const [isDark, setIsDark] = useState(true);
  const [selected, setSelected] = useState(clusters[0].name);
  const [activeTab, setActiveTab] = useState<Tab>('overview');

  useEffect(() => {
    document.documentElement.classList.toggle('dark', isDark);
  }, [isDark]);

  // Initialize dark mode
  useEffect(() => {
    document.documentElement.classList.add('dark');
  }, []);

  const cluster = clusters.find(c => c.name === selected)!;
  const clusterAlerts = useMemo(() => alerts.filter(a => a.cluster === selected), [selected]);
  const nodes = useMemo(() => getNodesForCluster(selected), [selected]);
  const pods = useMemo(() => getPodsForCluster(selected), [selected]);

  return (
    <div className="min-h-screen bg-background">
      <Navbar isDark={isDark} onToggleTheme={() => setIsDark(!isDark)} />
      <div className="flex">
        <ClusterSidebar clusters={clusters} selected={selected} onSelect={setSelected} />
        <main className="flex-1 p-3 sm:p-4 space-y-4 min-w-0">
          <MobileClusterSelect clusters={clusters} selected={selected} onSelect={setSelected} />

          {/* Tab bar */}
          <div className="flex gap-1 glass-card p-1 w-fit">
            {tabs.map(t => (
              <button
                key={t.id}
                onClick={() => setActiveTab(t.id)}
                className={cn(
                  'px-4 py-1.5 rounded-lg text-sm font-medium transition-all',
                  activeTab === t.id ? 'bg-primary text-primary-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
                )}
              >
                {t.label}
                {t.id === 'alerts' && alerts.length > 0 && (
                  <span className="ml-1.5 px-1.5 py-0.5 text-[10px] rounded-full bg-critical/20 text-critical">{alerts.length}</span>
                )}
              </button>
            ))}
          </div>

          {/* Content */}
          {cluster.connected ? (
            <>
              {activeTab === 'overview' && <OverviewTab cluster={cluster} alerts={clusterAlerts} />}
              {activeTab === 'nodes' && <NodesTab nodes={nodes} />}
              {activeTab === 'pods' && <PodsTab pods={pods} />}
              {activeTab === 'alerts' && <AlertsTab alerts={alerts} clusterNames={clusters.map(c => c.name)} />}
            </>
          ) : (
            <div className="glass-card p-12 text-center animate-fade-slide-in">
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-muted flex items-center justify-center">
                <span className="text-3xl opacity-40">☸</span>
              </div>
              <h3 className="text-lg font-semibold mb-2">Cluster Offline</h3>
              <p className="text-sm text-muted-foreground mb-4">This cluster is currently unreachable. Try restarting it:</p>
              <code className="glass-card px-4 py-2 text-xs font-mono inline-block">kind create cluster --name {selected.replace(/^(kind-|minikube-)/, '')}</code>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
