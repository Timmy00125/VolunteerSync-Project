/**
 * Volunteer API Hooks
 *
 * React Query hooks for volunteer-related data fetching.
 */

import { useQuery } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES } from '../query-client';
import { getVolunteerDashboard } from '../client';
import type { DashboardResponse } from '../types';

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
