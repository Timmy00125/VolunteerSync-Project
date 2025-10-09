/**
 * API Client
 *
 * Centralized API client with JWT token handling, automatic token refresh,
 * and comprehensive error handling.
 *
 * Features:
 * - JWT access + refresh token management
 * - Automatic token refresh on 401 responses
 * - Request/response interceptors
 * - Type-safe API calls
 * - Consistent error handling
 */

import type {
  ApiError,
  AuthTokens,
  DashboardResponse,
  VolunteerProfile,
  UpdateVolunteerProfileInput,
} from './types';

// ============================================================================
// Configuration
// ============================================================================

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const TOKEN_REFRESH_ENDPOINT = '/auth/refresh';
const TOKEN_STORAGE_KEY = 'volunteersync_tokens';

// ============================================================================
// Token Management
// ============================================================================

/**
 * Store authentication tokens in localStorage
 */
export function setTokens(tokens: AuthTokens): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(tokens));
}

/**
 * Retrieve authentication tokens from localStorage
 */
export function getTokens(): AuthTokens | null {
  if (typeof window === 'undefined') return null;

  try {
    const stored = localStorage.getItem(TOKEN_STORAGE_KEY);
    return stored ? JSON.parse(stored) : null;
  } catch (error) {
    console.error('Failed to parse stored tokens:', error);
    return null;
  }
}

/**
 * Remove authentication tokens from localStorage
 */
export function clearTokens(): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(TOKEN_STORAGE_KEY);
}

/**
 * Get current access token
 */
export function getAccessToken(): string | null {
  const tokens = getTokens();
  return tokens?.access_token || null;
}

/**
 * Get current refresh token
 */
export function getRefreshToken(): string | null {
  const tokens = getTokens();
  return tokens?.refresh_token || null;
}

// ============================================================================
// Token Refresh Logic
// ============================================================================

let isRefreshing = false;
let refreshSubscribers: Array<(token: string) => void> = [];

/**
 * Subscribe to token refresh completion
 */
function subscribeTokenRefresh(callback: (token: string) => void): void {
  refreshSubscribers.push(callback);
}

/**
 * Notify all subscribers that token refresh is complete
 */
function onTokenRefreshed(token: string): void {
  refreshSubscribers.forEach((callback) => callback(token));
  refreshSubscribers = [];
}

/**
 * Refresh the access token using the refresh token
 */
