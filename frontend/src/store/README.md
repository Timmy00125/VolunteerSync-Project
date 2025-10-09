# Zustand Stores

This directory contains all Zustand stores for global state management in the VolunteerSync frontend application.

## Overview

We use [Zustand](https://github.com/pmndrs/zustand) for state management because it's:

- Simple and unopinionated
- Small bundle size (~1KB)
- No boilerplate required
- Works with React hooks
- Built-in persistence support

## Stores

### 1. Auth Store (`auth-store.ts`)

Manages authentication state including user information and JWT tokens.

**State:**

- `user`: Current logged-in user or `null`
- `tokens`: Access and refresh tokens
- `isAuthenticated`: Boolean authentication status
- `isLoading`: Loading state for hydration

**Actions:**

- `setUser(user)`: Set the current user
- `setTokens(tokens)`: Set JWT tokens
- `login(user, tokens)`: Login and set user + tokens
- `logout()`: Clear all auth state
- `updateUser(updates)`: Update user data (e.g., after profile edit)
- `setLoading(loading)`: Set loading state

**Persistence:**
The auth store uses `localStorage` to persist user and tokens across page refreshes.

**Example Usage:**

```tsx
import { useAuthStore, useUser, useIsAuthenticated } from '@/store';

function MyComponent() {
  // Get the entire auth state and actions
  const { user, login, logout, isAuthenticated } = useAuthStore();

  // Or use specific selectors (more performant)
  const user = useUser();
  const isAuthenticated = useIsAuthenticated();

  // Login example
  const handleLogin = async () => {
    const response = await apiClient.post('/auth/login', credentials);
    login(response.data.user, response.data.tokens);
  };

  // Logout example
  const handleLogout = () => {
    logout();
    // Navigate to login page
  };

  return (
    <div>
      {isAuthenticated ? (
        <p>Welcome, {user?.first_name}!</p>
      ) : (
        <button onClick={handleLogin}>Login</button>
      )}
    </div>
  );
}
```

**Selectors:**

For better performance, use the provided selectors instead of accessing the full store:

```tsx
import {
  useUser, // Get current user
  useIsAuthenticated, // Get auth status
  useTokens, // Get both tokens
  useAccessToken, // Get access token only
  useRefreshToken, // Get refresh token only
  useUserType, // Get user type
  useIsVolunteer, // Check if volunteer
  useIsOrgAdmin, // Check if org admin
  useIsSuperAdmin, // Check if super admin
} from '@/store';
```

### 2. Notification Store (`notification-store.ts`)

Manages in-app notifications including unread count and notification list.

**State:**

- `notifications`: Array of all notifications
- `unreadCount`: Number of unread notifications
- `isLoading`: Loading state for fetching notifications

**Actions:**

- `setNotifications(notifications)`: Set all notifications
- `addNotification(notification)`: Add a new notification (real-time)
- `markAsRead(notificationId)`: Mark a single notification as read
- `markAllAsRead()`: Mark all notifications as read
- `removeNotification(notificationId)`: Remove a notification
- `setLoading(loading)`: Set loading state
- `refreshUnreadCount()`: Recalculate unread count
- `reset()`: Clear all notifications (on logout)

**Example Usage:**

```tsx
import { useNotificationStore, useUnreadCount, useHasUnreadNotifications } from '@/store';

function NotificationBadge() {
  const unreadCount = useUnreadCount();
  const hasUnread = useHasUnreadNotifications();

  return (
    <button>
      <BellIcon />
      {hasUnread && <span className="badge">{unreadCount}</span>}
    </button>
  );
}

function NotificationList() {
  const { notifications, markAsRead, markAllAsRead } = useNotificationStore();

  return (
    <div>
      <button onClick={markAllAsRead}>Mark all as read</button>
      {notifications.map((notification) => (
        <div key={notification.id} onClick={() => markAsRead(notification.id)}>
          <h4>{notification.title}</h4>
          <p>{notification.message}</p>
          {!notification.is_read && <span>NEW</span>}
        </div>
      ))}
    </div>
  );
}
```

**Selectors:**

```tsx
import {
  useUnreadNotifications, // Get only unread notifications
  useUnreadCount, // Get unread count only
  useNotificationsByType, // Filter by type
  useRecentNotifications, // Get recent notifications (limit)
  useHasUnreadNotifications, // Boolean: has unread
} from '@/store';
```

**Real-time Updates:**

To add new notifications in real-time (e.g., from WebSocket or polling):

```tsx
import { useNotificationStore } from '@/store';

function useNotificationPolling() {
  const addNotification = useNotificationStore((state) => state.addNotification);

  useEffect(() => {
    const interval = setInterval(async () => {
      const newNotifications = await fetchNewNotifications();
      newNotifications.forEach(addNotification);
    }, 30000); // Poll every 30 seconds

    return () => clearInterval(interval);
  }, [addNotification]);
}
```

## Integration with API Client

The stores work seamlessly with the API client. Here's a recommended pattern:

### Login Flow

```tsx
import { apiClient } from '@/lib/api/client';
import { useAuthStore } from '@/store';

async function handleLogin(credentials: LoginCredentials) {
  try {
    const response = await apiClient.post<AuthResponse>('/auth/login', credentials);
    const { user, tokens } = response.data;

    // Update store
    useAuthStore.getState().login(user, tokens);

    // API client will automatically use tokens from the store
  } catch (error) {
    console.error('Login failed:', error);
  }
}
```

### Logout Flow

```tsx
import { apiClient } from '@/lib/api/client';
import { useAuthStore, useNotificationStore } from '@/store';

async function handleLogout() {
  try {
    await apiClient.post('/auth/logout');
  } finally {
    // Clear stores
    useAuthStore.getState().logout();
    useNotificationStore.getState().reset();
  }
}
```

### Fetching Notifications

```tsx
import { apiClient } from '@/lib/api/client';
import { useNotificationStore } from '@/store';

async function loadNotifications() {
  const setLoading = useNotificationStore.getState().setLoading;
  const setNotifications = useNotificationStore.getState().setNotifications;

  setLoading(true);
  try {
    const response = await apiClient.get<Notification[]>('/notifications');
    setNotifications(response.data);
  } finally {
    setLoading(false);
  }
}
```

## Best Practices

### 1. Use Selectors

Instead of accessing the entire store, use selectors to prevent unnecessary re-renders:

```tsx
// ❌ Bad - component re-renders on any auth state change
const { user, tokens, isLoading } = useAuthStore();

// ✅ Good - component only re-renders when user changes
const user = useUser();
```

### 2. Access Outside Components

You can access stores outside of React components:

```tsx
import { useAuthStore } from '@/store';

// Get current state
const currentUser = useAuthStore.getState().user;

// Call an action
useAuthStore.getState().logout();
```

### 3. Combine with React Query

Use Zustand for client-side state and React Query for server state:

```tsx
import { useQuery } from '@tanstack/react-query';
import { useUser } from '@/store';

function MyComponent() {
  const user = useUser(); // Client state (Zustand)

  const { data: profile } = useQuery({
    queryKey: ['profile', user?.id],
    queryFn: () => fetchProfile(user?.id), // Server state (React Query)
    enabled: !!user,
  });
}
```

### 4. Reset on Logout

Always reset stores when logging out:

```tsx
function logout() {
  useAuthStore.getState().logout();
  useNotificationStore.getState().reset();
  // Clear any other stores
}
```

### 5. Persist Only Necessary Data

The auth store persists to localStorage, but:

- Don't persist sensitive data if avoidable
- Clear persisted data on logout
- Consider token expiration

## Testing

### Unit Testing Stores

```tsx
import { renderHook, act } from '@testing-library/react';
import { useAuthStore } from '@/store';

describe('Auth Store', () => {
  beforeEach(() => {
    // Reset store before each test
    useAuthStore.getState().logout();
  });

  it('should login user', () => {
    const { result } = renderHook(() => useAuthStore());

    act(() => {
      result.current.login(mockUser, mockTokens);
    });

    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.user).toEqual(mockUser);
  });

  it('should logout user', () => {
    const { result } = renderHook(() => useAuthStore());

    act(() => {
      result.current.login(mockUser, mockTokens);
      result.current.logout();
    });

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBeNull();
  });
});
```

### Integration Testing

```tsx
import { render, screen } from '@testing-library/react';
import { useAuthStore } from '@/store';

describe('Component with Auth Store', () => {
  beforeEach(() => {
    useAuthStore.getState().logout();
  });

  it('shows login button when not authenticated', () => {
    render(<MyComponent />);
    expect(screen.getByText('Login')).toBeInTheDocument();
  });

  it('shows user name when authenticated', () => {
    useAuthStore.getState().login(mockUser, mockTokens);
    render(<MyComponent />);
    expect(screen.getByText('Welcome, John!')).toBeInTheDocument();
  });
});
```

## TypeScript

All stores are fully typed. Import types from the API types:

```tsx
import type { User, AuthTokens, Notification } from '@/lib/api/types';
```

## Performance Considerations

1. **Selectors prevent unnecessary re-renders**: Use specific selectors instead of the full store
2. **Zustand is fast**: No Provider needed, direct store access
3. **Shallow equality**: Zustand uses shallow comparison for updates
4. **DevTools**: Install the [Zustand DevTools](https://github.com/pmndrs/zustand#devtools) extension for debugging

## Migration Notes

When migrating from another state management solution:

1. **From Redux**: Zustand is much simpler, no actions/reducers needed
2. **From Context API**: No Provider wrapper needed, better performance
3. **From MobX**: Similar observable pattern, but simpler syntax

## Resources

- [Zustand Documentation](https://github.com/pmndrs/zustand)
- [React Query Integration](https://tanstack.com/query/latest/docs/react/overview)
- [Zustand Best Practices](https://github.com/pmndrs/zustand#best-practices)
