# Tasks: VolunteerSync Platform

**Feature**: Build VolunteerSync - Volunteer Management Platform  
**Input**: Design documents from `/home/timmy/development/PROJECTS/VolunteerSync-Project/specs/001-build-volunteersync-an/`  
**Prerequisites**: plan.md, research.md, data-model.md, contracts/openapi.yaml, quickstart.md

## Overview

This task list implements the complete VolunteerSync platform following Test-Driven Development (TDD) principles. The platform connects nonprofit organizations with volunteers through a web application built with Next.js 15 (frontend) and Go 1.25 with Gin (backend).

**Tech Stack**:

- Frontend: Next.js 15 (App Router), TypeScript, React 19, Tailwind CSS, shadcn/ui
- Backend: Go 1.25, Gin framework, modular monolith with DDD
- Database: PostgreSQL 16 with GORM, Redis for caching/sessions
- Testing: Jest, React Testing Library, Playwright (E2E), Go testing, Testify

**Project Structure**: Web app with `backend/` and `frontend/` directories

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

---

## Phase 3.1: Project Setup & Infrastructure

### Backend Setup

- [x] T001 [P] Initialize Go module and project structure in `backend/`

  - Create `backend/go.mod` with Go 1.25
  - Create directory structure: `cmd/api/`, `internal/modules/`, `internal/middleware/`, `internal/pkg/`, `migrations/`, `tests/`
  - Initialize git repository if not exists

- [x] T002 [P] Install backend dependencies in `backend/go.mod`

  - Gin framework (github.com/gin-gonic/gin)
  - GORM (gorm.io/gorm, gorm.io/driver/postgres)
  - JWT libraries (github.com/golang-jwt/jwt/v5)
  - Argon2 (golang.org/x/crypto/argon2)
  - Redis client (github.com/redis/go-redis/v9)
  - golang-migrate (github.com/golang-migrate/migrate/v4)
  - Testing: testify (github.com/stretchr/testify), testcontainers-go

- [x] T003 [P] Configure backend linting and formatting tools
  - Create `backend/.golangci.yml` with linting rules
  - Configure gofmt, golangci-lint
  - Add pre-commit hooks

### Frontend Setup

- [x] T004 [P] Initialize Next.js 15 project in `frontend/`

  - Run `npx create-next-app@latest` with App Router, TypeScript, Tailwind
  - Create directory structure: `src/app/`, `src/components/`, `src/lib/`, `tests/`

- [x] T005 [P] Install frontend dependencies in `frontend/package.json`

  - React 19, Next.js 15
  - Tailwind CSS, shadcn/ui components
  - TanStack Query (React Query)
  - React Hook Form, Zustand
  - Leaflet (maps), Chart.js/Recharts
  - Testing: Jest, React Testing Library, Playwright

- [x] T006 [P] Configure frontend linting and formatting tools
  - Create `frontend/.eslintrc.json` with TypeScript rules
  - Create `frontend/.prettierrc` for code formatting
  - Configure strict TypeScript in `frontend/tsconfig.json`

### Infrastructure Setup

- [x] T007 [P] Create Docker configuration files

  - Create `docker/docker-compose.yml` with PostgreSQL 16, Redis, backend, frontend services
  - Create `docker/backend.Dockerfile`
  - Create `docker/frontend.Dockerfile`
  - Create `docker/docker-compose.prod.yml` for production

- [x] T008 [P] Create database migration structure

  - Create `backend/migrations/000001_initial_schema.up.sql` (empty placeholder)
  - Create `backend/migrations/000001_initial_schema.down.sql` (empty placeholder)
  - Create migration script `backend/scripts/migrate.sh`

- [x] T009 [P] Create GitHub Actions CI/CD pipelines
  - Create `.github/workflows/ci.yml` (tests, linting)
  - Create `.github/workflows/cd.yml` (build, deploy)
  - Create `.github/workflows/security.yml` (dependency scanning)

---

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Authentication Contract Tests

- [x] T010 [P] Contract test POST /api/v1/auth/register in `backend/tests/contract/auth_register_test.go`

  - Test valid registration with all required fields
  - Test missing required fields (400 error)
  - Test duplicate email (409 error)
  - Test weak password (400 error)
  - Test rate limiting (429 error after 5 attempts)

- [x] T011 [P] Contract test POST /api/v1/auth/login in `backend/tests/contract/auth_login_test.go`

  - Test valid login credentials
  - Test invalid credentials (401 error)
  - Test rate limiting (429 error after 5 attempts)
  - Test JWT token format and expiry

- [x] T012 [P] Contract test POST /api/v1/auth/refresh in `backend/tests/contract/auth_refresh_test.go`

  - Test valid refresh token
  - Test expired refresh token (401 error)
  - Test invalid refresh token (401 error)

- [x] T013 [P] Contract test POST /api/v1/auth/logout in `backend/tests/contract/auth_logout_test.go`

  - Test successful logout with valid token
  - Test logout without authentication (401 error)

- [x] T014 [P] Contract test password reset flow in `backend/tests/contract/auth_password_reset_test.go`
  - Test POST /api/v1/auth/password-reset/request
  - Test POST /api/v1/auth/password-reset/verify (2 of 3 security questions correct)
  - Test POST /api/v1/auth/password-reset/confirm

### Users Contract Tests

- [x] T015 [P] Contract test GET /api/v1/users/me in `backend/tests/contract/users_get_me_test.go`

  - Test authenticated user retrieval
  - Test without authentication (401 error)

- [x] T016 [P] Contract test PATCH /api/v1/users/me in `backend/tests/contract/users_update_me_test.go`

  - Test profile update with valid data
  - Test email change (requires reverification)
  - Test invalid data (400 error)

