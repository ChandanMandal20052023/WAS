export interface Tag {
  id: string;
  organization_id: string;
  name: string;
  color_hex: string;
  created_at: string;
}

export interface Contact {
  id: string;
  organization_id: string;
  phone_number: string;
  full_name: string;
  email?: string;
  avatar_url?: string;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface ContactWithTags extends Contact {
  tags: Tag[];
}

export interface CreateContactInput {
  phone_number: string;
  full_name: string;
  email?: string;
  avatar_url?: string;
  tag_ids?: string[];
}

export interface UpdateContactInput {
  full_name: string;
  email?: string;
  avatar_url?: string;
  tag_ids?: string[];
}

export interface ContactListResponse {
  contacts: ContactWithTags[];
  total: number;
  limit: number;
  offset: number;
}