async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken();

  if (!refreshToken) {
    clearTokens();
    return null;
  }

  try {
    const response = await fetch(`${API_BASE_URL}${TOKEN_REFRESH_ENDPOINT}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      // Refresh token is invalid or expired
      clearTokens();
      return null;
    }

    const data = await response.json();
    const newTokens: AuthTokens = data.tokens;

    setTokens(newTokens);
    return newTokens.access_token;
  } catch (error) {
    console.error('Token refresh failed:', error);
    clearTokens();
    return null;
  }
}

// ============================================================================
// Request Configuration
// ============================================================================

interface RequestConfig extends RequestInit {
  headers?: Record<string, string>;
  skipAuth?: boolean;
  skipRefresh?: boolean;
}

/**
 * Build request headers with authentication if available
 */
function buildHeaders(config?: RequestConfig): HeadersInit {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...config?.headers,
  };

  // Add authentication token if not explicitly skipped
  if (!config?.skipAuth) {
    const token = getAccessToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  return headers;
}

// ============================================================================
// Error Handling
// ============================================================================

/**
 * Custom error class for API errors
 */
export class ApiClientError extends Error {
  public statusCode: number;
  public details?: Record<string, any>;

  constructor(message: string, statusCode: number, details?: Record<string, any>) {
    super(message);
    this.name = 'ApiClientError';
    this.statusCode = statusCode;
    this.details = details;
  }
}

/**
 * Parse error response from API
 */
async function parseErrorResponse(response: Response): Promise<ApiError> {
  try {
    const error: ApiError = await response.json();
    return error;
  } catch {
    // Fallback if response is not JSON
    return {
      error: 'API Error',
      message: response.statusText || 'An unexpected error occurred',
      status_code: response.status,
    };
  }
}

/**
 * Handle API response and throw errors if needed
 */
async function handleResponse<T>(response: Response): Promise<T> {
  if (response.ok) {
    // Success response (2xx)
    if (response.status === 204) {
      // No content
      return null as T;
    }
    return response.json();
  }

  // Error response
  const error = await parseErrorResponse(response);
  throw new ApiClientError(error.message || 'Request failed', response.status, error.details);
}

// ============================================================================
// Core API Client
// ============================================================================

/**
 * Make an authenticated API request with automatic token refresh
 */
async function request<T>(endpoint: string, config?: RequestConfig): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = buildHeaders(config);

  try {
    let response = await fetch(url, {
      ...config,
      headers,
    });

    // Handle 401 Unauthorized - attempt token refresh
    if (response.status === 401 && !config?.skipRefresh) {
      // If already refreshing, wait for it to complete
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          subscribeTokenRefresh(async (token: string) => {
            try {
              const retryHeaders = {
                ...headers,
                Authorization: `Bearer ${token}`,
              };
              const retryResponse = await fetch(url, {
                ...config,
                headers: retryHeaders,
              });
              const data = await handleResponse<T>(retryResponse);
              resolve(data);
            } catch (error) {
              reject(error);
            }
          });
        });
      }

      // Start token refresh
      isRefreshing = true;
      const newToken = await refreshAccessToken();
      isRefreshing = false;

      if (newToken) {
        // Notify all waiting requests
        onTokenRefreshed(newToken);

        // Retry original request with new token
        const retryHeaders = {
          ...headers,
          Authorization: `Bearer ${newToken}`,
        };
        response = await fetch(url, {
          ...config,
          headers: retryHeaders,
        });
      } else {
        // Refresh failed - user needs to log in again
        // Redirect to login page or emit event
        if (typeof window !== 'undefined') {
          window.location.href = '/login?session_expired=true';
        }
        throw new ApiClientError('Session expired. Please log in again.', 401);
      }
    }

    return handleResponse<T>(response);
  } catch (error) {
    if (error instanceof ApiClientError) {
      throw error;
    }

    // Network error or other fetch failure
    throw new ApiClientError('Network error. Please check your connection.', 0);
  }
}

// ============================================================================
// HTTP Method Helpers
// ============================================================================

/**
 * GET request
 */
export async function get<T>(endpoint: string, config?: RequestConfig): Promise<T> {
  return request<T>(endpoint, {
    ...config,
    method: 'GET',
  });
}

/**
 * POST request
 */
export async function post<T>(endpoint: string, data?: any, config?: RequestConfig): Promise<T> {
  return request<T>(endpoint, {
    ...config,
    method: 'POST',
    body: data ? JSON.stringify(data) : undefined,
  });
}

/**
 * PATCH request
 */
export async function patch<T>(endpoint: string, data?: any, config?: RequestConfig): Promise<T> {
  return request<T>(endpoint, {
    ...config,
    method: 'PATCH',
    body: data ? JSON.stringify(data) : undefined,
  });
}

/**
 * PUT request
 */
export async function put<T>(endpoint: string, data?: any, config?: RequestConfig): Promise<T> {
  return request<T>(endpoint, {
    ...config,
    method: 'PUT',
    body: data ? JSON.stringify(data) : undefined,
  });
}

/**
 * DELETE request
 */
export async function del<T>(endpoint: string, config?: RequestConfig): Promise<T> {
  return request<T>(endpoint, {
    ...config,
    method: 'DELETE',
  });
}

// ============================================================================
// Convenience Functions
// ============================================================================

/**
 * Check if user is authenticated (has valid tokens)
 */
export function isAuthenticated(): boolean {
  return getAccessToken() !== null;
}

/**
 * Logout user by clearing tokens
 */
export async function logout(): Promise<void> {
  try {
    // Call logout endpoint to invalidate refresh token on server
    await post('/auth/logout', {});
  } catch (error) {
    // Log error but still clear local tokens
    console.error('Logout request failed:', error);
  } finally {
    clearTokens();
  }
}

// ============================================================================
// Volunteer API Methods
// ============================================================================

/**
 * Get volunteer dashboard data
 */
export async function getVolunteerDashboard(): Promise<DashboardResponse> {
  const response = await get<{ data: DashboardResponse }>('/volunteers/me/dashboard');
  return response.data;
}

/**
 * Get volunteer profile data
 */
export async function getVolunteerProfile(): Promise<VolunteerProfile> {
  const response = await get<{ data: VolunteerProfile }>('/volunteers/me');
  return response.data;
}

/**
 * Update volunteer profile
 */
export async function updateVolunteerProfile(
  data: UpdateVolunteerProfileInput
): Promise<VolunteerProfile> {
  const response = await patch<{ data: VolunteerProfile }>('/volunteers/me', data);
  return response.data;
}

// ============================================================================
// Default Export
// ============================================================================

export default {
  get,
  post,
  patch,
  put,
  delete: del,
  setTokens,
  getTokens,
  clearTokens,
  isAuthenticated,
  logout,
  getVolunteerDashboard,
  getVolunteerProfile,
  updateVolunteerProfile,
};
