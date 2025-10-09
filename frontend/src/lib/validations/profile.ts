/**
 * Profile Validation Schemas
 *
 * Zod schemas for user and volunteer profile management:
 * - User profile updates
 * - Volunteer profile creation and updates
 *
 * Based on OpenAPI spec and data model requirements
 */

import { z } from 'zod';
import {
  emailSchema,
  phoneSchema,
  nameSchema,
  shortTextSchema,
  mediumTextSchema,
  imageUrlSchema,
  makePartial,
} from './common';

/**
 * User Profile Update Schema
 *
 * For updating basic user information
 * Email changes require reverification (FR-007)
 *
 * PATCH /api/v1/users/me
 */
export const updateUserProfileSchema = z.object({
  first_name: nameSchema.optional(),
  last_name: nameSchema.optional(),
  phone: phoneSchema,
  email: emailSchema.optional(),
});

export type UpdateUserProfileData = z.infer<typeof updateUserProfileSchema>;

/**
 * Availability Day Schema
 */
export const availabilityDaySchema = z.object({
  day: z.enum(['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'], {
    message: 'Invalid day of week',
  }),
  available: z.boolean().default(false),
  time_slots: z
    .array(
      z.object({
        start: z.string().regex(/^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Invalid time format (HH:MM)'),
        end: z.string().regex(/^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Invalid time format (HH:MM)'),
      })
    )
    .optional(),
});

export type AvailabilityDay = z.infer<typeof availabilityDaySchema>;

/**
 * Emergency Contact Schema
 */
export const emergencyContactSchema = z.object({
  name: nameSchema,
  relationship: shortTextSchema,
  phone: z.string().min(10, 'Phone number is required'),
  email: emailSchema.optional(),
});

export type EmergencyContact = z.infer<typeof emergencyContactSchema>;

/**
 * Privacy Settings Schema
 */
export const privacySettingsSchema = z.object({
  show_profile_to_orgs: z.boolean().default(true),
  show_hours_publicly: z.boolean().default(true),
  show_location: z.boolean().default(true),
  allow_org_messages: z.boolean().default(true),
});

export type PrivacySettings = z.infer<typeof privacySettingsSchema>;

/**
 * Volunteer Profile Schema
 *
 * Complete volunteer profile with all fields
 *
 * PATCH /api/v1/volunteers/me
 */
export const volunteerProfileSchema = z.object({
  bio: mediumTextSchema.optional(),
  photo_url: imageUrlSchema,
  date_of_birth: z.coerce.date().optional(),

  // Location (triggers geocoding)
  address_line_1: z.string().optional(),
  address_line_2: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  postal_code: z.string().optional(),
  country: z.string().default('United States'),

  // Skills (array of skill IDs or names)
  skills: z.array(z.string()).optional(),

  // Interests/Causes (array of cause IDs or names)
  interests: z.array(z.string()).optional(),

  // Availability
  availability: z.array(availabilityDaySchema).optional(),

  // Emergency contact
  emergency_contact: emergencyContactSchema.optional(),

  // Privacy settings
  privacy_settings: privacySettingsSchema.optional(),

  // Additional fields
  languages: z.array(z.string()).optional(),
  t_shirt_size: z
    .enum(['XS', 'S', 'M', 'L', 'XL', 'XXL', '3XL'], {
      message: 'Invalid t-shirt size',
    })
    .optional(),
  dietary_restrictions: shortTextSchema.optional(),
  accessibility_needs: mediumTextSchema.optional(),
});

export type VolunteerProfileData = z.infer<typeof volunteerProfileSchema>;

/**
 * Volunteer Profile Update Schema (Partial)
 *
 * All fields optional for PATCH operations
 */
export const updateVolunteerProfileSchema = makePartial(volunteerProfileSchema);

export type UpdateVolunteerProfileData = z.infer<typeof updateVolunteerProfileSchema>;

/**
 * Skill Schema
 */
export const skillSchema = z.object({
  id: z.string().optional(),
  name: shortTextSchema,
  category: shortTextSchema.optional(),
});

export type Skill = z.infer<typeof skillSchema>;

/**
 * Cause/Interest Schema
 */
export const causeSchema = z.object({
  id: z.string().optional(),
  name: shortTextSchema,
  slug: z.string().optional(),
  description: mediumTextSchema.optional(),
});

export type Cause = z.infer<typeof causeSchema>;

/**
 * Common cause categories
 */
export const COMMON_CAUSES = [
  'Animal Welfare',
  'Arts & Culture',
  'Children & Youth',
  'Community Development',
  'Education & Literacy',
  'Environment',
  'Health & Wellness',
  'Homelessness & Housing',
  'Human Rights',
  'Hunger & Food Security',
  'LGBTQ+',
  'Seniors',
  'Sports & Recreation',
  'Veterans & Military',
  'Women & Girls',
] as const;

/**
 * Common skills
 */
export const COMMON_SKILLS = [
  'Administrative Support',
  'Arts & Crafts',
  'Coaching & Mentoring',
  'Communications & Marketing',
  'Computer Skills',
  'Construction & Repair',
  'Customer Service',
  'Data Entry',
  'Event Planning',
  'Fundraising',
  'Graphic Design',
  'Language Translation',
  'Legal Services',
  'Medical & Healthcare',
  'Photography',
  'Project Management',
  'Public Speaking',
  'Research & Analysis',
  'Social Media',
  'Teaching & Tutoring',
  'Video Production',
  'Web Development',
  'Writing & Editing',
] as const;

/**
 * Helper: Create default availability (all days unavailable)
 */
export const createDefaultAvailability = (): AvailabilityDay[] => [
  { day: 'monday', available: false },
  { day: 'tuesday', available: false },
  { day: 'wednesday', available: false },
  { day: 'thursday', available: false },
  { day: 'friday', available: false },
  { day: 'saturday', available: false },
  { day: 'sunday', available: false },
];
