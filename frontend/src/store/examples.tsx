/**
 * Store Usage Examples
 *
 * This file contains practical examples of how to use the Zustand stores
 * in the VolunteerSync application.
 */

import {
  useAuthStore,
  useUser,
  useIsAuthenticated,
  useNotificationStore,
  useUnreadCount,
} from '@/store';
import type { User, AuthTokens, Notification } from '@/lib/api/types';

// ============================================================================
// Example 1: Login Component
// ============================================================================

export function LoginExample() {
  const { login, isLoading } = useAuthStore();

  const handleLogin = async (email: string, password: string) => {
    try {
      // Call API (this would be in your API client)
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) throw new Error('Login failed');

      const data = await response.json();
      const { user, tokens } = data;

      // Update auth store
      login(user, tokens);

      // Navigate to dashboard (use your router)
      // router.push('/dashboard');
    } catch (error) {
      console.error('Login error:', error);
    }
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        handleLogin(formData.get('email') as string, formData.get('password') as string);
      }}
    >
      <input name="email" type="email" placeholder="Email" required />
      <input name="password" type="password" placeholder="Password" required />
      <button type="submit" disabled={isLoading}>
        {isLoading ? 'Logging in...' : 'Login'}
      </button>
    </form>
  );
}

// ============================================================================
// Example 2: Protected Route Component
// ============================================================================

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useIsAuthenticated();

  if (!isAuthenticated) {
    // Redirect to login (use your router)
    // return <Navigate to="/login" />;
    return <div>Please login to continue</div>;
  }

  return <>{children}</>;
}

// ============================================================================
// Example 3: User Profile Display
// ============================================================================

export function UserProfileDisplay() {
  const user = useUser();
  const { updateUser } = useAuthStore();

  const handleUpdateName = async (firstName: string, lastName: string) => {
    try {
      // Call API to update profile
      const response = await fetch('/api/v1/users/me', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ first_name: firstName, last_name: lastName }),
      });

      if (!response.ok) throw new Error('Update failed');

      // Update local store
      updateUser({ first_name: firstName, last_name: lastName });
    } catch (error) {
      console.error('Update error:', error);
    }
  };

  if (!user) return null;

  return (
    <div>
      <h2>Profile</h2>
      <p>
        Name: {user.first_name} {user.last_name}
      </p>
      <p>Email: {user.email}</p>
      <p>Type: {user.user_type}</p>
      <p>Status: {user.status}</p>
    </div>
  );
}

// ============================================================================
// Example 4: Logout Component
// ============================================================================

export function LogoutButton() {
  const { logout } = useAuthStore();
  const { reset: resetNotifications } = useNotificationStore();

  const handleLogout = async () => {
    try {
      // Call API to logout
      await fetch('/api/v1/auth/logout', { method: 'POST' });
    } finally {
      // Always clear local state
      logout();
      resetNotifications();

      // Navigate to login
      // router.push('/login');
    }
  };

  return <button onClick={handleLogout}>Logout</button>;
}

// ============================================================================
// Example 5: Notification Badge
// ============================================================================

export function NotificationBadge() {
  const unreadCount = useUnreadCount();

  return (
    <button className="notification-button">
      <BellIcon />
      {unreadCount > 0 && <span className="badge">{unreadCount}</span>}
    </button>
  );
}

// ============================================================================
// Example 6: Notification List
// ============================================================================

