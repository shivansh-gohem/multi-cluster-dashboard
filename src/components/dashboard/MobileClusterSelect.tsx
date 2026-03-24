import { Cluster } from '@/lib/mock-data';

export function MobileClusterSelect({ clusters, selected, onSelect }: { clusters: Cluster[]; selected: string; onSelect: (n: string) => void }) {
  return (
    <div className="lg:hidden">
      <select
        value={selected}
        onChange={e => onSelect(e.target.value)}
        className="w-full glass-card p-3 text-sm outline-none text-foreground bg-transparent"
      >
        {clusters.map(c => (
          <option key={c.name} value={c.name}>
            {c.status === 'healthy' ? '🟢' : c.status === 'warning' ? '🟡' : c.status === 'critical' ? '🔴' : '⚫'} {c.name}
          </option>
        ))}
      </select>
    </div>
  );
}
