/**
 * React Query (TanStack Query) Configuration
 *
 * Centralized configuration for TanStack Query with:
 * - Default cache times optimized for VolunteerSync data patterns
 * - Global error handling
 * - Retry logic for failed requests
 * - Query and mutation defaults
 *
 * Usage:
 * ```tsx
 * import { QueryClientProvider } from '@tanstack/react-query';
 * import { queryClient } from '@/lib/api/query-client';
 *
 * export default function RootLayout({ children }) {
 *   return (
 *     <QueryClientProvider client={queryClient}>
 *       {children}
 *     </QueryClientProvider>
 *   );
 * }
 * ```
 */

import { QueryClient, DefaultOptions } from '@tanstack/react-query';
import type { ApiError } from './types';

// ============================================================================
// Configuration Constants
// ============================================================================

/**
 * Cache time configuration (in milliseconds)
 *
 * Cache Strategy:
 * - Frequently changing data (notifications): 1 minute stale, 5 minutes cache
 * - Moderate changes (opportunities, hours): 5 minutes stale, 30 minutes cache
 * - Rarely changing (org profiles): 30 minutes stale, 1 hour cache
 */
export const CACHE_TIMES = {
  // How long data is considered "fresh" before refetch
  STALE_TIME: {
    SHORT: 1 * 60 * 1000, // 1 minute - notifications, unread counts
    MEDIUM: 5 * 60 * 1000, // 5 minutes - opportunities, registrations, hours logs
    LONG: 30 * 60 * 1000, // 30 minutes - organization profiles, user profiles
  },
  // How long inactive data stays in cache before garbage collection
  CACHE_TIME: {
    SHORT: 5 * 60 * 1000, // 5 minutes
    MEDIUM: 30 * 60 * 1000, // 30 minutes
    LONG: 60 * 60 * 1000, // 1 hour
  },
} as const;

/**
 * Retry configuration for failed queries
 */
const RETRY_CONFIG = {
  // Number of retry attempts for failed queries
  DEFAULT_RETRY: 2,
  // Retry delay function: exponential backoff
  RETRY_DELAY: (attemptIndex: number) => Math.min(1000 * 2 ** attemptIndex, 30000),
} as const;

// ============================================================================
// Error Handling
// ============================================================================

/**
 * Global error handler for React Query errors
 *
 * This function is called whenever a query or mutation fails after all retries.
 * Use it to:
 * - Log errors to monitoring services (e.g., Sentry)
 * - Show toast notifications for critical errors
 * - Track error patterns
 */
function onError(error: unknown): void {
  // Type-safe error handling
  if (isApiError(error)) {
    console.error('[React Query] API Error:', {
      status: error.status_code,
      error: error.error,
      message: error.message,
      details: error.details,
    });

    // Log to external monitoring service in production
    if (process.env.NODE_ENV === 'production') {
      // TODO: Integrate with error monitoring service (e.g., Sentry)
      // Sentry.captureException(error);
    }

    // Handle specific error cases
    if (error.status_code === 401) {
      // Token expired or invalid - handled by API client's token refresh logic
      console.warn('[React Query] Authentication error - token refresh should handle this');
    } else if (error.status_code === 403) {
      // Permission denied
      console.warn('[React Query] Permission denied:', error.message);
    } else if (error.status_code >= 500) {
      // Server error
      console.error('[React Query] Server error:', error.message);
    }
  } else {
    // Non-API error (network error, timeout, etc.)
    console.error('[React Query] Unexpected error:', error);
  }
}

/**
 * Type guard to check if error is an ApiError
 */
function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' && error !== null && 'status_code' in error && 'message' in error
  );
}

/**
 * Determine if a failed query should be retried
 *
 * Don't retry:
 * - 4xx errors (except 408 Request Timeout, 429 Too Many Requests)
 * - Authentication errors (handled by token refresh)
 */
function shouldRetry(failureCount: number, error: unknown): boolean {
  // Exceeded max retry attempts
  if (failureCount >= RETRY_CONFIG.DEFAULT_RETRY) {
    return false;
  }

  // Check if it's an API error
  if (isApiError(error)) {
    const status = error.status_code;

    // Don't retry client errors (4xx), except for:
    // - 408 Request Timeout
    // - 429 Too Many Requests (rate limiting)
    if (status >= 400 && status < 500) {
      return status === 408 || status === 429;
    }

    // Retry server errors (5xx)
    return status >= 500;
  }

  // Retry network errors, timeouts, etc.
  return true;
}

// ============================================================================
// Query Client Configuration
// ============================================================================

/**
 * Default options for all queries and mutations
 */