- [x] T017 [P] Contract test DELETE /api/v1/users/me/delete in `backend/tests/contract/users_delete_test.go`
  - Test account deletion request
  - Test data retention compliance

### Organizations Contract Tests

- [x] T018 [P] Contract test POST /api/v1/organizations in `backend/tests/contract/organizations_create_test.go`

  - Test organization creation with required fields
  - Test auto-verification on creation
  - Test slug generation from name
  - Test missing required fields (400 error)

- [x] T019 [P] Contract test GET /api/v1/organizations/{id} in `backend/tests/contract/organizations_get_test.go`

  - Test public organization retrieval
  - Test organization not found (404 error)

- [x] T020 [P] Contract test PATCH /api/v1/organizations/{id} in `backend/tests/contract/organizations_update_test.go`

  - Test organization update by admin
  - Test unauthorized update (403 error)

- [x] T021 [P] Contract test GET /api/v1/organizations in `backend/tests/contract/organizations_list_test.go`
  - Test list with pagination
  - Test filtering by cause, location
  - Test search by name

### Volunteers Contract Tests

- [x] T022 [P] Contract test PATCH /api/v1/volunteers/me in `backend/tests/contract/volunteers_update_test.go`

  - Test volunteer profile update
  - Test skill and interest associations
  - Test availability settings

- [x] T023 [P] Contract test GET /api/v1/volunteers/me/dashboard in `backend/tests/contract/volunteers_dashboard_test.go`
  - Test dashboard with impact metrics
  - Test empty state for new volunteers

### Opportunities Contract Tests

- [x] T024 [P] Contract test POST /api/v1/opportunities in `backend/tests/contract/opportunities_create_test.go`

  - Test opportunity creation by coordinator
  - Test immediate publish vs draft
  - Test geocoding on address save
  - Test recurring opportunity creation

- [x] T025 [P] Contract test GET /api/v1/opportunities in `backend/tests/contract/opportunities_list_test.go`

  - Test search with location radius
  - Test filtering by date range, cause, skills
  - Test performance (<2 seconds for up to 100 results)

- [x] T026 [P] Contract test GET /api/v1/opportunities/{id} in `backend/tests/contract/opportunities_get_test.go`

  - Test public opportunity detail retrieval
  - Test capacity display

- [x] T027 [P] Contract test PATCH /api/v1/opportunities/{id} in `backend/tests/contract/opportunities_update_test.go`
  - Test update by coordinator
  - Test cannot edit past events

### Registrations Contract Tests

- [x] T028 [P] Contract test POST /api/v1/registrations in `backend/tests/contract/registrations_create_test.go`

  - Test immediate registration
  - Test waitlist when at capacity
  - Test duplicate registration prevention
  - Test overlapping event warning

- [x] T029 [P] Contract test PATCH /api/v1/registrations/{id}/cancel in `backend/tests/contract/registrations_cancel_test.go`

  - Test cancellation
  - Test late cancellation warning (within 24 hours)

- [x] T030 [P] Contract test PATCH /api/v1/registrations/{id}/check-in in `backend/tests/contract/registrations_checkin_test.go`

  - Test volunteer check-in on event day

- [x] T031 [P] Contract test GET /api/v1/registrations/{id}/calendar.ics in `backend/tests/contract/registrations_calendar_test.go`
  - Test .ics file download

### Hours Tracking Contract Tests

- [x] T032 [P] Contract test POST /api/v1/hours/log in `backend/tests/contract/hours_log_test.go`

  - Test hours logging by coordinator
  - Test pending status on creation
  - Test notification to volunteer

- [x] T033 [P] Contract test POST /api/v1/hours/{id}/verify in `backend/tests/contract/hours_verify_test.go`

  - Test volunteer verification
  - Test total hours increment

- [x] T034 [P] Contract test POST /api/v1/hours/{id}/dispute in `backend/tests/contract/hours_dispute_test.go`
  - Test hours dispute by volunteer
  - Test coordinator notification

### Communications Contract Tests

- [x] T035 [P] Contract test POST /api/v1/messages in `backend/tests/contract/messages_create_test.go`

  - Test direct message creation
  - Test broadcast message to event volunteers

- [x] T036 [P] Contract test GET /api/v1/notifications in `backend/tests/contract/notifications_list_test.go`

  - Test notification list with pagination
  - Test unread count

- [x] T037 [P] Contract test PATCH /api/v1/notifications/{id}/read in `backend/tests/contract/notifications_read_test.go`
  - Test mark notification as read

### Integration Tests from Quickstart

- [x] T038 [P] Integration test Story 1 (Org Onboarding) in `backend/tests/integration/story1_org_onboarding_test.go`

  - Scenario 1.1: Organization admin registration
  - Scenario 1.2: Create organization profile (auto-verify, geocode)
  - Scenario 1.3: Create and publish volunteer opportunity

- [x] T039 [P] Integration test Story 2 (Volunteer Discovery) in `backend/tests/integration/story2_volunteer_discovery_test.go`

  - Scenario 2.1: Volunteer registration
  - Scenario 2.2: Complete volunteer profile (geocode, skills, interests)
  - Scenario 2.3: Search for opportunities (<2s performance)
  - Scenario 2.4: Register for opportunity (notification, capacity update)

- [x] T040 [P] Integration test Story 3 (Event Operations) in `backend/tests/integration/story3_event_operations_test.go`

  - Scenario 3.1: Event reminder notifications (24h, 2h)
  - Scenario 3.2: Volunteer check-in
  - Scenario 3.3: Log volunteer hours (pending status)
  - Scenario 3.4: Volunteer confirms hours (verified status)
  - Scenario 3.5: Auto-verify hours after 7 days
  - Scenario 3.6: Volunteer reviews event

