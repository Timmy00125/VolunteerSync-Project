# Zustand Stores - Quick Reference

## Import Stores

```tsx
// Import everything
import { useAuthStore, useNotificationStore } from '@/store';

// Import specific selectors (recommended for performance)
import { useUser, useIsAuthenticated, useUnreadCount } from '@/store';
```

## Auth Store

### Basic Usage

```tsx
// Get auth state and actions
const { user, isAuthenticated, login, logout } = useAuthStore();

// Use selectors (better performance)
const user = useUser();
const isAuthenticated = useIsAuthenticated();
```

### Login

```tsx
import { useAuthStore } from '@/store';

const login = useAuthStore((state) => state.login);

// In your login handler
login(userData, tokensData);
```

### Logout

```tsx
import { useAuthStore, useNotificationStore } from '@/store';

const handleLogout = () => {
  useAuthStore.getState().logout();
  useNotificationStore.getState().reset();
  // router.push('/login');
};
```

### Update User

```tsx
const updateUser = useAuthStore((state) => state.updateUser);

// After updating profile via API
updateUser({ first_name: 'John', last_name: 'Doe' });
```

### Access Outside Components

```tsx
import { useAuthStore } from '@/store';

// Get current token
const token = useAuthStore.getState().tokens?.access_token;

// Check authentication status
const isAuth = useAuthStore.getState().isAuthenticated;

// Logout
useAuthStore.getState().logout();
```

### Role Checks

```tsx
import { useIsVolunteer, useIsOrgAdmin, useIsSuperAdmin } from '@/store';

function MyComponent() {
  const isVolunteer = useIsVolunteer();
  const isAdmin = useIsOrgAdmin();
  const isSuperAdmin = useIsSuperAdmin();

  if (isAdmin) return <AdminPanel />;
  if (isVolunteer) return <VolunteerDashboard />;
  return null;
}
```

## Notification Store

### Basic Usage

```tsx
// Get notification state and actions
const { notifications, unreadCount, markAsRead } = useNotificationStore();

// Use selectors (better performance)
const unreadCount = useUnreadCount();
const hasUnread = useHasUnreadNotifications();
```

### Set Notifications

```tsx
const setNotifications = useNotificationStore((state) => state.setNotifications);

// After fetching from API
const response = await fetch('/api/v1/notifications');
const data = await response.json();
setNotifications(data);
```

### Add Single Notification

```tsx
const addNotification = useNotificationStore((state) => state.addNotification);

// Add new notification (e.g., from WebSocket)
addNotification(newNotification);
```

### Mark as Read

```tsx
const markAsRead = useNotificationStore((state) => state.markAsRead);

// Mark single notification as read
markAsRead(notificationId);

// Mark all as read
const markAllAsRead = useNotificationStore((state) => state.markAllAsRead);
markAllAsRead();
```

### Get Unread Notifications

```tsx
import { useUnreadNotifications } from '@/store';

function NotificationList() {
  const unreadNotifications = useUnreadNotifications();

  return (
    <ul>
      {unreadNotifications.map((notification) => (
        <li key={notification.id}>{notification.title}</li>
      ))}
    </ul>
  );
}
```

### Reset on Logout

```tsx
import { useNotificationStore } from '@/store';

const handleLogout = () => {
  // ... logout logic
  useNotificationStore.getState().reset();
};
```

## Common Patterns

### Protected Route

```tsx
import { useIsAuthenticated } from '@/store';
import { redirect } from 'next/navigation';

export default function ProtectedPage() {
  const isAuthenticated = useIsAuthenticated();

  if (!isAuthenticated) {
    redirect('/login');
  }

  return <div>Protected Content</div>;
}
```

### Auth Provider (with hydration)

```tsx
'use client';

import { useEffect, useState } from 'react';
import { useAuthStore } from '@/store';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    // Hydration happens automatically via Zustand persist
    setIsReady(true);
  }, []);

  if (!isReady) {
    return <div>Loading...</div>;
  }

  return <>{children}</>;
}
```

### Notification Polling

