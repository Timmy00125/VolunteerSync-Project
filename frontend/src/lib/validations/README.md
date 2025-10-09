# Form Validation Schemas

This directory contains Zod validation schemas for all forms and API requests in the VolunteerSync frontend application.

## Overview

All validation schemas are built using [Zod v4](https://zod.dev/) and are designed to work seamlessly with [React Hook Form](https://react-hook-form.com/) via `@hookform/resolvers`.

## File Structure

```
validations/
├── index.ts           # Central export point for all schemas
├── common.ts          # Shared validation utilities and base schemas
├── auth.ts            # Authentication-related schemas
├── profile.ts         # User and volunteer profile schemas
├── opportunity.ts     # Opportunity/event management schemas
└── organization.ts    # Organization management schemas
```

## Usage

### Basic Form Validation with React Hook Form

```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { loginSchema, type LoginFormData } from '@/lib/validations';

function LoginForm() {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = (data: LoginFormData) => {
    // data is fully typed and validated
    console.log(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('email')} />
      {errors.email && <span>{errors.email.message}</span>}

      <input type="password" {...register('password')} />
      {errors.password && <span>{errors.password.message}</span>}

      <button type="submit">Login</button>
    </form>
  );
}
```

### API Request Validation

```tsx
import { createOpportunitySchema } from '@/lib/validations';

async function createOpportunity(data: unknown) {
  // Validate before sending to API
  const validatedData = createOpportunitySchema.parse(data);

  // Now safely send to API
  const response = await fetch('/api/v1/opportunities', {
    method: 'POST',
    body: JSON.stringify(validatedData),
  });

  return response.json();
}
```

### Partial Updates (PATCH)

```tsx
import { updateVolunteerProfileSchema } from '@/lib/validations';

// All fields are optional for update operations
const updateData = updateVolunteerProfileSchema.parse({
  bio: 'New bio text',
  // Other fields are optional
});
```

## Schema Categories

### Common Schemas (`common.ts`)

Reusable validation schemas and utilities:

- **Basic types**: `emailSchema`, `passwordSchema`, `phoneSchema`, `nameSchema`
- **Address**: `addressSchema` (complete address), individual field schemas
- **Dates**: `dateSchema`, `futureDateSchema`, `pastDateSchema`
- **Numbers**: `positiveNumberSchema`, `nonNegativeNumberSchema`
- **Text**: `shortTextSchema` (max 255), `mediumTextSchema` (max 1000), `longTextSchema` (max 5000)
- **URLs**: `urlSchema`, `imageUrlSchema`
- **Utilities**: `makePartial()`, `createEnumSchema()`, `paginationSchema`

### Authentication Schemas (`auth.ts`)

User authentication and account management:

- `registerSchema` - User registration with security questions
- `loginSchema` - Email + password login
- `passwordResetRequestSchema` - Initiate password reset
- `passwordResetVerifySchema` - Verify security questions (2 of 3 correct)
- `passwordResetConfirmSchema` - Set new password
- `refreshTokenSchema` - Token refresh

**Enums**: `UserType` (volunteer, organization_admin)

**Constants**: `SECURITY_QUESTIONS` - Predefined security question list

### Profile Schemas (`profile.ts`)

User and volunteer profile management:

- `updateUserProfileSchema` - Basic user info updates
- `volunteerProfileSchema` - Complete volunteer profile
- `updateVolunteerProfileSchema` - Partial volunteer profile updates
- `emergencyContactSchema` - Emergency contact information
- `privacySettingsSchema` - Privacy preferences
- `availabilityDaySchema` - Weekly availability with time slots

**Enums**: Days of week for availability

**Constants**: `COMMON_CAUSES`, `COMMON_SKILLS` - Predefined lists for dropdowns

### Opportunity Schemas (`opportunity.ts`)

Volunteer opportunity/event management:

- `createOpportunitySchema` - Create new opportunity
- `updateOpportunitySchema` - Update existing opportunity
- `opportunitySearchSchema` - Search/filter opportunities
- `createRegistrationSchema` - Register for opportunity
- `cancelRegistrationSchema` - Cancel registration
- `checkInSchema` - Check-in on event day
- `logHoursSchema` - Log volunteer hours
- `disputeHoursSchema` - Dispute logged hours
- `opportunityReviewSchema` - Review completed event

**Enums**:

- `OpportunityStatus` (draft, published, cancelled, completed)
- `RecurrencePattern` (none, daily, weekly, monthly)
- `RegistrationStatus` (registered, waitlisted, checked_in, completed, cancelled, no_show)

**Helpers**: `calculateDuration()`, `validateTimeRange()`

### Organization Schemas (`organization.ts`)

Organization and team management:

- `createOrganizationSchema` - Create new organization
- `updateOrganizationSchema` - Update organization profile
- `organizationSearchSchema` - Search/filter organizations
- `inviteMemberSchema` - Invite team member
- `updateMemberRoleSchema` - Change member role
- `uploadDocumentSchema` - Upload required documents
- `sendMessageSchema` - Send messages to volunteers
- `createTeamSchema` - Create volunteer team/group

**Enums**:

- `OrganizationVerificationStatus` (verified, unverified)
- `OrganizationMemberRole` (admin, coordinator)

**Constants**: `ORGANIZATION_TYPES` - Common organization types

**Helpers**: `generateSlug()`, `validateEIN()`

## Validation Rules

### Password Requirements (FR-002)

- Minimum 8 characters
- Must contain at least one letter
- Must contain at least one number

### Email Validation

- Valid email format
- Automatically converted to lowercase
- Required for all auth operations

### Security Questions (FR-003)

- Exactly 3 questions required during registration
- Minimum 2 of 3 correct answers for password reset (FR-003a)
- Questions should be unique (use `validateUniqueSecurityQuestions()`)

### Phone Numbers

- Optional in most contexts
- Minimum 10 digits (international formats supported)
- Can include `+`, `-`, spaces, and parentheses

### Dates and Times

- All opportunity dates must be in the future
- End date/time must be after start date/time
- Time format: HH:MM (24-hour format)

### Location/Address

- Address line 1, city, state, postal code required
- Country defaults to "United States"
- Geocoding triggered automatically on save

### Text Fields

- **Short** (255 chars): Names, titles, single-line inputs
- **Medium** (1000 chars): Descriptions, bio, messages
- **Long** (5000 chars): Detailed descriptions, policies

### Arrays

- Use `nonEmptyArraySchema()` for required arrays
- Skills, interests, causes: Array of strings
- Security questions: Fixed array of 3 items

## TypeScript Integration

Every schema has a corresponding TypeScript type:

```tsx
import { loginSchema, type LoginFormData } from '@/lib/validations';

// Type is automatically inferred from schema
type LoginData = z.infer<typeof loginSchema>;

// Or use the exported type
const data: LoginFormData = {
  email: 'user@example.com',
  password: 'password123',
};
```

## Custom Error Messages

All schemas include user-friendly error messages:

```tsx
const result = emailSchema.safeParse('invalid-email');

if (!result.success) {
  console.log(result.error.issues[0].message);
  // Output: "Must be a valid email address"
}
```

## Testing

Schemas can be easily unit tested:

```tsx
import { describe, it, expect } from '@jest/globals';
import { registerSchema } from '@/lib/validations';

describe('registerSchema', () => {
  it('should validate correct registration data', () => {
    const data = {
      email: 'test@example.com',
      password: 'password123',
      first_name: 'John',
      last_name: 'Doe',
      user_type: 'volunteer',
      security_questions: [
        { question: 'Q1', answer: 'A1' },
        { question: 'Q2', answer: 'A2' },
        { question: 'Q3', answer: 'A3' },
      ],
    };

    const result = registerSchema.safeParse(data);
    expect(result.success).toBe(true);
  });

  it('should reject weak passwords', () => {
    const data = {
      email: 'test@example.com',
      password: 'weak', // Too short, no numbers
      // ... other fields
    };

    const result = registerSchema.safeParse(data);
    expect(result.success).toBe(false);
  });
});
```

## OpenAPI Spec Alignment

All schemas are aligned with the OpenAPI specification at:

- `/specs/001-build-volunteersync-an/contracts/openapi.yaml`

Field names, types, and validation rules match the API contract exactly.

## Best Practices

1. **Import from index**: Always import from `@/lib/validations`, not individual files
2. **Use types**: Import both schema and type for full TypeScript support
3. **Validate early**: Validate on the frontend before API calls to reduce server load
4. **Reuse schemas**: Use common schemas for consistency across the app
5. **Update constraints**: When API requirements change, update schemas first
6. **Test validation**: Write unit tests for complex validation logic

## Related Documentation

- [OpenAPI Spec](../../../specs/001-build-volunteersync-an/contracts/openapi.yaml)
- [Data Model](../../../specs/001-build-volunteersync-an/data-model.md)
- [Zod Documentation](https://zod.dev/)
- [React Hook Form](https://react-hook-form.com/)