- [x] T041 [P] Integration test Story 4 (Impact Tracking) in `backend/tests/integration/story4_impact_tracking_test.go`

  - Scenario 4.1: View personal dashboard with metrics
  - Scenario 4.2: Achievement badge earned (First Event badge)
  - Scenario 4.3: Download impact report PDF

- [x] T042 [P] Integration test Story 5 (Edge Cases) in `backend/tests/integration/story5_edge_cases_test.go`
  - Scenario 5.1: Event at capacity → waitlist
  - Scenario 5.2: Late cancellation warning
  - Scenario 5.3: Overlapping events warning
  - Scenario 5.4: Hours dispute workflow

### Frontend E2E Tests

- [x] T043 [P] E2E test authentication flow in `frontend/tests/e2e/auth.spec.ts`

  - Test registration, login, logout
  - Test password reset with security questions

- [x] T044 [P] E2E test volunteer journey in `frontend/tests/e2e/volunteer_journey.spec.ts`

  - Test profile creation, opportunity search, registration, hours verification

- [x] T045 [P] E2E test organization journey in `frontend/tests/e2e/org_journey.spec.ts`
  - Test org creation, opportunity posting, hours logging

---

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Database Schema & Migrations

- [x] T046 Create initial database migration in `backend/migrations/000001_initial_schema.up.sql`

  - Users table with security questions
  - Organizations table with geocoding
  - Organization_Member junction table
  - Volunteer_Profile table with availability
  - Opportunity table with recurrence support
  - Registration table with hours tracking
  - Hours_Log audit table
  - Message and Message_Recipient tables
  - Notification table
  - Skill, Cause_Category lookup tables
  - Achievement and Volunteer_Achievement tables
  - Document, Team, Team_Member tables
  - All junction tables for N:M relationships
  - Indexes on foreign keys and frequently queried fields

- [x] T047 Create migration down script in `backend/migrations/000001_initial_schema.down.sql`
  - Drop all tables in reverse dependency order

### Shared Infrastructure

- [x] T048 [P] Database connection utilities in `backend/internal/pkg/database/connection.go`

  - PostgreSQL connection with GORM
  - Connection pooling configuration
  - Health check endpoint

- [x] T049 [P] Redis cache utilities in `backend/internal/pkg/cache/redis.go`

  - Redis client initialization
  - Cache helper functions (Get, Set, Delete)
  - Session storage

- [x] T050 [P] JWT utilities in `backend/internal/pkg/jwt/jwt.go`

  - Generate access token (15 minutes)
  - Generate refresh token (7 days)
  - Validate and parse tokens
  - Token rotation logic

- [x] T051 [P] Error handling utilities in `backend/internal/pkg/errors/errors.go`

  - Custom error types
  - Error response formatting
  - Logging integration

- [x] T052 [P] Structured logging setup in `backend/internal/pkg/logger/logger.go`
  - Initialize zerolog or zap
  - Contextual logging with request IDs
  - No PII in logs

### Authentication Module

- [x] T053 [P] User model in `backend/internal/modules/auth/models/user.go`

  - User struct with GORM tags
  - Password hashing with Argon2
  - Security questions hashing

- [x] T054 [P] Auth repository in `backend/internal/modules/auth/repositories/auth_repository.go`

  - CreateUser with transaction
  - FindUserByEmail
  - UpdatePassword
  - GetSecurityQuestions

- [x] T055 Auth service in `backend/internal/modules/auth/services/auth_service.go`

  - Register user (validate, hash password, create user, generate tokens)
  - Login user (validate credentials, generate tokens)
  - RefreshToken (rotate tokens)
  - Logout (invalidate refresh token)
  - Password reset flow (verify security questions, reset password)
  - Account status checks (inactive after 12 months)

- [x] T056 Auth handlers in `backend/internal/modules/auth/handlers/auth_handlers.go`
  - POST /auth/register handler
  - POST /auth/login handler with rate limiting
  - POST /auth/refresh handler
  - POST /auth/logout handler
  - POST /auth/password-reset/request handler
  - POST /auth/password-reset/verify handler (2 of 3 questions)
  - POST /auth/password-reset/confirm handler

### Users Module

- [x] T057 [P] User service in `backend/internal/modules/users/services/user_service.go`

  - GetCurrentUser
  - UpdateUserProfile
  - DeleteUserAccount (soft delete with data retention)

- [x] T058 User handlers in `backend/internal/modules/users/handlers/user_handlers.go`
  - GET /users/me handler
  - PATCH /users/me handler
  - DELETE /users/me/delete handler

### Organizations Module

- [x] T059 [P] Organization model in `backend/internal/modules/organizations/models/organization.go`

  - Organization struct with GORM tags
  - Slug generation from name

- [x] T060 [P] Organization repository in `backend/internal/modules/organizations/repositories/org_repository.go`

  - CreateOrganization
  - FindOrganizationById
  - FindOrganizationBySlug
  - ListOrganizations with filters
  - UpdateOrganization
  - AddMember, RemoveMember

- [x] T061 Organization service in `backend/internal/modules/organizations/services/org_service.go`

  - CreateOrganization (auto-verify, geocode address, create member record)
  - GetOrganization (with member count)
  - UpdateOrganization (admin only)
  - ListOrganizations (search, filter by cause/location)
  - InviteMember (FR-014)
  - RemoveMember

- [x] T062 Organization handlers in `backend/internal/modules/organizations/handlers/org_handlers.go`
  - POST /organizations handler
  - GET /organizations/{id} handler
  - PATCH /organizations/{id} handler
  - GET /organizations handler with pagination

