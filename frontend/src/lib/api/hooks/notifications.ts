/**
 * Notification API Hooks
 *
 * React Query hooks for notification-related data fetching.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES } from '../query-client';
import { getNotifications, getUnreadCount, markNotificationAsRead } from '../client';
import type { NotificationsResponse, NotificationFilters } from '../types';

// ============================================================================
// Notifications List Hook
// ============================================================================

/**
 * Hook to fetch notifications for the authenticated user
 *
 * Fetches paginated notifications with optional filters:
 * - unread: filter by read/unread status
 * - type: filter by notification type
 * - priority: filter by priority level
 * - page/limit: pagination controls
 *
 * @param filters - Optional filters for notifications
 * @returns React Query result with notifications data
 *
 * @example
 * ```tsx
 * function NotificationsList() {
 *   const { data, isLoading, error } = useNotifications({ unread: true, limit: 10 });
 *
 *   if (isLoading) return <Skeleton />;
 *   if (error) return <ErrorMessage error={error} />;
 *
 *   return (
 *     <ul>
 *       {data.notifications.map(notif => (
 *         <NotificationItem key={notif.id} notification={notif} />
 *       ))}
 *     </ul>
 *   );
 * }
 * ```
 */
export function useNotifications(filters?: NotificationFilters) {
  return useQuery<NotificationsResponse>({
    queryKey: queryKeys.notifications.list(filters as Record<string, unknown>),
    queryFn: () => getNotifications(filters),
    staleTime: CACHE_TIMES.STALE_TIME.SHORT, // 1 minute - notifications change frequently
    gcTime: CACHE_TIMES.CACHE_TIME.SHORT, // 5 minutes
    refetchInterval: 30000, // Poll every 30 seconds for real-time updates
  });
}

// ============================================================================
// Unread Count Hook
// ============================================================================

/**
 * Hook to fetch unread notification count
 *
 * Optimized for frequent polling to show real-time unread badge count.
 * Uses shorter cache times and automatic polling.
 *
 * @returns React Query result with unread count
 *
 * @example
 * ```tsx
 * function NotificationBadge() {
 *   const { data, isLoading } = useUnreadCount();
 *
 *   if (isLoading) return null;
 *   if (!data || data.unread_count === 0) return null;
 *
 *   return (
 *     <Badge variant="destructive">
 *       {data.unread_count > 99 ? '99+' : data.unread_count}
 *     </Badge>
 *   );
 * }
 * ```
 */
export function useUnreadCount() {
  return useQuery<{ unread_count: number }>({
    queryKey: queryKeys.notifications.unreadCount(),
    queryFn: getUnreadCount,
    staleTime: CACHE_TIMES.STALE_TIME.SHORT, // 1 minute
    gcTime: CACHE_TIMES.CACHE_TIME.SHORT, // 5 minutes
    refetchInterval: 30000, // Poll every 30 seconds
    refetchOnWindowFocus: true, // Refetch when user returns to tab
  });
}

// ============================================================================
// Mark as Read Mutation
// ============================================================================

/**
 * Hook to mark a notification as read
 *
 * Optimistically updates the cache and invalidates notification queries
 * to ensure UI consistency.
 *
 * @returns React Query mutation for marking notification as read
 *
 * @example
 * ```tsx
 * function NotificationItem({ notification }) {
 *   const markAsRead = useMarkNotificationAsRead();
 *
 *   const handleClick = () => {
 *     markAsRead.mutate(notification.id, {
 *       onSuccess: () => {
 *         // Navigate to related page
 *         if (notification.action_url) {
 *           router.push(notification.action_url);
 *         }
 *       },
 *       onError: (error) => {
 *         toast.error('Failed to mark notification as read');
 *       },
 *     });
 *   };
 *
 *   return (
 *     <button onClick={handleClick}>
 *       {notification.title}
 *     </button>
 *   );
 * }
 * ```
 */
export function useMarkNotificationAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (notificationId: string) => markNotificationAsRead(notificationId),
    onSuccess: () => {
      // Invalidate notification queries to refetch updated data
      queryClient.invalidateQueries({
        queryKey: queryKeys.notifications.all,
      });
    },
  });
}
