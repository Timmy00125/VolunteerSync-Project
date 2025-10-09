/**
 * NotificationCenter Component
 *
 * Dropdown menu displaying notifications with filtering and actions.
 *
 * Features:
 * - Dropdown with unread notifications
 * - Shows different notification types with icons
 * - Mark as read action
 * - Click to navigate to relevant page
 * - Empty state for no notifications
 * - Auto-refreshes every 30 seconds
 * - Loading and error states
 *
 * Usage:
 * ```tsx
 * import { NotificationCenter } from '@/components/features/notifications/NotificationCenter';
 *
 * function Header() {
 *   return (
 *     <div className="flex items-center gap-4">
 *       <NotificationCenter />
 *     </div>
 *   );
 * }
 * ```
 */

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import {
  Bell,
  Calendar,
  Clock,
  Trophy,
  AlertCircle,
  MessageSquare,
  CheckCircle,
  XCircle,
  Loader2,
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { NotificationBadge } from './NotificationBadge';
import { useNotifications, useMarkNotificationAsRead } from '@/lib/api/hooks/notifications';
import type { Notification } from '@/lib/api/types';
import { cn } from '@/lib/utils';

export interface NotificationCenterProps {
  /**
   * Additional CSS classes for the trigger button
   */
  className?: string;
  /**
   * Show all notifications or only unread
   * @default false (show only unread)
   */
  showAll?: boolean;
  /**
   * Maximum notifications to display
   * @default 10
   */
  maxNotifications?: number;
}

/**
 * Get icon for notification type
 */
function getNotificationIcon(type: string) {
  switch (type) {
    case 'registration_confirmed':
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    case 'event_reminder':
      return <Clock className="h-4 w-4 text-blue-500" />;
    case 'hours_logged':
      return <Calendar className="h-4 w-4 text-purple-500" />;
    case 'achievement_earned':
      return <Trophy className="h-4 w-4 text-yellow-500" />;
    case 'message_received':
      return <MessageSquare className="h-4 w-4 text-blue-500" />;
    case 'event_cancelled':
      return <XCircle className="h-4 w-4 text-red-500" />;
    case 'waitlist_notification':
      return <AlertCircle className="h-4 w-4 text-orange-500" />;
    case 'hours_disputed':
      return <AlertCircle className="h-4 w-4 text-orange-500" />;
    case 'hours_verified':
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    default:
      return <Bell className="h-4 w-4 text-muted-foreground" />;
  }
}

/**
 * Format relative time from timestamp
 */
function formatRelativeTime(timestamp: string): string {
  const now = new Date();
  const date = new Date(timestamp);
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
}

/**
 * Individual notification item
 */
function NotificationItem({
  notification,
  onMarkAsRead,
}: {
  notification: Notification;
  onMarkAsRead: (id: string) => void;
}) {
  const router = useRouter();
  const isUnread = !notification.read_at;

  const handleClick = () => {
    // Mark as read
    if (isUnread) {
      onMarkAsRead(notification.id);
    }

    // Navigate if action URL exists
    if (notification.action_url) {
      router.push(notification.action_url);
    }
  };

  return (
    <DropdownMenuItem
      className={cn('flex items-start gap-3 p-3 cursor-pointer', isUnread && 'bg-accent/50')}
      onClick={handleClick}
    >
      <div className="mt-0.5 shrink-0">{getNotificationIcon(notification.notification_type)}</div>

      <div className="flex-1 space-y-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <p className={cn('text-sm leading-tight', isUnread && 'font-semibold')}>
            {notification.title}
          </p>
          {isUnread && <div className="h-2 w-2 rounded-full bg-blue-500 shrink-0 mt-1" />}
        </div>

        <p className="text-xs text-muted-foreground line-clamp-2">{notification.content}</p>

        <p className="text-xs text-muted-foreground">{formatRelativeTime(notification.sent_at)}</p>
      </div>
    </DropdownMenuItem>
  );
}

/**
 * Empty state for no notifications
 */
function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
      <Bell className="h-12 w-12 text-muted-foreground/50 mb-3" />
      <p className="text-sm font-medium text-muted-foreground">No notifications</p>
      <p className="text-xs text-muted-foreground mt-1">You&apos;re all caught up!</p>
    </div>
  );
}

/**
 * Notification center dropdown
 */
export function NotificationCenter({
  className,
  showAll = false,
  maxNotifications = 10,
}: NotificationCenterProps) {
  const [isOpen, setIsOpen] = React.useState(false);
  const router = useRouter();

  // Fetch notifications
  const { data, isLoading, error } = useNotifications({
    unread: !showAll,
    limit: maxNotifications,
  });

  // Mark as read mutation
  const markAsRead = useMarkNotificationAsRead();

  const handleMarkAsRead = (notificationId: string) => {
    markAsRead.mutate(notificationId);
  };

  const handleViewAll = () => {
    setIsOpen(false);
    router.push('/notifications');
  };

  const notifications = data?.notifications ?? [];
  const hasNotifications = notifications.length > 0;

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className={cn('relative', className)}
          aria-label="Notifications"
        >
          <Bell className="h-4 w-4" />
          <NotificationBadge />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-80 md:w-96 p-0">
        <DropdownMenuLabel className="px-4 py-3 border-b">
          <div className="flex items-center justify-between">
            <span className="font-semibold">Notifications</span>
            {data?.unread_count ? (
              <span className="text-xs text-muted-foreground">{data.unread_count} unread</span>
            ) : null}
          </div>
        </DropdownMenuLabel>

        <ScrollArea className="max-h-[400px]">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : error ? (
            <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
              <AlertCircle className="h-12 w-12 text-destructive/50 mb-3" />
              <p className="text-sm font-medium text-destructive">Failed to load notifications</p>
              <p className="text-xs text-muted-foreground mt-1">Please try again later</p>
            </div>
          ) : hasNotifications ? (
            <div className="divide-y">
              {notifications.map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  onMarkAsRead={handleMarkAsRead}
                />
              ))}
            </div>
          ) : (
            <EmptyState />
          )}
        </ScrollArea>

        {hasNotifications && (
          <>
            <DropdownMenuSeparator />
            <div className="p-2">
              <Button
                variant="ghost"
                className="w-full justify-center text-sm"
                onClick={handleViewAll}
              >
                View all notifications
              </Button>
            </div>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