### Volunteers Module

- [ ] T063 [P] Volunteer profile model in `backend/internal/modules/volunteers/models/volunteer_profile.go`

  - VolunteerProfile struct with availability fields
  - Privacy settings

- [ ] T064 [P] Volunteer repository in `backend/internal/modules/volunteers/repositories/volunteer_repository.go`

  - CreateVolunteerProfile (on user registration)
  - FindVolunteerProfileByUserId
  - UpdateVolunteerProfile
  - AddSkills, RemoveSkills
  - AddInterests, RemoveInterests

- [ ] T065 Volunteer service in `backend/internal/modules/volunteers/services/volunteer_service.go`

  - GetVolunteerProfile
  - UpdateVolunteerProfile (geocode location, manage skills/interests)
  - GetDashboard (total hours, events, organizations)
  - GetAnalytics (hours over time chart)
  - GenerateImpactReport (PDF)

- [ ] T066 Volunteer handlers in `backend/internal/modules/volunteers/handlers/volunteer_handlers.go`
  - PATCH /volunteers/me handler
  - GET /volunteers/me/dashboard handler
  - GET /volunteers/me/analytics handler
  - GET /volunteers/me/report handler (PDF download)

### Opportunities Module

- [ ] T067 [P] Opportunity model in `backend/internal/modules/opportunities/models/opportunity.go`

  - Opportunity struct with recurrence fields
  - Status enum (draft, published, cancelled, completed)

- [ ] T068 [P] Opportunity repository in `backend/internal/modules/opportunities/repositories/opp_repository.go`

  - CreateOpportunity
  - FindOpportunityById
  - ListOpportunities with complex filters
  - UpdateOpportunity
  - IncrementRegistrations, DecrementRegistrations
  - CreateRecurringInstances

- [ ] T069 Opportunity service in `backend/internal/modules/opportunities/services/opp_service.go`

  - CreateOpportunity (geocode, publish immediately or draft)
  - GetOpportunity (with registration count)
  - UpdateOpportunity (prevent editing past events)
  - ListOpportunities (search with location radius, date range, cause, skills)
  - CancelOpportunity (notify registered volunteers)
  - CompleteOpportunity (auto-complete 7 days after end)
  - CreateRecurringOpportunities (generate child instances)

- [ ] T070 Opportunity handlers in `backend/internal/modules/opportunities/handlers/opp_handlers.go`
  - POST /opportunities handler
  - GET /opportunities/{id} handler
  - PATCH /opportunities/{id} handler
  - GET /opportunities handler with search/filters (<2s performance)

### Registrations Module

- [ ] T071 [P] Registration model in `backend/internal/modules/registrations/models/registration.go`

  - Registration struct with status enum
  - Hours tracking fields

- [ ] T072 [P] Registration repository in `backend/internal/modules/registrations/repositories/reg_repository.go`

  - CreateRegistration
  - FindRegistrationById
  - FindRegistrationsByVolunteer
  - FindRegistrationsByOpportunity
  - UpdateRegistrationStatus
  - CheckIn
  - CancelRegistration

- [ ] T073 Registration service in `backend/internal/modules/registrations/services/reg_service.go`

  - RegisterVolunteer (check capacity, check duplicates, check overlaps, add to waitlist if full)
  - CancelRegistration (late cancellation warning, notify org)
  - CheckInVolunteer
  - GetRegistration
  - ListVolunteerRegistrations
  - GenerateCalendarFile (.ics download)

- [ ] T074 Registration handlers in `backend/internal/modules/registrations/handlers/reg_handlers.go`
  - POST /registrations handler
  - PATCH /registrations/{id}/cancel handler
  - PATCH /registrations/{id}/check-in handler
  - GET /registrations/{id}/calendar.ics handler

### Hours Tracking Module

- [ ] T075 [P] Hours log model in `backend/internal/modules/hours/models/hours_log.go`

  - HoursLog struct (immutable audit trail)
  - Status enum (pending, verified, disputed)

- [ ] T076 [P] Hours repository in `backend/internal/modules/hours/repositories/hours_repository.go`

  - CreateHoursLog (immutable)
  - FindHoursLogById
  - UpdateHoursStatus (only status, not hours amount)
  - FindPendingHoursOlderThan7Days

- [ ] T077 Hours service in `backend/internal/modules/hours/services/hours_service.go`

  - LogHours (create log, update registration, notify volunteer)
  - VerifyHours (update status to verified, increment volunteer total_hours)
  - DisputeHours (update status to disputed, notify coordinator)
  - AutoVerifyOldHours (cron job, auto-verify after 7 days)
  - ResolveDispute

- [ ] T078 Hours handlers in `backend/internal/modules/hours/handlers/hours_handlers.go`
  - POST /hours/log handler
  - POST /hours/{id}/verify handler
  - POST /hours/{id}/dispute handler

### Communications Module

- [ ] T079 [P] Message and notification models in `backend/internal/modules/communications/models/`

  - Message struct (direct, broadcast types)
  - MessageRecipient struct
  - Notification struct (type enum)

- [ ] T080 [P] Communications repository in `backend/internal/modules/communications/repositories/comm_repository.go`

  - CreateMessage
  - CreateNotification
  - FindNotificationsByUser
  - MarkNotificationAsRead
  - GetUnreadCount

- [ ] T081 Communications service in `backend/internal/modules/communications/services/comm_service.go`

  - SendDirectMessage
  - SendBroadcastMessage (to all event registrants)
  - CreateNotification (various types: registration_confirmed, hours_logged, achievement_earned, etc.)
  - SendEventReminders (cron job, 24h and 2h before event)
  - GetUserNotifications
  - MarkNotificationRead

