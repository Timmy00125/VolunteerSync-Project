/**
 * Organization Validation Schemas
 *
 * Zod schemas for organization management:
 * - Organization creation
 * - Organization updates
 * - Organization search/filters
 * - Team member management
 *
 * Based on OpenAPI spec and FR-010 through FR-016 requirements
 */

import { z } from 'zod';
import {
  nameSchema,
  emailSchema,
  phoneSchema,
  shortTextSchema,
  mediumTextSchema,
  longTextSchema,
  urlSchema,
  slugSchema,
  imageUrlSchema,
  addressLine1Schema,
  addressLine2Schema,
  citySchema,
  stateSchema,
  postalCodeSchema,
  countrySchema,
  makePartial,
} from './common';

/**
 * Organization Verification Status Enum
 */
export const OrganizationVerificationStatus = z.enum(['verified', 'unverified'], {
  message: 'Invalid verification status',
});

export type OrganizationVerificationStatusEnum = z.infer<typeof OrganizationVerificationStatus>;

/**
 * Organization Member Role Enum
 */
export const OrganizationMemberRole = z.enum(['admin', 'coordinator'], {
  message: 'Invalid member role',
});

export type OrganizationMemberRoleEnum = z.infer<typeof OrganizationMemberRole>;

/**
 * Organization Create Schema
 *
 * For creating new organization profiles
 * Auto-verified on creation (FR-015)
 * Slug auto-generated from name
 * Geocoding triggered on address save
 *
 * POST /api/v1/organizations
 */
export const createOrganizationSchema = z.object({
  name: nameSchema.min(3, 'Organization name must be at least 3 characters'),
  mission_statement: mediumTextSchema.optional(),
  description: longTextSchema.optional(),

  // Contact information
  email: emailSchema,
  phone: phoneSchema,
  website: urlSchema,

  // Address
  address_line_1: addressLine1Schema,
  address_line_2: addressLine2Schema,
  city: citySchema,
  state: stateSchema,
  postal_code: postalCodeSchema,
  country: countrySchema,

  // Branding
  logo_url: imageUrlSchema,
  banner_url: imageUrlSchema,

  // Categories
  cause_categories: z.array(z.string()).min(1, 'At least one cause category is required'),
});

export type CreateOrganizationData = z.infer<typeof createOrganizationSchema>;

/**
 * Organization Update Schema (Partial)
 *
 * All fields optional for PATCH operations
 * Only admins can update (enforced by middleware)
 *
 * PATCH /api/v1/organizations/{id}
 */
export const updateOrganizationSchema = makePartial(createOrganizationSchema).extend({
  slug: slugSchema.optional(),
});

export type UpdateOrganizationData = z.infer<typeof updateOrganizationSchema>;

/**
 * Organization Search/Filter Schema
 *
 * For searching and filtering organizations
 *
 * GET /api/v1/organizations
 */
export const organizationSearchSchema = z.object({
  // Text search
  search: z.string().optional(),
  q: z.string().optional(),

  // Location filters
  city: z.string().optional(),
  state: z.string().optional(),

  // Category filter
  cause: z.string().optional(),
  causes: z.array(z.string()).optional(),

  // Verification filter
  verified: z.boolean().optional(),

  // Pagination
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().positive().max(100).default(20),

  // Sorting
  sort_by: z
    .enum(['name', 'created_at', 'total_volunteers', 'total_hours'], {
      message: 'Invalid sort option',
    })
    .default('name'),
  sort_order: z.enum(['asc', 'desc'], { message: 'Invalid sort order' }).default('asc'),
});

export type OrganizationSearchParams = z.infer<typeof organizationSearchSchema>;

/**
 * Invite Team Member Schema
 *
 * For inviting users to join organization team
 *
 * POST /api/v1/organizations/{id}/members/invite
 */
export const inviteMemberSchema = z.object({
  email: emailSchema,
  role: OrganizationMemberRole,
  message: mediumTextSchema.optional(),
});

export type InviteMemberData = z.infer<typeof inviteMemberSchema>;

/**
 * Update Member Role Schema
 *
 * For changing team member roles
 *
 * PATCH /api/v1/organizations/{id}/members/{userId}
 */
export const updateMemberRoleSchema = z.object({
  role: OrganizationMemberRole,
});

