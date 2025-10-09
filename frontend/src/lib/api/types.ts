/**
 * API Types and Interfaces
 *
 * Centralized type definitions for API requests and responses.
 */

// ============================================================================
// Common Types
// ============================================================================

export interface ApiError {
  error: string;
  message: string;
  status_code: number;
  details?: Record<string, any>;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

// ============================================================================
// Authentication Types
// ============================================================================

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  user_type: 'volunteer' | 'organization_admin';
  security_questions: SecurityQuestion[];
}

export interface SecurityQuestion {
  question: string;
  answer: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface AuthResponse {
  user: User;
  tokens: AuthTokens;
}

// ============================================================================
// User Types
// ============================================================================

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: 'volunteer' | 'organization_admin' | 'super_admin';
  status: 'active' | 'inactive' | 'suspended';
  created_at: string;
  updated_at: string;
  last_login_at: string | null;
  profile_photo_url?: string;
}

export interface UserProfile extends User {
  phone?: string;
  date_of_birth?: string;
  emergency_contact_name?: string;
  emergency_contact_phone?: string;
}

// ============================================================================
// Organization Types
// ============================================================================

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description: string;
  mission?: string;
  website?: string;
  email: string;
  phone?: string;
  address: string;
  city: string;
  state: string;
  zip_code: string;
  country: string;
  latitude?: number;
  longitude?: number;
  logo_url?: string;
  verification_status: 'verified' | 'pending' | 'rejected';
  created_at: string;
  updated_at: string;
}

// ============================================================================
// Volunteer Types
// ============================================================================

export interface VolunteerProfile {
  id: string;
  user_id: string;
  bio?: string;
  location?: string;
  latitude?: number;
  longitude?: number;
  skills: string[];
  interests: string[];
  availability: {
    monday?: boolean;
    tuesday?: boolean;
    wednesday?: boolean;
    thursday?: boolean;
    friday?: boolean;
    saturday?: boolean;
    sunday?: boolean;
  };
  total_hours: number;
  privacy_settings: {
    show_profile: boolean;
    show_hours: boolean;
    show_location: boolean;
  };
  created_at: string;
  updated_at: string;
}

export interface VolunteerDashboard {
  total_hours: number;
  total_events: number;
  total_organizations: number;
  upcoming_events: number;
  recent_achievements: Achievement[];
  hours_by_month: Array<{
    month: string;
    hours: number;
  }>;
}

// ============================================================================
// Opportunity Types
// ============================================================================

export interface Opportunity {
  id: string;
  organization_id: string;
  title: string;
  description: string;
  status: 'draft' | 'published' | 'cancelled' | 'completed';
  start_time: string;
  end_time: string;
  location: string;
  address: string;
  city: string;
  state: string;
  zip_code: string;
  latitude?: number;
  longitude?: number;
  capacity: number;
  registered_count: number;
  waitlist_count: number;
  min_age?: number;
  causes: string[];
  required_skills: string[];
  is_recurring: boolean;
  recurrence_pattern?: string;
  created_at: string;
  updated_at: string;
}

export interface OpportunitySearchParams extends PaginationParams {
  search?: string;
  latitude?: number;
  longitude?: number;
  radius?: number;
  cause?: string;
  start_date?: string;
  end_date?: string;
  skills?: string[];
  min_age?: number;
}

// ============================================================================
// Registration Types
// ============================================================================

export interface Registration {
  id: string;
  opportunity_id: string;
  volunteer_id: string;
  status: 'registered' | 'checked_in' | 'completed' | 'cancelled' | 'waitlisted';
  checked_in_at?: string;
  cancelled_at?: string;
  cancellation_reason?: string;
  hours_logged?: number;
  hours_status?: 'pending' | 'verified' | 'disputed';
  created_at: string;
  updated_at: string;
}

// ============================================================================
// Hours Types
// ============================================================================

export interface HoursLog {
  id: string;
  registration_id: string;
  volunteer_id: string;
  opportunity_id: string;
  hours: number;
  status: 'pending' | 'verified' | 'disputed';
  logged_by_user_id: string;
  coordinator_notes?: string;
  volunteer_notes?: string;
  verified_at?: string;
  disputed_at?: string;
  created_at: string;
  updated_at: string;
}

// ============================================================================
// Communication Types
// ============================================================================

export interface Message {
  id: string;
  sender_id: string;
  subject: string;
  body: string;
  message_type: 'direct' | 'broadcast';
  created_at: string;
}

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  message: string;
  is_read: boolean;
  related_entity_type?: string;
  related_entity_id?: string;
  created_at: string;
}

// ============================================================================
// Analytics Types
// ============================================================================

export interface Achievement {
  id: string;
  name: string;
  description: string;
  icon_url?: string;
  criteria: string;
  earned_at?: string;
}

export interface ImpactReport {
  volunteer_id: string;
  total_hours: number;
  total_events: number;
  total_organizations: number;
  achievements: Achievement[];
  hours_by_cause: Array<{
    cause: string;
    hours: number;
  }>;
  generated_at: string;
}
