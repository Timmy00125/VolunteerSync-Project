# Quickstart: Integration Test Scenarios

**Feature**: VolunteerSync Platform  
**Date**: October 1, 2025  
**Purpose**: End-to-end integration test scenarios validating primary user stories

## Overview

This document defines integration test scenarios that validate the complete user journeys from the feature specification. These scenarios should be implemented as automated integration tests using the testing stack (Playwright for E2E, Go integration tests with testcontainers).

**Test Execution Order**: Tests must be executed in order as they build on each other (e.g., Story 1 creates data used by Story 2).

**Prerequisites**:

- Local development environment running (`docker-compose up`)
- Database seeded with lookup data (Cause_Category, Skill tables)
- Clean database state before test suite execution

---

## Test Suite 1: Organization Onboarding & Opportunity Creation

### Story 1: Organization Posts Volunteer Opportunity

**As an** organization administrator  
**I want to** create and publish a volunteer opportunity  
**So that** volunteers can discover and sign up for our event

**Functional Requirements Validated**: FR-001, FR-002, FR-010, FR-011, FR-017, FR-018, FR-021, NFR-017

#### Test Steps

**Setup**:

```
GIVEN a clean database
AND cause categories exist (Environment, Education, Healthcare)
AND skills exist (Web Development, Teaching, Event Planning)
```

**Scenario 1.1: Organization Admin Registration**

```gherkin
GIVEN I am on the registration page
WHEN I fill in the registration form:
  | Field                 | Value                          |
  | Email                 | admin@greenearth.org           |
  | Password              | SecurePass123                  |
  | First Name            | Jane                           |
  | Last Name             | Smith                          |
  | User Type             | organization_admin             |
  | Security Question 1   | What is your mother's maiden name? |
  | Security Answer 1     | Johnson                        |
  | Security Question 2   | What city were you born in?    |
  | Security Answer 2     | Boston                         |
  | Security Question 3   | What was your first pet's name?|
  | Security Answer 3     | Max                            |
AND I click "Register"
THEN I should see my dashboard
AND my account should be immediately active (FR-002)
AND I should receive an access token and refresh token
AND I should be logged in automatically
```

**Expected Database State**:

- User record created with `account_status = 'active'`
- Security questions and hashed answers stored
- Last_login_at timestamp set

**API Calls**:

- `POST /api/v1/auth/register`

**Assertions**:

```javascript
// Frontend (Playwright E2E)
expect(page.url()).toContain('/dashboard')
expect(page.locator('[data-testid="welcome-message"]')).toContainText('Welcome, Jane')

// Backend (Go Integration Test)
user := getUserByEmail("admin@greenearth.org")
assert.Equal(t, "active", user.AccountStatus)
assert.NotNil(t, user.LastLoginAt)
assert.NotEmpty(t, user.SecurityQuestion1)
```

---

**Scenario 1.2: Create Organization Profile**

```gherkin
GIVEN I am logged in as admin@greenearth.org
WHEN I navigate to "Create Organization"
AND I fill in the organization form:
  | Field              | Value                                    |
  | Name               | Green Earth Initiative                   |
  | Mission Statement  | Protecting the environment through action|
  | Description        | We organize community cleanups...        |
  | Email              | contact@greenearth.org                   |
  | Phone              | +1-555-0100                             |
  | Address Line 1     | 123 Eco Street                          |
  | City               | Portland                                |
  | State              | Oregon                                  |
  | Postal Code        | 97201                                   |
  | Website            | https://greenearth.org                  |
  | Cause Categories   | Environment                             |
AND I upload a logo image
AND I click "Create Organization"
THEN I should see the organization profile page
AND the organization should be auto-verified (FR-015)
AND the profile should have a unique public URL slug
AND the organization should be geocoded (latitude/longitude set)
```

**Expected Database State**:

- Organization record created
- `verification_status = 'verified'`, `verified_at` timestamp set
- `slug = 'green-earth-initiative'` (auto-generated)
- `latitude` and `longitude` populated from geocoding
- Organization_Member record created linking admin user with role 'admin'
- Organization_Cause junction record created

**API Calls**:

- `POST /api/v1/organizations`
- Geocoding service call (to convert address to lat/lng)

**Assertions**:

```javascript
// Frontend
expect(page.url()).toMatch(/\/organizations\/[a-f0-9-]+/)
expect(page.locator('[data-testid="org-name"]')).toHaveText('Green Earth Initiative')
expect(page.locator('[data-testid="verified-badge"]')).toBeVisible()

// Backend
org := getOrganizationBySlug("green-earth-initiative")
assert.Equal(t, "verified", org.VerificationStatus)
assert.NotNil(t, org.Latitude)
assert.NotNil(t, org.Longitude)
assert.Equal(t, 45.5202, *org.Latitude, 0.01) // Portland coords
```

**Performance Target**: Organization creation completes within 5 minutes (NFR-017)

---

**Scenario 1.3: Create Volunteer Opportunity**

```gherkin
GIVEN I am logged in as admin@greenearth.org
AND my organization "Green Earth Initiative" exists
WHEN I navigate to "Create Opportunity"
AND I fill in the opportunity form:
  | Field              | Value                                    |
  | Title              | Beach Cleanup - Ocean Park               |
  | Description        | Join us for a morning beach cleanup...   |
  | Start Date         | 2025-10-15                              |
  | Start Time         | 09:00                                   |
  | End Date           | 2025-10-15                              |
  | End Time           | 12:00                                   |
  | Timezone           | America/Los_Angeles                     |
  | Address Line 1     | Ocean Park Beach                        |
  | City               | San Francisco                           |
  | State              | California                              |
  | Postal Code        | 94121                                   |
  | Capacity           | 20                                      |
  | Min Age            | 16                                      |
  | Cause Categories   | Environment                             |
  | Required Skills    | None                                    |
AND I click "Publish Immediately"
THEN I should see the opportunity detail page
AND the opportunity status should be "published"
AND the opportunity should appear in search results immediately
AND the opportunity should be geocoded
```

**Expected Database State**:

- Opportunity record created
- `status = 'published'`
- `published_at` timestamp set
- `current_registrations = 0`
- `latitude` and `longitude` populated
- Opportunity_Cause junction record created
- `created_by_user_id` set to admin user

**API Calls**:

- `POST /api/v1/opportunities`
- Geocoding service call

**Assertions**:

```javascript
// Frontend
expect(page.locator('[data-testid="opportunity-status"]')).toHaveText('Published')
expect(page.locator('[data-testid="capacity-display"]')).toHaveText('0 of 20 spots filled')

// Backend
opp := getOpportunityByTitle("Beach Cleanup - Ocean Park")
assert.Equal(t, "published", opp.Status)
assert.NotNil(t, opp.PublishedAt)
assert.Equal(t, 20, opp.Capacity)
assert.Equal(t, 0, opp.CurrentRegistrations)
```

**Performance Target**: Opportunity creation + geocoding completes within 5 seconds

---

## Test Suite 2: Volunteer Discovery & Registration

### Story 2: Volunteer Discovers and Registers for Event

**As a** volunteer  
**I want to** find volunteer opportunities matching my interests and availability  
**So that** I can contribute to causes I care about

**Functional Requirements Validated**: FR-028, FR-029, FR-030, FR-031, FR-039, FR-040, FR-055, FR-064, NFR-018

#### Test Steps

**Setup**:

```
GIVEN the organization "Green Earth Initiative" exists
AND the opportunity "Beach Cleanup - Ocean Park" is published
```

**Scenario 2.1: Volunteer Registration**

```gherkin
GIVEN I am on the registration page
WHEN I fill in the registration form:
  | Field                 | Value                          |
  | Email                 | john.volunteer@example.com     |
  | Password              | VolunteerPass123               |
  | First Name            | John                           |
  | Last Name             | Doe                            |
  | User Type             | volunteer                      |
  | Security Questions    | [3 questions with answers]     |
AND I click "Register"
THEN I should be redirected to volunteer profile setup
AND my account should be immediately active
```

