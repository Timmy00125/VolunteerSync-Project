/**
 * Notification Store
 *
 * Zustand store for managing in-app notifications including:
 * - Unread notification count
 * - Notifications list
 * - Mark as read functionality
 * - Real-time updates
 */

import { create } from 'zustand';
import type { Notification } from '@/lib/api/types';

// ============================================================================
// Types
// ============================================================================

interface NotificationState {
  // State
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;

  // Actions
  setNotifications: (notifications: Notification[]) => void;
  addNotification: (notification: Notification) => void;
  markAsRead: (notificationId: string) => void;
  markAllAsRead: () => void;
  removeNotification: (notificationId: string) => void;
  setLoading: (loading: boolean) => void;
  refreshUnreadCount: () => void;
  reset: () => void;
}

// ============================================================================
// Store
// ============================================================================

export const useNotificationStore = create<NotificationState>((set, get) => ({
  // Initial state
  notifications: [],
  unreadCount: 0,
  isLoading: false,

  // Set all notifications
  setNotifications: (notifications) => {
    const unreadCount = notifications.filter((n) => !n.is_read).length;
    set({
      notifications,
      unreadCount,
    });
  },

  // Add a new notification (e.g., from real-time update)
  addNotification: (notification) => {
    const currentNotifications = get().notifications;

    // Check if notification already exists
    const exists = currentNotifications.some((n) => n.id === notification.id);
    if (exists) {
      return;
    }

    // Add to the beginning of the list
    const newNotifications = [notification, ...currentNotifications];
    const unreadCount = newNotifications.filter((n) => !n.is_read).length;

    set({
      notifications: newNotifications,
      unreadCount,
    });
  },

  // Mark a single notification as read
  markAsRead: (notificationId) => {
    const currentNotifications = get().notifications;
    const updatedNotifications = currentNotifications.map((n) =>
      n.id === notificationId ? { ...n, is_read: true } : n
    );

    const unreadCount = updatedNotifications.filter((n) => !n.is_read).length;

    set({
      notifications: updatedNotifications,
      unreadCount,
    });
  },

  // Mark all notifications as read
  markAllAsRead: () => {
    const currentNotifications = get().notifications;
    const updatedNotifications = currentNotifications.map((n) => ({
      ...n,
      is_read: true,
    }));

    set({
      notifications: updatedNotifications,
      unreadCount: 0,
    });
  },

  // Remove a notification
  removeNotification: (notificationId) => {
    const currentNotifications = get().notifications;
    const updatedNotifications = currentNotifications.filter((n) => n.id !== notificationId);

    const unreadCount = updatedNotifications.filter((n) => !n.is_read).length;

    set({
      notifications: updatedNotifications,
      unreadCount,
    });
  },

  // Set loading state
  setLoading: (loading) =>
    set({
      isLoading: loading,
    }),

  // Refresh unread count from current notifications
  refreshUnreadCount: () => {
    const currentNotifications = get().notifications;
    const unreadCount = currentNotifications.filter((n) => !n.is_read).length;

    set({
      unreadCount,
    });
  },

  // Reset store (e.g., on logout)
  reset: () =>
    set({
      notifications: [],
      unreadCount: 0,
      isLoading: false,
    }),
}));

// ============================================================================
// Selectors
// ============================================================================

/**
 * Select only unread notifications
 */
export const useUnreadNotifications = () =>
  useNotificationStore((state) => state.notifications.filter((n) => !n.is_read));

/**
 * Select only the unread count
 */
export const useUnreadCount = () => useNotificationStore((state) => state.unreadCount);

/**
 * Select notifications by type
 */
export const useNotificationsByType = (type: string) =>
  useNotificationStore((state) => state.notifications.filter((n) => n.type === type));

/**
 * Select the most recent notifications (limit)
 */
export const useRecentNotifications = (limit = 10) =>
  useNotificationStore((state) => state.notifications.slice(0, limit));

/**
 * Check if there are any unread notifications
 */
export const useHasUnreadNotifications = () =>
  useNotificationStore((state) => state.unreadCount > 0);
