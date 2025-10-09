# React Query (TanStack Query) Configuration

This directory contains the React Query setup for VolunteerSync, providing efficient server state management, caching, and synchronization.

## Overview

**File**: `query-client.ts`

The query client is configured with:

- **Smart caching** with optimized stale/cache times for different data types
- **Automatic retries** with exponential backoff
- **Global error handling** for consistent error management
- **Query key factories** for type-safe cache key generation
- **Invalidation helpers** for easy cache updates
- **Optimistic update utilities** for immediate UI feedback

## Usage

### 1. Setup in App Layout

Wrap your application with `QueryClientProvider`:

```tsx
// app/layout.tsx
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from '@/lib/api/query-client';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      </body>
    </html>
  );
}
```

### 2. Enable DevTools (Development Only)

Add React Query DevTools for debugging:

```tsx
// app/layout.tsx
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

<QueryClientProvider client={queryClient}>
  {children}
  {process.env.NODE_ENV === 'development' && <ReactQueryDevtools initialIsOpen={false} />}
</QueryClientProvider>;
```

### 3. Using Queries

Use the query key factories for consistent cache keys:

```tsx
// components/opportunities/OpportunityList.tsx
import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '@/lib/api/query-client';
import { fetchOpportunities } from '@/lib/api/opportunities';

export function OpportunityList() {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.opportunities.list({ status: 'open' }),
    queryFn: () => fetchOpportunities({ status: 'open' }),
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {data.opportunities.map((opp) => (
        <OpportunityCard key={opp.id} opportunity={opp} />
      ))}
    </div>
  );
}
```

### 4. Using Mutations

Handle mutations with automatic cache invalidation:

```tsx
// components/opportunities/CreateOpportunityForm.tsx
import { useMutation } from '@tanstack/react-query';
import { invalidateQueries } from '@/lib/api/query-client';
import { createOpportunity } from '@/lib/api/opportunities';

export function CreateOpportunityForm() {
  const mutation = useMutation({
    mutationFn: createOpportunity,
    onSuccess: () => {
      // Invalidate and refetch opportunities list
      invalidateQueries.opportunities();
    },
  });

  const handleSubmit = (data: OpportunityInput) => {
    mutation.mutate(data);
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* form fields */}
      <button disabled={mutation.isPending}>{mutation.isPending ? 'Creating...' : 'Create'}</button>
    </form>
  );
}
```

### 5. Optimistic Updates

Provide immediate feedback with optimistic updates:

```tsx
// components/hours/LogHoursForm.tsx
import { useMutation } from '@tanstack/react-query';
import { queryKeys, createOptimisticUpdate, queryClient } from '@/lib/api/query-client';
import { logHours } from '@/lib/api/hours';

export function LogHoursForm() {
  const mutation = useMutation({
    mutationFn: logHours,
    onMutate: async (variables) => {
      // Optimistically update the hours list
      return createOptimisticUpdate(queryKeys.hours.byVolunteer(variables.volunteerId), (old) => {
        if (!old) return old;
        return {
          ...old,
          hours: [...old.hours, { ...variables, id: 'temp-id', status: 'pending' }],
        };
      });
    },
    onError: (err, variables, context) => {
      // Rollback on error
      if (context?.previousData) {
        queryClient.setQueryData(
          queryKeys.hours.byVolunteer(variables.volunteerId),
          context.previousData
        );
      }
    },
    onSuccess: () => {
      // Refetch to get the real data
      invalidateQueries.hours();
    },
  });

  return <form onSubmit={(data) => mutation.mutate(data)}>...</form>;
}
```

### 6. Prefetching Data

Prefetch data before navigation for instant page loads:

```tsx
// components/opportunities/OpportunityCard.tsx
import { useRouter } from 'next/navigation';
import { prefetchOpportunity } from '@/lib/api/query-client';

export function OpportunityCard({ opportunity }) {
  const router = useRouter();

  const handleClick = async () => {
    // Prefetch opportunity details
    await prefetchOpportunity(opportunity.id);
    // Navigate (data is already cached)
    router.push(`/opportunities/${opportunity.id}`);
  };

  return (
    <div onClick={handleClick}>
      <h3>{opportunity.title}</h3>
    </div>
  );
}
```

## Cache Strategy

Different data types have different cache configurations:

| Data Type     | Stale Time | Cache Time | Reasoning                                                   |
| ------------- | ---------- | ---------- | ----------------------------------------------------------- |
| Notifications | 1 minute   | 5 minutes  | Frequently changing, users expect fresh data                |
| Opportunities | 5 minutes  | 30 minutes | Moderate changes, balance between freshness and performance |
| Organizations | 30 minutes | 1 hour     | Rarely changes, safe to cache longer                        |
| User Profiles | 30 minutes | 1 hour     | Rarely changes, safe to cache longer                        |
| Hours Logs    | 5 minutes  | 30 minutes | Important data, users expect relatively fresh info          |

