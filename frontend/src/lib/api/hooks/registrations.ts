/**
 * Registrations API Hooks
 *
 * React Query hooks for registration-related data fetching.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES } from '../query-client';
import apiClient from '../client';
import type { Registration, PaginatedResponse } from '../types';

// ============================================================================
// List My Registrations Hook
// ============================================================================

interface MyRegistrationsParams {
  status?: 'registered' | 'checked_in' | 'completed' | 'cancelled' | 'waitlisted';
  page?: number;
  limit?: number;
}

/**
 * Hook to fetch volunteer's registrations
 *
 * @param params - Filter parameters
 * @returns React Query result with paginated registrations
 */
export function useMyRegistrations(params?: MyRegistrationsParams) {
  return useQuery<PaginatedResponse<Registration>>({
    queryKey: queryKeys.registrations.list((params as Record<string, unknown>) || {}),
    queryFn: async () => {
      const queryParams = new URLSearchParams();
      if (params?.status) queryParams.append('status', params.status);
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.limit) queryParams.append('limit', params.limit.toString());

      const queryString = queryParams.toString();
      const endpoint = queryString
        ? `/volunteers/me/registrations?${queryString}`
        : '/volunteers/me/registrations';

      return apiClient.get<PaginatedResponse<Registration>>(endpoint);
    },
    staleTime: CACHE_TIMES.STALE_TIME.SHORT, // 1 minute
    gcTime: CACHE_TIMES.CACHE_TIME.SHORT, // 5 minutes
  });
}

// ============================================================================
// Check-in Mutation Hook
// ============================================================================

/**
 * Hook to check in to an event
 *
 * @returns Mutation hook for checking in
 */
export function useCheckIn() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (registrationId: string) => {
      return apiClient.post<{ data: Registration }>(
        `/registrations/${registrationId}/check-in`,
        {}
      );
    },
    onSuccess: () => {
      // Invalidate registrations list to refresh data
      queryClient.invalidateQueries({ queryKey: queryKeys.registrations.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.volunteers.dashboard() });
    },
  });
}

// ============================================================================
// Cancel Registration Mutation Hook
// ============================================================================

interface CancelRegistrationInput {
  registrationId: string;
  reason?: string;
}

/**
 * Hook to cancel a registration
 *
 * @returns Mutation hook for cancelling registration
 */
export function useCancelRegistration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ registrationId, reason }: CancelRegistrationInput) => {
      return apiClient.post<{ data: Registration }>(`/registrations/${registrationId}/cancel`, {
        reason,
      });
    },
    onSuccess: () => {
      // Invalidate registrations list to refresh data
      queryClient.invalidateQueries({ queryKey: queryKeys.registrations.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.volunteers.dashboard() });
    },
  });
}

// ============================================================================
// Get Registrations by Opportunity Hook
// ============================================================================

/**
 * Hook to fetch registrations for a specific opportunity
 *
 * @param opportunityId - The opportunity ID
 * @returns React Query result with registrations
 */
export function useOpportunityRegistrations(opportunityId: string) {
  return useQuery<Registration[]>({
    queryKey: queryKeys.registrations.byOpportunity(opportunityId),
    queryFn: async () => {
      const response = await apiClient.get<{ data: Registration[] }>(
        `/opportunities/${opportunityId}/registrations`
      );
      return response.data;
    },
    enabled: !!opportunityId,
    staleTime: CACHE_TIMES.STALE_TIME.SHORT, // 1 minute
    gcTime: CACHE_TIMES.CACHE_TIME.SHORT, // 5 minutes
  });
}
