export type Role = 'Owner' | 'Admin' | 'Manager' | 'Agent';

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone?: string;
  avatar_url?: string;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  status: string;
  settings: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface OrganizationMember {
  id: string;
  organization_id: string;
  user_id: string;
  role: Role;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  organization?: Organization;
  user?: User;
}

export interface AuthResponse {
  user: User;
  organization: Organization;
  role: Role;
  access_token: string;
  refresh_token: string;
  expires_in: number;
  memberships?: OrganizationMember[];
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface RegisterPayload {
  business_name: string;
  full_name: string;
  email: string;
  password: string;
}