export type UpdateMemberRoleData = z.infer<typeof updateMemberRoleSchema>;

/**
 * Document Upload Schema
 *
 * For uploading required documents (background check, orientation materials, etc.)
 *
 * POST /api/v1/organizations/{id}/documents
 */
export const uploadDocumentSchema = z.object({
  title: shortTextSchema.min(3, 'Document title is required'),
  description: mediumTextSchema.optional(),
  document_type: z.enum(
    ['background_check_form', 'orientation_material', 'waiver', 'policy', 'other'],
    { message: 'Invalid document type' }
  ),
  file_url: z.string().url('Must be a valid file URL'),
  required_for_volunteers: z.boolean().default(false),
});

export type UploadDocumentData = z.infer<typeof uploadDocumentSchema>;

/**
 * Organization Analytics Request Schema
 *
 * For requesting organization analytics/reports
 *
 * GET /api/v1/analytics/organization/{id}
 */
export const organizationAnalyticsSchema = z.object({
  // Date range
  start_date: z.coerce.date().optional(),
  end_date: z.coerce.date().optional(),

  // Metrics to include
  include_volunteers: z.boolean().default(true),
  include_hours: z.boolean().default(true),
  include_opportunities: z.boolean().default(true),
  include_retention: z.boolean().default(false),

  // Format
  format: z.enum(['json', 'pdf'], { message: 'Invalid format' }).default('json'),
});

export type OrganizationAnalyticsParams = z.infer<typeof organizationAnalyticsSchema>;

/**
 * Message Broadcast Schema
 *
 * For sending messages to volunteers
 *
 * POST /api/v1/messages
 */
export const sendMessageSchema = z
  .object({
    subject: shortTextSchema.min(3, 'Subject is required'),
    body: longTextSchema.min(10, 'Message body is required'),

    // Recipients
    recipient_type: z.enum(['direct', 'opportunity', 'all_volunteers'], {
      message: 'Invalid recipient type',
    }),
    recipient_ids: z.array(z.string().uuid()).optional(),
    opportunity_id: z.string().uuid().optional(),

    // Options
    send_email: z.boolean().default(true),
    send_notification: z.boolean().default(true),
  })
  .refine(
    (data) => {
      // If direct message, must have recipient_ids
      if (data.recipient_type === 'direct') {
        return data.recipient_ids && data.recipient_ids.length > 0;
      }
      // If opportunity message, must have opportunity_id
      if (data.recipient_type === 'opportunity') {
        return data.opportunity_id;
      }
      return true;
    },
    {
      message: 'Invalid recipient configuration',
      path: ['recipient_ids'],
    }
  );

export type SendMessageData = z.infer<typeof sendMessageSchema>;

/**
 * Team Schema
 *
 * For creating volunteer teams/groups within an organization
 *
 * POST /api/v1/teams
 */
export const createTeamSchema = z.object({
  organization_id: z.string().uuid('Invalid organization ID'),
  name: nameSchema.min(3, 'Team name must be at least 3 characters'),
  description: mediumTextSchema.optional(),
  is_public: z.boolean().default(true),
  max_members: z.number().int().positive().optional(),
});

export type CreateTeamData = z.infer<typeof createTeamSchema>;

/**
 * Helper: Generate slug from organization name
 */
export const generateSlug = (name: string): string => {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '') // Remove special characters
    .replace(/\s+/g, '-') // Replace spaces with hyphens
    .replace(/-+/g, '-') // Replace multiple hyphens with single hyphen
    .replace(/^-+|-+$/g, ''); // Remove leading/trailing hyphens
};

/**
 * Helper: Validate EIN (Employer Identification Number) format
 * Optional for organizations
 */
export const validateEIN = (ein: string): boolean => {
  // Format: XX-XXXXXXX (9 digits with hyphen after 2nd digit)
  return /^\d{2}-\d{7}$/.test(ein);
};

/**
 * Common organization types
 */
export const ORGANIZATION_TYPES = [
  '501(c)(3) Nonprofit',
  'Charitable Organization',
  'Community Group',
  'Educational Institution',
  'Faith-Based Organization',
  'Government Agency',
  'Healthcare Organization',
  'Professional Association',
  'Social Service Organization',
  'Youth Organization',
] as const;
