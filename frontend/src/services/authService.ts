import { apiRequest } from './apiClient';
import { AuthResponse, LoginPayload, RegisterPayload } from '../types/auth';

export const authService = {
  login: async (payload: LoginPayload): Promise<AuthResponse> => {
    return apiRequest<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  register: async (payload: RegisterPayload): Promise<AuthResponse> => {
    return apiRequest<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  getMe: async (): Promise<AuthResponse> => {
    return apiRequest<AuthResponse>('/auth/me', {
      method: 'GET',
    });
  },

  logout: async (): Promise<void> => {
    try {
      await apiRequest('/auth/logout', { method: 'POST' });
    } catch {
      // Ignore network errors on logout
    } finally {
      localStorage.removeItem('was_access_token');
      localStorage.removeItem('was_refresh_token');
      localStorage.removeItem('was_user');
      localStorage.removeItem('was_org');
      localStorage.removeItem('was_role');
    }
  },
};

