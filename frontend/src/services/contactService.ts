import { apiClient } from './apiClient';
import {
  ContactWithTags,
  ContactListResponse,
  CreateContactInput,
  UpdateContactInput,
  Tag,
} from '../types/contacts';

export const contactService = {
  async getContacts(params?: {
    search?: string;
    tag_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<ContactListResponse> {
    const { data } = await apiClient.get<ContactListResponse>('/contacts', { params });
    return data;
  },

  async getContact(id: string): Promise<ContactWithTags> {
    const { data } = await apiClient.get<ContactWithTags>(`/contacts/${id}`);
    return data;
  },

  async createContact(input: CreateContactInput): Promise<ContactWithTags> {
    const { data } = await apiClient.post<ContactWithTags>('/contacts', input);
    return data;
  },

  async updateContact(id: string, input: UpdateContactInput): Promise<ContactWithTags> {
    const { data } = await apiClient.put<ContactWithTags>(`/contacts/${id}`, input);
    return data;
  },

  async deleteContact(id: string): Promise<void> {
    await apiClient.delete(`/contacts/${id}`);
  },

  async getTags(): Promise<Tag[]> {
    const { data } = await apiClient.get<Tag[]>('/tags');
    return data;
  },

  async createTag(input: { name: string; color_hex?: string }): Promise<Tag> {
    const { data } = await apiClient.post<Tag>('/tags', input);
    return data;
  },
};

