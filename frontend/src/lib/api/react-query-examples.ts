/**
 * Example: Using React Query with API Client
 *
 * This file demonstrates how to use TanStack Query with the API client
 * for common data fetching patterns in VolunteerSync.
 *
 * NOTE: This is an example/reference file. Some types and API functions
 * referenced here may not exist yet and will be implemented in future tasks.
 * The examples show the intended patterns and usage.
 */

// @ts-nocheck - This is an example file with placeholder types
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES, invalidateQueries } from './query-client';
import apiClient from './client';
import type { Opportunity, OpportunityCreateInput } from './types';

// ============================================================================
// Example 1: Basic Query - Fetch Opportunities List
// ============================================================================

/**
 * Fetch opportunities from the API
 */
async function fetchOpportunities(filters?: { status?: string }): Promise<Opportunity[]> {
  const params = new URLSearchParams();
  if (filters?.status) params.append('status', filters.status);

  const response = await apiClient.get(`/opportunities?${params.toString()}`);
  return response.data;
}

/**
 * Hook to fetch opportunities with React Query
 */
export function useOpportunities(filters?: { status?: string }) {
  return useQuery({
    queryKey: queryKeys.opportunities.list(filters),
    queryFn: () => fetchOpportunities(filters),
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM, // 5 minutes
  });
}

/**
 * Usage in component:
 *
 * function OpportunityList() {
 *   const { data, isLoading, error } = useOpportunities({ status: 'open' });
 *
 *   if (isLoading) return <Skeleton />;
 *   if (error) return <ErrorMessage error={error} />;
 *
 *   return (
 *     <div>
 *       {data?.map(opp => <OpportunityCard key={opp.id} opportunity={opp} />)}
 *     </div>
 *   );
 * }
 */

// ============================================================================
// Example 2: Query with Detail - Fetch Single Opportunity
// ============================================================================

/**
 * Fetch a single opportunity by ID
 */
async function fetchOpportunity(id: string): Promise<Opportunity> {
  const response = await apiClient.get(`/opportunities/${id}`);
  return response.data;
}

/**
 * Hook to fetch a single opportunity
 */
export function useOpportunity(id: string) {
  return useQuery({
    queryKey: queryKeys.opportunities.detail(id),
    queryFn: () => fetchOpportunity(id),
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM,
    enabled: !!id, // Only fetch if ID is provided
  });
}

/**
 * Usage in component:
 *
 * function OpportunityDetail({ id }: { id: string }) {
 *   const { data: opportunity, isLoading } = useOpportunity(id);
 *
 *   if (isLoading) return <Skeleton />;
 *
 *   return (
 *     <div>
 *       <h1>{opportunity?.title}</h1>
 *       <p>{opportunity?.description}</p>
 *     </div>
 *   );
 * }
 */

// ============================================================================
// Example 3: Mutation - Create Opportunity
// ============================================================================

/**
 * Create a new opportunity
 */
async function createOpportunity(input: OpportunityCreateInput): Promise<Opportunity> {
  const response = await apiClient.post('/opportunities', input);
  return response.data;
}

/**
 * Hook to create an opportunity with automatic cache invalidation
 */
export function useCreateOpportunity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createOpportunity,
    onSuccess: (newOpportunity) => {
      // Invalidate opportunities list to refetch with new data
      invalidateQueries.opportunities();

      // Optionally, set the new opportunity in cache
      queryClient.setQueryData(queryKeys.opportunities.detail(newOpportunity.id), newOpportunity);
    },
  });
}

/**
 * Usage in component:
 *
 * function CreateOpportunityForm() {
 *   const createMutation = useCreateOpportunity();
 *   const { register, handleSubmit } = useForm<OpportunityCreateInput>();
 *
 *   const onSubmit = (data: OpportunityCreateInput) => {
 *     createMutation.mutate(data, {
 *       onSuccess: () => {
 *         toast.success('Opportunity created!');
 *         router.push('/opportunities');
 *       },
 *       onError: (error) => {
 *         toast.error(error.message);
 *       },
 *     });
 *   };
 *
 *   return (
 *     <form onSubmit={handleSubmit(onSubmit)}>
 *       <input {...register('title')} />
 *       <button disabled={createMutation.isPending}>
 *         {createMutation.isPending ? 'Creating...' : 'Create'}
 *       </button>
 *     </form>
 *   );
 * }
 */

// ============================================================================
// Example 4: Optimistic Updates - Log Volunteer Hours
// ============================================================================

/**
 * Log volunteer hours
 */
async function logHours(input: {
  volunteerId: string;
  opportunityId: string;
  hours: number;
  date: string;
}): Promise<any> {
  const response = await apiClient.post('/hours', input);
  return response.data;
}

/**
 * Hook to log hours with optimistic updates
 */
export function useLogHours() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: logHours,
    onMutate: async (variables) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: queryKeys.hours.byVolunteer(variables.volunteerId),
      });

      // Snapshot previous value
      const previousHours = queryClient.getQueryData(
        queryKeys.hours.byVolunteer(variables.volunteerId)
      );

      // Optimistically update with new hours
      queryClient.setQueryData(queryKeys.hours.byVolunteer(variables.volunteerId), (old: any) => {
        if (!old) return old;
        return {
          ...old,
          hours: [
            ...old.hours,
            {
              ...variables,
              id: 'temp-id',
              status: 'pending',
              created_at: new Date().toISOString(),
            },
          ],
        };
      });

      // Return context for rollback
      return { previousHours };
    },
    onError: (err, variables, context) => {
      // Rollback on error
      if (context?.previousHours) {
        queryClient.setQueryData(
          queryKeys.hours.byVolunteer(variables.volunteerId),
          context.previousHours
        );
      }
    },
    onSettled: (data, error, variables) => {
      // Refetch to get real data from server
      queryClient.invalidateQueries({
        queryKey: queryKeys.hours.byVolunteer(variables.volunteerId),
      });
    },
  });
}

