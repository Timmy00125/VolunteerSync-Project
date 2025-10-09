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

export interface CreateOrganizationInput {
  name: string;
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
  logo_url?: string;
  cause_ids?: string[];
}

export interface UpdateOrganizationInput {
  name?: string;
  description?: string;
  mission?: string;
  website?: string;
  email?: string;
  phone?: string;
  address?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  country?: string;
  logo_url?: string;
  cause_ids?: string[];
}

export interface OrganizationDashboard {
  organization: Organization;
  metrics: {
    total_volunteers: number;
    active_volunteers: number;
    total_hours: number;
    hours_this_month: number;
    total_events: number;
    upcoming_events: number;
    events_this_month: number;
    volunteer_retention_rate: number;
  };
  upcoming_events: UpcomingOrganizationEvent[];
  recent_registrations: RecentRegistration[];
}

export interface UpcomingOrganizationEvent {
  id: string;
  title: string;
  date: string;
  start_time: string;
  registered_count: number;
  capacity: number;
  status: string;
}

export interface RecentRegistration {
  id: string;
  volunteer_name: string;
  opportunity_title: string;
  registered_at: string;
  status: string;
}

export interface OrganizationAnalytics {
  volunteers_by_cause: Array<{
    cause: string;
    count: number;
  }>;
  hours_over_time: Array<{
    month: string;
    hours: number;
  }>;
  volunteer_retention_rate: number;
  event_completion_rate: number;
  average_volunteers_per_event: number;
  top_volunteers: Array<{
    id: string;
    name: string;
    hours: number;
    events: number;
  }>;
}

export interface TeamMember {
  id: string;
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: 'admin' | 'coordinator' | 'volunteer';
  joined_at: string;
}

export interface InviteMemberInput {
  email: string;
  role: 'admin' | 'coordinator';
}

// ============================================================================
// Volunteer Types
// ============================================================================

export interface VolunteerProfile {
  id: string;
  user_id: string;
  profile_photo_url?: string;
  biography?: string;
  location?: string;
  latitude?: number;
  longitude?: number;
  availability_monday: boolean;
  availability_tuesday: boolean;
  availability_wednesday: boolean;
  availability_thursday: boolean;
  availability_friday: boolean;
  availability_saturday: boolean;
  availability_sunday: boolean;
  preferred_time?: 'morning' | 'afternoon' | 'evening' | 'flexible';
  total_hours: number;
  total_events: number;
  emergency_contact_name?: string;
  emergency_contact_phone?: string;
  privacy_show_hours: boolean;
  privacy_show_events: boolean;
  privacy_show_organizations: boolean;
  notification_in_app: boolean;
  notification_browser_push: boolean;
  skills: Skill[];
  interests: Cause[];
  created_at: string;
  updated_at: string;
}

export interface Skill {
  id: string;
  name: string;
  category?: string;
}

export interface Cause {
  id: string;
  name: string;
  slug?: string;
  description?: string;
}

export interface UpdateVolunteerProfileInput {
  profile_photo_url?: string;
  biography?: string;
  location?: string;
  availability?: {
    monday?: boolean;
    tuesday?: boolean;
    wednesday?: boolean;
    thursday?: boolean;
    friday?: boolean;
    saturday?: boolean;
    sunday?: boolean;
  };
  preferred_time?: 'morning' | 'afternoon' | 'evening' | 'flexible';
  emergency_contact_name?: string;
  emergency_contact_phone?: string;
  privacy_settings?: {
    show_hours?: boolean;
    show_events?: boolean;
    show_organizations?: boolean;
  };
  notification_settings?: {
    in_app?: boolean;
    browser_push?: boolean;
  };
  skill_ids?: string[];
  interest_ids?: string[];
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

// Enhanced Dashboard types matching backend API response
export interface DashboardResponse {
  profile: VolunteerProfile;
  total_hours: number;
  total_events: number;
  total_organizations: number;
  recent_events: RecentEvent[];
  upcoming_events: UpcomingEvent[];
  achievements: Achievement[];
  hours_this_month: number;
  events_this_month: number;
}

export interface RecentEvent {
  id: string;
  opportunity_title: string;
  organization_name: string;
  date: string;
  hours_logged: number;
  status: string;
}

export interface UpcomingEvent {
  id: string;
  opportunity_title: string;
  organization_name: string;
  date: string;
  start_time: string;
  end_time: string;
  location: string;
  status: string;
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
