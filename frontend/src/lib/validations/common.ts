/**
 * Common Validation Schemas and Utilities
 *
 * Shared validation schemas and helper functions for form validation
 * using Zod. These can be reused across different form schemas.
 */

import { z } from 'zod';

/**
 * Email validation schema
 * - Must be valid email format
 * - Case insensitive
 */
export const emailSchema = z
  .string()
  .min(1, 'Email is required')
  .email('Must be a valid email address')
  .toLowerCase();

/**
 * Password validation schema
 * - Minimum 8 characters
 * - Must contain letters and numbers
 * - Per FR-002 requirements
 */
export const passwordSchema = z
  .string()
  .min(8, 'Password must be at least 8 characters')
  .regex(/^(?=.*[A-Za-z])(?=.*\d)/, 'Password must contain at least one letter and one number');

/**
 * Phone number validation schema (optional)
 * - Allows various international formats
 * - Can be null or empty
 */
export const phoneSchema = z
  .string()
  .optional()
  .nullable()
  .refine(
    (val) => {
      if (!val) return true;
      // Basic phone validation - allows +, -, spaces, and parentheses
      return /^[+]?[\d\s\-()]+$/.test(val) && val.replace(/\D/g, '').length >= 10;
    },
    { message: 'Must be a valid phone number with at least 10 digits' }
  );

/**
 * Name validation schema
 * - Non-empty string
 * - Trims whitespace
 */
export const nameSchema = z.string().min(1, 'This field is required').trim();

/**
 * URL validation schema (optional)
 * - Must be valid URL if provided
 * - Can be null or empty
 */
export const urlSchema = z
  .string()
  .url('Must be a valid URL')
  .optional()
  .nullable()
  .or(z.literal(''));

/**
 * Location/Address schemas
 */
export const addressLine1Schema = z.string().min(1, 'Street address is required').trim();
export const addressLine2Schema = z.string().optional().nullable();
export const citySchema = z.string().min(1, 'City is required').trim();
export const stateSchema = z.string().min(1, 'State/Province is required').trim();
export const postalCodeSchema = z.string().min(1, 'Postal code is required').trim();
export const countrySchema = z.string().default('United States');

/**
 * Complete address schema
 */
export const addressSchema = z.object({
  address_line_1: addressLine1Schema,
  address_line_2: addressLine2Schema,
  city: citySchema,
  state: stateSchema,
  postal_code: postalCodeSchema,
  country: countrySchema,
});

/**
 * Date/Time validation schemas
 */
export const dateSchema = z.coerce.date({
  message: 'Must be a valid date',
});

export const futureDateSchema = dateSchema.refine((date) => date > new Date(), {
  message: 'Date must be in the future',
});

export const pastDateSchema = dateSchema.refine((date) => date < new Date(), {
  message: 'Date must be in the past',
});

/**
 * Numeric validation schemas
 */
export const positiveNumberSchema = z
  .number({ message: 'Must be a number' })
  .positive('Must be greater than 0');

export const nonNegativeNumberSchema = z
  .number({ message: 'Must be a number' })
  .nonnegative('Must be 0 or greater');

/**
 * Text content schemas
 */
export const shortTextSchema = z.string().max(255, 'Maximum 255 characters').trim();
export const mediumTextSchema = z.string().max(1000, 'Maximum 1000 characters').trim();
export const longTextSchema = z.string().max(5000, 'Maximum 5000 characters').trim();

/**
 * Slug validation
 * - Lowercase
 * - Alphanumeric with hyphens
 * - No spaces
 */
export const slugSchema = z
  .string()
  .min(1, 'Slug is required')
  .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, 'Slug must be lowercase letters, numbers, and hyphens only');

/**
 * Image URL validation (for uploads or external URLs)
 */
export const imageUrlSchema = z
  .string()
  .url('Must be a valid URL')
  .optional()
  .nullable()
  .or(z.literal(''));

/**
 * Array validation helpers
 */
export const nonEmptyArraySchema = <T extends z.ZodTypeAny>(itemSchema: T) =>
  z.array(itemSchema).min(1, 'At least one item is required');

/**
 * Enum helpers
 */
export const createEnumSchema = <T extends readonly [string, ...string[]]>(
  values: T,
  errorMessage?: string
) =>
  z.enum(values, {
    message: errorMessage || 'Invalid selection',
  });

/**
 * Security question validation
 */
export const securityQuestionSchema = z.object({
  question: z.string().min(1, 'Security question is required'),
  answer: z.string().min(1, 'Answer is required'),
});

/**
 * Pagination parameters
 */
export const paginationSchema = z.object({
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().positive().max(100).default(20),
});

/**
 * Search/Filter parameters
 */
export const searchSchema = z.object({
  q: z.string().optional(),
  search: z.string().optional(),
});

/**
 * Helper: Make all fields optional (for PATCH operations)
 */
export type PartialSchema<T extends z.ZodRawShape> = {
  [K in keyof T]: z.ZodOptional<T[K]>;
};

/**
 * Helper: Create a partial version of a schema for updates
 */
export const makePartial = <T extends z.ZodRawShape>(schema: z.ZodObject<T>) => schema.partial();