const defaultOptions: DefaultOptions = {
  queries: {
    // Stale time: how long data is considered fresh
    // Use MEDIUM as default, override per query as needed
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM,

    // Cache time: how long inactive data stays in memory
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM,

    // Retry configuration
    retry: shouldRetry,
    retryDelay: RETRY_CONFIG.RETRY_DELAY,

    // Refetch configuration
    refetchOnWindowFocus: true, // Refetch stale data when user returns to tab
    refetchOnReconnect: true, // Refetch stale data when network reconnects
    refetchOnMount: true, // Refetch stale data when component mounts

    // Error handling
    throwOnError: false, // Don't throw errors, handle them in components

    // Structural sharing: optimize re-renders by reusing unchanged data
    structuralSharing: true,
  },

  mutations: {
    // Retry configuration for mutations
    retry: false, // Don't retry mutations by default (to avoid duplicate actions)

    // Error handling
    throwOnError: false,

    // Global mutation error handler
    onError,
  },
};

/**
 * Create and export the QueryClient instance
 *
 * This is the main QueryClient used throughout the application.
 * Wrap your app with QueryClientProvider and pass this client.
 */
export const queryClient = new QueryClient({
  defaultOptions,
});

// ============================================================================
// Query Key Factories
// ============================================================================

/**
 * Query key factories for consistent cache key generation
 *
 * Benefits:
 * - Type-safe query keys
 * - Consistent invalidation patterns
 * - Easy to refactor and maintain
 *
 * Usage:
 * ```tsx
 * const { data } = useQuery({
 *   queryKey: queryKeys.opportunities.list({ status: 'open' }),
 *   queryFn: () => fetchOpportunities({ status: 'open' }),
 * });
 * ```
 */
export const queryKeys = {
  // Authentication
  auth: {
    all: ['auth'] as const,
    currentUser: () => [...queryKeys.auth.all, 'current-user'] as const,
  },

  // Opportunities
  opportunities: {
    all: ['opportunities'] as const,
    lists: () => [...queryKeys.opportunities.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.opportunities.lists(), filters] as const,
    details: () => [...queryKeys.opportunities.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.opportunities.details(), id] as const,
  },

  // Organizations
  organizations: {
    all: ['organizations'] as const,
    lists: () => [...queryKeys.organizations.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.organizations.lists(), filters] as const,
    details: () => [...queryKeys.organizations.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.organizations.details(), id] as const,
    members: (id: string) => [...queryKeys.organizations.detail(id), 'members'] as const,
    dashboard: (id: string) => [...queryKeys.organizations.detail(id), 'dashboard'] as const,
    analytics: (id: string) => [...queryKeys.organizations.detail(id), 'analytics'] as const,
    team: (id: string) => [...queryKeys.organizations.detail(id), 'team'] as const,
  },

  // Users
  users: {
    all: ['users'] as const,
    details: () => [...queryKeys.users.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.users.details(), id] as const,
    profile: (id: string) => [...queryKeys.users.detail(id), 'profile'] as const,
  },

  // Volunteers
  volunteers: {
    all: ['volunteers'] as const,
    lists: () => [...queryKeys.volunteers.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.volunteers.lists(), filters] as const,
    details: () => [...queryKeys.volunteers.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.volunteers.details(), id] as const,
    profile: (id: string) => [...queryKeys.volunteers.detail(id), 'profile'] as const,
    myProfile: () => [...queryKeys.volunteers.all, 'my-profile'] as const,
    dashboard: () => [...queryKeys.volunteers.all, 'dashboard'] as const,
  },

  // Registrations
  registrations: {
    all: ['registrations'] as const,
    lists: () => [...queryKeys.registrations.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.registrations.lists(), filters] as const,
    details: () => [...queryKeys.registrations.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.registrations.details(), id] as const,
    byOpportunity: (opportunityId: string) =>
      [...queryKeys.registrations.all, 'by-opportunity', opportunityId] as const,
    byVolunteer: (volunteerId: string) =>
      [...queryKeys.registrations.all, 'by-volunteer', volunteerId] as const,
  },

  // Hours
  hours: {
    all: ['hours'] as const,
    lists: () => [...queryKeys.hours.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) => [...queryKeys.hours.lists(), filters] as const,
    details: () => [...queryKeys.hours.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.hours.details(), id] as const,
    byVolunteer: (volunteerId: string) =>
      [...queryKeys.hours.all, 'by-volunteer', volunteerId] as const,
    summary: (volunteerId: string) =>
      [...queryKeys.hours.byVolunteer(volunteerId), 'summary'] as const,
  },

  // Notifications
  notifications: {
    all: ['notifications'] as const,
    lists: () => [...queryKeys.notifications.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.notifications.lists(), filters] as const,
    unreadCount: () => [...queryKeys.notifications.all, 'unread-count'] as const,
  },

  // Communications (Messages)
  messages: {
    all: ['messages'] as const,
    lists: () => [...queryKeys.messages.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) => [...queryKeys.messages.lists(), filters] as const,
    thread: (recipientId: string) => [...queryKeys.messages.all, 'thread', recipientId] as const,
  },

  // Achievements
  achievements: {
    all: ['achievements'] as const,
    lists: () => [...queryKeys.achievements.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.achievements.lists(), filters] as const,
    byVolunteer: (volunteerId: string) =>
      [...queryKeys.achievements.all, 'by-volunteer', volunteerId] as const,
  },

  // Analytics
  analytics: {
    all: ['analytics'] as const,
    volunteerStats: (volunteerId: string) =>
      [...queryKeys.analytics.all, 'volunteer-stats', volunteerId] as const,
    organizationStats: (orgId: string) =>
      [...queryKeys.analytics.all, 'organization-stats', orgId] as const,
  },
} as const;