**API Calls**:

- `POST /api/v1/auth/register`

**Assertions**:

```javascript
user := getUserByEmail("john.volunteer@example.com")
assert.Equal(t, "active", user.AccountStatus)

volunteerProfile := getVolunteerProfileByUserId(user.ID)
assert.NotNil(t, volunteerProfile)
```

---

**Scenario 2.2: Complete Volunteer Profile**

```gherkin
GIVEN I am logged in as john.volunteer@example.com
WHEN I complete my volunteer profile:
  | Field              | Value                          |
  | Location           | San Francisco, CA              |
  | Biography          | Passionate about environment...|
  | Skills             | Event Planning                 |
  | Interests          | Environment                    |
  | Availability       | Weekends (Sat, Sun)            |
  | Preferred Time     | Morning                        |
  | Profile Photo      | [upload image]                 |
AND I click "Save Profile"
THEN my profile should be updated
AND my location should be geocoded
```

**Expected Database State**:

- Volunteer_Profile record updated
- `availability_saturday = true`, `availability_sunday = true`
- `preferred_time = 'morning'`
- `latitude` and `longitude` set
- Volunteer_Skill junction record created
- Volunteer_Interest junction record created

**API Calls**:

- `PATCH /api/v1/volunteers/me`

---

**Scenario 2.3: Search for Opportunities**

```gherkin
GIVEN I am logged in as john.volunteer@example.com
WHEN I navigate to "Find Opportunities"
AND I apply filters:
  | Filter           | Value                |
  | Location         | San Francisco, CA    |
  | Radius           | 25 miles             |
  | Cause            | Environment          |
  | Date Range       | Oct 1 - Oct 31, 2025 |
AND I click "Search"
THEN I should see search results within 2 seconds (NFR-002)
AND "Beach Cleanup - Ocean Park" should appear in results
AND the result should show "0 of 20 spots filled"
AND I should see the result on an interactive map
```

**API Calls**:

- `GET /api/v1/opportunities?location=San+Francisco,CA&radius=25&cause=environment&start_date=2025-10-01&end_date=2025-10-31`

**Assertions**:

```javascript
// Frontend
const results = page.locator('[data-testid="opportunity-card"]');
await expect(results).toHaveCount(1);
await expect(results.first()).toContainText("Beach Cleanup - Ocean Park");

// Performance check
const searchStartTime = Date.now();
await page.click('[data-testid="search-button"]');
await page.waitForSelector('[data-testid="search-results"]');
const searchDuration = Date.now() - searchStartTime;
expect(searchDuration).toBeLessThan(2000); // <2s requirement
```

---

**Scenario 2.4: Register for Opportunity**

```gherkin
GIVEN I am viewing "Beach Cleanup - Ocean Park" opportunity
WHEN I click "Register for Event"
THEN my registration should be confirmed immediately
AND I should receive an in-platform notification (FR-055)
AND the opportunity should show "1 of 20 spots filled"
AND the event should be added to my calendar
AND I should be able to download a .ics calendar file (FR-064)
```

**Expected Database State**:

- Registration record created
- `status = 'confirmed'`
- `registered_at` timestamp set
- Opportunity `current_registrations = 1`
- Notification record created with `notification_type = 'registration_confirmed'`

**API Calls**:

- `POST /api/v1/registrations`
- `GET /api/v1/registrations/{id}/calendar.ics`

**Assertions**:

```javascript
// Frontend
expect(page.locator('[data-testid="registration-status"]')).toHaveText('Registered')
expect(page.locator('[data-testid="capacity-display"]')).toHaveText('1 of 20 spots filled')

// Backend
reg := getRegistrationByVolunteerAndOpportunity(volunteerProfile.ID, opportunity.ID)
assert.Equal(t, "confirmed", reg.Status)
assert.NotNil(t, reg.RegisteredAt)

opp := getOpportunityById(opportunity.ID)
assert.Equal(t, 1, opp.CurrentRegistrations)

notif := getLatestNotificationForUser(user.ID)
assert.Equal(t, "registration_confirmed", notif.NotificationType)
assert.False(t, notif.ReadAt.Valid) // Unread notification
```

**Performance Target**: Registration completes within 1 second (NFR-006)

---

## Test Suite 3: Event Day Operations & Hours Tracking

### Story 3: Coordinator Manages Event Day Operations

**As a** volunteer coordinator  
**I want to** track volunteer attendance and log hours  
**So that** I can recognize volunteer contributions and generate reports

**Functional Requirements Validated**: FR-046, FR-047, FR-048, FR-049, FR-051, FR-054, FR-056, FR-068

#### Test Steps

**Setup**:

```
GIVEN the volunteer "John Doe" is registered for "Beach Cleanup - Ocean Park"
AND the event date is 2025-10-15 (today in test context)
```

**Scenario 3.1: Event Reminder Notifications**

```gherkin
GIVEN the event "Beach Cleanup - Ocean Park" is tomorrow (24 hours away)
WHEN the notification system runs
THEN John Doe should receive a 24-hour reminder notification (FR-056)
AND the notification should appear in his notification center

GIVEN the event is in 2 hours
WHEN the notification system runs
THEN John Doe should receive a 2-hour reminder notification
```

**Expected Database State**:

- 2 Notification records created for John Doe
- `notification_type = 'event_reminder_24h'` and `'event_reminder_2h'`
- `sent_at` timestamps set

**API Calls**:

- Background job: `POST /api/v1/notifications/send-reminders` (cron job)

**Assertions**:

```javascript
notifications := getNotificationsForUser(user.ID, "event_reminder_24h")
assert.Len(t, notifications, 1)

notifications2h := getNotificationsForUser(user.ID, "event_reminder_2h")
assert.Len(t, notifications2h, 1)
```

---

**Scenario 3.2: Volunteer Check-In**

```gherkin
GIVEN I am logged in as admin@greenearth.org
AND I am viewing the event roster for "Beach Cleanup - Ocean Park"
WHEN I click "Check In" next to John Doe
THEN John Doe's status should show "Checked In"
AND the check-in timestamp should be recorded
```

**Expected Database State**:

- Registration `checked_in_at` timestamp set

**API Calls**:

- `PATCH /api/v1/registrations/{id}/check-in`

---

**Scenario 3.3: Log Volunteer Hours**

```gherkin
GIVEN the event "Beach Cleanup - Ocean Park" has ended
AND John Doe was checked in
WHEN I navigate to "Log Hours" for this event
AND I enter hours for John Doe:
  | Volunteer  | Hours | Notes                          |
  | John Doe   | 3.0   | Great work, very enthusiastic  |
AND I click "Save Hours"
THEN John Doe should receive a notification that hours were logged (FR-047)
AND the hours should be in "pending" status (FR-046)
AND a Hours_Log audit record should be created (FR-054)
```

**Expected Database State**:

- Registration record updated: `hours_worked = 3.0`, `hours_status = 'pending'`, `hours_logged_at` set, `coordinator_notes` set
- Hours_Log record created with `status = 'pending'`, immutable audit trail
- Notification created: `notification_type = 'hours_logged'`
- Volunteer_Profile `total_hours` NOT yet incremented (pending verification)

**API Calls**:

- `POST /api/v1/hours/log`

**Assertions**:

```javascript
// Backend
reg := getRegistrationById(registration.ID)
assert.Equal(t, 3.0, reg.HoursWorked)
assert.Equal(t, "pending", reg.HoursStatus)

hoursLog := getHoursLogByRegistration(registration.ID)
assert.Equal(t, "pending", hoursLog.Status)
assert.Equal(t, adminUser.ID, hoursLog.LoggedByUserID)

notif := getLatestNotificationForUser(user.ID)
assert.Equal(t, "hours_logged", notif.NotificationType)

// Hours not yet added to profile
profile := getVolunteerProfileById(volunteerProfile.ID)
assert.Equal(t, 0.0, profile.TotalHours) // Still 0, pending verification
```

---

**Scenario 3.4: Volunteer Confirms Hours**

```gherkin
GIVEN I am logged in as john.volunteer@example.com
AND I have 3.0 hours logged in "pending" status
WHEN I navigate to my notifications
AND I click on the "Hours Logged" notification
AND I review the hours (3.0 hours)
AND I click "Confirm Hours"
THEN the hours should be verified (FR-048)
AND my total hours should increase to 3.0
AND the registration should show "verified" status (FR-051)
```

**Expected Database State**:

- Registration `hours_status = 'verified'`, `hours_verified_at` timestamp set
- Hours_Log `status = 'verified'`, `verified_at` timestamp set
- Volunteer_Profile `total_hours = 3.0`, `total_events = 1`
- Registration `status = 'completed'`

**API Calls**:

- `POST /api/v1/hours/{id}/verify`

**Assertions**:

```javascript
reg := getRegistrationById(registration.ID)
assert.Equal(t, "verified", reg.HoursStatus)
assert.Equal(t, "completed", reg.Status)

profile := getVolunteerProfileById(volunteerProfile.ID)
assert.Equal(t, 3.0, profile.TotalHours)
assert.Equal(t, 1, profile.TotalEvents)
```

---

**Scenario 3.5: Auto-Verify Hours After 7 Days (FR-049)**

```gherkin
GIVEN a volunteer has hours logged in "pending" status
AND 7 days have passed since hours were logged
WHEN the auto-verification cron job runs
THEN the hours should be auto-verified
AND the volunteer should receive a final notification
```

**Expected Database State**:

- Hours_Log `status = 'verified'`, `auto_verified_at` timestamp set

---

**Scenario 3.6: Volunteer Reviews Event (FR-068)**

```gherkin
GIVEN I am logged in as john.volunteer@example.com
AND I have completed "Beach Cleanup - Ocean Park"
WHEN I navigate to my event history
AND I click "Write Review" for the event
AND I submit a review:
  | Rating  | 5 stars                              |
  | Review  | Amazing experience! Well organized... |
AND I click "Submit Review"
THEN my review should be published immediately (FR-071, no moderation)
AND the organization's average rating should update
```

**Expected Database State**:

- Registration `volunteer_rating = 5`, `volunteer_review` set, `review_submitted_at` timestamp
- Organization `avg_rating` recalculated

**API Calls**:

- `POST /api/v1/registrations/{id}/review`

---

## Test Suite 4: Volunteer Impact Tracking

### Story 4: Volunteer Tracks Personal Impact

**As a** volunteer  
**I want to** view my volunteer history and cumulative impact  
**So that** I can see my contribution to the community

**Functional Requirements Validated**: FR-042, FR-043, FR-045, FR-073, FR-074, FR-076, FR-078

#### Test Steps

**Scenario 4.1: View Personal Dashboard**

```gherkin
GIVEN I am logged in as john.volunteer@example.com
AND I have completed 1 event with 3.0 verified hours
WHEN I navigate to my volunteer dashboard
THEN I should see my impact metrics:
  | Metric                 | Value                    |
  | Total Hours            | 3.0 hours                |
  | Events Attended        | 1 event                  |
  | Organizations Supported| 1 (Green Earth Initiative)|
AND I should see a chart of my hours over time
AND I should see my event history list
```

**API Calls**:

- `GET /api/v1/volunteers/me/dashboard`
- `GET /api/v1/volunteers/me/analytics`

---

**Scenario 4.2: Achievement Badge Earned (FR-073)**