- [ ] T082 Communications handlers in `backend/internal/modules/communications/handlers/comm_handlers.go`
  - POST /messages handler
  - GET /notifications handler
  - PATCH /notifications/{id}/read handler

### Analytics Module

- [ ] T083 [P] Analytics service in `backend/internal/modules/analytics/services/analytics_service.go`

  - GetVolunteerAnalytics (hours over time, events by cause)
  - GetOrganizationAnalytics (volunteers recruited, hours contributed, retention rate)
  - GetPlatformAnalytics (total volunteers, orgs, hours, growth trends)
  - GenerateReport (PDF generation)

- [ ] T084 Analytics handlers in `backend/internal/modules/analytics/handlers/analytics_handlers.go`
  - GET /analytics/volunteer/{id} handler
  - GET /analytics/organization/{id} handler
  - GET /analytics/platform handler (admin only)

### Achievements Module

- [ ] T085 [P] Achievement model in `backend/internal/modules/achievements/models/achievement.go`

  - Achievement struct (badge definition)
  - VolunteerAchievement junction

- [ ] T086 [P] Achievement repository in `backend/internal/modules/achievements/repositories/achievement_repository.go`

  - CreateAchievement
  - AwardAchievement
  - FindVolunteerAchievements

- [ ] T087 Achievement service in `backend/internal/modules/achievements/services/achievement_service.go`

  - CheckAndAwardAchievements (cron job, check criteria like "First Event", "10 Hours", "5 Organizations")
  - GetVolunteerAchievements

- [ ] T088 Achievement handlers in `backend/internal/modules/achievements/handlers/achievement_handlers.go`
  - GET /achievements handler (list all available)
  - GET /volunteers/{id}/achievements handler

### Middleware

- [ ] T089 [P] Authentication middleware in `backend/internal/middleware/auth.go`

  - Validate JWT token
  - Extract user from token
  - Add user to request context

- [ ] T090 [P] RBAC middleware in `backend/internal/middleware/rbac.go`

  - Check user roles (Super Admin, Org Admin, Coordinator, Volunteer)
  - Verify organization membership for org-specific actions

- [ ] T091 [P] Rate limiting middleware in `backend/internal/middleware/rate_limit.go`

  - 100 requests per minute per user (general)
  - 5 login attempts per 15 minutes per IP

- [ ] T092 [P] CORS middleware in `backend/internal/middleware/cors.go`

  - Configure allowed origins
  - Configure allowed methods and headers

- [ ] T093 [P] Logging middleware in `backend/internal/middleware/logging.go`

  - Log all requests (method, path, status, duration)
  - Add request ID to context
  - No PII in logs

- [ ] T094 [P] Recovery middleware in `backend/internal/middleware/recovery.go`
  - Catch panics
  - Return 500 error
  - Log stack trace

### Backend Main Application

- [ ] T095 Create main application in `backend/cmd/api/main.go`

  - Initialize database connection
  - Initialize Redis connection
  - Initialize logger
  - Setup middleware chain (logging → recovery → CORS → rate limiting → auth → RBAC)
  - Register all module routes
  - Start HTTP server with graceful shutdown
  - Health check endpoint

- [ ] T096 Create configuration management in `backend/internal/config/config.go`
  - Load from environment variables
  - Database connection string
  - Redis connection string
  - JWT secret
  - Port configuration

### Frontend - Shared Infrastructure

- [ ] T097 [P] API client setup in `frontend/src/lib/api/client.ts`

  - Axios or fetch wrapper
  - JWT token handling (access + refresh)
  - Token refresh on 401
  - Error handling

- [ ] T098 [P] React Query setup in `frontend/src/lib/api/query-client.ts`

  - Configure TanStack Query
  - Default cache times
  - Error handling

- [ ] T099 [P] Zustand stores in `frontend/src/store/`

  - Auth store (user, tokens, login, logout)
  - Notification store (unread count, notifications list)

- [ ] T100 [P] shadcn/ui component installation

  - Install and configure shadcn/ui
  - Add Button, Input, Card, Dialog, Form components to `frontend/src/components/ui/`

- [ ] T101 [P] Tailwind configuration in `frontend/tailwind.config.ts`

  - Custom theme (colors, spacing, typography)
  - Dark mode support
  - Responsive breakpoints

- [ ] T102 [P] Form validation schemas in `frontend/src/lib/validations/`
  - Zod schemas for registration, login, profile, opportunity forms

### Frontend - Authentication Pages

- [ ] T103 [P] Registration page in `frontend/src/app/(auth)/register/page.tsx`

  - User type selection (volunteer, organization_admin)
  - Form with email, password, name, security questions
  - Client-side validation with React Hook Form + Zod
  - API call to POST /auth/register

- [ ] T104 [P] Login page in `frontend/src/app/(auth)/login/page.tsx`

  - Email and password form
  - Remember me checkbox
  - Link to password reset
  - API call to POST /auth/login

- [ ] T105 [P] Password reset pages in `frontend/src/app/(auth)/reset-password/`
  - Request page (enter email)
  - Verify page (answer security questions)
  - Confirm page (set new password)

### Frontend - Dashboard Layouts

- [ ] T106 Create root layout in `frontend/src/app/layout.tsx`

  - HTML structure
  - Global styles
  - React Query provider
  - Auth context provider

- [ ] T107 [P] Volunteer dashboard layout in `frontend/src/app/(dashboard)/volunteer/layout.tsx`

  - Navigation sidebar (Dashboard, Find Opportunities, My Events, Impact, Profile)
  - Header with notifications bell
  - User menu