/**
 * Usage in component:
 *
 * function LogHoursForm({ volunteerId, opportunityId }) {
 *   const logHoursMutation = useLogHours();
 *
 *   const handleSubmit = (hours: number, date: string) => {
 *     logHoursMutation.mutate(
 *       { volunteerId, opportunityId, hours, date },
 *       {
 *         onSuccess: () => {
 *           toast.success('Hours logged!');
 *         },
 *       }
 *     );
 *   };
 *
 *   return (
 *     <form>
 *       {logHoursMutation.isPending && <Spinner />}
 *       {logHoursMutation.isError && <ErrorMessage error={logHoursMutation.error} />}
 *       <button onClick={() => handleSubmit(2, '2025-10-09')}>
 *         Log 2 Hours
 *       </button>
 *     </form>
 *   );
 * }
 */

// ============================================================================
// Example 5: Dependent Queries - Fetch Organization Members
// ============================================================================

/**
 * Fetch organization members (requires organization ID)
 */
async function fetchOrganizationMembers(orgId: string): Promise<any[]> {
  const response = await apiClient.get(`/organizations/${orgId}/members`);
  return response.data;
}

/**
 * Hook to fetch organization members (dependent on organization being loaded)
 */
export function useOrganizationMembers(orgId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.organizations.members(orgId!),
    queryFn: () => fetchOrganizationMembers(orgId!),
    enabled: !!orgId, // Only fetch when orgId is available
    staleTime: CACHE_TIMES.STALE_TIME.LONG, // Members change infrequently
  });
}

/**
 * Usage with dependent queries:
 *
 * function OrganizationMembersList() {
 *   const { data: currentUser } = useQuery({
 *     queryKey: queryKeys.auth.currentUser(),
 *     queryFn: fetchCurrentUser,
 *   });
 *
 *   // This query waits until currentUser.organizationId is available
 *   const { data: members, isLoading } = useOrganizationMembers(
 *     currentUser?.organizationId
 *   );
 *
 *   if (!currentUser) return <div>Loading user...</div>;
 *   if (isLoading) return <div>Loading members...</div>;
 *
 *   return (
 *     <ul>
 *       {members?.map(member => <li key={member.id}>{member.name}</li>)}
 *     </ul>
 *   );
 * }
 */

// ============================================================================
// Example 6: Pagination with React Query
// ============================================================================

/**
 * Fetch paginated opportunities
 */
async function fetchOpportunitiesPaginated(page: number, limit: number = 10) {
  const response = await apiClient.get(`/opportunities?page=${page}&limit=${limit}`);
  return response.data;
}

/**
 * Hook for paginated opportunities
 */
export function useOpportunitiesPaginated(page: number) {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: queryKeys.opportunities.list({ page }),
    queryFn: () => fetchOpportunitiesPaginated(page),
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM,
  });

  // Prefetch next page for instant navigation
  const prefetchNextPage = () => {
    if (query.data?.pagination.total_pages > page) {
      queryClient.prefetchQuery({
        queryKey: queryKeys.opportunities.list({ page: page + 1 }),
        queryFn: () => fetchOpportunitiesPaginated(page + 1),
      });
    }
  };

  return { ...query, prefetchNextPage };
}

/**
 * Usage in paginated component:
 *
 * function PaginatedOpportunities() {
 *   const [page, setPage] = useState(1);
 *   const { data, isLoading, prefetchNextPage } = useOpportunitiesPaginated(page);
 *
 *   // Prefetch next page on hover
 *   const handleNextHover = () => prefetchNextPage();
 *
 *   return (
 *     <div>
 *       {isLoading ? (
 *         <Skeleton />
 *       ) : (
 *         <div>
 *           {data.opportunities.map(opp => <OpportunityCard key={opp.id} opportunity={opp} />)}
 *           <button onClick={() => setPage(p => p - 1)} disabled={page === 1}>
 *             Previous
 *           </button>
 *           <button
 *             onClick={() => setPage(p => p + 1)}
 *             onMouseEnter={handleNextHover}
 *             disabled={page >= data.pagination.total_pages}
 *           >
 *             Next
 *           </button>
 *         </div>
 *       )}
 *     </div>
 *   );
 * }
 */

// ============================================================================
// Example 7: Real-time Data with Polling
// ============================================================================

/**
 * Fetch unread notifications count
 */
async function fetchUnreadCount(): Promise<number> {
  const response = await apiClient.get('/notifications/unread-count');
  return response.data.count;
}

/**
 * Hook to fetch unread count with polling for real-time updates
 */
export function useUnreadNotificationsCount() {
  return useQuery({
    queryKey: queryKeys.notifications.unreadCount(),
    queryFn: fetchUnreadCount,
    staleTime: CACHE_TIMES.STALE_TIME.SHORT, // 1 minute
    refetchInterval: 60000, // Poll every 60 seconds
    refetchIntervalInBackground: false, // Don't poll when tab is inactive
  });
}

/**
 * Usage in notification badge:
 *
 * function NotificationBadge() {
 *   const { data: unreadCount } = useUnreadNotificationsCount();
 *
 *   return (
 *     <div>
 *       <BellIcon />
 *       {unreadCount > 0 && <Badge>{unreadCount}</Badge>}
 *     </div>
 *   );
 * }
 */
