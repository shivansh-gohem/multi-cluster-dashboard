import { useState, useMemo } from 'react';
import { Search } from 'lucide-react';
import type { Pod, PodStatus } from '@/lib/hooks';

const statusBadgeClass = (s: PodStatus) => {
  switch (s) {
    case 'Running': return 'status-badge-healthy';
    case 'Succeeded': return 'status-badge-info';
    case 'Pending': return 'status-badge-warning';
    case 'Failed': case 'CrashLoopBackOff': return 'status-badge-critical';
  }
};

export function PodsTab({ pods }: { pods: Pod[] }) {
  const [search, setSearch] = useState('');
  const [nsFilter, setNsFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');

  const namespaces = useMemo(() => [...new Set(pods.map(p => p.namespace))], [pods]);
  const statuses = useMemo(() => [...new Set(pods.map(p => p.status))], [pods]);

  const filtered = useMemo(() => pods.filter(p =>
    (search === '' || p.name.toLowerCase().includes(search.toLowerCase())) &&
    (nsFilter === 'all' || p.namespace === nsFilter) &&
    (statusFilter === 'all' || p.status === statusFilter)
  ), [pods, search, nsFilter, statusFilter]);

  return (
    <div className="space-y-3 animate-fade-slide-in">
      <div className="glass-card p-3 flex flex-wrap gap-2">
        <div className="flex items-center gap-2 flex-1 min-w-[200px] bg-muted/50 rounded-lg px-3 py-1.5">
          <Search className="w-4 h-4 text-muted-foreground" />
          <input value={search} onChange={e => setSearch(e.target.value)} placeholder="Search pods..." className="bg-transparent text-sm outline-none flex-1 placeholder:text-muted-foreground" />
        </div>
        <select value={nsFilter} onChange={e => setNsFilter(e.target.value)} className="bg-muted/50 text-sm rounded-lg px-3 py-1.5 outline-none border-0 text-foreground">
          <option value="all">All Namespaces</option>
          {namespaces.map(ns => <option key={ns} value={ns}>{ns}</option>)}
        </select>
        <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="bg-muted/50 text-sm rounded-lg px-3 py-1.5 outline-none border-0 text-foreground">
          <option value="all">All Statuses</option>
          {statuses.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>

      <div className="glass-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Pod Name</th>
                <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Namespace</th>
                <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Status</th>
                <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Restarts</th>
                <th className="text-left p-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Age</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(p => (
                <tr key={p.name} className="border-b last:border-b-0 hover:bg-muted/30 transition-colors">
                  <td className="p-3 font-mono text-xs">{p.name}</td>
                  <td className="p-3"><span className="status-badge status-badge-info">{p.namespace}</span></td>
                  <td className="p-3"><span className={`status-badge ${statusBadgeClass(p.status)}`}>{p.status}</span></td>
                  <td className={`p-3 ${p.restarts > 5 ? 'text-critical font-medium' : 'text-muted-foreground'}`}>{p.restarts}</td>
                  <td className="p-3 text-muted-foreground">{p.age}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
