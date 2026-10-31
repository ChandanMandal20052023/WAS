import { apiRequest } from './apiClient';
import {
  WhatsAppChannel,
  CreateChannelPayload,
  QRCodeResponse,
  ChannelStatusResponse,
} from '../types/channels';

export const channelService = {
  listChannels: async (): Promise<WhatsAppChannel[]> => {
    return apiRequest<WhatsAppChannel[]>('/channels');
  },

  getChannel: async (id: string): Promise<WhatsAppChannel> => {
    return apiRequest<WhatsAppChannel>(`/channels/${id}`);
  },

  createChannel: async (payload: CreateChannelPayload): Promise<WhatsAppChannel> => {
    return apiRequest<WhatsAppChannel>('/channels', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  getQRCode: async (id: string): Promise<QRCodeResponse> => {
    return apiRequest<QRCodeResponse>(`/channels/${id}/qr`);
  },

  connectChannel: async (id: string): Promise<void> => {
    await apiRequest(`/channels/${id}/connect`, {
      method: 'POST',
    });
  },

  syncStatus: async (id: string): Promise<ChannelStatusResponse> => {
    return apiRequest<ChannelStatusResponse>(`/channels/${id}/status`);
  },

  disconnectChannel: async (id: string): Promise<void> => {
    await apiRequest(`/channels/${id}/disconnect`, {
      method: 'POST',
    });
  },

  reconnectChannel: async (id: string): Promise<void> => {
    await apiRequest(`/channels/${id}/reconnect`, {
      method: 'POST',
    });
  },

  deleteChannel: async (id: string): Promise<void> => {
    await apiRequest(`/channels/${id}`, {
      method: 'DELETE',
    });
  },

  pairPhone: async (id: string, phone: string): Promise<{ pairing_code: string }> => {
    return apiRequest<{ pairing_code: string }>(`/channels/${id}/pair`, {
      method: 'POST',
      body: JSON.stringify({ phone }),
    });
  },
};