// ============================================================================
// Cache Invalidation Helpers
// ============================================================================

/**
 * Invalidate all queries for a specific entity type
 *
 * Usage:
 * ```tsx
 * import { invalidateQueries } from '@/lib/api/query-client';
 *
 * // After creating a new opportunity
 * await invalidateQueries.opportunities();
 * ```
 */
export const invalidateQueries = {
  opportunities: () => queryClient.invalidateQueries({ queryKey: queryKeys.opportunities.all }),
  organizations: () => queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all }),
  users: () => queryClient.invalidateQueries({ queryKey: queryKeys.users.all }),
  volunteers: () => queryClient.invalidateQueries({ queryKey: queryKeys.volunteers.all }),
  registrations: () => queryClient.invalidateQueries({ queryKey: queryKeys.registrations.all }),
  hours: () => queryClient.invalidateQueries({ queryKey: queryKeys.hours.all }),
  notifications: () => queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all }),
  messages: () => queryClient.invalidateQueries({ queryKey: queryKeys.messages.all }),
  achievements: () => queryClient.invalidateQueries({ queryKey: queryKeys.achievements.all }),
  analytics: () => queryClient.invalidateQueries({ queryKey: queryKeys.analytics.all }),
  auth: () => queryClient.invalidateQueries({ queryKey: queryKeys.auth.all }),
} as const;

// ============================================================================
// Prefetch Helpers
// ============================================================================

/**
 * Prefetch data before navigation
 *
 * Usage:
 * ```tsx
 * // Before navigating to opportunity detail page
 * await prefetchOpportunity('opp_123');
 * router.push('/opportunities/opp_123');
 * ```
 */
export async function prefetchOpportunity(id: string): Promise<void> {
  // Implementation will be added when opportunity API functions are created
  // This is a placeholder for the pattern
  await queryClient.prefetchQuery({
    queryKey: queryKeys.opportunities.detail(id),
    // queryFn: () => fetchOpportunity(id),
  });
}

/**
 * Prefetch organization data
 */
export async function prefetchOrganization(id: string): Promise<void> {
  await queryClient.prefetchQuery({
    queryKey: queryKeys.organizations.detail(id),
    // queryFn: () => fetchOrganization(id),
  });
}

// ============================================================================
// Optimistic Update Helpers
// ============================================================================

/**
 * Type for optimistic update context
 * Used to rollback changes if mutation fails
 */
export interface OptimisticUpdateContext<T> {
  previousData?: T;
}

/**
 * Create an optimistic update helper for a specific query
 *
 * Usage:
 * ```tsx
 * const mutation = useMutation({
 *   mutationFn: updateOpportunity,
 *   onMutate: async (variables) => {
 *     return createOptimisticUpdate(
 *       queryKeys.opportunities.detail(variables.id),
 *       (old) => ({ ...old, ...variables })
 *     );
 *   },
 *   onError: (err, variables, context) => {
 *     if (context?.previousData) {
 *       queryClient.setQueryData(
 *         queryKeys.opportunities.detail(variables.id),
 *         context.previousData
 *       );
 *     }
 *   },
 * });
 * ```
 */
export async function createOptimisticUpdate<T>(
  queryKey: readonly unknown[],
  updater: (old: T | undefined) => T
): Promise<OptimisticUpdateContext<T>> {
  // Cancel outgoing refetches
  await queryClient.cancelQueries({ queryKey });

  // Snapshot previous value
  const previousData = queryClient.getQueryData<T>(queryKey);

  // Optimistically update
  queryClient.setQueryData(queryKey, updater);

  // Return context for rollback
  return { previousData };
}

// ============================================================================
// DevTools Configuration
// ============================================================================

/**
 * React Query DevTools configuration
 *
 * Enable in development for debugging:
 * ```tsx
 * import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
 *
 * <QueryClientProvider client={queryClient}>
 *   {children}
 *   <ReactQueryDevtools initialIsOpen={false} />
 * </QueryClientProvider>
 * ```
 */
export const devToolsConfig = {
  initialIsOpen: false,
  position: 'bottom-right' as const,
  toggleButtonProps: {
    style: {
      marginLeft: '5.5rem', // Adjust based on UI
    },
  },
};