- [ ] T108 [P] Organization dashboard layout in `frontend/src/app/(dashboard)/organization/layout.tsx`
  - Navigation sidebar (Dashboard, Opportunities, Team, Analytics, Settings)
  - Header with notifications
  - Organization switcher (if member of multiple)

### Frontend - Volunteer Features

- [ ] T109 [P] Volunteer dashboard page in `frontend/src/app/(dashboard)/volunteer/page.tsx`

  - Impact metrics cards (total hours, events, organizations)
  - Hours over time chart (Chart.js)
  - Recent events list
  - Upcoming events

- [ ] T110 [P] Volunteer profile page in `frontend/src/app/(dashboard)/volunteer/profile/page.tsx`

  - Profile form (bio, location, photo, availability, skills, interests)
  - Privacy settings
  - Emergency contact
  - Save/update with API call

- [ ] T111 Opportunity search page in `frontend/src/app/(dashboard)/volunteer/opportunities/page.tsx`

  - Search filters (location, radius, cause, date range, skills)
  - Map view with Leaflet
  - List view with pagination
  - Performance: <2s search results
  - Click to view opportunity detail

- [ ] T112 Opportunity detail page in `frontend/src/app/opportunities/[id]/page.tsx`

  - Opportunity information (title, description, date, location, capacity)
  - Organization information
  - Map with location
  - "Register" button
  - Reviews from past volunteers

- [ ] T113 [P] My events page in `frontend/src/app/(dashboard)/volunteer/events/page.tsx`

  - Tabs: Upcoming, Past, Cancelled
  - Event cards with status
  - Check-in button (on event day)
  - Hours verification
  - Write review (after completion)

- [ ] T114 [P] Impact page in `frontend/src/app/(dashboard)/volunteer/impact/page.tsx`
  - Total impact metrics
  - Achievement badges display
  - Hours breakdown by cause
  - Download impact report button

### Frontend - Organization Features

- [ ] T115 [P] Organization dashboard page in `frontend/src/app/(dashboard)/organization/page.tsx`

  - Metrics: volunteers recruited, hours contributed, events hosted
  - Upcoming events
  - Recent registrations
  - Analytics charts

- [ ] T116 Create organization form in `frontend/src/app/(dashboard)/organization/new/page.tsx`

  - Organization profile form (name, mission, description, address, logo)
  - Cause selection
  - Submit to POST /organizations

- [ ] T117 Create opportunity form in `frontend/src/app/(dashboard)/organization/opportunities/new/page.tsx`

  - Opportunity form (title, description, date/time, location, capacity, min age)
  - Skill requirements
  - Required documents
  - Recurring event options
  - Publish immediately or save as draft

- [ ] T118 Opportunity management page in `frontend/src/app/(dashboard)/organization/opportunities/page.tsx`

  - List all opportunities (tabs: Published, Draft, Completed, Cancelled)
  - Edit, cancel, complete actions
  - View registrations

- [ ] T119 Event roster page in `frontend/src/app/(dashboard)/organization/opportunities/[id]/roster/page.tsx`

  - List of registered volunteers
  - Check-in volunteers
  - Send broadcast message to all
  - Log hours after event

- [ ] T120 Hours logging page in `frontend/src/app/(dashboard)/organization/opportunities/[id]/hours/page.tsx`

  - Table of checked-in volunteers
  - Input hours for each volunteer
  - Add coordinator notes
  - Submit hours (creates pending records)

- [ ] T121 [P] Team management page in `frontend/src/app/(dashboard)/organization/team/page.tsx`

  - List team members with roles
  - Invite new members (send email invitation)
  - Remove members

- [ ] T122 [P] Organization analytics page in `frontend/src/app/(dashboard)/organization/analytics/page.tsx`
  - Volunteers by cause chart
  - Hours contributed over time
  - Volunteer retention rate
  - Event completion rate

### Frontend - Notifications

- [ ] T123 [P] Notifications component in `frontend/src/components/features/notifications/NotificationCenter.tsx`

  - Dropdown with unread notifications
  - Notification types: registration_confirmed, hours_logged, achievement_earned, event_reminder, etc.
  - Mark as read action
  - Click to navigate to relevant page

- [ ] T124 [P] Notification badge in `frontend/src/components/features/notifications/NotificationBadge.tsx`
  - Show unread count
  - Update in real-time with React Query polling

### Frontend - Shared Components

- [ ] T125 [P] Opportunity card component in `frontend/src/components/features/opportunities/OpportunityCard.tsx`

  - Display opportunity summary
  - Show organization info
  - Capacity indicator
  - Cause badges

- [ ] T126 [P] Map component in `frontend/src/components/shared/Map.tsx`

  - Leaflet integration
  - Display opportunity markers
  - Click marker to view detail

- [ ] T127 [P] Calendar integration in `frontend/src/lib/utils/calendar.ts`
  - Generate .ics file for event
  - Download or add to Google Calendar

---

## Phase 3.4: Integration & Cron Jobs

### Background Jobs

- [ ] T128 [P] Event reminder cron job in `backend/internal/jobs/event_reminders.go`

  - Run every hour
  - Find events 24 hours away, send reminder notifications
  - Find events 2 hours away, send reminder notifications

- [ ] T129 [P] Auto-verify hours cron job in `backend/internal/jobs/auto_verify_hours.go`

  - Run daily
  - Find pending hours older than 7 days
  - Auto-verify and update volunteer total_hours

- [ ] T130 [P] Auto-complete events cron job in `backend/internal/jobs/auto_complete_events.go`

  - Run daily
  - Find events 7 days past end_date
  - Mark as completed

- [ ] T131 [P] Check achievements cron job in `backend/internal/jobs/check_achievements.go`

  - Run daily
  - Check all volunteers for achievement criteria
  - Award new badges and send notifications