```gherkin
GIVEN the system has achievement badges defined
AND "First Event" badge is awarded for completing 1 event
WHEN John Doe completes his first event
THEN he should automatically earn the "First Event" badge
AND he should receive a congratulatory notification (FR-076)
AND the badge should appear on his profile (FR-074)
```

**Expected Database State**:

- Volunteer_Achievement record created
- `earned_at` timestamp set
- Notification created: `notification_type = 'achievement_earned'`

**API Calls**:

- Background job: Check and award achievements

**Assertions**:

```javascript
achievement := getVolunteerAchievements(volunteerProfile.ID)
assert.Len(t, achievement, 1)
assert.Equal(t, "First Event", achievement[0].Name)

notif := getLatestNotificationForUser(user.ID)
assert.Equal(t, "achievement_earned", notif.NotificationType)
```

---

**Scenario 4.3: Download Impact Report (FR-045)**

```gherkin
GIVEN I am on my volunteer dashboard
WHEN I click "Download Impact Report"
THEN I should receive a PDF report with:
  - Total hours volunteered
  - Events attended
  - Organizations supported
  - Achievement badges earned
  - Date range covered
```

**API Calls**:

- `GET /api/v1/volunteers/me/report?format=pdf`

---

## Test Suite 5: Edge Cases & Error Handling

### Scenario 5.1: Event at Capacity (FR-030, FR-035)

```gherkin
GIVEN an opportunity has capacity of 2
AND 2 volunteers are already registered
WHEN a third volunteer attempts to register
THEN they should be added to the waitlist
AND they should receive a notification that they are waitlisted
AND the opportunity should show "Full (2 of 2)"
```

---

### Scenario 5.2: Late Cancellation (FR-034)

```gherkin
GIVEN I am registered for an event starting in 12 hours
WHEN I cancel my registration
THEN I should see a warning about late cancellation
AND my cancellation should be recorded with reason
AND the organization should receive an immediate notification
```

---

### Scenario 5.3: Overlapping Events (FR-032)

```gherkin
GIVEN I am registered for Event A from 9:00-12:00
WHEN I attempt to register for Event B from 11:00-14:00
THEN I should see a warning about overlapping times
AND I should be able to override with confirmation
```

---

### Scenario 5.4: Hours Dispute (FR-050)

```gherkin
GIVEN I have 3.0 hours logged by coordinator
WHEN I review the hours and click "Dispute Hours"
AND I provide a reason: "I worked 4 hours, not 3"
THEN the hours status should change to "disputed"
AND the coordinator should receive a notification
AND a dispute resolution workflow should initiate
```

---

## Performance Benchmarks

All scenarios must meet these performance targets:

- **Page Load**: <2 seconds on 3G network (NFR-002)
- **Search Results**: <2 seconds (NFR-002)
- **Dashboard Load**: <3 seconds (NFR-003)
- **API Responses**: <500ms p95 (NFR)
- **Notification Delivery**: <1 second (NFR-005)
- **Registration Processing**: <1 second (NFR-006)

---

## Test Execution

### Manual Execution (Development)

```bash
# Start local environment
docker-compose up

# Run backend integration tests
cd backend
go test ./tests/integration/... -v

# Run frontend E2E tests
cd frontend
npx playwright test tests/e2e/quickstart.spec.ts
```

### CI Pipeline Execution

```yaml
# .github/workflows/ci.yml
- name: Run Integration Tests
  run: |
    docker-compose up -d
    go test ./tests/integration/...
    npx playwright test
```

---

## Success Criteria

**Quickstart tests PASS when**:

- ✅ All 4 primary user stories execute successfully end-to-end
- ✅ All database state assertions pass
- ✅ All API responses match expected schemas
- ✅ All performance benchmarks are met
- ✅ All constitutional requirements validated (security, accessibility, performance)
- ✅ Tests can run repeatedly in CI with clean state

**Status**: Quickstart scenarios defined ✅  
**Next Step**: Implement automated tests for each scenario using Playwright (frontend) and Go integration tests (backend)
