/**
 * API Client Usage Examples
 *
 * This file demonstrates how to use the API client in various scenarios.
 * These examples can be used as reference when implementing features.
 */

import { get, post, patch, ApiClientError } from './client';
import type {
  AuthResponse,
  User,
  Organization,
  Opportunity,
  Registration,
  LoginCredentials,
  RegisterData,
  PaginatedResponse,
  OpportunitySearchParams,
} from './types';

// ============================================================================
// Authentication Examples
// ============================================================================

/**
 * Example: User login
 */
export async function loginExample(credentials: LoginCredentials): Promise<AuthResponse> {
  try {
    const response = await post<AuthResponse>('/auth/login', credentials);
    // Tokens are automatically stored by the client
    return response;
  } catch (error) {
    if (error instanceof ApiClientError) {
      if (error.statusCode === 401) {
        throw new Error('Invalid email or password');
      } else if (error.statusCode === 429) {
        throw new Error('Too many login attempts. Please try again later.');
      }
    }
    throw error;
  }
}

/**
 * Example: User registration
 */
export async function registerExample(data: RegisterData): Promise<AuthResponse> {
  try {
    const response = await post<AuthResponse>('/auth/register', data);
    return response;
  } catch (error) {
    if (error instanceof ApiClientError) {
      if (error.statusCode === 409) {
        throw new Error('Email already exists');
      } else if (error.statusCode === 400) {
        throw new Error('Invalid registration data: ' + error.message);
      }
    }
    throw error;
  }
}

// ============================================================================
// User Examples
// ============================================================================

/**
 * Example: Get current user profile
 */
export async function getCurrentUserExample(): Promise<User> {
  return get<User>('/users/me');
}

/**
 * Example: Update user profile
 */
export async function updateUserProfileExample(updates: Partial<User>): Promise<User> {
  return patch<User>('/users/me', updates);
}

// ============================================================================
// Organization Examples
// ============================================================================

/**
 * Example: Create organization
 */
export async function createOrganizationExample(orgData: {
  name: string;
  description: string;
  email: string;
  address: string;
  city: string;
  state: string;
  zip_code: string;
  country: string;
}): Promise<Organization> {
  return post<Organization>('/organizations', orgData);
}

/**
 * Example: Get organization by ID
 */
export async function getOrganizationExample(id: string): Promise<Organization> {
  return get<Organization>(`/organizations/${id}`);
}

/**
 * Example: List organizations with pagination
 */
export async function listOrganizationsExample(
  page = 1,
  limit = 20
): Promise<PaginatedResponse<Organization>> {
  return get<PaginatedResponse<Organization>>(`/organizations?page=${page}&limit=${limit}`);
}

// ============================================================================
// Opportunity Examples
// ============================================================================

/**
 * Example: Search opportunities with filters
 */
export async function searchOpportunitiesExample(
  params: OpportunitySearchParams
): Promise<PaginatedResponse<Opportunity>> {
  const queryParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      if (Array.isArray(value)) {
        value.forEach((v) => queryParams.append(key, v.toString()));
      } else {
        queryParams.append(key, value.toString());
      }
    }
  });

  return get<PaginatedResponse<Opportunity>>(`/opportunities?${queryParams.toString()}`);
}

/**
 * Example: Create opportunity
 */
export async function createOpportunityExample(opportunityData: {
  title: string;
  description: string;
  start_time: string;
  end_time: string;
  location: string;
  address: string;
  city: string;
  state: string;
  zip_code: string;
  capacity: number;
  causes: string[];
}): Promise<Opportunity> {
  return post<Opportunity>('/opportunities', opportunityData);
}

// ============================================================================
// Registration Examples
// ============================================================================

/**
 * Example: Register for opportunity
 */
export async function registerForOpportunityExample(opportunityId: string): Promise<Registration> {
  try {
    const response = await post<Registration>('/registrations', {
      opportunity_id: opportunityId,
    });
    return response;
  } catch (error) {
    if (error instanceof ApiClientError) {
      if (error.statusCode === 409) {
        throw new Error('Already registered for this opportunity');
      } else if (error.statusCode === 400) {
        // Check if it's a capacity or overlap issue
        if (error.details?.reason === 'at_capacity') {
          throw new Error('Event is at capacity. You have been added to the waitlist.');
        } else if (error.details?.reason === 'overlapping_event') {
          throw new Error('You have another event scheduled at this time.');
        }
      }
    }
    throw error;
  }
}

/**
 * Example: Cancel registration
 */
export async function cancelRegistrationExample(
  registrationId: string,
  reason?: string
): Promise<void> {
  return patch(`/registrations/${registrationId}/cancel`, { reason });
}

/**
 * Example: Check in to event
 */
export async function checkInExample(registrationId: string): Promise<Registration> {
  return patch<Registration>(`/registrations/${registrationId}/check-in`, {});
}

// ============================================================================
// Hours Tracking Examples
// ============================================================================

/**
 * Example: Log volunteer hours (coordinator)
 */
export async function logHoursExample(data: {
  registration_id: string;
  hours: number;
  coordinator_notes?: string;
}): Promise<any> {
  return post('/hours/log', data);
}

/**
 * Example: Verify hours (volunteer)
 */
export async function verifyHoursExample(hoursId: string): Promise<void> {
  return post(`/hours/${hoursId}/verify`, {});
}

/**
 * Example: Dispute hours (volunteer)
 */
export async function disputeHoursExample(hoursId: string, reason: string): Promise<void> {
  return post(`/hours/${hoursId}/dispute`, { reason });
}

// ============================================================================
// Notifications Examples
// ============================================================================

/**
 * Example: Get notifications with pagination
 */
export async function getNotificationsExample(page = 1, limit = 20) {
  return get(`/notifications?page=${page}&limit=${limit}`);
}

/**
 * Example: Mark notification as read
 */
export async function markNotificationReadExample(notificationId: string): Promise<void> {
  return patch(`/notifications/${notificationId}/read`, {});
}

// ============================================================================
// Error Handling Pattern
// ============================================================================

/**
 * Example: Generic error handler wrapper
 */
export async function withErrorHandling<T>(
  apiCall: () => Promise<T>,
  errorMessages?: Record<number, string>
): Promise<T> {
  try {
    return await apiCall();
  } catch (error) {
    if (error instanceof ApiClientError) {
      const customMessage = errorMessages?.[error.statusCode];
      if (customMessage) {
        throw new Error(customMessage);
      }

      // Default error messages by status code
      switch (error.statusCode) {
        case 400:
          throw new Error('Invalid request. Please check your input.');
        case 401:
          throw new Error('Please log in to continue.');
        case 403:
          throw new Error('You do not have permission to perform this action.');
        case 404:
          throw new Error('The requested resource was not found.');
        case 409:
          throw new Error('This action conflicts with existing data.');
        case 429:
          throw new Error('Too many requests. Please slow down.');
        case 500:
          throw new Error('Server error. Please try again later.');
        default:
          throw new Error(error.message);
      }
    }
    throw error;
  }
}

/**
 * Example usage with error handling wrapper
 */
export async function safeLoginExample(credentials: LoginCredentials): Promise<AuthResponse> {
  return withErrorHandling(() => post<AuthResponse>('/auth/login', credentials), {
    401: 'Invalid email or password',
    429: 'Too many login attempts. Please try again in 15 minutes.',
  });
}