- [ ] T132 [P] Inactive account checker in `backend/internal/jobs/inactive_accounts.go`
  - Run weekly
  - Mark accounts inactive after 12 months of no login

### Geocoding Integration

- [ ] T133 [P] Geocoding service in `backend/internal/pkg/geocoding/geocoding.go`
  - Use OpenStreetMap Nominatim API (free)
  - Convert address to lat/lng
  - Cache results in Redis
  - Fallback handling

### Email Templates (V1: In-platform notifications only)

- [ ] T134 [P] Notification templates in `backend/internal/modules/communications/templates/`
  - Template for each notification type
  - Placeholder replacement
  - Note: V1 uses in-platform notifications only, no email/SMS

---

## Phase 3.5: Polish & Testing

### Unit Tests

- [ ] T135 [P] Unit tests for auth service in `backend/internal/modules/auth/services/auth_service_test.go`

  - Test password hashing
  - Test security question validation (2 of 3)
  - Test token generation

- [ ] T136 [P] Unit tests for hours service in `backend/internal/modules/hours/services/hours_service_test.go`

  - Test auto-verify logic
  - Test dispute handling

- [ ] T137 [P] Unit tests for registration service in `backend/internal/modules/registrations/services/reg_service_test.go`

  - Test capacity check
  - Test duplicate prevention
  - Test overlapping event detection

- [ ] T138 [P] Frontend unit tests for form validation in `frontend/tests/unit/validations.test.ts`

  - Test Zod schemas
  - Test error messages

- [ ] T139 [P] Frontend component tests in `frontend/tests/components/`
  - Test OpportunityCard component
  - Test NotificationCenter component
  - Test authentication forms

### Performance Optimization

- [ ] T140 [P] Backend performance optimization

  - Add database indexes (verify migration)
  - Optimize N+1 queries with GORM Preload
  - Add Redis caching for frequently accessed data (organizations, opportunities)
  - Profile API endpoints (<500ms p95)

- [ ] T141 [P] Frontend performance optimization
  - Lazy load routes with Next.js dynamic imports
  - Optimize images with Next.js Image component
  - Code splitting
  - Bundle size analysis
  - Performance testing (<2s page load on 3G)

### Security Hardening

- [ ] T142 [P] Security audit in `backend/`

  - Verify rate limiting works
  - Test JWT token expiry and rotation
  - Test password strength requirements
  - Verify no PII in logs
  - Test CORS configuration
  - XSS prevention (sanitize message content)
  - SQL injection prevention (parameterized queries)

- [ ] T143 [P] Accessibility audit in `frontend/`
  - WCAG 2.1 Level AA compliance check
  - Keyboard navigation testing
  - Screen reader testing
  - Color contrast validation

### Documentation

- [ ] T144 [P] API documentation in `backend/docs/api.md`

  - Document all endpoints from OpenAPI spec
  - Authentication requirements
  - Rate limiting details
  - Example requests/responses

- [ ] T145 [P] Deployment documentation in `docs/deployment.md`

  - Production Docker Compose configuration
  - Environment variables
  - Database migration process
  - Backup and restore procedures

- [ ] T146 [P] User guides in `docs/user-guide/`
  - Volunteer user guide
  - Organization admin guide
  - Screenshots

### Final Testing

- [ ] T147 Run all contract tests and verify pass

  - Execute `go test ./tests/contract/... -v`
  - All tests must pass (T010-T037)

- [ ] T148 Run all integration tests and verify pass

  - Execute `go test ./tests/integration/... -v`
  - All quickstart scenarios must pass (T038-T042)

- [ ] T149 Run all frontend E2E tests and verify pass

  - Execute `npx playwright test`
  - All user journeys must pass (T043-T045)

- [ ] T150 Performance validation

  - Verify page load <2s on 3G
  - Verify search results <2s
  - Verify API p95 <500ms
  - Verify dashboard load <3s

- [ ] T151 Manual testing from quickstart.md
  - Execute all 5 test suites manually
  - Verify all scenarios work end-to-end
  - Document any issues

---

## Dependencies

**Critical Path**:

1. Setup (T001-T009) must complete before anything else
2. Tests (T010-T045) must be written and failing before implementation
3. Database migration (T046-T047) blocks all implementation
4. Shared infrastructure (T048-T052) blocks module implementation
5. Auth module (T053-T056) blocks all protected endpoints
6. Middleware (T089-T094) required before main application (T095)
7. Backend API (T053-T095) must be functional before frontend pages
8. Cron jobs (T128-T132) require service implementations
9. All implementation before polish (T135-T151)

**Module Dependencies**:

- Auth module → Users, Organizations, Volunteers (all require authentication)
- Organizations → Opportunities (opportunities belong to organizations)
- Volunteers → Registrations (registrations require volunteer profiles)
- Opportunities → Registrations (registrations link to opportunities)
- Registrations → Hours (hours tracked per registration)
- All modules → Communications (notifications sent from various modules)

**Frontend Dependencies**:

- API client (T097-T099) blocks all pages
- Auth pages (T103-T105) before dashboard pages
- Layouts (T106-T108) before feature pages
- Shared components (T125-T127) can be used across features

---

## Parallel Execution Examples

### Phase 3.1 - Parallel Setup

```bash
# All setup tasks can run in parallel:
Task: "Initialize Go module in backend/" # T001
Task: "Install backend dependencies in backend/go.mod" # T002
Task: "Configure backend linting in backend/.golangci.yml" # T003
Task: "Initialize Next.js project in frontend/" # T004
Task: "Install frontend dependencies in frontend/package.json" # T005
Task: "Configure frontend linting in frontend/.eslintrc.json" # T006
Task: "Create Docker configuration in docker/docker-compose.yml" # T007
Task: "Create migration structure in backend/migrations/" # T008
Task: "Create CI/CD pipelines in .github/workflows/" # T009
```

