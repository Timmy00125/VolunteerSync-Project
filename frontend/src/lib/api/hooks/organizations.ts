/**
 * Organization API Hooks
 *
 * React Query hooks for organization-related data fetching.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES } from '../query-client';
import type {
  Organization,
  OrganizationDashboard,
  OrganizationAnalytics,
  CreateOrganizationInput,
  UpdateOrganizationInput,
  TeamMember,
  InviteMemberInput,
} from '../types';
import apiClient from '../client';

// ============================================================================
// Dashboard Hook
// ============================================================================

/**
 * Hook to fetch organization dashboard data
 *
 * Fetches comprehensive dashboard metrics including:
 * - Volunteers recruited, hours contributed, events hosted
 * - Upcoming events
 * - Recent registrations
 * - Analytics summaries
 *
 * @param organizationId - The organization ID
 * @returns React Query result with dashboard data
 */
export function useOrganizationDashboard(organizationId: string) {
  return useQuery<OrganizationDashboard>({
    queryKey: queryKeys.organizations.dashboard(organizationId),
    queryFn: async () => {
      const response = await apiClient.get<{ data: OrganizationDashboard }>(
        `/organizations/${organizationId}/dashboard`
      );
      return response.data;
    },
    enabled: !!organizationId,
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM, // 5 minutes
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM, // 30 minutes
  });
}

// ============================================================================
// Organization CRUD Hooks
// ============================================================================

/**
 * Hook to fetch single organization
 */
export function useOrganization(organizationId: string) {
  return useQuery<Organization>({
    queryKey: queryKeys.organizations.detail(organizationId),
    queryFn: async () => {
      const response = await apiClient.get<{ data: Organization }>(
        `/organizations/${organizationId}`
      );
      return response.data;
    },
    enabled: !!organizationId,
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM,
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM,
  });
}

/**
 * Hook to create organization
 */
export function useCreateOrganization() {
  const queryClient = useQueryClient();

  return useMutation<Organization, Error, CreateOrganizationInput>({
    mutationFn: async (data: CreateOrganizationInput) => {
      const response = await apiClient.post<{ data: Organization }>('/organizations', data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
    },
  });
}

/**
 * Hook to update organization
 */
export function useUpdateOrganization(organizationId: string) {
  const queryClient = useQueryClient();

  return useMutation<Organization, Error, UpdateOrganizationInput>({
    mutationFn: async (data: UpdateOrganizationInput) => {
      const response = await apiClient.patch<{ data: Organization }>(
        `/organizations/${organizationId}`,
        data
      );
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.detail(organizationId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
    },
  });
}

// ============================================================================
// Analytics Hook
// ============================================================================

/**
 * Hook to fetch organization analytics
 *
 * @param timePeriod - Time period for analytics (e.g., '30d', '3m', '6m', '1y', 'all')
 */
export function useOrganizationAnalytics(timePeriod?: string) {
  return useQuery<OrganizationAnalytics>({
    queryKey: ['organizations', 'analytics', timePeriod],
    queryFn: async () => {
      const params = timePeriod ? `?period=${timePeriod}` : '';
      const response = await apiClient.get<{ data: OrganizationAnalytics }>(
        `/analytics/organization${params}`
      );
      return response.data;
    },
    staleTime: CACHE_TIMES.STALE_TIME.LONG, // 10 minutes
    gcTime: CACHE_TIMES.CACHE_TIME.LONG, // 60 minutes
  });
}

// ============================================================================
// Team Management Hooks
// ============================================================================

/**
 * Hook to fetch organization team members
 */
export function useOrganizationTeam(organizationId: string) {
  return useQuery<TeamMember[]>({
    queryKey: queryKeys.organizations.team(organizationId),
    queryFn: async () => {
      const response = await apiClient.get<{ data: TeamMember[] }>(
        `/organizations/${organizationId}/team`
      );
      return response.data;
    },
    enabled: !!organizationId,
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM,
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM,
  });
}

/**
 * Hook to invite team member
 */
export function useInviteTeamMember(organizationId: string) {
  const queryClient = useQueryClient();

  return useMutation<void, Error, InviteMemberInput>({
    mutationFn: async (data: InviteMemberInput) => {
      await apiClient.post(`/organizations/${organizationId}/team/invite`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.team(organizationId) });
    },
  });
}

/**
 * Hook to remove team member
 */
export function useRemoveTeamMember(organizationId: string) {
  const queryClient = useQueryClient();

  return useMutation<void, Error, string>({
    mutationFn: async (userId: string) => {
      await apiClient.delete(`/organizations/${organizationId}/team/${userId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.team(organizationId) });
    },
  });
}
