/**
 * Store Index
 *
 * Centralized exports for all Zustand stores.
 */

// Auth store
export {
  useAuthStore,
  useUser,
  useIsAuthenticated,
  useTokens,
  useAccessToken,
  useRefreshToken,
  useUserType,
  useIsVolunteer,
  useIsOrgAdmin,
  useIsSuperAdmin,
} from './auth-store';

// Notification store
export {
  useNotificationStore,
  useUnreadNotifications,
  useUnreadCount,
  useNotificationsByType,
  useRecentNotifications,
  useHasUnreadNotifications,
} from './notification-store';
