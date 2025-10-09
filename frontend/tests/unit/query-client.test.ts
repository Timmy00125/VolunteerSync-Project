/**
 * Tests for React Query Configuration
 *
 * Verifies that the query client is properly configured with:
 * - Correct default options
 * - Query key factories
 * - Cache invalidation helpers
 */

import { queryClient, queryKeys, CACHE_TIMES } from '@/lib/api/query-client';

describe('Query Client Configuration', () => {
  beforeEach(() => {
    // Clear query cache before each test
    queryClient.clear();
  });

  afterAll(() => {
    // Clean up query client to avoid Jest hanging
    queryClient.clear();
  });

  describe('Query Client Instance', () => {
    it('should be defined', () => {
      expect(queryClient).toBeDefined();
    });

    it('should have default options configured', () => {
      const defaultOptions = queryClient.getDefaultOptions();

      expect(defaultOptions.queries?.staleTime).toBe(CACHE_TIMES.STALE_TIME.MEDIUM);
      expect(defaultOptions.queries?.gcTime).toBe(CACHE_TIMES.CACHE_TIME.MEDIUM);
      expect(defaultOptions.queries?.refetchOnWindowFocus).toBe(true);
      expect(defaultOptions.queries?.refetchOnReconnect).toBe(true);
    });
  });

  describe('Query Key Factories', () => {
    it('should generate correct auth query keys', () => {
      expect(queryKeys.auth.all).toEqual(['auth']);
      expect(queryKeys.auth.currentUser()).toEqual(['auth', 'current-user']);
    });

    it('should generate correct opportunities query keys', () => {
      expect(queryKeys.opportunities.all).toEqual(['opportunities']);
      expect(queryKeys.opportunities.lists()).toEqual(['opportunities', 'list']);
      expect(queryKeys.opportunities.list({ status: 'open' })).toEqual([
        'opportunities',
        'list',
        { status: 'open' },
      ]);
      expect(queryKeys.opportunities.detail('opp_123')).toEqual([
        'opportunities',
        'detail',
        'opp_123',
      ]);
    });

    it('should generate correct organizations query keys', () => {
      expect(queryKeys.organizations.all).toEqual(['organizations']);
      expect(queryKeys.organizations.detail('org_123')).toEqual([
        'organizations',
        'detail',
        'org_123',
      ]);
      expect(queryKeys.organizations.members('org_123')).toEqual([
        'organizations',
        'detail',
        'org_123',
        'members',
      ]);
    });

    it('should generate correct hours query keys', () => {
      expect(queryKeys.hours.byVolunteer('vol_123')).toEqual(['hours', 'by-volunteer', 'vol_123']);
      expect(queryKeys.hours.summary('vol_123')).toEqual([
        'hours',
        'by-volunteer',
        'vol_123',
        'summary',
      ]);
    });

    it('should generate correct notifications query keys', () => {
      expect(queryKeys.notifications.unreadCount()).toEqual(['notifications', 'unread-count']);
    });
  });

  describe('Cache Time Constants', () => {
    it('should have SHORT stale time of 1 minute', () => {
      expect(CACHE_TIMES.STALE_TIME.SHORT).toBe(60000);
    });

    it('should have MEDIUM stale time of 5 minutes', () => {
      expect(CACHE_TIMES.STALE_TIME.MEDIUM).toBe(300000);
    });

    it('should have LONG stale time of 30 minutes', () => {
      expect(CACHE_TIMES.STALE_TIME.LONG).toBe(1800000);
    });

    it('should have SHORT cache time of 5 minutes', () => {
      expect(CACHE_TIMES.CACHE_TIME.SHORT).toBe(300000);
    });

    it('should have MEDIUM cache time of 30 minutes', () => {
      expect(CACHE_TIMES.CACHE_TIME.MEDIUM).toBe(1800000);
    });

    it('should have LONG cache time of 1 hour', () => {
      expect(CACHE_TIMES.CACHE_TIME.LONG).toBe(3600000);
    });
  });

  describe('Query Data Management', () => {
    it('should set and get query data', () => {
      const testData = { id: '1', name: 'Test' };
      const queryKey = queryKeys.opportunities.detail('1');

      queryClient.setQueryData(queryKey, testData);

      const retrievedData = queryClient.getQueryData(queryKey);
      expect(retrievedData).toEqual(testData);
    });

    it('should invalidate queries', async () => {
      const testData = { id: '1', name: 'Test' };
      const queryKey = queryKeys.opportunities.detail('1');

      queryClient.setQueryData(queryKey, testData);

      await queryClient.invalidateQueries({ queryKey: queryKeys.opportunities.all });

      // After invalidation, queries should be marked as stale
      const queryState = queryClient.getQueryState(queryKey);
      expect(queryState?.isInvalidated).toBe(true);
    });

    it('should clear all queries', () => {
      queryClient.setQueryData(queryKeys.opportunities.detail('1'), { id: '1' });
      queryClient.setQueryData(queryKeys.organizations.detail('1'), { id: '1' });

      queryClient.clear();

      expect(queryClient.getQueryData(queryKeys.opportunities.detail('1'))).toBeUndefined();
      expect(queryClient.getQueryData(queryKeys.organizations.detail('1'))).toBeUndefined();
    });
  });

  describe('Hierarchical Query Keys', () => {
    it('should support hierarchical invalidation', async () => {
      // Set multiple opportunities queries
      queryClient.setQueryData(queryKeys.opportunities.list({ status: 'open' }), []);
      queryClient.setQueryData(queryKeys.opportunities.list({ status: 'closed' }), []);
      queryClient.setQueryData(queryKeys.opportunities.detail('opp_123'), {});

      // Invalidate all opportunities
      await queryClient.invalidateQueries({ queryKey: queryKeys.opportunities.all });

      // All opportunities queries should be invalidated
      const listOpenState = queryClient.getQueryState(
        queryKeys.opportunities.list({ status: 'open' })
      );
      const listClosedState = queryClient.getQueryState(
        queryKeys.opportunities.list({ status: 'closed' })
      );
      const detailState = queryClient.getQueryState(queryKeys.opportunities.detail('opp_123'));

      expect(listOpenState?.isInvalidated).toBe(true);
      expect(listClosedState?.isInvalidated).toBe(true);
      expect(detailState?.isInvalidated).toBe(true);
    });

    it('should support granular invalidation', async () => {
      // Set multiple opportunities queries
      queryClient.setQueryData(queryKeys.opportunities.list({ status: 'open' }), []);
      queryClient.setQueryData(queryKeys.opportunities.detail('opp_123'), {});

      // Invalidate only list queries
      await queryClient.invalidateQueries({ queryKey: queryKeys.opportunities.lists() });

      // Only list queries should be invalidated
      const listState = queryClient.getQueryState(queryKeys.opportunities.list({ status: 'open' }));
      const detailState = queryClient.getQueryState(queryKeys.opportunities.detail('opp_123'));

      expect(listState?.isInvalidated).toBe(true);
      expect(detailState?.isInvalidated).toBe(false);
    });
  });
});
