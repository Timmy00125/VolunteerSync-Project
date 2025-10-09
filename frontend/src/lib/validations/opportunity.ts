/**
 * Opportunity Validation Schemas
 *
 * Zod schemas for volunteer opportunity/event management:
 * - Opportunity creation
 * - Opportunity updates
 * - Opportunity search/filters
 * - Registration
 *
 * Based on OpenAPI spec and FR-017 through FR-033 requirements
 */

import { z } from 'zod';
import {
  shortTextSchema,
  mediumTextSchema,
  longTextSchema,
  futureDateSchema,
  positiveNumberSchema,
  nonNegativeNumberSchema,
  addressLine1Schema,
  addressLine2Schema,
  citySchema,
  stateSchema,
  postalCodeSchema,
  countrySchema,
  imageUrlSchema,
  makePartial,
} from './common';

/**
 * Opportunity Status Enum
 */
export const OpportunityStatus = z.enum(['draft', 'published', 'cancelled', 'completed'], {
  message: 'Invalid opportunity status',
});

export type OpportunityStatusEnum = z.infer<typeof OpportunityStatus>;

/**
 * Recurrence Pattern Enum
 */
export const RecurrencePattern = z.enum(['none', 'daily', 'weekly', 'monthly'], {
  message: 'Invalid recurrence pattern',
});

export type RecurrencePatternEnum = z.infer<typeof RecurrencePattern>;

/**
 * Opportunity Create Schema
 *
 * For creating new volunteer opportunities
 * Geocoding triggered on address save
 * Can publish immediately or save as draft
 *
 * POST /api/v1/opportunities
 */
