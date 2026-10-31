import { Tag } from './contacts';

export type ConversationStatus = 'open' | 'pending' | 'resolved';

export type MessageSenderType = 'customer' | 'agent' | 'system';

export type MessageDeliveryStatus = 'pending' | 'sent' | 'delivered' | 'read' | 'failed';

export interface Conversation {
  id: string;
  organization_id: string;
  channel_id: string;
  contact_id: string;
  assigned_agent_id?: string;
  status: ConversationStatus;
  last_message_preview?: string;
  last_message_time: string;
  unread_count: number;
  created_at: string;
  updated_at: string;
}

export interface ConversationDetail extends Conversation {
  contact_name: string;
  contact_phone: string;
  contact_avatar?: string;
  channel_name: string;
  assigned_agent_name?: string;
  assigned_agent_email?: string;
  tags: Tag[];
}

export interface Message {
  id: string;
  organization_id: string;
  conversation_id: string;
  whatsapp_message_id?: string;
  sender_type: MessageSenderType;
  sender_user_id?: string;
  sender_user_name?: string;
  message_type: string;
  body?: string;
  media_url?: string;
  media_mime_type?: string;
  delivery_status: MessageDeliveryStatus;
  sent_at: string;
  created_at: string;
}

export interface InternalNote {
  id: string;
  organization_id: string;
  conversation_id?: string;
  contact_id?: string;
  user_id: string;
  user_name: string;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface ConversationListResponse {
  conversations: ConversationDetail[];
  total: number;
  limit: number;
  offset: number;
}