export function NotificationList() {
  const { notifications, markAsRead, markAllAsRead } = useNotificationStore();

  const handleNotificationClick = async (notification: Notification) => {
    // Mark as read in API
    try {
      await fetch(`/api/v1/notifications/${notification.id}/read`, {
        method: 'PATCH',
      });

      // Update local store
      markAsRead(notification.id);

      // Navigate to related entity if exists
      if (notification.related_entity_type && notification.related_entity_id) {
        // router.push(`/${notification.related_entity_type}/${notification.related_entity_id}`);
      }
    } catch (error) {
      console.error('Error marking notification as read:', error);
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      // Call API to mark all as read
      await fetch('/api/v1/notifications/read-all', { method: 'PATCH' });

      // Update local store
      markAllAsRead();
    } catch (error) {
      console.error('Error marking all as read:', error);
    }
  };

  return (
    <div className="notification-list">
      <div className="notification-header">
        <h3>Notifications</h3>
        {notifications.length > 0 && (
          <button onClick={handleMarkAllAsRead}>Mark all as read</button>
        )}
      </div>

      <div className="notification-items">
        {notifications.length === 0 ? (
          <p>No notifications</p>
        ) : (
          notifications.map((notification) => (
            <div
              key={notification.id}
              className={`notification-item ${!notification.is_read ? 'unread' : ''}`}
              onClick={() => handleNotificationClick(notification)}
            >
              <h4>{notification.title}</h4>
              <p>{notification.message}</p>
              <small>{new Date(notification.created_at).toLocaleString()}</small>
              {!notification.is_read && <span className="unread-indicator" />}
            </div>
          ))
        )}
      </div>
    </div>
  );
}

// ============================================================================
// Example 7: Real-time Notification Polling
// ============================================================================

export function useNotificationPolling(intervalMs = 30000) {
  const { addNotification, setNotifications } = useNotificationStore();
  const isAuthenticated = useIsAuthenticated();

  // Poll for new notifications
  React.useEffect(() => {
    if (!isAuthenticated) return;

    const fetchNotifications = async () => {
      try {
        const response = await fetch('/api/v1/notifications');
        if (!response.ok) return;

        const notifications: Notification[] = await response.json();
        setNotifications(notifications);
      } catch (error) {
        console.error('Error fetching notifications:', error);
      }
    };

    // Fetch immediately
    fetchNotifications();

    // Then poll at interval
    const interval = setInterval(fetchNotifications, intervalMs);

    return () => clearInterval(interval);
  }, [isAuthenticated, intervalMs, setNotifications]);
}

// ============================================================================
// Example 8: Auth Persistence and Hydration
// ============================================================================

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { isLoading, setLoading, tokens } = useAuthStore();

  // Hydrate auth state on mount
  React.useEffect(() => {
    const hydrateAuth = async () => {
      // Check if tokens exist (from localStorage via Zustand persist)
      if (tokens?.access_token) {
        try {
          // Verify token is still valid by fetching user
          const response = await fetch('/api/v1/users/me', {
            headers: {
              Authorization: `Bearer ${tokens.access_token}`,
            },
          });

          if (!response.ok) {
            // Token invalid, logout
            useAuthStore.getState().logout();
          }
        } catch (error) {
          // Network error or token invalid
          console.error('Auth hydration error:', error);
          useAuthStore.getState().logout();
        }
      }

      setLoading(false);
    };

    hydrateAuth();
  }, []);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return <>{children}</>;
}

// ============================================================================
// Example 9: Role-Based Access Control
// ============================================================================

import { useIsVolunteer, useIsOrgAdmin, useIsSuperAdmin } from '@/store';

export function AdminOnlyButton() {
  const isAdmin = useIsOrgAdmin();

  if (!isAdmin) return null;

  return <button>Admin Action</button>;
}

export function VolunteerOnlySection() {
  const isVolunteer = useIsVolunteer();

  if (!isVolunteer) return null;

  return (
    <div>
      <h2>Volunteer Dashboard</h2>
      {/* Volunteer-specific content */}
    </div>
  );
}

// ============================================================================
// Example 10: Direct Store Access (outside components)
// ============================================================================

// Useful for API client, middleware, or utilities
export function getAuthToken(): string | undefined {
  return useAuthStore.getState().tokens?.access_token;
}

export function isUserAuthenticated(): boolean {
  return useAuthStore.getState().isAuthenticated;
}

export function logoutUser(): void {
  useAuthStore.getState().logout();
  useNotificationStore.getState().reset();
}

// ============================================================================
// Helper Components
// ============================================================================

function BellIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
      <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
    </svg>
  );
}

// Placeholder for React import
declare const React: any;
