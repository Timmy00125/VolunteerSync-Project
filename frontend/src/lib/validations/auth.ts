/**
 * Authentication Form Validation Schemas
 *
 * Zod schemas for user authentication flows:
 * - Registration
 * - Login
 * - Password Reset (Request, Verify, Confirm)
 *
 * Based on OpenAPI spec and FR-001, FR-002, FR-003 requirements
 */

import { z } from 'zod';
import {
  emailSchema,
  passwordSchema,
  phoneSchema,
  nameSchema,
  securityQuestionSchema,
} from './common';

/**
 * User Type enum for registration
 */
export const UserType = z.enum(['volunteer', 'organization_admin'], {
  message: 'Please select a user type',
});

export type UserTypeEnum = z.infer<typeof UserType>;

/**
 * Registration Form Schema
 *
 * Requirements:
 * - Email (valid format, required)
 * - Password (min 8 chars, letters + numbers)
 * - First name and last name
 * - User type (volunteer or organization_admin)
 * - Exactly 3 security questions with answers (FR-003)
 * - Optional phone number
 *
 * POST /api/v1/auth/register
 */
export const registerSchema = z.object({
  email: emailSchema,
  password: passwordSchema,
  first_name: nameSchema,
  last_name: nameSchema,
  phone: phoneSchema,
  user_type: UserType,
  security_questions: z
    .array(securityQuestionSchema)
    .length(3, 'Exactly 3 security questions are required'),
});

export type RegisterFormData = z.infer<typeof registerSchema>;

/**
 * Login Form Schema
 *
 * Simple email + password authentication
 *
 * POST /api/v1/auth/login
 */
export const loginSchema = z.object({
  email: emailSchema,
  password: z.string().min(1, 'Password is required'),
});

export type LoginFormData = z.infer<typeof loginSchema>;

/**
 * Password Reset Request Schema
 *
 * Initiate password reset by providing email
 * Returns security questions for verification
 *
 * POST /api/v1/auth/password-reset/request
 */
export const passwordResetRequestSchema = z.object({
  email: emailSchema,
});

export type PasswordResetRequestData = z.infer<typeof passwordResetRequestSchema>;

/**
 * Password Reset Verify Schema
 *
 * Verify security question answers
 * Minimum 2 of 3 correct answers required (FR-003a)
 *
 * POST /api/v1/auth/password-reset/verify
 */
export const passwordResetVerifySchema = z.object({
  reset_token: z.string().min(1, 'Reset token is required'),
  answers: z.array(z.string().min(1, 'Answer cannot be empty')).length(3, 'Must provide 3 answers'),
});

export type PasswordResetVerifyData = z.infer<typeof passwordResetVerifySchema>;

/**
 * Password Reset Confirm Schema
 *
 * Set new password after security question verification
 *
 * POST /api/v1/auth/password-reset/confirm
 */
export const passwordResetConfirmSchema = z
  .object({
    verified_token: z.string().min(1, 'Verified token is required'),
    new_password: passwordSchema,
    confirm_password: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: 'Passwords do not match',
    path: ['confirm_password'],
  });

export type PasswordResetConfirmData = z.infer<typeof passwordResetConfirmSchema>;

/**
 * Token Refresh Schema
 *
 * POST /api/v1/auth/refresh
 */
export const refreshTokenSchema = z.object({
  refresh_token: z.string().min(1, 'Refresh token is required'),
});

export type RefreshTokenData = z.infer<typeof refreshTokenSchema>;

/**
 * Predefined Security Questions
 *
 * Common security questions users can choose from
 */
export const SECURITY_QUESTIONS = [
  "What is your mother's maiden name?",
  'What was the name of your first pet?',
  'What city were you born in?',
  'What is the name of your favorite teacher?',
  'What was the make of your first car?',
  'What is your favorite book?',
  'What street did you grow up on?',
  "What is your father's middle name?",
  'What was the name of your elementary school?',
  'What is your favorite movie?',
] as const;

/**
 * Helper: Default security questions structure
 */
export const createEmptySecurityQuestions = () => [
  { question: '', answer: '' },
  { question: '', answer: '' },
  { question: '', answer: '' },
];

/**
 * Helper: Validate security questions are unique
 */
export const validateUniqueSecurityQuestions = (
  questions: Array<{ question: string; answer: string }>
): boolean => {
  const questionTexts = questions.map((q) => q.question);
  const uniqueQuestions = new Set(questionTexts);
  return uniqueQuestions.size === questionTexts.length;
};
