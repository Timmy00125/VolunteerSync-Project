# VolunteerSync Project - TODOs and Status

**Last Updated**: October 9, 2025  
**Project**: VolunteerSync - Volunteer Management Platform  
**Branch**: 001-build-volunteersync-an

---

## Summary

This document tracks all TODO items, both from the task list and inline code comments. Items are categorized by priority and status.

**Overall Progress**:

- **Completed**: 95 tasks (Phase 3.1-3.3 backend core + tests)
- **In Progress**: Frontend implementation
- **Remaining**: 56 tasks (Frontend + Polish + Optimization)

---

## Critical TODOs (High Priority)

### 1. Authorization & Security

**Status**: ✅ **COMPLETE**  
**Priority**: HIGH  
**Impact**: Security vulnerabilities, unauthorized access

**Progress**: 7 of 7 items complete (all authorization checks implemented ✅)

#### Backend Issues:

1. **Analytics Authorization Checks** ✅ **COMPLETE**

   - File: `backend/internal/modules/analytics/handlers/analytics_handlers.go`
   - ✅ Line 124: User can only access their own analytics or be admin/coordinator
   - ✅ Line 180: User must be a member/admin of organization to view org analytics
   - ✅ Line 224: Proper role-based access control for platform analytics (super admin only)
   - **Implementation Details**:
     - Created `OrganizationMember` model with roles (admin, coordinator)
     - Added repository methods: `IsMember`, `GetMemberRole`, `FindMemberByOrgAndUser`, `AddMember`, `RemoveMember`
     - Volunteer analytics: Checks profile ownership or staff role
     - Organization analytics: Verifies organization membership or super admin role
     - Platform analytics: Requires super admin role only

2. **Achievements Authorization** ✅ **COMPLETE**

   - File: `backend/internal/modules/achievements/handlers/achievement_handlers.go`
   - ✅ Line 176: Check if user is authorized (org admin/coordinator) to create custom achievements
   - ✅ Line 225: Get current user from context (authenticated user) for awarding achievements
   - **Implementation Details**:
     - Injected `OrganizationRepository` into `AchievementHandler` for membership checks
     - CreateCustomAchievement: Verifies user is admin/coordinator of organization or super admin
     - AwardCustomAchievement:
       - Fetches achievement to determine if it's custom (org-specific) or platform-wide
       - Custom achievements: Requires admin/coordinator of the organization or super admin
       - Platform achievements: Requires super admin only
     - Uses `middleware.MustGetUserUUID()` to get authenticated user from context
     - Updated `main.go` to wire organization repository to achievement handler

3. **RBAC Middleware Organization Membership** ✅ **COMPLETE**

   - File: `backend/internal/middleware/rbac.go`
   - ✅ Line 131: Implement actual organization membership check from database
   - **Implementation Details**:
     - Created `OrganizationMembershipChecker` interface for dependency injection
     - Updated `RequireOrgMembership` to accept optional checker parameter
     - Performs actual database lookup when checker is provided
     - Super admins bypass membership checks
     - Validates UUID formats and handles errors gracefully
     - Backward compatible: passes `nil` for checker to defer to handler level
     - **Usage**: `middleware.RequireOrgMembership("org_id", orgRepo)` in route groups

4. **Organizations Service Authorization** ✅ **COMPLETE**
   - File: `backend/internal/modules/organizations/services/org_service.go`
   - ✅ Line 216: Create organization member record for creator as admin
   - ✅ Line 295: Verify user is admin before allowing organization updates
   - ✅ Line 431: Verify user is admin before allowing organization deletion
   - **Implementation Details**:
     - CreateOrganization: Automatically creates `OrganizationMember` record with `OrgRoleAdmin` for creator
     - UpdateOrganization: Validates user has `OrgRoleAdmin` role via `GetMemberRole` before allowing updates
     - DeleteOrganization: Validates user has `OrgRoleAdmin` role via `GetMemberRole` before allowing deletion
     - Proper error logging for authorization failures
     - Returns `ForbiddenError` when user lacks admin privileges

**Action Items**:

- [x] Implement organization membership repository/service
- [x] Add database queries to verify user roles and org membership
- [x] Update RBAC middleware to perform actual DB lookups
- [x] Add proper authorization checks to all analytics handlers
- [x] Add proper authorization checks to achievement handlers
- [x] Create organization member records on org creation
- [x] Add authorization checks before org updates/deletes

---

### 2. PDF Generation

**Status**: ❌ Not Started  
**Priority**: MEDIUM  
**Impact**: Missing feature for impact reports

#### Backend Issues:

1. **Volunteer Impact Reports**

   - File: `backend/internal/modules/volunteers/services/volunteer_service.go`
   - Line 489: Implement PDF generation using a library like gopdf or wkhtmltopdf

2. **Analytics Reports**
   - File: `backend/internal/modules/analytics/services/analytics_service.go`
   - Line 372: Implement PDF generation using a PDF library (e.g., gofpdf)

**Action Items**:

- [ ] Research and select PDF generation library (gofpdf, wkhtmltopdf, gotenberg)
- [ ] Install PDF generation dependencies
- [ ] Create PDF template for volunteer impact reports
- [ ] Create PDF template for analytics reports
- [ ] Implement PDF generation in volunteer service
- [ ] Implement PDF generation in analytics service
- [ ] Add tests for PDF generation

---

### 3. Notifications Integration

**Status**: ⚠️ Partially Complete  
**Priority**: MEDIUM  
**Impact**: Users not receiving important notifications

#### Backend Issues:

1. **Achievement Notifications**

   - File: `backend/internal/modules/achievements/services/achievement_service.go`
   - Line 356: Send notification to volunteer when achievement is awarded (FR-076)

2. **Registration Hours Update**
   - File: `backend/internal/modules/hours/services/registration_adapter.go`
   - Line 52: Registration service needs UpdateHoursInformation method

**Action Items**:

- [ ] Integrate communications service with achievements module
- [ ] Send notification when achievement is awarded
- [ ] Implement UpdateHoursInformation in registration service
- [ ] Update registration_adapter to call registration service
- [ ] Test notification flow end-to-end

---

### 4. Integration Test Infrastructure

**Status**: ⚠️ Partially Complete  
**Priority**: HIGH  
**Impact**: Cannot run full integration tests

#### Test Issues:

1. **Test Environment Setup**
   - File: `backend/tests/integration/helpers/test_helpers.go`
   - Line 31: Initialize testcontainers for PostgreSQL and Redis
   - Line 44: Initialize HTTP server with Gin router

**Current State**: Using placeholder connection string and test server

**Action Items**:

- [ ] Set up testcontainers for PostgreSQL
- [ ] Set up testcontainers for Redis
- [ ] Initialize actual Gin router for integration tests
- [ ] Wire up all modules to test server
- [ ] Update test helpers to use real services
- [ ] Verify all integration tests pass with real infrastructure

---

### 5. Volunteer Dashboard Data Integration

**Status**: ⚠️ Partially Complete  
**Priority**: MEDIUM  
**Impact**: Dashboard shows placeholder data

#### Backend Issues:

1. **Service Dependencies**
   - File: `backend/internal/modules/volunteers/services/volunteer_service.go`
   - Line 171: Add dependencies for fetching registrations, hours, achievements
   - Line 411: Fetch registrations, hours, achievements for dashboard
   - Line 449: Fetch analytics data from registrations and hours modules

**Current State**: Dashboard returns placeholder values, not real data

**Action Items**:

- [ ] Add registrations service dependency to volunteer service
- [ ] Add hours service dependency to volunteer service
- [ ] Add achievements service dependency to volunteer service
- [ ] Implement GetDashboard to fetch real data from all modules
- [ ] Implement GetAnalytics to aggregate data from all modules
- [ ] Add tests for dashboard data aggregation

---

## Task List Status (from tasks.md)

### Phase 3.1: Project Setup & Infrastructure ✅ COMPLETE

All 9 tasks completed:

- ✅ T001-T009: Backend setup, frontend setup, Docker, migrations, CI/CD

---

### Phase 3.2: Tests First (TDD) ✅ COMPLETE

All 36 tests completed:

- ✅ T010-T014: Authentication contract tests
- ✅ T015-T017: Users contract tests
- ✅ T018-T021: Organizations contract tests
- ✅ T022-T023: Volunteers contract tests
- ✅ T024-T027: Opportunities contract tests
- ✅ T028-T031: Registrations contract tests
- ✅ T032-T034: Hours tracking contract tests
- ✅ T035-T037: Communications contract tests
- ✅ T038-T042: Integration tests (5 user stories)
- ✅ T043-T045: Frontend E2E tests

---

### Phase 3.3: Core Implementation ⚠️ MOSTLY COMPLETE

Backend: 50/50 tasks completed ✅

