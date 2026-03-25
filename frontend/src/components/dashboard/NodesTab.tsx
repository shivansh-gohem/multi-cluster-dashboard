import type { Node } from '@/lib/hooks';

export function NodesTab({ nodes }: { nodes: Node[] }) {
  return (
    <div className="glass-card overflow-hidden animate-fade-slide-in">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b">
              <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Node Name</th>
              <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Status</th>
              <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Roles</th>
              <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">CPU %</th>
              <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Memory %</th>
              <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Age</th>
            </tr>
          </thead>
          <tbody>
            {nodes.map(n => (
              <tr key={n.name} className={`border-b last:border-b-0 hover:bg-muted/30 transition-colors border-l-2 ${n.status === 'Ready' ? 'border-l-healthy' : 'border-l-critical'}`}>
                <td className="p-3 font-mono text-xs">{n.name}</td>
                <td className="p-3"><span className={`status-badge ${n.status === 'Ready' ? 'status-badge-healthy' : 'status-badge-critical'}`}>{n.status}</span></td>
                <td className="p-3"><div className="flex gap-1 flex-wrap">{n.roles.map(r => <span key={r} className="status-badge status-badge-info">{r}</span>)}</div></td>
                <td className="p-3"><div className="flex items-center gap-2"><div className="w-16 h-1.5 rounded-full bg-muted"><div className="h-full rounded-full bg-metric-cpu" style={{ width: `${n.cpuUsage}%` }} /></div><span className="text-xs text-muted-foreground">{n.cpuUsage}%</span></div></td>
                <td className="p-3"><div className="flex items-center gap-2"><div className="w-16 h-1.5 rounded-full bg-muted"><div className="h-full rounded-full bg-metric-memory" style={{ width: `${n.memoryUsage}%` }} /></div><span className="text-xs text-muted-foreground">{n.memoryUsage}%</span></div></td>
                <td className="p-3 text-muted-foreground">{n.age}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