### Phase 3.2 - Parallel Contract Tests

```bash
# Authentication contract tests (different files):
Task: "Contract test POST /auth/register" # T010
Task: "Contract test POST /auth/login" # T011
Task: "Contract test POST /auth/refresh" # T012
Task: "Contract test POST /auth/logout" # T013
Task: "Contract test password reset flow" # T014

# Users contract tests:
Task: "Contract test GET /users/me" # T015
Task: "Contract test PATCH /users/me" # T016
Task: "Contract test DELETE /users/me" # T017

# Organizations contract tests:
Task: "Contract test POST /organizations" # T018
Task: "Contract test GET /organizations/{id}" # T019
Task: "Contract test PATCH /organizations/{id}" # T020
Task: "Contract test GET /organizations" # T021

# All integration tests can run in parallel:
Task: "Integration test Story 1 (Org Onboarding)" # T038
Task: "Integration test Story 2 (Volunteer Discovery)" # T039
Task: "Integration test Story 3 (Event Operations)" # T040
Task: "Integration test Story 4 (Impact Tracking)" # T041
Task: "Integration test Story 5 (Edge Cases)" # T042

# Frontend E2E tests:
Task: "E2E test authentication flow" # T043
Task: "E2E test volunteer journey" # T044
Task: "E2E test organization journey" # T045
```

### Phase 3.3 - Parallel Infrastructure

```bash
# Shared infrastructure (different files):
Task: "Database connection utilities" # T048
Task: "Redis cache utilities" # T049
Task: "JWT utilities" # T050
Task: "Error handling utilities" # T051
Task: "Structured logging setup" # T052

# Auth module models and repos (different files):
Task: "User model in auth/models/" # T053
Task: "Auth repository in auth/repositories/" # T054

# Middleware (different files):
Task: "Authentication middleware" # T089
Task: "RBAC middleware" # T090
Task: "Rate limiting middleware" # T091
Task: "CORS middleware" # T092
Task: "Logging middleware" # T093
Task: "Recovery middleware" # T094
```

### Phase 3.3 - Parallel Frontend Components

```bash
# Frontend infrastructure (different files):
Task: "API client setup" # T097
Task: "React Query setup" # T098
Task: "Zustand stores" # T099
Task: "shadcn/ui installation" # T100
Task: "Tailwind configuration" # T101
Task: "Form validation schemas" # T102

# Auth pages (different files):
Task: "Registration page" # T103
Task: "Login page" # T104
Task: "Password reset pages" # T105

# Dashboard layouts (different files):
Task: "Volunteer dashboard layout" # T107
Task: "Organization dashboard layout" # T108
```

### Phase 3.5 - Parallel Polish

```bash
# Unit tests (different files):
Task: "Unit tests for auth service" # T135
Task: "Unit tests for hours service" # T136
Task: "Unit tests for registration service" # T137
Task: "Frontend unit tests for validation" # T138
Task: "Frontend component tests" # T139

# Optimization and audits:
Task: "Backend performance optimization" # T140
Task: "Frontend performance optimization" # T141
Task: "Security audit" # T142
Task: "Accessibility audit" # T143

# Documentation (different files):
Task: "API documentation" # T144
Task: "Deployment documentation" # T145
Task: "User guides" # T146
```

---

## Validation Checklist

### Completeness

- [x] All entities from data-model.md have model tasks
- [x] All endpoints from openapi.yaml have contract tests and handler tasks
- [x] All quickstart scenarios have integration test tasks
- [x] All modules (auth, users, orgs, volunteers, opportunities, registrations, hours, communications, analytics, achievements) have implementation tasks
- [x] Middleware, cron jobs, and infrastructure tasks included

### Test-Driven Development

- [x] All contract tests come before implementation (Phase 3.2 before 3.3)
- [x] Integration tests written before core implementation
- [x] E2E tests included
- [x] Unit tests in polish phase

### Dependencies

- [x] Setup tasks first
- [x] Tests before implementation
- [x] Shared infrastructure before modules
- [x] Auth before protected endpoints
- [x] Backend before frontend
- [x] Implementation before polish

### Parallel Execution

- [x] Tasks marked [P] are in different files
- [x] No [P] tasks modify the same file
- [x] Dependencies respected (no parallel tasks with dependencies)

### File Paths

- [x] All tasks specify exact file paths
- [x] Paths follow project structure (backend/, frontend/)
- [x] Module-based organization clear

---

## Estimated Timeline

**Total Tasks**: 151  
**Estimated Effort**: 400-500 developer hours

**Phase Breakdown**:

- Phase 3.1 (Setup): 1-2 days (9 tasks)
- Phase 3.2 (Tests): 5-7 days (36 tasks) - can parallelize significantly
- Phase 3.3 (Implementation): 20-30 days (100 tasks) - largest phase
- Phase 3.4 (Integration): 2-3 days (6 tasks)
- Phase 3.5 (Polish): 3-5 days (6 tasks)

**Critical Path**: ~35-45 days with optimal parallelization

---

## Next Steps

1. Start with Phase 3.1 setup tasks (T001-T009)
2. Write failing tests in Phase 3.2 (T010-T045)
3. Implement core functionality module by module (T046-T095)
4. Build frontend in parallel with backend API maturity (T097-T127)
5. Add background jobs and integrations (T128-T134)
6. Polish and finalize (T135-T151)

**Ready for execution!** Each task is specific, actionable, and follows TDD principles. Use parallel execution where marked [P] to maximize development velocity.