- ✅ T046-T052: Database & shared infrastructure
- ✅ T053-T056: Authentication module
- ✅ T057-T058: Users module
- ✅ T059-T062: Organizations module
- ✅ T063-T066: Volunteers module
- ✅ T067-T070: Opportunities module
- ✅ T071-T074: Registrations module
- ✅ T075-T078: Hours tracking module
- ✅ T079-T082: Communications module
- ✅ T083-T084: Analytics module
- ✅ T085-T088: Achievements module
- ✅ T089-T094: Middleware
- ✅ T095: Backend main application

**Backend Notes**:

- Core functionality is complete
- Authorization checks need enhancement (see Critical TODOs)
- PDF generation is stubbed out

Frontend: 0/33 tasks completed ❌

- ❌ T096: Configuration management
- ❌ T097-T102: Shared infrastructure (API client, React Query, stores, components)
- ❌ T103-T105: Authentication pages
- ❌ T106-T108: Dashboard layouts
- ❌ T109-T114: Volunteer features
- ❌ T115-T122: Organization features
- ❌ T123-T124: Notifications
- ❌ T125-T127: Shared components

---

### Phase 3.4: Integration & Cron Jobs ❌ NOT STARTED

0/7 tasks completed:

- ❌ T128: Event reminder cron job
- ❌ T129: Auto-verify hours cron job
- ❌ T130: Auto-complete events cron job
- ❌ T131: Check achievements cron job
- ❌ T132: Inactive account checker
- ❌ T133: Geocoding service
- ❌ T134: Notification templates

**Priority**: HIGH for cron jobs, MEDIUM for geocoding

---

### Phase 3.5: Polish & Testing ❌ NOT STARTED

0/16 tasks completed:

**Unit Tests**:

- ❌ T135: Auth service unit tests
- ❌ T136: Hours service unit tests
- ❌ T137: Registration service unit tests
- ❌ T138: Frontend validation unit tests
- ❌ T139: Frontend component tests

**Performance**:

- ❌ T140: Backend performance optimization
- ❌ T141: Frontend performance optimization

**Security**:

- ❌ T142: Security audit
- ❌ T143: Accessibility audit

**Documentation**:

- ❌ T144: API documentation
- ❌ T145: Deployment documentation
- ❌ T146: User guides

**Final Validation**:

- ❌ T147: Run all contract tests
- ❌ T148: Run all integration tests
- ❌ T149: Run all E2E tests
- ❌ T150: Performance validation
- ❌ T151: Manual testing from quickstart.md

---

## Non-Code TODOs (from prompts/docs)

### Documentation TODOs

1. **Constitution/Governance** (`.github/prompts/constitution.prompt.md`)

   - Placeholder for ratification date if unknown
   - Deferred items tracking

2. **Analysis Prompts** (`.github/prompts/analyze.prompt.md`)

   - Flag unresolved placeholders (TODO, TKTK, ???, <placeholder>)

3. **Clarification Prompts** (`.github/prompts/clarify.prompt.md`)
   - TODO markers / unresolved decisions tracking

---

## Quick Reference: Code TODOs by File

### Backend Code TODOs

| File                                            | Line | Description                                  | Priority | Status      |
| ----------------------------------------------- | ---- | -------------------------------------------- | -------- | ----------- |
| `analytics/handlers/analytics_handlers.go`      | 124  | Add authorization check - user own analytics | HIGH     | ✅ COMPLETE |
| `analytics/handlers/analytics_handlers.go`      | 180  | Add authorization check - org member         | HIGH     | ✅ COMPLETE |
| `analytics/handlers/analytics_handlers.go`      | 224  | Implement proper RBAC                        | HIGH     | ✅ COMPLETE |
| `analytics/services/analytics_service.go`       | 372  | Implement PDF generation                     | MEDIUM   | ❌          |
| `achievements/handlers/achievement_handlers.go` | 176  | Check org admin/coordinator auth             | HIGH     | ✅ COMPLETE |
| `achievements/handlers/achievement_handlers.go` | 225  | Get current user from context                | HIGH     | ✅ COMPLETE |
| `achievements/services/achievement_service.go`  | 356  | Send notification on achievement award       | MEDIUM   | ❌          |
| `middleware/rbac.go`                            | 131  | Implement org membership DB check            | HIGH     | ✅ COMPLETE |
| `hours/services/registration_adapter.go`        | 52   | Registration service UpdateHoursInformation  | MEDIUM   | ❌          |
| `volunteers/services/volunteer_service.go`      | 171  | Add dependencies for registrations/hours     | MEDIUM   | ❌          |
| `volunteers/services/volunteer_service.go`      | 411  | Fetch real dashboard data                    | MEDIUM   | ❌          |
| `volunteers/services/volunteer_service.go`      | 449  | Fetch real analytics data                    | MEDIUM   | ❌          |
| `volunteers/services/volunteer_service.go`      | 489  | Implement PDF generation                     | MEDIUM   | ❌          |
| `organizations/services/org_service.go`         | 216  | Create org member record for creator         | HIGH     | ✅ COMPLETE |
| `organizations/services/org_service.go`         | 295  | Verify user is admin before update           | HIGH     | ✅ COMPLETE |
| `organizations/services/org_service.go`         | 431  | Verify user is admin before delete           | HIGH     | ✅ COMPLETE |
| `tests/integration/helpers/test_helpers.go`     | 31   | Initialize testcontainers                    | HIGH     | ❌          |
| `tests/integration/helpers/test_helpers.go`     | 44   | Initialize HTTP server with Gin              | HIGH     | ❌          |

