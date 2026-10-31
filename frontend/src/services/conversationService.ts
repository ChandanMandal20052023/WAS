import { apiClient } from './apiClient';
import {
  ConversationDetail,
  ConversationListResponse,
  ConversationStatus,
  Message,
  InternalNote,
} from '../types/conversations';

export const conversationService = {
  async getConversations(params?: {
    search?: string;
    status?: ConversationStatus;
    channel_id?: string;
    agent_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<ConversationListResponse> {
    const { data } = await apiClient.get<ConversationListResponse>('/conversations', { params });
    return data;
  },

  async getConversation(id: string): Promise<ConversationDetail> {
    const { data } = await apiClient.get<ConversationDetail>(`/conversations/${id}`);
    return data;
  },

  async updateStatus(id: string, status: ConversationStatus): Promise<void> {
    await apiClient.patch(`/conversations/${id}/status`, { status });
  },

  async assignAgent(id: string, agentId: string | null, reason?: string): Promise<void> {
    await apiClient.post(`/conversations/${id}/assign`, {
      agent_id: agentId,
      reason,
    });
  },

  async markAsRead(id: string): Promise<void> {
    await apiClient.post(`/conversations/${id}/read`);
  },

  async getMessages(conversationId: string, limit = 50, offset = 0): Promise<Message[]> {
    const { data } = await apiClient.get<{ messages: Message[]; count: number }>(
      `/conversations/${conversationId}/messages`,
      { params: { limit, offset } }
    );
    return data.messages;
  },

  async sendMessage(conversationId: string, body: string): Promise<Message> {
    const { data } = await apiClient.post<Message>(`/conversations/${conversationId}/messages`, {
      body,
    });
    return data;
  },

  async getNotes(conversationId: string): Promise<InternalNote[]> {
    const { data } = await apiClient.get<InternalNote[]>(`/conversations/${conversationId}/notes`);
    return data;
  },

  async createNote(conversationId: string, content: string): Promise<InternalNote> {
    const { data } = await apiClient.post<InternalNote>(`/conversations/${conversationId}/notes`, {
      content,
    });
    return data;
  },
};
