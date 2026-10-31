import { create } from 'zustand';
import { User, Organization, Role, AuthResponse } from '../types/auth';

interface AuthState {
  user: User | null;
  organization: Organization | null;
  role: Role | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setAuth: (response: AuthResponse) => void;
  logout: () => void;
}

const getStoredUser = (): User | null => {
  const data = localStorage.getItem('was_user');
  if (!data) return null;
  try {
    return JSON.parse(data);
  } catch {
    return null;
  }
};

const getStoredOrg = (): Organization | null => {
  const data = localStorage.getItem('was_org');
  if (!data) return null;
  try {
    return JSON.parse(data);
  } catch {
    return null;
  }
};

export const useAuthStore = create<AuthState>((set) => ({
  user: getStoredUser(),
  organization: getStoredOrg(),
  role: (localStorage.getItem('was_role') as Role) || null,
  accessToken: localStorage.getItem('was_access_token'),
  isAuthenticated: Boolean(localStorage.getItem('was_access_token')),
  isLoading: false,

  setAuth: (response: AuthResponse) => {
    localStorage.setItem('was_access_token', response.access_token);
    localStorage.setItem('was_refresh_token', response.refresh_token);
    localStorage.setItem('was_user', JSON.stringify(response.user));
    localStorage.setItem('was_org', JSON.stringify(response.organization));
    localStorage.setItem('was_role', response.role);

    set({
      user: response.user,
      organization: response.organization,
      role: response.role,
      accessToken: response.access_token,
      isAuthenticated: true,
      isLoading: false,
    });
  },

  logout: () => {
    localStorage.removeItem('was_access_token');
    localStorage.removeItem('was_refresh_token');
    localStorage.removeItem('was_user');
    localStorage.removeItem('was_org');
    localStorage.removeItem('was_role');

    set({
      user: null,
      organization: null,
      role: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
  },
}));

