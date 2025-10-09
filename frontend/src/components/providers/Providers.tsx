/**
 * Providers Component
 *
 * Central wrapper for all application-level providers including:
 * - React Query Provider
 * - Auth initialization
 * - Theme provider (if needed)
 *
 * This component should wrap all application content in the root layout.
 */

'use client';

import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from '@/lib/api/query-client';
import { useEffect } from 'react';
import { useAuthStore } from '@/store/auth-store';

interface ProvidersProps {
  children: React.ReactNode;
}

/**
 * Providers component that wraps the entire application
 *
 * Features:
 * - React Query client for server state management
 * - Auth initialization on mount
 *
 * Note: React Query DevTools can be added by installing @tanstack/react-query-devtools
 */
export function Providers({ children }: ProvidersProps) {
  const setLoading = useAuthStore((state) => state.setLoading);

  // Initialize auth state on mount
  useEffect(() => {
    // The auth store will automatically hydrate from localStorage
    // We just need to mark loading as complete after hydration
    setLoading(false);
  }, [setLoading]);

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