### Customizing Cache Times

Override default cache times per query:

```tsx
import { useQuery } from '@tanstack/react-query';
import { CACHE_TIMES, queryKeys } from '@/lib/api/query-client';

// Short cache for real-time data
const { data } = useQuery({
  queryKey: queryKeys.notifications.unreadCount(),
  queryFn: fetchUnreadCount,
  staleTime: CACHE_TIMES.STALE_TIME.SHORT,
  gcTime: CACHE_TIMES.CACHE_TIME.SHORT,
});

// Long cache for rarely changing data
const { data } = useQuery({
  queryKey: queryKeys.organizations.detail(orgId),
  queryFn: () => fetchOrganization(orgId),
  staleTime: CACHE_TIMES.STALE_TIME.LONG,
  gcTime: CACHE_TIMES.CACHE_TIME.LONG,
});
```

## Query Key Factories

Query key factories ensure consistent cache keys across the application:

```typescript
// All query keys follow a hierarchical structure
queryKeys.opportunities.all; // ['opportunities']
queryKeys.opportunities.lists(); // ['opportunities', 'list']
queryKeys.opportunities.list({ status: 'open' }); // ['opportunities', 'list', { status: 'open' }]
queryKeys.opportunities.detail('opp_123'); // ['opportunities', 'detail', 'opp_123']
```

Benefits:

- **Type-safe**: TypeScript ensures correct key structure
- **Consistent**: Same key structure everywhere
- **Easy invalidation**: Invalidate all related queries at once
- **Refactor-friendly**: Change key structure in one place

## Error Handling

Errors are handled globally by the query client:

1. **Automatic retries** for network errors and 5xx server errors
2. **No retries** for 4xx client errors (except 408, 429)
3. **Global error logging** for monitoring and debugging
4. **Component-level error handling** via `error` property

```tsx
const { data, error, isLoading } = useQuery({
  queryKey: queryKeys.opportunities.detail(id),
  queryFn: () => fetchOpportunity(id),
});

if (error) {
  // Error is typed as ApiError
  return <ErrorMessage message={error.message} />;
}
```

## Best Practices

### 1. Use Query Key Factories

Always use the provided query key factories instead of string keys:

```tsx
// ✅ Good
queryKey: queryKeys.opportunities.list({ status: 'open' });

// ❌ Bad
queryKey: ['opportunities', 'list', { status: 'open' }];
```

### 2. Invalidate Related Queries

After mutations, invalidate related queries:

```tsx
onSuccess: () => {
  // Invalidate all opportunity queries
  invalidateQueries.opportunities();

  // Or be more specific
  queryClient.invalidateQueries({
    queryKey: queryKeys.opportunities.list(),
  });
};
```

### 3. Handle Loading and Error States

Always handle loading and error states in components:

```tsx
if (isLoading) return <Skeleton />;
if (error) return <ErrorBoundary error={error} />;
return <Content data={data} />;
```

### 4. Use Optimistic Updates for Better UX

For actions that are likely to succeed (like logging hours, submitting registrations):

```tsx
// Show immediate feedback, rollback on error
onMutate: async (variables) => {
  return createOptimisticUpdate(queryKey, updater);
},
onError: (err, variables, context) => {
  queryClient.setQueryData(queryKey, context.previousData);
},
```

### 5. Prefetch for Instant Navigation

Prefetch data when user hovers over links:

```tsx
<Link href={`/opportunities/${id}`} onMouseEnter={() => prefetchOpportunity(id)}>
  View Details
</Link>
```

## Troubleshooting

### Query Not Updating After Mutation

**Problem**: Data doesn't refresh after creating/updating

**Solution**: Ensure you're invalidating queries:

```tsx
onSuccess: () => {
  invalidateQueries.opportunities(); // Invalidate all opportunity queries
};
```

### Stale Data Showing

**Problem**: Old data showing even after update

**Solution**: Check stale time configuration or force refetch:

```tsx
queryClient.invalidateQueries({ queryKey: queryKeys.opportunities.all });
// or
refetch(); // In component
```

### Too Many Network Requests

**Problem**: Same data fetched multiple times

**Solution**: Increase stale time or use prefetching:

```tsx
staleTime: CACHE_TIMES.STALE_TIME.LONG, // Cache for 30 minutes
```

### Memory Issues

**Problem**: Too much data in cache

**Solution**: Reduce cache time or use pagination:

```tsx
gcTime: CACHE_TIMES.CACHE_TIME.SHORT, // Clean up after 5 minutes
```

## Related Files

- `client.ts` - API client with fetch wrapper and token management
- `types.ts` - TypeScript types for API responses and errors
- `examples.ts` - Example API implementations using React Query

## Resources

- [TanStack Query Documentation](https://tanstack.com/query/latest)
- [Best Practices Guide](https://tanstack.com/query/latest/docs/react/guides/important-defaults)
- [Community Best Practices](https://tkdodo.eu/blog/practical-react-query)
