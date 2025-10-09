/**
 * Auth Store
 *
 * Zustand store for managing authentication state including:
 * - Current user information
 * - JWT tokens (access and refresh)
 * - Login/logout actions
 * - Token persistence
 */

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { User, AuthTokens } from '@/lib/api/types';

// ============================================================================
// Types
// ============================================================================

interface AuthState {
  // State
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setTokens: (tokens: AuthTokens | null) => void;
  login: (user: User, tokens: AuthTokens) => void;
  logout: () => void;
  updateUser: (updates: Partial<User>) => void;
  setLoading: (loading: boolean) => void;
}

// ============================================================================
// Store
// ============================================================================

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      tokens: null,
      isAuthenticated: false,
      isLoading: true,

      // Set user
      setUser: (user) =>
        set({
          user,
          isAuthenticated: !!user,
        }),

      // Set tokens
      setTokens: (tokens) =>
        set({
          tokens,
        }),

      // Login - set user and tokens
      login: (user, tokens) =>
        set({
          user,
          tokens,
          isAuthenticated: true,
          isLoading: false,
        }),

      // Logout - clear all auth state
      logout: () =>
        set({
          user: null,
          tokens: null,
          isAuthenticated: false,
          isLoading: false,
        }),

      // Update user data (e.g., after profile edit)
      updateUser: (updates) => {
        const currentUser = get().user;
        if (currentUser) {
          set({
            user: { ...currentUser, ...updates },
          });
        }
      },

      // Set loading state
      setLoading: (loading) =>
        set({
          isLoading: loading,
        }),
    }),
    {
      name: 'auth-storage', // localStorage key
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        // Only persist user and tokens, not isLoading
        user: state.user,
        tokens: state.tokens,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// ============================================================================
// Selectors
// ============================================================================

/**
 * Select only the user from the store
 */
export const useUser = () => useAuthStore((state) => state.user);

/**
 * Select only the authentication status
 */
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated);

/**
 * Select only the tokens
 */
export const useTokens = () => useAuthStore((state) => state.tokens);

/**
 * Select the access token
 */
export const useAccessToken = () => useAuthStore((state) => state.tokens?.access_token);

/**
 * Select the refresh token
 */
export const useRefreshToken = () => useAuthStore((state) => state.tokens?.refresh_token);

/**
 * Select user type
 */
export const useUserType = () => useAuthStore((state) => state.user?.user_type);

/**
 * Check if user is a specific type
 */
export const useIsVolunteer = () => useAuthStore((state) => state.user?.user_type === 'volunteer');

export const useIsOrgAdmin = () =>
  useAuthStore(
    (state) =>
      state.user?.user_type === 'organization_admin' || state.user?.user_type === 'super_admin'
  );

export const useIsSuperAdmin = () =>
  useAuthStore((state) => state.user?.user_type === 'super_admin');
