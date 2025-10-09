/**
 * NotificationBadge Component
 *
 * Displays unread notification count with real-time polling updates.
 *
 * Features:
 * - Shows unread count badge
 * - Auto-updates every 30 seconds
 * - Formats counts over 99 as "99+"
 * - Hides badge when count is 0
 * - Accessible and responsive
 *
 * Usage:
 * ```tsx
 * import { NotificationBadge } from '@/components/features/notifications/NotificationBadge';
 *
 * function Header() {
 *   return (
 *     <button>
 *       <BellIcon />
 *       <NotificationBadge />
 *     </button>
 *   );
 * }
 * ```
 */

'use client';

import * as React from 'react';
import { Badge } from '@/components/ui/badge';
import { useUnreadCount } from '@/lib/api/hooks/notifications';
import { cn } from '@/lib/utils';

export interface NotificationBadgeProps {
  /**
   * Additional CSS classes
   */
  className?: string;
  /**
   * Show badge even when count is 0
   * @default false
   */
  showWhenEmpty?: boolean;
  /**
   * Maximum number to display before showing "+"
   * @default 99
   */
  maxCount?: number;
  /**
   * Badge variant
   * @default "destructive"
   */
  variant?: 'default' | 'secondary' | 'destructive' | 'outline';
}

/**
 * Notification badge showing unread count
 */
export function NotificationBadge({
  className,
  showWhenEmpty = false,
  maxCount = 99,
  variant = 'destructive',
}: NotificationBadgeProps) {
  const { data, isLoading, error } = useUnreadCount();

  // Don't render while loading
  if (isLoading) {
    return null;
  }

  // Don't render if there's an error
  if (error) {
    console.error('Failed to load unread count:', error);
    return null;
  }

  // Get unread count
  const unreadCount = data?.unread_count ?? 0;

  // Don't render if count is 0 and showWhenEmpty is false
  if (unreadCount === 0 && !showWhenEmpty) {
    return null;
  }

  // Format count for display
  const displayCount = unreadCount > maxCount ? `${maxCount}+` : unreadCount.toString();

  return (
    <Badge
      variant={variant}
      className={cn(
        'absolute -top-1 -right-1 h-5 min-w-[1.25rem] px-1 text-[10px] font-semibold',
        'flex items-center justify-center',
        'animate-in fade-in zoom-in duration-200',
        className
      )}
      aria-label={`${unreadCount} unread notification${unreadCount === 1 ? '' : 's'}`}
    >
      {displayCount}
    </Badge>
  );
}

/**
 * Inline notification badge (not absolutely positioned)
 */
export function NotificationBadgeInline({
  className,
  showWhenEmpty = false,
  maxCount = 99,
  variant = 'destructive',
}: NotificationBadgeProps) {
  const { data, isLoading, error } = useUnreadCount();

  if (isLoading) {
    return null;
  }

  if (error) {
    console.error('Failed to load unread count:', error);
    return null;
  }

  const unreadCount = data?.unread_count ?? 0;

  if (unreadCount === 0 && !showWhenEmpty) {
    return null;
  }

  const displayCount = unreadCount > maxCount ? `${maxCount}+` : unreadCount.toString();

  return (
    <Badge
      variant={variant}
      className={cn(
        'h-5 min-w-[1.25rem] px-1.5 text-[10px] font-semibold',
        'animate-in fade-in zoom-in duration-200',
        className
      )}
      aria-label={`${unreadCount} unread notification${unreadCount === 1 ? '' : 's'}`}
    >
      {displayCount}
    </Badge>
  );
}
