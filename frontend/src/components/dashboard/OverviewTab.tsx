import { ArrowUp, ArrowDown, Server, Container, AlertTriangle, Cpu } from 'lucide-react';
import type { Cluster, Alert, TimeSeriesPoint } from '@/lib/hooks';
import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

interface OverviewTabProps {
  cluster: Cluster;
  alerts: Alert[];
  history?: { cpu: TimeSeriesPoint[]; memory: TimeSeriesPoint[] };
}

function StatCard({ label, value, icon: Icon, change, variant }: { label: string; value: string | number; icon: React.ElementType; change?: number; variant?: 'critical' }) {
  return (
    <div className={`glass-card p-4 animate-fade-slide-in ${variant === 'critical' ? 'border-critical/30' : ''}`}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-muted-foreground font-medium">{label}</span>
        <Icon className={`w-4 h-4 ${variant === 'critical' ? 'text-critical' : 'text-muted-foreground'}`} />
      </div>
      <div className={`text-2xl font-bold ${variant === 'critical' ? 'text-critical' : 'text-foreground'}`}>{value}</div>
      {change !== undefined && (
        <div className={`flex items-center gap-1 mt-1 text-xs ${change >= 0 ? 'text-healthy' : 'text-critical'}`}>
          {change >= 0 ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />}
          <span>{Math.abs(change)}%</span>
        </div>
      )}
    </div>
  );
}

function MetricChart({ title, data, color }: { title: string; data: { time: string; value: number }[]; color: string }) {
  if (data.length === 0) {
    return (
      <div className="glass-card p-4 animate-fade-slide-in">
        <h3 className="text-sm font-medium mb-3">{title}</h3>
        <div className="h-[200px] flex items-center justify-center text-muted-foreground text-sm">
          Collecting metrics data...
        </div>
      </div>
    );
  }

  return (
    <div className="glass-card p-4 animate-fade-slide-in">
      <h3 className="text-sm font-medium mb-3">{title}</h3>
      <ResponsiveContainer width="100%" height={200}>
        <AreaChart data={data}>
          <defs>
            <linearGradient id={`grad-${color}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.3} />
              <stop offset="100%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <XAxis dataKey="time" tick={{ fontSize: 10, fill: 'hsl(220, 9%, 46%)' }} axisLine={false} tickLine={false} interval={3} />
          <YAxis tick={{ fontSize: 10, fill: 'hsl(220, 9%, 46%)' }} axisLine={false} tickLine={false} domain={[0, 100]} />
          <Tooltip contentStyle={{ background: 'hsl(224, 25%, 10%)', border: '1px solid hsl(224, 20%, 18%)', borderRadius: '8px', fontSize: '12px', color: '#fff' }} />
          <Area type="monotone" dataKey="value" stroke={color} strokeWidth={2} fill={`url(#grad-${color})`} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}

export function OverviewTab({ cluster, alerts: clusterAlerts, history }: OverviewTabProps) {
  const cpuData = history?.cpu || [];
  const memData = history?.memory || [];
  const failedPods = cluster.failedPods || 0;

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard label="Total Nodes" value={cluster.nodeCount} icon={Server} />
        <StatCard label="Running Pods" value={cluster.runningPods || cluster.podCount} icon={Container} />
        <StatCard label="Failed Pods" value={failedPods} icon={AlertTriangle} variant={failedPods > 0 ? 'critical' : undefined} />
        <StatCard label="CPU Usage" value={`${cluster.cpuUsage}%`} icon={Cpu} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <MetricChart title="CPU Usage — 24h" data={cpuData} color="#3B82F6" />
        <MetricChart title="Memory Usage — 24h" data={memData} color="#8B5CF6" />
      </div>

      {clusterAlerts.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-semibold">Active Alerts</h3>
          {clusterAlerts.map(a => (
            <div key={a.id} className={`glass-card p-3 flex items-start gap-3 animate-fade-slide-in ${a.severity === 'critical' ? 'border-critical/30' : 'border-warning/30'}`}>
              <AlertTriangle className={`w-4 h-4 mt-0.5 shrink-0 ${a.severity === 'critical' ? 'text-critical animate-pulse-dot' : 'text-warning'}`} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className={`status-badge ${a.severity === 'critical' ? 'status-badge-critical' : 'status-badge-warning'}`}>{a.severity}</span>
                  <span className="text-xs text-muted-foreground">{a.resource}</span>
                </div>
                <p className="text-sm mt-1">{a.message}</p>
                <span className="text-xs text-muted-foreground">{a.timestamp}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
