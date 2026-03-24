import { useState, useMemo } from 'react';
import { AlertTriangle } from 'lucide-react';
import { Alert } from '@/lib/mock-data';

export function AlertsTab({ alerts, clusterNames }: { alerts: Alert[]; clusterNames: string[] }) {
  const [sevFilter, setSevFilter] = useState('all');
  const [clusterFilter, setClusterFilter] = useState('all');

  const filtered = useMemo(() => alerts.filter(a =>
    (sevFilter === 'all' || a.severity === sevFilter) &&
    (clusterFilter === 'all' || a.cluster === clusterFilter)
  ), [alerts, sevFilter, clusterFilter]);

  return (
    <div className="space-y-3 animate-fade-slide-in">
      <div className="glass-card p-3 flex flex-wrap gap-2">
        <select value={sevFilter} onChange={e => setSevFilter(e.target.value)} className="bg-muted/50 text-sm rounded-lg px-3 py-1.5 outline-none border-0 text-foreground">
          <option value="all">All Severities</option>
          <option value="critical">Critical</option>
          <option value="warning">Warning</option>
        </select>
        <select value={clusterFilter} onChange={e => setClusterFilter(e.target.value)} className="bg-muted/50 text-sm rounded-lg px-3 py-1.5 outline-none border-0 text-foreground">
          <option value="all">All Clusters</option>
          {clusterNames.map(n => <option key={n} value={n}>{n}</option>)}
        </select>
      </div>

      <div className="space-y-2">
        {filtered.length === 0 && (
          <div className="glass-card p-8 text-center text-muted-foreground">
            <AlertTriangle className="w-8 h-8 mx-auto mb-2 opacity-40" />
            <p className="text-sm">No alerts match your filters</p>
          </div>
        )}
        {filtered.map(a => (
          <div key={a.id} className={`glass-card p-4 flex items-start gap-3 animate-fade-slide-in ${a.severity === 'critical' ? 'border-critical/30' : 'border-warning/30'}`}>
            <AlertTriangle className={`w-4 h-4 mt-0.5 shrink-0 ${a.severity === 'critical' ? 'text-critical animate-pulse-dot' : 'text-warning'}`} />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap mb-1">
                <span className={`status-badge ${a.severity === 'critical' ? 'status-badge-critical' : 'status-badge-warning'}`}>{a.severity}</span>
                <span className="status-badge status-badge-info">{a.cluster}</span>
              </div>
              <p className="text-sm font-medium">{a.resource}</p>
              <p className="text-sm text-muted-foreground mt-0.5">{a.message}</p>
              <span className="text-xs text-muted-foreground">{a.timestamp}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
