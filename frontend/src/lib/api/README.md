# API Client Documentation

This directory contains the API client for the VolunteerSync frontend application.

## Features

- ✅ **JWT Authentication**: Automatic token handling (access + refresh)
- ✅ **Token Refresh**: Automatic token refresh on 401 responses
- ✅ **Type Safety**: Full TypeScript support with typed responses
- ✅ **Error Handling**: Consistent error handling with custom error types
- ✅ **Request Interceptors**: Automatic Authorization header injection
- ✅ **Network Error Handling**: Graceful handling of network failures

## Usage

### Basic Requests

```typescript
import { get, post, patch, del } from '@/lib/api';

// GET request
const users = await get<User[]>('/users');

// POST request
const newUser = await post<User>('/users', {
  name: 'John Doe',
  email: 'john@example.com',
});

// PATCH request
const updatedUser = await patch<User>('/users/123', {
  name: 'Jane Doe',
});

// DELETE request
await del('/users/123');
```

### Authentication

```typescript
import { post, setTokens, clearTokens, isAuthenticated } from '@/lib/api';

// Login
const response = await post<AuthResponse>('/auth/login', {
  email: 'user@example.com',
  password: 'password123',
});

// Store tokens
setTokens(response.tokens);

// Check authentication status
if (isAuthenticated()) {
  console.log('User is logged in');
}

// Logout
await logout();
```

### Error Handling

```typescript
import { get, ApiClientError } from '@/lib/api';

try {
  const data = await get<User>('/users/me');
} catch (error) {
  if (error instanceof ApiClientError) {
    console.error('Status:', error.statusCode);
    console.error('Message:', error.message);
    console.error('Details:', error.details);
  } else {
    console.error('Unexpected error:', error);
  }
}
```

### Skipping Authentication

For public endpoints that don't require authentication:

```typescript
import { get } from '@/lib/api';

const publicData = await get<PublicData>('/public/data', {
  skipAuth: true,
});
```

### Custom Headers

```typescript
import { post } from '@/lib/api';

const data = await post<Data>('/endpoint', payload, {
  headers: {
    'X-Custom-Header': 'value',
  },
});
```

## Files

### `client.ts`

Core API client with:

- Token management (localStorage)
- Automatic token refresh
- Request/response interceptors
- HTTP method helpers (get, post, patch, put, delete)
- Error handling

### `types.ts`

TypeScript type definitions for:

- API requests and responses
- Domain models (User, Organization, Volunteer, etc.)
- Error types
- Pagination types

### `index.ts`

Convenience re-exports for easier imports.

## Configuration

The API base URL is configured via environment variable:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

If not set, defaults to `http://localhost:8080/api/v1`.

## Token Management

Tokens are stored in `localStorage` under the key `volunteersync_tokens`.

### Token Structure

```typescript
interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}
```

### Automatic Token Refresh

When an API request receives a 401 Unauthorized response:

1. The client automatically attempts to refresh the access token using the refresh token
2. If refresh succeeds, the original request is retried with the new token
3. If refresh fails, the user is redirected to the login page
4. Multiple simultaneous requests wait for a single refresh to complete

## Error Handling

### ApiClientError

Custom error class with:

- `statusCode`: HTTP status code
- `message`: Error message
- `details`: Additional error details (optional)

### Common Status Codes

- **400**: Bad Request (validation errors)
- **401**: Unauthorized (invalid or expired token)
- **403**: Forbidden (insufficient permissions)
- **404**: Not Found
- **409**: Conflict (duplicate resource)
- **429**: Too Many Requests (rate limited)
- **500**: Internal Server Error

## Best Practices

1. **Always use TypeScript types**: Specify the expected response type

   ```typescript
   const user = await get<User>('/users/me');
   ```

2. **Handle errors appropriately**: Use try-catch blocks

   ```typescript
   try {
     await post('/endpoint', data);
   } catch (error) {
     // Handle error
   }
   ```

3. **Use React Query for data fetching**: Combine with TanStack Query for caching and state management

   ```typescript
   import { useQuery } from '@tanstack/react-query';
   import { get } from '@/lib/api';

   const { data, error, isLoading } = useQuery({
     queryKey: ['user', 'me'],
     queryFn: () => get<User>('/users/me'),
   });
   ```

4. **Don't store sensitive data in state**: Let the API client manage tokens

5. **Check authentication before protected routes**: Use `isAuthenticated()`

## Security Considerations

- Tokens are stored in localStorage (XSS-safe with proper CSP headers)
- Access tokens expire after 15 minutes
- Refresh tokens expire after 7 days
- Automatic logout on refresh token expiry
- No PII logged in console errors
- HTTPS required in production

## Testing

When writing tests, you can mock the API client:

```typescript
import * as apiClient from '@/lib/api/client';

jest.mock('@/lib/api/client', () => ({
  get: jest.fn(),
  post: jest.fn(),
  // ... other methods
}));

// In test
(apiClient.get as jest.Mock).mockResolvedValue({ data: 'test' });
```