export const createOpportunitySchema = z
  .object({
    title: shortTextSchema.min(3, 'Title must be at least 3 characters'),
    description: longTextSchema.min(10, 'Description must be at least 10 characters'),

    // Date and time
    start_date: futureDateSchema,
    end_date: futureDateSchema,
    start_time: z
      .string()
      .regex(/^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Invalid time format (HH:MM)'),
    end_time: z.string().regex(/^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Invalid time format (HH:MM)'),

    // Location
    address_line_1: addressLine1Schema,
    address_line_2: addressLine2Schema,
    city: citySchema,
    state: stateSchema,
    postal_code: postalCodeSchema,
    country: countrySchema,

    // Capacity
    capacity: positiveNumberSchema.int('Capacity must be a whole number'),

    // Categories and requirements
    cause_categories: z.array(z.string()).min(1, 'At least one cause is required'),
    required_skills: z.array(z.string()).optional(),

    // Status
    status: OpportunityStatus.default('draft'),

    // Optional fields
    image_url: imageUrlSchema,
    contact_email: z.string().email('Must be a valid email').optional(),
    contact_phone: z.string().optional(),

    // Requirements
    minimum_age: nonNegativeNumberSchema.int().optional(),
    background_check_required: z.boolean().default(false),
    orientation_required: z.boolean().default(false),

    // Recurrence
    is_recurring: z.boolean().default(false),
    recurrence_pattern: RecurrencePattern.default('none'),
    recurrence_end_date: z.coerce.date().optional(),
  })
  .refine((data) => data.end_date >= data.start_date, {
    message: 'End date must be after start date',
    path: ['end_date'],
  })
  .refine(
    (data) => {
      // If recurring, must have recurrence pattern and end date
      if (data.is_recurring) {
        return data.recurrence_pattern !== 'none' && data.recurrence_end_date;
      }
      return true;
    },
    {
      message: 'Recurring events must have a pattern and end date',
      path: ['recurrence_pattern'],
    }
  );

export type CreateOpportunityData = z.infer<typeof createOpportunitySchema>;

/**
 * Opportunity Update Schema (Partial)
 *
 * Cannot edit past events (enforced in backend)
 * All fields optional for PATCH operations
 *
 * PATCH /api/v1/opportunities/{id}
 */
export const updateOpportunitySchema = makePartial(createOpportunitySchema);

export type UpdateOpportunityData = z.infer<typeof updateOpportunitySchema>;

/**
 * Opportunity Search/Filter Schema
 *
 * For searching and filtering opportunities
 * Performance requirement: <2 seconds for up to 100 results
 *
 * GET /api/v1/opportunities
 */
export const opportunitySearchSchema = z.object({
  // Text search
  search: z.string().optional(),
  q: z.string().optional(),

  // Location filters
  city: z.string().optional(),
  state: z.string().optional(),
  latitude: z.number().optional(),
  longitude: z.number().optional(),
  radius: z.number().positive().default(25), // miles

  // Date filters
  start_date: z.coerce.date().optional(),
  end_date: z.coerce.date().optional(),

  // Category filters
  cause: z.string().optional(),
  causes: z.array(z.string()).optional(),

  // Skill filters
  skills: z.array(z.string()).optional(),

  // Status filter
  status: OpportunityStatus.optional(),

  // Organization filter
  organization_id: z.string().uuid().optional(),

  // Capacity filter
  has_availability: z.boolean().optional(),

  // Pagination
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().positive().max(100).default(20),

  // Sorting
  sort_by: z
    .enum(['date', 'relevance', 'distance', 'capacity'], {
      message: 'Invalid sort option',
    })
    .default('date'),
  sort_order: z.enum(['asc', 'desc'], { message: 'Invalid sort order' }).default('asc'),
});

export type OpportunitySearchParams = z.infer<typeof opportunitySearchSchema>;

/**
 * Registration Create Schema
 *
 * For volunteers registering for opportunities
 * Duplicate registration prevention
 * Waitlist when at capacity
 * Overlapping event warning
 *
 * POST /api/v1/registrations
 */
export const createRegistrationSchema = z.object({
  opportunity_id: z.string().uuid('Invalid opportunity ID'),
  notes: mediumTextSchema.optional(),
  accept_terms: z.boolean().refine((val) => val === true, {
    message: 'You must accept the terms and conditions',
  }),
});

export type CreateRegistrationData = z.infer<typeof createRegistrationSchema>;

/**
 * Registration Status Enum
 */
export const RegistrationStatus = z.enum(
  ['registered', 'waitlisted', 'checked_in', 'completed', 'cancelled', 'no_show'],
  {
    message: 'Invalid registration status',
  }
);

export type RegistrationStatusEnum = z.infer<typeof RegistrationStatus>;

/**
 * Cancel Registration Schema
 *
 * PATCH /api/v1/registrations/{id}/cancel
 */
export const cancelRegistrationSchema = z.object({
  reason: mediumTextSchema.optional(),
});

export type CancelRegistrationData = z.infer<typeof cancelRegistrationSchema>;

/**
 * Check-in Schema
 *
 * PATCH /api/v1/registrations/{id}/check-in
 */
export const checkInSchema = z.object({
  notes: shortTextSchema.optional(),
});

export type CheckInData = z.infer<typeof checkInSchema>;

/**
 * Hours Log Schema
 *
 * For coordinators logging volunteer hours
 *
 * POST /api/v1/hours/log
 */
export const logHoursSchema = z.object({
  registration_id: z.string().uuid('Invalid registration ID'),
  hours: positiveNumberSchema.max(24, 'Hours cannot exceed 24 in a single entry'),
  date: z.coerce.date(),
  notes: mediumTextSchema.optional(),
});

export type LogHoursData = z.infer<typeof logHoursSchema>;

/**
 * Hours Dispute Schema
 *
 * For volunteers disputing logged hours
 *
 * POST /api/v1/hours/{id}/dispute
 */
export const disputeHoursSchema = z.object({
  reason: mediumTextSchema.min(10, 'Please provide a reason for the dispute'),
  proposed_hours: positiveNumberSchema.max(24).optional(),
});

export type DisputeHoursData = z.infer<typeof disputeHoursSchema>;

/**
 * Opportunity Review Schema
 *
 * For volunteers reviewing completed events
 */
export const opportunityReviewSchema = z.object({
  rating: z.number().int().min(1).max(5, 'Rating must be between 1 and 5'),
  review: mediumTextSchema.optional(),
  would_recommend: z.boolean(),
});

export type OpportunityReviewData = z.infer<typeof opportunityReviewSchema>;

/**
 * Helper: Calculate duration in hours from start and end times
 */
export const calculateDuration = (startTime: string, endTime: string): number => {
  const [startHour, startMin] = startTime.split(':').map(Number);
  const [endHour, endMin] = endTime.split(':').map(Number);

  if (
    startHour === undefined ||
    startMin === undefined ||
    endHour === undefined ||
    endMin === undefined
  ) {
    return 0;
  }

  const startMinutes = startHour * 60 + startMin;
  const endMinutes = endHour * 60 + endMin;

  return (endMinutes - startMinutes) / 60;
};

/**
 * Helper: Validate time range (end time after start time)
 */
export const validateTimeRange = (startTime: string, endTime: string): boolean => {
  const duration = calculateDuration(startTime, endTime);
  return duration > 0;
};
