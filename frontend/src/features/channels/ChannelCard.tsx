import React, { useState } from 'react';
import { WhatsAppChannel } from '../../types/channels';
import { StatusBadge } from './StatusBadge';
import {
  QrCode,
  Power,
  RotateCw,
  Trash2,
  Calendar,
  Smartphone,
  Server,
  Loader2,
} from 'lucide-react';

interface ChannelCardProps {
  channel: WhatsAppChannel;
  onOpenQR: (channel: WhatsAppChannel) => void;
  onDisconnect: (id: string) => Promise<void>;
  onReconnect: (id: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onSync: (id: string) => Promise<void>;
}

export const ChannelCard: React.FC<ChannelCardProps> = ({
  channel,
  onOpenQR,
  onDisconnect,
  onReconnect,
  onDelete,
  onSync,
}) => {
  const [isActionLoading, setIsActionLoading] = useState(false);

  const handleAction = async (action: () => Promise<void>) => {
    try {
      setIsActionLoading(true);
      await action();
    } finally {
      setIsActionLoading(false);
    }
  };

  const formattedDate = channel.last_connected_at
    ? new Date(channel.last_connected_at).toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : 'Never connected';

  return (
    <div className="bg-card border border-border rounded-2xl p-5 shadow-sm hover:shadow-md transition-all flex flex-col justify-between space-y-4">
      {/* Header */}
      <div className="space-y-2">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-whatsapp/15 text-whatsapp flex items-center justify-center font-bold text-lg flex-shrink-0">
              <Smartphone className="w-5 h-5" />
            </div>
            <div>
              <h3 className="font-semibold text-base text-foreground leading-snug">
                {channel.name}
              </h3>
              <p className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                <Server className="w-3 h-3" />
                <span className="font-mono">{channel.instance_name}</span>
              </p>
            </div>
          </div>
          <StatusBadge status={channel.status} />
        </div>

        {/* Channel Details */}
        <div className="pt-2 grid grid-cols-2 gap-2 text-xs text-muted-foreground border-t border-border">
          <div>
            <span className="block font-medium text-foreground">Phone Number</span>
            <span>{channel.phone_number || 'Not connected'}</span>
          </div>
          <div>
            <span className="block font-medium text-foreground">Last Active</span>
            <span className="flex items-center gap-1">
              <Calendar className="w-3 h-3" />
              {formattedDate}
            </span>
          </div>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="pt-3 border-t border-border flex items-center justify-between gap-2">
        <button
          onClick={() => handleAction(() => onSync(channel.id))}
          disabled={isActionLoading}
          className="p-2 text-muted-foreground hover:bg-muted hover:text-foreground rounded-lg transition-colors"
          title="Refresh Status"
        >
          <RotateCw className={`w-4 h-4 ${isActionLoading ? 'animate-spin' : ''}`} />
        </button>

        <div className="flex items-center gap-2">
          {channel.status !== 'Connected' && (
            <button
              onClick={() => onOpenQR(channel)}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-whatsapp/10 text-whatsapp hover:bg-whatsapp hover:text-white transition-all"
            >
              <QrCode className="w-3.5 h-3.5" />
              <span>Pair Device</span>
            </button>
          )}

          {channel.status === 'Connected' ? (
            <button
              onClick={() => handleAction(() => onDisconnect(channel.id))}
              disabled={isActionLoading}
              className="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium text-destructive hover:bg-destructive/10 transition-colors"
              title="Disconnect"
            >
              <Power className="w-3.5 h-3.5" />
              <span>Disconnect</span>
            </button>
          ) : (
            <button
              onClick={() => handleAction(() => onReconnect(channel.id))}
              disabled={isActionLoading}
              className="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium text-primary hover:bg-primary/10 transition-colors"
              title="Reconnect Session"
            >
              <RotateCw className="w-3.5 h-3.5" />
              <span>Reconnect</span>
            </button>
          )}

          <button
            onClick={() => handleAction(() => onDelete(channel.id))}
            disabled={isActionLoading}
            className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
            title="Delete Channel"
          >
            {isActionLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Trash2 className="w-4 h-4" />}
          </button>
        </div>
      </div>
    </div>
  );
};
