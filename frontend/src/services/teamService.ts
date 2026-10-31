import { apiClient } from './apiClient';
import { Role } from '../types/auth';

export interface TeamMember {
  id: string;
  organization_id: string;
  user_id: string;
  role: Role;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  user: {
    id: string;
    email: string;
    full_name: string;
    phone?: string;
    avatar_url?: string;
    is_active: boolean;
  };
  open_conversations: number;
  pending_conversations: number;
  resolved_conversations: number;
}

export const teamService = {
  async getMembers(): Promise<TeamMember[]> {
    const { data } = await apiClient.get<TeamMember[]>('/organizations/members');
    return data;
  },

  async updateMemberRole(memberId: string, role: Role): Promise<void> {
    await apiClient.put(`/organizations/members/${memberId}/role`, { role });
  },

  async updateMemberStatus(memberId: string, isActive: boolean): Promise<void> {
    await apiClient.put(`/organizations/members/${memberId}/status`, { is_active: isActive });
  },

  async inviteMember(email: string, role: Role): Promise<any> {
    const { data } = await apiClient.post('/auth/invite', { email, role });
    return data;
  },
};
