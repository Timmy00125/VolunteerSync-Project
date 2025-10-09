/**
 * Validation Schemas Index
 *
 * Central export point for all Zod validation schemas.
 * Import schemas from this file to use in forms and API calls.
 *
 * Usage:
 * ```ts
 * import { loginSchema, registerSchema } from '@/lib/validations';
 * import type { LoginFormData } from '@/lib/validations';
 * ```
 */

// ============================================================================
// Common Schemas and Utilities
// ============================================================================
export {
  emailSchema,
  passwordSchema,
  phoneSchema,
  nameSchema,
  urlSchema,
  addressSchema,
  addressLine1Schema,
  addressLine2Schema,
  citySchema,
  stateSchema,
  postalCodeSchema,
  countrySchema,
  dateSchema,
  futureDateSchema,
  pastDateSchema,
  positiveNumberSchema,
  nonNegativeNumberSchema,
  shortTextSchema,
  mediumTextSchema,
  longTextSchema,
  slugSchema,
  imageUrlSchema,
  nonEmptyArraySchema,
  createEnumSchema,
  securityQuestionSchema,
  paginationSchema,
  searchSchema,
  makePartial,
} from './common';

export type { PartialSchema } from './common';

// ============================================================================
// Authentication Schemas
// ============================================================================
export {
  UserType,
  registerSchema,
  loginSchema,
  passwordResetRequestSchema,
  passwordResetVerifySchema,
  passwordResetConfirmSchema,
  refreshTokenSchema,
  SECURITY_QUESTIONS,
  createEmptySecurityQuestions,
  validateUniqueSecurityQuestions,
} from './auth';

export type {
  UserTypeEnum,
  RegisterFormData,
  LoginFormData,
  PasswordResetRequestData,
  PasswordResetVerifyData,
  PasswordResetConfirmData,
  RefreshTokenData,
} from './auth';

// ============================================================================
// Profile Schemas
// ============================================================================
export {
  updateUserProfileSchema,
  availabilityDaySchema,
  emergencyContactSchema,
  privacySettingsSchema,
  volunteerProfileSchema,
  updateVolunteerProfileSchema,
  skillSchema,
  causeSchema,
  COMMON_CAUSES,
  COMMON_SKILLS,
  createDefaultAvailability,
} from './profile';

export type {
  UpdateUserProfileData,
  AvailabilityDay,
  EmergencyContact,
  PrivacySettings,
  VolunteerProfileData,
  UpdateVolunteerProfileData,
  Skill,
  Cause,
} from './profile';

// ============================================================================
// Opportunity Schemas
// ============================================================================
export {
  OpportunityStatus,
  RecurrencePattern,
  createOpportunitySchema,
  updateOpportunitySchema,
  opportunitySearchSchema,
  createRegistrationSchema,
  RegistrationStatus,
  cancelRegistrationSchema,
  checkInSchema,
  logHoursSchema,
  disputeHoursSchema,
  opportunityReviewSchema,
  calculateDuration,
  validateTimeRange,
} from './opportunity';

export type {
  OpportunityStatusEnum,
  RecurrencePatternEnum,
  CreateOpportunityData,
  UpdateOpportunityData,
  OpportunitySearchParams,
  CreateRegistrationData,
  RegistrationStatusEnum,
  CancelRegistrationData,
  CheckInData,
  LogHoursData,
  DisputeHoursData,
  OpportunityReviewData,
} from './opportunity';

// ============================================================================
// Organization Schemas
// ============================================================================
export {
  OrganizationVerificationStatus,
  OrganizationMemberRole,
  createOrganizationSchema,
  updateOrganizationSchema,
  organizationSearchSchema,
  inviteMemberSchema,
  updateMemberRoleSchema,
  uploadDocumentSchema,
  organizationAnalyticsSchema,
  sendMessageSchema,
  createTeamSchema,
  generateSlug,
  validateEIN,
  ORGANIZATION_TYPES,
} from './organization';

export type {
  OrganizationVerificationStatusEnum,
  OrganizationMemberRoleEnum,
  CreateOrganizationData,
  UpdateOrganizationData,
  OrganizationSearchParams,
  InviteMemberData,
  UpdateMemberRoleData,
  UploadDocumentData,
  OrganizationAnalyticsParams,
  SendMessageData,
  CreateTeamData,
} from './organization';
