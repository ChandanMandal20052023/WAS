import React from 'react';
import { ChannelStatus } from '../../types/channels';
import { CheckCircle2, AlertCircle, RefreshCw, Radio } from 'lucide-react';

interface StatusBadgeProps {
  status: ChannelStatus;
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  switch (status) {
    case 'Connected':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
          <CheckCircle2 className="w-3.5 h-3.5" />
          <span>Connected</span>
        </span>
      );

    case 'Waiting for QR Scan':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20">
          <Radio className="w-3.5 h-3.5 animate-pulse" />
          <span>Waiting for Scan</span>
        </span>
      );

    case 'Connecting':
    case 'Creating':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
          <RefreshCw className="w-3.5 h-3.5 animate-spin" />
          <span>{status}</span>
        </span>
      );

    case 'Reconnecting':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
          <RefreshCw className="w-3.5 h-3.5 animate-spin" />
          <span>Reconnecting</span>
        </span>
      );

    case 'Connection Failed':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-destructive/10 text-destructive border border-destructive/20">
          <AlertCircle className="w-3.5 h-3.5" />
          <span>Failed</span>
        </span>
      );

    case 'Disconnected':
    default:
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-muted text-muted-foreground border border-border">
          <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground/50"></span>
          <span>Disconnected</span>
        </span>
      );
  }
};
