/**
 * Opportunities API Hooks
 *
 * React Query hooks for opportunity-related data fetching.
 */

import { useQuery } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES } from '../query-client';
import { searchOpportunities, getOpportunityById } from '../client';
import type { Opportunity, OpportunitySearchParams, PaginatedResponse } from '../types';

// ============================================================================
// Search Hook
// ============================================================================

/**
 * Hook to search opportunities with filters
 *
 * Fetches paginated opportunities based on search criteria:
 * - Text search
 * - Location-based search (lat/lng/radius)
 * - Cause filter
 * - Date range filter
 * - Skills filter
 * - Age requirements
 *
 * @param params - Search parameters
 * @returns React Query result with paginated opportunities
 *
 * @example
 * ```tsx
 * function OpportunitySearchPage() {
 *   const [filters, setFilters] = useState<OpportunitySearchParams>({
 *     search: '',
 *     latitude: 37.7749,
 *     longitude: -122.4194,
 *     radius: 25,
 *     page: 1,
 *     limit: 20,
 *   });
 *
 *   const { data, isLoading, error } = useSearchOpportunities(filters);
 *
 *   if (isLoading) return <Skeleton />;
 *   if (error) return <ErrorMessage error={error} />;
 *
 *   return (
 *     <div>
 *       <OpportunityFilters onChange={setFilters} />
 *       <OpportunityList opportunities={data.data} />
 *       <Pagination {...data.pagination} />
 *     </div>
 *   );
 * }
 * ```
 */
export function useSearchOpportunities(params?: OpportunitySearchParams) {
  return useQuery<PaginatedResponse<Opportunity>>({
    queryKey: queryKeys.opportunities.list((params as Record<string, unknown>) || {}),
    queryFn: () => searchOpportunities(params),
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM, // 5 minutes
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM, // 30 minutes
    // Keep previous data while fetching new results for smoother UX
    placeholderData: (previousData) => previousData,
  });
}

// ============================================================================
// Detail Hook
// ============================================================================

/**
 * Hook to fetch a single opportunity by ID
 *
 * @param id - Opportunity ID
 * @param enabled - Whether the query should execute (default: true)
 * @returns React Query result with opportunity details
 *
 * @example
 * ```tsx
 * function OpportunityDetailPage({ opportunityId }: { opportunityId: string }) {
 *   const { data: opportunity, isLoading, error } = useOpportunity(opportunityId);
 *
 *   if (isLoading) return <Skeleton />;
 *   if (error) return <ErrorMessage error={error} />;
 *   if (!opportunity) return <NotFound />;
 *
 *   return (
 *     <div>
 *       <h1>{opportunity.title}</h1>
 *       <p>{opportunity.description}</p>
 *       <RegisterButton opportunityId={opportunity.id} />
 *     </div>
 *   );
 * }
 * ```
 */
export function useOpportunity(id: string, enabled = true) {
  return useQuery<Opportunity>({
    queryKey: queryKeys.opportunities.detail(id),
    queryFn: () => getOpportunityById(id),
    staleTime: CACHE_TIMES.STALE_TIME.MEDIUM, // 5 minutes
    gcTime: CACHE_TIMES.CACHE_TIME.MEDIUM, // 30 minutes
    enabled: enabled && !!id,
  });
}
