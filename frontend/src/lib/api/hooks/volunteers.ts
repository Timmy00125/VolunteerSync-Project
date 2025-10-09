/**
 * Volunteer API Hooks
 *
 * React Query hooks for volunteer-related data fetching.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES } from '../query-client';
import { getVolunteerDashboard, getVolunteerProfile, updateVolunteerProfile } from '../client';
import type { DashboardResponse, VolunteerProfile, UpdateVolunteerProfileInput } from '../types';

// ============================================================================
// Dashboard Hook
// ============================================================================

/**
 * Hook to fetch volunteer dashboard data
 *
 * Fetches comprehensive dashboard metrics including:
 * - Total hours, events, organizations
 * - Recent and upcoming events
 * - Achievements
 * - Monthly statistics
 *
 * @returns React Query result with dashboard data
 *
 * @example
 * ```tsx
 * function VolunteerDashboard() {
 *   const { data, isLoading, error } = useVolunteerDashboard();
 *
 *   if (isLoading) return <Skeleton />;
 *   if (error) return <ErrorMessage error={error} />;
 *
 *   return (
 *     <div>
 *       <h1>Total Hours: {data.total_hours}</h1>
 *       <EventsList events={data.upcoming_events} />
 *     </div>
 *   );
 * }
 * ```
 */
export function useVolunteerDashboard() {
  return useQuery<DashboardResponse>({
    queryKey: queryKeys.volunteers.dashboard(),
    queryFn: getVolunteerDashboard,
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM, // 5 minutes
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM, // 30 minutes
  });
}

// ============================================================================
// Profile Hooks
// ============================================================================

/**
 * Hook to fetch volunteer profile
 *
 * @returns React Query result with profile data
 *
 * @example
 * ```tsx
 * function ProfilePage() {
 *   const { data: profile, isLoading, error } = useVolunteerProfile();
 *
 *   if (isLoading) return <Skeleton />;
 *   if (error) return <ErrorMessage error={error} />;
 *
 *   return <ProfileForm defaultValues={profile} />;
 * }
 * ```
 */
export function useVolunteerProfile() {
  return useQuery<VolunteerProfile>({
    queryKey: queryKeys.volunteers.myProfile(),
    queryFn: getVolunteerProfile,
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM, // 5 minutes
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM, // 30 minutes
  });
}

/**
 * Hook to update volunteer profile
 *
 * @returns React Query mutation for updating profile
 *
 * @example
 * ```tsx
 * function ProfileForm() {
 *   const updateProfile = useUpdateVolunteerProfile();
 *
 *   const handleSubmit = (data) => {
 *     updateProfile.mutate(data, {
 *       onSuccess: () => toast.success('Profile updated'),
 *       onError: (error) => toast.error(error.message),
 *     });
 *   };
 * }
 * ```
 */
export function useUpdateVolunteerProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateVolunteerProfileInput) => updateVolunteerProfile(data),
    onSuccess: (data) => {
      // Invalidate and refetch profile data
      queryClient.setQueryData(queryKeys.volunteers.myProfile(), data);
      queryClient.invalidateQueries({ queryKey: queryKeys.volunteers.dashboard() });
    },
  });
}
