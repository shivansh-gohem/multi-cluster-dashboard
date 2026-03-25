import { Plus, Hexagon, Box } from 'lucide-react';
import type { Cluster } from '@/lib/hooks';
import { cn } from '@/lib/utils';

interface ClusterSidebarProps {
  clusters: Cluster[];
  selected: string;
  onSelect: (name: string) => void;
}

export function ClusterSidebar({ clusters, selected, onSelect }: ClusterSidebarProps) {
  const statusClass = (s: Cluster['status']) =>
    s === 'healthy' ? 'status-dot-healthy' : s === 'warning' ? 'status-dot-warning' : s === 'critical' ? 'status-dot-critical' : 'status-dot-offline';

  return (
    <aside className="w-64 shrink-0 p-3 flex flex-col gap-2 overflow-y-auto hidden lg:flex">
      <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-2 mb-1">Clusters</h2>
      {clusters.map(c => (
        <button
          key={c.name}
          onClick={() => onSelect(c.name)}
          className={cn(
            'glass-card-hover p-3 text-left w-full animate-fade-slide-in',
            selected === c.name && 'border-primary/50 bg-primary/5'
          )}
        >
          <div className="flex items-center gap-2.5">
            <div className={cn('w-8 h-8 rounded-lg flex items-center justify-center', c.type === 'kind' ? 'bg-primary/10' : 'bg-accent/10')}>
              {c.type === 'kind' ? <Hexagon className="w-4 h-4 text-primary" /> : <Box className="w-4 h-4 text-accent" />}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5">
                <span className={`status-dot ${statusClass(c.status)}`} />
                <span className="text-sm font-medium truncate">{c.name}</span>
              </div>
              <div className="flex items-center gap-2 mt-1">
                {c.connected ? (
                  <span className="text-[10px] font-semibold text-healthy flex items-center gap-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-healthy animate-pulse-dot" /> LIVE
                  </span>
                ) : (
                  <span className="status-badge status-badge-offline text-[10px]">Offline</span>
                )}
                <span className="text-[10px] text-muted-foreground">{c.nodeCount}N · {c.podCount}P</span>
              </div>
            </div>
          </div>
        </button>
      ))}
      <button className="glass-card-hover p-3 flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground">
        <Plus className="w-4 h-4" />
        <span className="text-sm">Add Cluster</span>
      </button>
    </aside>
  );
}
