import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Radio,
  Plus,
  RotateCw,
  Search,
  ShieldAlert,
  Loader2,
  X,
} from 'lucide-react';

import { channelService } from '../../services/channelService';
import { WhatsAppChannel } from '../../types/channels';
import { ChannelCard } from '../../features/channels/ChannelCard';
import { ConnectChannelModal } from '../../features/channels/ConnectChannelModal';

export const WhatsAppChannelsPage: React.FC = () => {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedChannel, setSelectedChannel] = useState<WhatsAppChannel | null>(null);
  const [isQRModalOpen, setIsQRModalOpen] = useState(false);
  const [isNewChannelModalOpen, setIsNewChannelModalOpen] = useState(false);
  const [newChannelName, setNewChannelName] = useState('');
  const [isCreatingChannel, setIsCreatingChannel] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  // Fetch channels list
  const {
    data: channels = [],
    isLoading,
    refetch,
    isRefetching,
  } = useQuery<WhatsAppChannel[]>({
    queryKey: ['channels'],
    queryFn: channelService.listChannels,
  });

  // Action Mutations
  const disconnectMutation = useMutation({
    mutationFn: channelService.disconnectChannel,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['channels'] }),
  });

  const reconnectMutation = useMutation({
    mutationFn: channelService.reconnectChannel,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['channels'] }),
  });

  const deleteMutation = useMutation({
    mutationFn: channelService.deleteChannel,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['channels'] }),
  });

  const syncMutation = useMutation({
    mutationFn: channelService.syncStatus,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['channels'] }),
  });

  const handleCreateChannel = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newChannelName.trim()) return;

    try {
      setIsCreatingChannel(true);
      setCreateError(null);
      const created = await channelService.createChannel({ name: newChannelName.trim() });
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      setIsNewChannelModalOpen(false);
      setNewChannelName('');
      // Open QR modal for the newly created channel
      setSelectedChannel(created);
      setIsQRModalOpen(true);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setCreateError(err.message);
      } else {
        setCreateError('Failed to create channel');
      }
    } finally {
      setIsCreatingChannel(false);
    }
  };

  const filteredChannels = channels.filter((ch) =>
    ch.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (ch.phone_number && ch.phone_number.includes(searchQuery))
  );

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Top Banner & Action Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <Radio className="w-6 h-6 text-whatsapp" />
            <span>WhatsApp Channels</span>
          </h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            Link and manage multi-device WhatsApp accounts for your organization.
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            onClick={() => refetch()}
            disabled={isRefetching}
            className="p-2 border border-border rounded-xl text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="Refresh Channels"
          >
            <RotateCw className={`w-4 h-4 ${isRefetching ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={() => setIsNewChannelModalOpen(true)}
            className="inline-flex items-center gap-2 bg-whatsapp hover:bg-whatsapp-dark text-white font-medium px-4 py-2 rounded-xl text-sm transition-all shadow-sm hover:shadow"
          >
            <Plus className="w-4 h-4" />
            <span>Connect WhatsApp</span>
          </button>
        </div>
      </div>

      {/* Search Bar */}
      <div className="relative max-w-md">
        <Search className="w-4 h-4 absolute left-3 top-3 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search by channel name or phone..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full bg-card border border-input pl-9 pr-4 py-2 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all shadow-sm"
        />
      </div>

      {/* Loading State */}
      {isLoading ? (
        <div className="py-20 flex flex-col items-center justify-center gap-3 text-center">
          <Loader2 className="w-8 h-8 text-whatsapp animate-spin" />
          <p className="text-sm text-muted-foreground">Loading WhatsApp channels...</p>
        </div>
      ) : filteredChannels.length > 0 ? (
        /* Channel Cards Grid */
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {filteredChannels.map((channel) => (
            <ChannelCard
              key={channel.id}
              channel={channel}
              onOpenQR={(ch) => {
                setSelectedChannel(ch);
                setIsQRModalOpen(true);
              }}
              onDisconnect={async (id) => {
                await disconnectMutation.mutateAsync(id);
              }}
              onReconnect={async (id) => {
                await reconnectMutation.mutateAsync(id);
              }}
              onDelete={async (id) => {
                if (window.confirm('Are you sure you want to delete this channel?')) {
                  await deleteMutation.mutateAsync(id);
                }
              }}
              onSync={async (id) => {
                await syncMutation.mutateAsync(id);
              }}
            />
          ))}
        </div>
      ) : (
        /* Empty State */
        <div className="bg-card border border-dashed border-border rounded-2xl p-12 text-center space-y-4 max-w-lg mx-auto mt-8">
          <div className="w-16 h-16 rounded-2xl bg-whatsapp/10 text-whatsapp flex items-center justify-center mx-auto">
            <Radio className="w-8 h-8" />
          </div>
          <div className="space-y-1">
            <h3 className="font-bold text-lg text-foreground">No WhatsApp Channels Connected</h3>
            <p className="text-sm text-muted-foreground">
              Connect your first WhatsApp account by scanning a QR code with your phone.
            </p>
          </div>
          <button
            onClick={() => setIsNewChannelModalOpen(true)}
            className="inline-flex items-center gap-2 bg-whatsapp hover:bg-whatsapp-dark text-white font-medium px-4 py-2 rounded-xl text-sm transition-all shadow-sm"
          >
            <Plus className="w-4 h-4" />
            <span>Connect WhatsApp Now</span>
          </button>
        </div>
      )}

      {/* Create Channel Modal */}
      {isNewChannelModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="w-full max-w-md bg-card border border-border rounded-2xl shadow-xl overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b border-border">
              <h3 className="font-bold text-base text-foreground">New WhatsApp Channel</h3>
              <button
                onClick={() => setIsNewChannelModalOpen(false)}
                className="text-muted-foreground hover:text-foreground p-1"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateChannel} className="p-6 space-y-4">
              {createError && (
                <div className="flex items-center gap-2 p-3 text-xs text-destructive bg-destructive/10 border border-destructive/20 rounded-xl">
                  <ShieldAlert className="w-4 h-4 flex-shrink-0" />
                  <span>{createError}</span>
                </div>
              )}

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-foreground">Channel Name / Label</label>
                <input
                  type="text"
                  placeholder="e.g., Customer Support, Sales Direct"
                  value={newChannelName}
                  onChange={(e) => setNewChannelName(e.target.value)}
                  className="w-full bg-background border border-input px-3.5 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                  autoFocus
                  required
                />
                <p className="text-[11px] text-muted-foreground">
                  A recognizable label for your team to identify this phone line.
                </p>
              </div>

              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsNewChannelModalOpen(false)}
                  className="px-4 py-2 rounded-xl text-xs font-medium text-muted-foreground hover:bg-muted"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreatingChannel || !newChannelName.trim()}
                  className="bg-whatsapp hover:bg-whatsapp-dark text-white font-medium px-4 py-2 rounded-xl text-xs flex items-center gap-1.5 shadow-sm disabled:opacity-50"
                >
                  {isCreatingChannel ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <span>Continue to QR Pairing</span>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* QR Code / Device Link Modal */}
      <ConnectChannelModal
        channel={selectedChannel}
        isOpen={isQRModalOpen}
        onClose={() => {
          setIsQRModalOpen(false);
          setSelectedChannel(null);
        }}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: ['channels'] });
        }}
      />
    </div>
  );
};
