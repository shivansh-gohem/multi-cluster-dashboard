import { RefreshCw, Settings, Sun, Moon, Hexagon } from 'lucide-react';
import { useEffect, useState } from 'react';
import { alerts } from '@/lib/mock-data';

interface NavbarProps {
  isDark: boolean;
  onToggleTheme: () => void;
}

export function Navbar({ isDark, onToggleTheme }: NavbarProps) {
  const [countdown, setCountdown] = useState(5);
  const [refreshing, setRefreshing] = useState(false);
  const criticalCount = alerts.filter(a => a.severity === 'critical').length;
  const hasAlerts = criticalCount > 0;

  useEffect(() => {
    const timer = setInterval(() => {
      setCountdown(prev => {
        if (prev <= 1) {
          setRefreshing(true);
          setTimeout(() => setRefreshing(false), 600);
          return 5;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  return (
    <header className="glass-card h-14 px-4 flex items-center justify-between gap-4 sticky top-0 z-50 rounded-none border-x-0 border-t-0">
      <div className="flex items-center gap-2.5">
        <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
          <Hexagon className="w-5 h-5 text-primary" />
        </div>
        <span className="font-semibold text-foreground hidden sm:inline">K8s Monitor</span>
      </div>

      <div className="flex items-center gap-2">
        <span className={`status-dot ${hasAlerts ? 'status-dot-critical animate-pulse-dot' : 'status-dot-healthy'}`} />
        <span className="text-sm font-medium">
          {hasAlerts ? `${alerts.length} Alerts Active` : 'All Systems Healthy'}
        </span>
      </div>

      <div className="flex items-center gap-1.5">
        <button
          onClick={onToggleTheme}
          className="p-2 rounded-lg hover:bg-muted transition-colors"
          aria-label="Toggle theme"
        >
          {isDark ? <Sun className="w-4 h-4 text-muted-foreground" /> : <Moon className="w-4 h-4 text-muted-foreground" />}
        </button>
        <div className="flex items-center gap-1.5 px-2 py-1.5 rounded-lg hover:bg-muted transition-colors cursor-pointer">
          <RefreshCw className={`w-4 h-4 text-muted-foreground ${refreshing ? 'animate-spin-slow' : ''}`} />
          <span className="text-xs text-muted-foreground hidden md:inline">{countdown}s</span>
        </div>
        <button className="p-2 rounded-lg hover:bg-muted transition-colors" aria-label="Settings">
          <Settings className="w-4 h-4 text-muted-foreground" />
        </button>
      </div>
    </header>
  );
}
