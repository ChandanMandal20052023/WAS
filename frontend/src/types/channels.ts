export type ChannelStatus =
  | 'Disconnected'
  | 'Creating'
  | 'Waiting for QR Scan'
  | 'Connecting'
  | 'Connected'
  | 'Reconnecting'
  | 'Connection Failed';

export interface WhatsAppChannel {
  id: string;
  organization_id: string;
  name: string;
  phone_number?: string;
  instance_name: string;
  status: ChannelStatus;
  qrcode_base64?: string;
  settings: Record<string, unknown>;
  last_connected_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateChannelPayload {
  name: string;
}

export interface QRCodeResponse {
  qrcode: string;
  status: ChannelStatus;
  expires_in: number;
}

export interface ChannelStatusResponse {
  id: string;
  status: ChannelStatus;
  phone_number?: string;
  connected: boolean;
}

export interface PairPhonePayload {
  phone: string;
}