### Frontend Code TODOs

**All frontend tasks are pending** - see Phase 3.3 Frontend section above

---

## Recommended Action Plan

### Sprint 1: Critical Security & Infrastructure (1 week)

**Priority**: Unblock testing and fix security issues

1. **Integration Test Infrastructure** (T031, T044 from test_helpers.go)

   - Set up testcontainers
   - Wire up real Gin router
   - Verify all integration tests pass

2. **Authorization System** (All HIGH priority auth TODOs)

   - Create organization_member table/repository if missing
   - Implement RBAC middleware DB checks
   - Add authorization to analytics handlers
   - Add authorization to achievements handlers
   - Add authorization to organization updates/deletes

3. **Module Integration** (volunteers service TODOs)
   - Wire up registrations/hours/achievements dependencies
   - Implement real dashboard data fetching
   - Test end-to-end data flow

### Sprint 2: Frontend Core (2 weeks)

**Priority**: Get basic frontend working

1. **Infrastructure** (T096-T102)

   - API client, React Query, stores
   - shadcn/ui setup
   - Validation schemas

2. **Authentication** (T103-T105)

   - Registration, login, password reset pages

3. **Layouts & Navigation** (T106-T108)
   - Root layout, volunteer layout, org layout

### Sprint 3: Frontend Features (2 weeks)

**Priority**: Complete user journeys

1. **Volunteer Features** (T109-T114)

   - Dashboard, profile, opportunity search, my events, impact

2. **Organization Features** (T115-T122)

   - Dashboard, create org/opportunity, manage roster, log hours

3. **Notifications** (T123-T124)
   - Notification center and badge

### Sprint 4: Background Jobs & Polish (1 week)

**Priority**: Automation and optimization

1. **Cron Jobs** (T128-T132)

   - Event reminders, auto-verify hours, auto-complete events
   - Check achievements, inactive accounts

2. **PDF Generation** (volunteers & analytics services)

   - Select library and implement

3. **Geocoding** (T133)
   - Implement OpenStreetMap integration

### Sprint 5: Testing & Launch Prep (1 week)

**Priority**: Quality assurance and documentation

1. **Unit Tests** (T135-T139)
2. **Performance Optimization** (T140-T141)
3. **Security & Accessibility Audits** (T142-T143)
4. **Documentation** (T144-T146)
5. **Final Validation** (T147-T151)

---

## Completion Criteria

### Definition of Done:

- [ ] All HIGH priority TODOs resolved
- [ ] All 151 tasks from tasks.md completed
- [ ] All contract tests passing
- [ ] All integration tests passing
- [ ] All E2E tests passing
- [ ] Performance targets met (<2s search, <500ms API p95)
- [ ] Security audit passed
- [ ] Accessibility WCAG 2.1 AA compliance
- [ ] Documentation complete
- [ ] Manual testing scenarios verified

### Blocked Items:

- Frontend E2E tests (T149) - blocked by frontend implementation
- Performance validation (T150) - blocked by frontend implementation
- Manual testing (T151) - blocked by complete implementation

---

## Notes

### Design Decisions Captured:

- Using placeholder data in volunteer dashboard until module integration complete
- RBAC middleware defers to service layer for org membership checks
- PDF generation deferred to sprint 4 (non-blocking for MVP)
- Integration tests use placeholder connection until testcontainers ready

### Technical Debt:

- Registration adapter cannot update registration hours information
- Test infrastructure using placeholders instead of real containers
- Authorization checks incomplete across multiple handlers
- Dashboard and analytics returning mock data

### Dependencies to Track:

- Organization_Member table/repository for RBAC
- Communications service for achievement notifications
- Registration service UpdateHoursInformation method
- Testcontainers for integration tests

---

**Next Steps**: Start with Sprint 1 to fix critical security issues and unblock testing infrastructure.