```tsx
import { useEffect } from 'react';
import { useNotificationStore, useIsAuthenticated } from '@/store';

export function useNotificationPolling(intervalMs = 30000) {
  const setNotifications = useNotificationStore((state) => state.setNotifications);
  const isAuthenticated = useIsAuthenticated();

  useEffect(() => {
    if (!isAuthenticated) return;

    const fetchNotifications = async () => {
      const response = await fetch('/api/v1/notifications');
      if (response.ok) {
        const data = await response.json();
        setNotifications(data);
      }
    };

    fetchNotifications();
    const interval = setInterval(fetchNotifications, intervalMs);

    return () => clearInterval(interval);
  }, [isAuthenticated, intervalMs, setNotifications]);
}
```

### Notification Badge

```tsx
import { useUnreadCount } from '@/store';

export function NotificationBadge() {
  const unreadCount = useUnreadCount();

  return (
    <button className="relative">
      <BellIcon />
      {unreadCount > 0 && (
        <span className="absolute top-0 right-0 bg-red-500 text-white rounded-full px-2 py-1 text-xs">
          {unreadCount}
        </span>
      )}
    </button>
  );
}
```

### User Avatar

```tsx
import { useUser } from '@/store';

export function UserAvatar() {
  const user = useUser();

  if (!user) return null;

  return (
    <div className="flex items-center gap-2">
      {user.profile_photo_url ? (
        <img
          src={user.profile_photo_url}
          alt={`${user.first_name} ${user.last_name}`}
          className="w-10 h-10 rounded-full"
        />
      ) : (
        <div className="w-10 h-10 rounded-full bg-blue-500 flex items-center justify-center text-white">
          {user.first_name[0]}
          {user.last_name[0]}
        </div>
      )}
      <span>
        {user.first_name} {user.last_name}
      </span>
    </div>
  );
}
```

## API Integration

### API Client with Auth

```tsx
// In your API client (lib/api/client.ts)
import { useAuthStore } from '@/store';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
});

// Request interceptor - add auth token
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().tokens?.access_token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor - handle 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Token expired, logout
      useAuthStore.getState().logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### React Query Integration

```tsx
import { useQuery } from '@tanstack/react-query';
import { useUser } from '@/store';

function MyComponent() {
  const user = useUser(); // Zustand (client state)

  // React Query (server state)
  const { data: profile } = useQuery({
    queryKey: ['profile', user?.id],
    queryFn: () => fetchProfile(user?.id),
    enabled: !!user,
  });

  return <div>{profile?.bio}</div>;
}
```

## Testing

### Test Setup

```tsx
import { renderHook, act } from '@testing-library/react';
import { useAuthStore, useNotificationStore } from '@/store';

beforeEach(() => {
  // Reset stores before each test
  useAuthStore.getState().logout();
  useNotificationStore.getState().reset();
});
```

### Test Login

```tsx
it('should login user', () => {
  const { result } = renderHook(() => useAuthStore());

  act(() => {
    result.current.login(mockUser, mockTokens);
  });

  expect(result.current.isAuthenticated).toBe(true);
  expect(result.current.user).toEqual(mockUser);
});
```

### Test Notifications

```tsx
it('should calculate unread count', () => {
  const { result } = renderHook(() => useNotificationStore());

  act(() => {
    result.current.setNotifications([
      { ...mockNotification, is_read: false },
      { ...mockNotification, is_read: true },
    ]);
  });

  expect(result.current.unreadCount).toBe(1);
});
```

## Performance Tips

1. **Use selectors** instead of the full store:

   ```tsx
   // ❌ Bad - re-renders on any auth change
   const { user, tokens, isLoading } = useAuthStore();

   // ✅ Good - re-renders only when user changes
   const user = useUser();
   ```

2. **Access outside components** for utilities:

   ```tsx
   // Get token in API client
   const token = useAuthStore.getState().tokens?.access_token;
   ```

3. **Reset on logout** to clear memory:
   ```tsx
   useAuthStore.getState().logout();
   useNotificationStore.getState().reset();
   ```

## Troubleshooting

### Store not persisting

Check that localStorage is available and not blocked.

### Hydration mismatch

Wrap your app with an AuthProvider to handle hydration.

### Type errors

Make sure `@/lib/api/types` exports all required types.

### Store not updating

Make sure you're calling the action functions, not just setting state.

---

For more details, see [README.md](./README.md) and [examples.tsx](./examples.tsx).
