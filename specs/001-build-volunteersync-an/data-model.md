# Phase 1: Data Model

**Feature**: VolunteerSync Platform  
**Date**: October 1, 2025  
**Status**: Complete

## Overview

This document defines the database schema for VolunteerSync, extracted from the feature specification and functional requirements. The data model supports 111 functional requirements across 16 major feature modules.

**Database**: PostgreSQL 16  
**ORM**: GORM (Go)  
**Design Principles**:

- Normalized schema to 3NF (Third Normal Form)
- Soft deletes for audit trail (deleted_at timestamp)
- Timestamps on all entities (created_at, updated_at)
- Foreign key constraints for referential integrity
- Indexes on foreign keys and frequently queried fields

---

## Entity Relationship Diagram (Conceptual)

```
User (1) ──── (*) Organization_Member ──── (1) Organization
  │                                              │
  │ (1)                                          │ (1)
  │                                              │
  ▼ (0..1)                                       ▼ (*)
Volunteer_Profile                           Opportunity
  │                                              │
  │ (*)                                          │ (*)
  │                                              │
  └──────────── Registration ◄──────────────────┘
                    │ (1)
                    │
                    ▼ (*)
                Hours_Log

User (1) ────── (*) Message ────── (*) User
User (1) ────── (*) Notification
User (1) ────── (*) Team_Member ──── (1) Team
Volunteer_Profile (*) ────── (*) Achievement
Organization (1) ────── (*) Document
Volunteer_Profile (1) ────── (*) Saved_Search
```

---

## Core Entities

### 1. User

**Purpose**: Core entity representing all platform users with authentication credentials.

**Attributes**:

| Field                    | Type         | Constraints                | Description                                             |
| ------------------------ | ------------ | -------------------------- | ------------------------------------------------------- |
| `id`                     | UUID         | PRIMARY KEY                | Unique identifier                                       |
| `email`                  | VARCHAR(255) | UNIQUE, NOT NULL, INDEX    | User email (used for login)                             |
| `password_hash`          | VARCHAR(255) | NOT NULL                   | Argon2id hashed password                                |
| `first_name`             | VARCHAR(100) | NOT NULL                   | User's first name                                       |
| `last_name`              | VARCHAR(100) | NOT NULL                   | User's last name                                        |
| `phone`                  | VARCHAR(20)  | NULL                       | Optional phone number                                   |
| `account_status`         | ENUM         | NOT NULL, DEFAULT 'active' | active, inactive, suspended                             |
| `last_login_at`          | TIMESTAMP    | NULL                       | Last successful login timestamp                         |
| `email_verified`         | BOOLEAN      | NOT NULL, DEFAULT false    | Email verification status (V1: immediate true per spec) |
| `security_question_1`    | VARCHAR(255) | NOT NULL                   | First security question                                 |
| `security_answer_1_hash` | VARCHAR(255) | NOT NULL                   | Hashed answer to question 1                             |
| `security_question_2`    | VARCHAR(255) | NOT NULL                   | Second security question                                |
| `security_answer_2_hash` | VARCHAR(255) | NOT NULL                   | Hashed answer to question 2                             |
| `security_question_3`    | VARCHAR(255) | NOT NULL                   | Third security question                                 |
| `security_answer_3_hash` | VARCHAR(255) | NOT NULL                   | Hashed answer to question 3                             |
| `created_at`             | TIMESTAMP    | NOT NULL                   | Record creation timestamp                               |
| `updated_at`             | TIMESTAMP    | NOT NULL                   | Last update timestamp                                   |
| `deleted_at`             | TIMESTAMP    | NULL, INDEX                | Soft delete timestamp                                   |

**Relationships**:

- 1:1 with `Volunteer_Profile` (optional, only for volunteer users)
- 1:N with `Organization_Member` (user can be member of multiple organizations)
- 1:N with `Message` (as sender or recipient)
- 1:N with `Notification` (recipient)
- 1:N with `Team_Member` (member of multiple teams)

**Indexes**:

- `email` (unique, for login)
- `deleted_at` (for soft delete queries)
- `account_status` (for active user queries)

**Business Rules** (enforced in application layer):

- Email must be valid format
- Password must meet complexity requirements (8+ chars, letters + numbers)
- Account marked inactive after 12 months of no login (FR-008)
- All 3 security questions must be set during registration (FR-003)
- Minimum 2 correct security question answers required for password reset (FR-003a)

---

### 2. Organization

**Purpose**: Represents nonprofit organizations, community groups managing volunteer programs.

**Attributes**:

| Field                 | Type          | Constraints                       | Description                                       |
| --------------------- | ------------- | --------------------------------- | ------------------------------------------------- |
| `id`                  | UUID          | PRIMARY KEY                       | Unique identifier                                 |
| `name`                | VARCHAR(200)  | NOT NULL, INDEX                   | Organization name                                 |
| `slug`                | VARCHAR(200)  | UNIQUE, NOT NULL                  | URL-friendly name (for public profile page)       |
| `mission_statement`   | TEXT          | NULL                              | Organization's mission                            |
| `description`         | TEXT          | NULL                              | Full description of organization                  |
| `website`             | VARCHAR(255)  | NULL                              | Organization website URL                          |
| `email`               | VARCHAR(255)  | NOT NULL                          | Primary contact email                             |
| `phone`               | VARCHAR(20)   | NULL                              | Primary contact phone                             |
| `address_line_1`      | VARCHAR(255)  | NULL                              | Street address                                    |
| `address_line_2`      | VARCHAR(255)  | NULL                              | Apartment, suite, etc.                            |
| `city`                | VARCHAR(100)  | NULL                              | City                                              |
| `state`               | VARCHAR(50)   | NULL                              | State/province                                    |
| `postal_code`         | VARCHAR(20)   | NULL                              | ZIP/postal code                                   |
| `country`             | VARCHAR(100)  | NOT NULL, DEFAULT 'United States' | Country                                           |
| `latitude`            | DECIMAL(10,7) | NULL                              | Geocoded latitude                                 |
| `longitude`           | DECIMAL(10,7) | NULL                              | Geocoded longitude                                |
| `logo_url`            | VARCHAR(500)  | NULL                              | Organization logo image URL                       |
| `banner_url`          | VARCHAR(500)  | NULL                              | Profile banner image URL                          |
| `verification_status` | ENUM          | NOT NULL, DEFAULT 'verified'      | verified, unverified (V1: auto-verified per spec) |
| `verified_at`         | TIMESTAMP     | NULL                              | Verification timestamp                            |
| `total_volunteers`    | INT           | NOT NULL, DEFAULT 0               | Cached count of unique volunteers                 |
| `total_hours`         | DECIMAL(10,2) | NOT NULL, DEFAULT 0               | Cached sum of verified volunteer hours            |
| `avg_rating`          | DECIMAL(3,2)  | NULL                              | Average rating from volunteer reviews             |
| `created_at`          | TIMESTAMP     | NOT NULL                          | Record creation timestamp                         |
| `updated_at`          | TIMESTAMP     | NOT NULL                          | Last update timestamp                             |
| `deleted_at`          | TIMESTAMP     | NULL, INDEX                       | Soft delete timestamp                             |

**Relationships**:

- 1:N with `Organization_Member` (team members)
- 1:N with `Opportunity` (opportunities posted)
- 1:N with `Document` (required documents)
- 1:N with `Achievement` (custom badges)
- N:M with `Cause_Category` (via `Organization_Cause` join table)

**Indexes**:

- `slug` (unique, for public URLs)
- `name` (for search)
- `city, state` (composite, for location-based search)
- `verification_status` (for filtering verified orgs)
- `deleted_at` (for soft delete queries)

**Business Rules**:

- Slug auto-generated from name (lowercase, hyphens)
- All organizations auto-verified on registration (FR-015)
- Geocoding triggered on address save (lat/lng from address)
- Analytics counters (total_volunteers, total_hours) updated via triggers or batch jobs

---

### 3. Organization_Member

**Purpose**: Junction table linking users to organizations with role assignments.

**Attributes**:

| Field             | Type      | Constraints                 | Description                     |
| ----------------- | --------- | --------------------------- | ------------------------------- |
| `id`              | UUID      | PRIMARY KEY                 | Unique identifier               |
| `organization_id` | UUID      | NOT NULL, FK → Organization | Organization reference          |
| `user_id`         | UUID      | NOT NULL, FK → User         | User reference                  |
| `role`            | ENUM      | NOT NULL                    | admin, coordinator              |
| `invited_by`      | UUID      | NULL, FK → User             | User who sent invitation        |
| `invited_at`      | TIMESTAMP | NULL                        | Invitation timestamp            |
| `joined_at`       | TIMESTAMP | NOT NULL                    | When user joined organization   |
| `created_at`      | TIMESTAMP | NOT NULL                    | Record creation timestamp       |
| `updated_at`      | TIMESTAMP | NOT NULL                    | Last update timestamp           |
| `deleted_at`      | TIMESTAMP | NULL, INDEX                 | Soft delete (removed from team) |

**Relationships**:

- N:1 with `Organization`
- N:1 with `User`

**Indexes**:

- `organization_id, user_id` (composite unique, for membership lookup)
- `user_id` (for user's organization list)

**Business Rules**:

- User can have different roles in different organizations (FR-006)
- Admin role can invite members, edit organization, delete organization
- Coordinator role can create/edit opportunities, log hours, send messages
- When coordinator leaves, their opportunities remain (reassign or mark as created_by: [departed])

---

### 4. Volunteer_Profile

**Purpose**: Extended profile data specific to volunteer users.

**Attributes**:

| Field                        | Type          | Constraints                 | Description                           |
| ---------------------------- | ------------- | --------------------------- | ------------------------------------- |
| `id`                         | UUID          | PRIMARY KEY                 | Unique identifier                     |
| `user_id`                    | UUID          | UNIQUE, NOT NULL, FK → User | One-to-one with User                  |
| `profile_photo_url`          | VARCHAR(500)  | NULL                        | Profile photo URL                     |
| `biography`                  | TEXT          | NULL                        | Personal bio                          |
| `location`                   | VARCHAR(255)  | NULL                        | City, state for search                |
| `latitude`                   | DECIMAL(10,7) | NULL                        | Geocoded latitude                     |
| `longitude`                  | DECIMAL(10,7) | NULL                        | Geocoded longitude                    |
| `availability_monday`        | BOOLEAN       | NOT NULL, DEFAULT false     | Available on Mondays                  |
| `availability_tuesday`       | BOOLEAN       | NOT NULL, DEFAULT false     | Available on Tuesdays                 |
| `availability_wednesday`     | BOOLEAN       | NOT NULL, DEFAULT false     | Available on Wednesdays               |
| `availability_thursday`      | BOOLEAN       | NOT NULL, DEFAULT false     | Available on Thursdays                |
| `availability_friday`        | BOOLEAN       | NOT NULL, DEFAULT false     | Available on Fridays                  |
| `availability_saturday`      | BOOLEAN       | NOT NULL, DEFAULT false     | Available on Saturdays                |
| `availability_sunday`        | BOOLEAN       | NOT NULL, DEFAULT false     | Available on Sundays                  |
| `preferred_time`             | ENUM          | NULL                        | morning, afternoon, evening, flexible |
| `total_hours`                | DECIMAL(10,2) | NOT NULL, DEFAULT 0         | Cached sum of verified hours          |
| `total_events`               | INT           | NOT NULL, DEFAULT 0         | Cached count of completed events      |
| `emergency_contact_name`     | VARCHAR(200)  | NULL                        | Emergency contact name                |
| `emergency_contact_phone`    | VARCHAR(20)   | NULL                        | Emergency contact phone               |
| `privacy_show_hours`         | BOOLEAN       | NOT NULL, DEFAULT true      | Show hours publicly                   |
| `privacy_show_events`        | BOOLEAN       | NOT NULL, DEFAULT true      | Show event history publicly           |
| `privacy_show_organizations` | BOOLEAN       | NOT NULL, DEFAULT true      | Show supported orgs publicly          |
| `notification_in_app`        | BOOLEAN       | NOT NULL, DEFAULT true      | In-app notifications enabled          |
| `notification_browser_push`  | BOOLEAN       | NOT NULL, DEFAULT false     | Browser push notifications enabled    |
| `created_at`                 | TIMESTAMP     | NOT NULL                    | Record creation timestamp             |
| `updated_at`                 | TIMESTAMP     | NOT NULL                    | Last update timestamp                 |
| `deleted_at`                 | TIMESTAMP     | NULL                        | Soft delete timestamp                 |

**Relationships**:

- 1:1 with `User`
- 1:N with `Registration` (event registrations)
- 1:N with `Saved_Search` (saved searches)
- N:M with `Skill` (via `Volunteer_Skill` join table)
- N:M with `Cause_Category` (via `Volunteer_Interest` join table)
- N:M with `Achievement` (via `Volunteer_Achievement` join table)
- N:M with `Organization` (via favorites join table)

**Indexes**:

- `user_id` (unique, for profile lookup)
- `location` (for location-based search)
- `latitude, longitude` (for geospatial queries)

**Business Rules**:

- Created automatically when user registers as volunteer
- Geocoding triggered on location save
- Analytics counters (total_hours, total_events) updated when hours verified
- Privacy settings control public profile visibility (FR-044)

---

### 5. Opportunity

**Purpose**: Represents individual volunteer events or shifts.

**Attributes**:

| Field                   | Type          | Constraints                       | Description                                   |
| ----------------------- | ------------- | --------------------------------- | --------------------------------------------- |
| `id`                    | UUID          | PRIMARY KEY                       | Unique identifier                             |
| `organization_id`       | UUID          | NOT NULL, FK → Organization       | Organization hosting event                    |
| `created_by_user_id`    | UUID          | NOT NULL, FK → User               | Coordinator who created opportunity           |
| `title`                 | VARCHAR(200)  | NOT NULL, INDEX                   | Opportunity title                             |
| `description`           | TEXT          | NOT NULL                          | Full description                              |
| `status`                | ENUM          | NOT NULL, DEFAULT 'draft'         | draft, published, cancelled, completed        |
| `start_date`            | DATE          | NOT NULL, INDEX                   | Event start date                              |
| `start_time`            | TIME          | NOT NULL                          | Event start time                              |
| `end_date`              | DATE          | NOT NULL                          | Event end date (same as start for single-day) |
| `end_time`              | TIME          | NOT NULL                          | Event end time                                |
| `timezone`              | VARCHAR(50)   | NOT NULL                          | Timezone (e.g., America/New_York)             |
| `address_line_1`        | VARCHAR(255)  | NOT NULL                          | Event location street address                 |
| `address_line_2`        | VARCHAR(255)  | NULL                              | Apartment, suite, etc.                        |
| `city`                  | VARCHAR(100)  | NOT NULL, INDEX                   | City                                          |
| `state`                 | VARCHAR(50)   | NOT NULL, INDEX                   | State/province                                |
| `postal_code`           | VARCHAR(20)   | NOT NULL                          | ZIP/postal code                               |
| `country`               | VARCHAR(100)  | NOT NULL, DEFAULT 'United States' | Country                                       |
| `latitude`              | DECIMAL(10,7) | NULL                              | Geocoded latitude                             |
| `longitude`             | DECIMAL(10,7) | NULL                              | Geocoded longitude                            |
| `capacity`              | INT           | NOT NULL                          | Maximum volunteers                            |
| `current_registrations` | INT           | NOT NULL, DEFAULT 0               | Cached current registration count             |
| `min_age`               | INT           | NULL                              | Minimum age requirement                       |
| `is_recurring`          | BOOLEAN       | NOT NULL, DEFAULT false           | Whether opportunity is recurring              |
| `recurrence_pattern`    | VARCHAR(50)   | NULL                              | daily, weekly, monthly, custom                |
| `recurrence_end_date`   | DATE          | NULL                              | When recurrence ends                          |
| `parent_opportunity_id` | UUID          | NULL, FK → Opportunity            | Parent if this is recurrence instance         |
| `published_at`          | TIMESTAMP     | NULL                              | When opportunity was published                |
| `cancelled_at`          | TIMESTAMP     | NULL                              | When opportunity was cancelled                |
| `cancellation_reason`   | TEXT          | NULL                              | Reason for cancellation                       |
| `completed_at`          | TIMESTAMP     | NULL                              | When marked complete                          |
| `auto_complete_at`      | TIMESTAMP     | NULL                              | Auto-complete 7 days after end_date           |
| `created_at`            | TIMESTAMP     | NOT NULL, INDEX                   | Record creation timestamp                     |
| `updated_at`            | TIMESTAMP     | NOT NULL                          | Last update timestamp                         |
| `deleted_at`            | TIMESTAMP     | NULL                              | Soft delete timestamp                         |

**Relationships**:

- N:1 with `Organization`
- N:1 with `User` (creator/coordinator)
- 1:N with `Opportunity` (parent to recurring instances)
- 1:N with `Registration` (volunteer registrations)
- N:M with `Skill` (via `Opportunity_Skill` join table - required skills)
- N:M with `Document` (via `Opportunity_Document` join table - required docs)
- N:M with `Cause_Category` (via `Opportunity_Cause` join table)

**Indexes**:

- `organization_id` (for org's opportunities)
- `created_by_user_id` (for coordinator's opportunities)
- `status` (for filtering by status)
- `start_date, start_time` (composite, for date range queries)
- `city, state` (composite, for location search)
- `latitude, longitude` (for geospatial queries)
- `created_at` (for recent opportunities)

**Business Rules**:

- Cannot edit past events (start_date < today) except for hour logging (FR-023)
- Hour logging allowed for 7 days after completion (FR-053)
- Auto-complete 7 days after end_date if not manually completed (FR-025)
- Geocoding triggered on address save
- Current_registrations incremented/decremented on registration changes
- When capacity reached, opportunity marked full (FR-030)
- Recurring opportunities generate child instances (FR-018)

---

### 6. Registration

**Purpose**: Junction table linking volunteers to opportunities (event sign-ups).

**Attributes**:

| Field                  | Type         | Constraints                      | Description                                 |
| ---------------------- | ------------ | -------------------------------- | ------------------------------------------- |
| `id`                   | UUID         | PRIMARY KEY                      | Unique identifier                           |
| `opportunity_id`       | UUID         | NOT NULL, FK → Opportunity       | Opportunity reference                       |
| `volunteer_profile_id` | UUID         | NOT NULL, FK → Volunteer_Profile | Volunteer reference                         |
| `status`               | ENUM         | NOT NULL, DEFAULT 'confirmed'    | confirmed, cancelled, waitlisted, completed |
| `registered_at`        | TIMESTAMP    | NOT NULL                         | Registration timestamp                      |
| `checked_in_at`        | TIMESTAMP    | NULL                             | Check-in timestamp (event day)              |
| `cancelled_at`         | TIMESTAMP    | NULL                             | Cancellation timestamp                      |
| `cancellation_reason`  | TEXT         | NULL                             | Reason for cancellation                     |
| `hours_worked`         | DECIMAL(5,2) | NULL                             | Hours worked (logged by coordinator)        |
| `hours_status`         | ENUM         | NULL                             | pending, verified, disputed                 |
| `hours_logged_at`      | TIMESTAMP    | NULL                             | When hours were logged                      |
| `hours_verified_at`    | TIMESTAMP    | NULL                             | When volunteer verified hours               |
| `volunteer_rating`     | INT          | NULL                             | Volunteer's rating (1-5 stars)              |
| `volunteer_review`     | TEXT         | NULL                             | Volunteer's written review                  |
| `review_submitted_at`  | TIMESTAMP    | NULL                             | Review submission timestamp                 |
| `coordinator_notes`    | TEXT         | NULL                             | Private notes from coordinator              |
| `created_at`           | TIMESTAMP    | NOT NULL                         | Record creation timestamp                   |
| `updated_at`           | TIMESTAMP    | NOT NULL                         | Last update timestamp                       |
| `deleted_at`           | TIMESTAMP    | NULL                             | Soft delete timestamp                       |

**Relationships**:

- N:1 with `Opportunity`
- N:1 with `Volunteer_Profile`
- 1:N with `Hours_Log` (hour log entries for audit trail)

**Indexes**:

- `opportunity_id, volunteer_profile_id` (composite unique, prevent duplicate registrations)
- `volunteer_profile_id` (for volunteer's registrations)
- `status` (for filtering by status)
- `registered_at` (for recent registrations)

**Business Rules**:

- Volunteer cannot register twice for same opportunity (unique constraint)
- Status transitions: confirmed → cancelled, confirmed → waitlisted → confirmed, confirmed → completed
- Late cancellation (within 24 hours) triggers warning (FR-034)
- Hours auto-verify after 7 days if volunteer doesn't respond (FR-049)
- Review only allowed if status = completed and volunteer attended (FR-072)
- Hours logged by coordinator become pending until volunteer verifies (FR-046, FR-047)

---

### 7. Hours_Log

**Purpose**: Detailed audit trail of volunteer hours worked and verification.

**Attributes**:

| Field               | Type         | Constraints                 | Description                  |
| ------------------- | ------------ | --------------------------- | ---------------------------- |
| `id`                | UUID         | PRIMARY KEY                 | Unique identifier            |
| `registration_id`   | UUID         | NOT NULL, FK → Registration | Registration reference       |
| `hours`             | DECIMAL(5,2) | NOT NULL                    | Hours amount                 |
| `logged_by_user_id` | UUID         | NOT NULL, FK → User         | Coordinator who logged hours |
| `status`            | ENUM         | NOT NULL                    | pending, verified, disputed  |
| `coordinator_notes` | TEXT         | NULL                        | Notes from coordinator       |
| `volunteer_notes`   | TEXT         | NULL                        | Notes from volunteer         |
| `dispute_reason`    | TEXT         | NULL                        | Reason for dispute           |
| `disputed_at`       | TIMESTAMP    | NULL                        | When dispute was filed       |
| `resolved_at`       | TIMESTAMP    | NULL                        | When dispute was resolved    |
| `resolution_notes`  | TEXT         | NULL                        | Resolution details           |
| `logged_at`         | TIMESTAMP    | NOT NULL                    | When hours were logged       |
| `verified_at`       | TIMESTAMP    | NULL                        | When volunteer verified      |
| `auto_verified_at`  | TIMESTAMP    | NULL                        | When auto-verified (7 days)  |
| `created_at`        | TIMESTAMP    | NOT NULL                    | Record creation timestamp    |
| `updated_at`        | TIMESTAMP    | NOT NULL                    | Last update timestamp        |

**Relationships**:

- N:1 with `Registration`
- N:1 with `User` (logged_by coordinator)

**Indexes**:

- `registration_id` (for registration's hour logs)
- `logged_by_user_id` (for coordinator's logged hours)
- `status` (for filtering by status)

**Business Rules**:

- Immutable audit log (no deletes, only status updates) (FR-054)
- Hours must be positive (reject negative values) (FR edge case)
- Status transitions: pending → verified, pending → disputed → verified
- Auto-verify after 7 days if volunteer doesn't respond (FR-049)
- Disputed hours trigger notification to coordinator (FR-050)

---

### 8. Message

**Purpose**: Communication records between users (direct and broadcast messages).

**Attributes**:

| Field            | Type         | Constraints            | Description                         |
| ---------------- | ------------ | ---------------------- | ----------------------------------- |
| `id`             | UUID         | PRIMARY KEY            | Unique identifier                   |
| `sender_id`      | UUID         | NOT NULL, FK → User    | Sender user                         |
| `opportunity_id` | UUID         | NULL, FK → Opportunity | Related opportunity (for broadcast) |
| `message_type`   | ENUM         | NOT NULL               | direct, broadcast                   |
| `subject`        | VARCHAR(255) | NULL                   | Message subject                     |
| `content`        | TEXT         | NOT NULL               | Message body                        |
| `sent_at`        | TIMESTAMP    | NOT NULL, INDEX        | Sent timestamp                      |
| `created_at`     | TIMESTAMP    | NOT NULL               | Record creation timestamp           |
| `updated_at`     | TIMESTAMP    | NOT NULL               | Last update timestamp               |

**Relationships**:

- N:1 with `User` (sender)
- N:1 with `Opportunity` (optional, for event-related messages)
- 1:N with `Message_Recipient` (recipients for direct messages)

**Indexes**:

- `sender_id` (for sent messages)
- `opportunity_id` (for event-related messages)
- `sent_at` (for recent messages)

**Business Rules**:

- Direct messages have individual recipients in Message_Recipient table
- Broadcast messages sent to all registered volunteers for opportunity (FR-058)
- Message content sanitized before rendering (XSS prevention)

---

### 9. Message_Recipient

**Purpose**: Junction table for message recipients (for direct messages).

**Attributes**:

| Field          | Type      | Constraints            | Description               |
| -------------- | --------- | ---------------------- | ------------------------- |
| `id`           | UUID      | PRIMARY KEY            | Unique identifier         |
| `message_id`   | UUID      | NOT NULL, FK → Message | Message reference         |
| `recipient_id` | UUID      | NOT NULL, FK → User    | Recipient user            |
| `read_at`      | TIMESTAMP | NULL                   | When message was read     |
| `created_at`   | TIMESTAMP | NOT NULL               | Record creation timestamp |

**Relationships**:

- N:1 with `Message`
- N:1 with `User` (recipient)

**Indexes**:

- `message_id` (for message recipients)
- `recipient_id` (for user's received messages)
- `read_at` (for unread messages: WHERE read_at IS NULL)

**Business Rules**:

- Read status tracked per recipient
- Unread count displayed in notification center

---

### 10. Notification

**Purpose**: System-generated notifications for users.

**Attributes**:

| Field                 | Type         | Constraints                | Description                                                                                   |
| --------------------- | ------------ | -------------------------- | --------------------------------------------------------------------------------------------- |
| `id`                  | UUID         | PRIMARY KEY                | Unique identifier                                                                             |
| `recipient_id`        | UUID         | NOT NULL, FK → User, INDEX | Recipient user                                                                                |
| `notification_type`   | ENUM         | NOT NULL                   | registration_confirmed, event_reminder, hours_logged, message_received, event_cancelled, etc. |
| `title`               | VARCHAR(255) | NOT NULL                   | Notification title                                                                            |
| `content`             | TEXT         | NOT NULL                   | Notification body                                                                             |
| `action_url`          | VARCHAR(500) | NULL                       | Deep link to related page                                                                     |
| `priority`            | ENUM         | NOT NULL, DEFAULT 'normal' | low, normal, high, critical                                                                   |
| `related_entity_type` | VARCHAR(50)  | NULL                       | opportunity, registration, message, etc.                                                      |
| `related_entity_id`   | UUID         | NULL                       | Related entity ID (polymorphic)                                                               |
| `read_at`             | TIMESTAMP    | NULL, INDEX                | When notification was read                                                                    |
| `delivered_at`        | TIMESTAMP    | NULL                       | When notification was delivered to user                                                       |
| `delivery_method`     | ENUM         | NOT NULL, DEFAULT 'in_app' | in_app, browser_push                                                                          |
| `sent_at`             | TIMESTAMP    | NOT NULL, INDEX            | Notification creation timestamp                                                               |
| `created_at`          | TIMESTAMP    | NOT NULL                   | Record creation timestamp                                                                     |
| `updated_at`          | TIMESTAMP    | NOT NULL                   | Last update timestamp                                                                         |

**Relationships**:

- N:1 with `User` (recipient)

**Indexes**:

- `recipient_id` (for user's notifications)
- `recipient_id, read_at` (composite, for unread notifications: WHERE read_at IS NULL)
- `sent_at` (for recent notifications)
- `notification_type` (for filtering by type)

**Business Rules**:

- In-app notifications always created (FR-055)
- Browser push notifications only if user enabled preference (FR-059)
- Notifications created for: registration confirmation (FR-055), event reminders (FR-056), hours logged (FR-047), messages (FR-057), event cancellations (FR-024), waitlist notifications (FR-036)
- Unread count displayed in header (FR-063)
- Critical notifications (cancellations, urgent updates) always show in notification center even if user disabled alerts (FR-062)

---

### 11. Achievement

**Purpose**: Recognition badges for volunteer milestones and custom organization awards.

**Attributes**:

| Field             | Type         | Constraints             | Description                                            |
| ----------------- | ------------ | ----------------------- | ------------------------------------------------------ |
| `id`              | UUID         | PRIMARY KEY             | Unique identifier                                      |
| `organization_id` | UUID         | NULL, FK → Organization | Org if custom badge, NULL if system badge              |
| `name`            | VARCHAR(100) | NOT NULL                | Badge name                                             |
| `description`     | TEXT         | NOT NULL                | Badge description                                      |
| `icon_url`        | VARCHAR(500) | NULL                    | Badge icon image URL                                   |
| `badge_type`      | ENUM         | NOT NULL                | system, organization_custom                            |
| `criteria_type`   | ENUM         | NULL                    | hours_milestone, events_milestone, consistency, custom |
| `criteria_value`  | INT          | NULL                    | Threshold value (e.g., 50 for 50 hours)                |
| `created_at`      | TIMESTAMP    | NOT NULL                | Record creation timestamp                              |
| `updated_at`      | TIMESTAMP    | NOT NULL                | Last update timestamp                                  |
| `deleted_at`      | TIMESTAMP    | NULL                    | Soft delete timestamp                                  |

**Relationships**:

- N:1 with `Organization` (optional, for custom badges)
- N:M with `Volunteer_Profile` (via `Volunteer_Achievement` join table)

**Indexes**:

- `organization_id` (for org's custom badges)
- `badge_type` (for system vs custom badges)

**Business Rules**:

- System badges awarded automatically based on criteria (FR-073)
- Organization custom badges awarded manually by coordinators (FR-075)
- Examples: "10 Hours", "25 Hours", "100 Hours", "First Event", "10 Events", "3 Months Consistent"

---

### 12. Volunteer_Achievement

**Purpose**: Junction table tracking which volunteers earned which badges.

**Attributes**:

| Field                  | Type      | Constraints                      | Description                          |
| ---------------------- | --------- | -------------------------------- | ------------------------------------ |
| `id`                   | UUID      | PRIMARY KEY                      | Unique identifier                    |
| `volunteer_profile_id` | UUID      | NOT NULL, FK → Volunteer_Profile | Volunteer reference                  |
| `achievement_id`       | UUID      | NOT NULL, FK → Achievement       | Badge reference                      |
| `earned_at`            | TIMESTAMP | NOT NULL                         | When badge was earned                |
| `awarded_by_user_id`   | UUID      | NULL, FK → User                  | User who awarded (for custom badges) |
| `created_at`           | TIMESTAMP | NOT NULL                         | Record creation timestamp            |

**Relationships**:

- N:1 with `Volunteer_Profile`
- N:1 with `Achievement`
- N:1 with `User` (awarded_by, optional)

**Indexes**:

- `volunteer_profile_id, achievement_id` (composite unique, prevent duplicate badges)
- `volunteer_profile_id` (for volunteer's badges)
- `achievement_id` (for badge holders)

**Business Rules**:

- Badge cannot be earned twice by same volunteer (unique constraint)
- Congratulatory notification sent when badge earned (FR-076)

---

### 13. Document

**Purpose**: Files uploaded by organizations for volunteer requirements (waivers, policies).

**Attributes**:

| Field                 | Type         | Constraints                 | Description                                    |
| --------------------- | ------------ | --------------------------- | ---------------------------------------------- |
| `id`                  | UUID         | PRIMARY KEY                 | Unique identifier                              |
| `organization_id`     | UUID         | NOT NULL, FK → Organization | Organization reference                         |
| `filename`            | VARCHAR(255) | NOT NULL                    | Original filename                              |
| `file_url`            | VARCHAR(500) | NOT NULL                    | Stored file URL/path                           |
| `file_size`           | BIGINT       | NOT NULL                    | File size in bytes                             |
| `mime_type`           | VARCHAR(100) | NOT NULL                    | File MIME type                                 |
| `document_type`       | ENUM         | NOT NULL                    | waiver, policy, certification, guidelines      |
| `version`             | INT          | NOT NULL, DEFAULT 1         | Document version number                        |
| `is_required`         | BOOLEAN      | NOT NULL, DEFAULT false     | Whether document is required for opportunities |
| `uploaded_by_user_id` | UUID         | NOT NULL, FK → User         | User who uploaded document                     |
| `uploaded_at`         | TIMESTAMP    | NOT NULL                    | Upload timestamp                               |
| `created_at`          | TIMESTAMP    | NOT NULL                    | Record creation timestamp                      |
| `updated_at`          | TIMESTAMP    | NOT NULL                    | Last update timestamp                          |
| `deleted_at`          | TIMESTAMP    | NULL                        | Soft delete (replaced by new version)          |

**Relationships**:

- N:1 with `Organization`
- N:1 with `User` (uploader)
- N:M with `Opportunity` (via `Opportunity_Document` join table - which opportunities require doc)
- 1:N with `Document_Acknowledgment` (which volunteers acknowledged)

**Indexes**:

- `organization_id` (for org's documents)
- `document_type` (for filtering by type)

**Business Rules**:

- Documents support versioning (FR-087)
- When new version uploaded, re-acknowledgment required (FR-087)
- Required documents block registration if not acknowledged (FR-086)

---

### 14. Document_Acknowledgment

**Purpose**: Tracks which volunteers have signed/acknowledged required documents.

**Attributes**:

| Field                  | Type        | Constraints                      | Description                          |
| ---------------------- | ----------- | -------------------------------- | ------------------------------------ |
| `id`                   | UUID        | PRIMARY KEY                      | Unique identifier                    |
| `document_id`          | UUID        | NOT NULL, FK → Document          | Document reference                   |
| `volunteer_profile_id` | UUID        | NOT NULL, FK → Volunteer_Profile | Volunteer reference                  |
| `acknowledged_at`      | TIMESTAMP   | NOT NULL                         | Acknowledgment timestamp             |
| `ip_address`           | VARCHAR(45) | NULL                             | IP address of acknowledgment (audit) |
| `created_at`           | TIMESTAMP   | NOT NULL                         | Record creation timestamp            |

**Relationships**:

- N:1 with `Document`
- N:1 with `Volunteer_Profile`

**Indexes**:

- `document_id, volunteer_profile_id` (composite unique, one acknowledgment per volunteer per version)
- `volunteer_profile_id` (for volunteer's acknowledged docs)

**Business Rules**:

- Acknowledgment required before registering for opportunities requiring document (FR-086)
- New document version requires re-acknowledgment (FR-087)
- IP address logged for audit trail

---

### 15. Team

**Purpose**: Volunteer groups created for coordinated participation.

**Attributes**:

| Field            | Type          | Constraints                      | Description                       |
| ---------------- | ------------- | -------------------------------- | --------------------------------- |
| `id`             | UUID          | PRIMARY KEY                      | Unique identifier                 |
| `name`           | VARCHAR(200)  | NOT NULL                         | Team name                         |
| `description`    | TEXT          | NULL                             | Team description                  |
| `team_leader_id` | UUID          | NOT NULL, FK → Volunteer_Profile | Team leader                       |
| `total_hours`    | DECIMAL(10,2) | NOT NULL, DEFAULT 0              | Cached sum of team members' hours |
| `total_events`   | INT           | NOT NULL, DEFAULT 0              | Cached count of team events       |
| `created_at`     | TIMESTAMP     | NOT NULL                         | Record creation timestamp         |
| `updated_at`     | TIMESTAMP     | NOT NULL                         | Last update timestamp             |
| `deleted_at`     | TIMESTAMP     | NULL                             | Soft delete timestamp             |

**Relationships**:

- N:1 with `Volunteer_Profile` (team leader)
- 1:N with `Team_Member` (team members)

**Indexes**:

- `team_leader_id` (for leader's teams)
- `name` (for search)

**Business Rules**:

- Team leader can invite members (FR-088)
- Team leader can register entire team for opportunities (FR-089)
- Team analytics track collective impact (FR-090)

---

### 16. Team_Member

**Purpose**: Junction table linking volunteers to teams.

**Attributes**:

| Field                  | Type      | Constraints                      | Description               |
| ---------------------- | --------- | -------------------------------- | ------------------------- |
| `id`                   | UUID      | PRIMARY KEY                      | Unique identifier         |
| `team_id`              | UUID      | NOT NULL, FK → Team              | Team reference            |
| `volunteer_profile_id` | UUID      | NOT NULL, FK → Volunteer_Profile | Volunteer reference       |
| `joined_at`            | TIMESTAMP | NOT NULL                         | When member joined team   |
| `created_at`           | TIMESTAMP | NOT NULL                         | Record creation timestamp |
| `deleted_at`           | TIMESTAMP | NULL                             | Soft delete (left team)   |

**Relationships**:

- N:1 with `Team`
- N:1 with `Volunteer_Profile`

**Indexes**:

- `team_id, volunteer_profile_id` (composite unique, prevent duplicate membership)
- `volunteer_profile_id` (for volunteer's teams)

**Business Rules**:

- Volunteer can be member of multiple teams
- Team leader automatically member of their team

---

### 17. Saved_Search

**Purpose**: User-saved search criteria for opportunity alerts.

**Attributes**:

| Field                  | Type         | Constraints                      | Description                                       |
| ---------------------- | ------------ | -------------------------------- | ------------------------------------------------- |
| `id`                   | UUID         | PRIMARY KEY                      | Unique identifier                                 |
| `volunteer_profile_id` | UUID         | NOT NULL, FK → Volunteer_Profile | Volunteer reference                               |
| `name`                 | VARCHAR(100) | NOT NULL                         | Search name (user-defined)                        |
| `search_filters`       | JSONB        | NOT NULL                         | Search criteria (location, causes, skills, dates) |
| `alert_frequency`      | ENUM         | NOT NULL                         | immediate, daily, weekly                          |
| `last_notification_at` | TIMESTAMP    | NULL                             | Last time alert was sent                          |
| `is_active`            | BOOLEAN      | NOT NULL, DEFAULT true           | Whether alerts are enabled                        |
| `created_at`           | TIMESTAMP    | NOT NULL                         | Record creation timestamp                         |
| `updated_at`           | TIMESTAMP    | NOT NULL                         | Last update timestamp                             |
| `deleted_at`           | TIMESTAMP    | NULL                             | Soft delete timestamp                             |

**Relationships**:

- N:1 with `Volunteer_Profile`

**Indexes**:

- `volunteer_profile_id` (for volunteer's saved searches)
- `is_active` (for active searches to check for new matches)

**Business Rules**:

- When new opportunity matches filters, send alert notification (FR-038)
- Alert frequency controls notification cadence
- Search filters stored as JSON for flexibility (location radius, causes, skills, date ranges)

---

## Supporting Entities (Junction Tables & Lookups)

### 18. Cause_Category

**Purpose**: Lookup table for cause areas (environment, education, healthcare, etc.).

**Attributes**:

| Field         | Type         | Constraints      | Description               |
| ------------- | ------------ | ---------------- | ------------------------- |
| `id`          | UUID         | PRIMARY KEY      | Unique identifier         |
| `name`        | VARCHAR(100) | UNIQUE, NOT NULL | Cause name                |
| `slug`        | VARCHAR(100) | UNIQUE, NOT NULL | URL-friendly name         |
| `description` | TEXT         | NULL             | Cause description         |
| `icon_name`   | VARCHAR(50)  | NULL             | Icon identifier           |
| `created_at`  | TIMESTAMP    | NOT NULL         | Record creation timestamp |

**Relationships**:

- N:M with `Organization` (via `Organization_Cause`)
- N:M with `Opportunity` (via `Opportunity_Cause`)
- N:M with `Volunteer_Profile` (via `Volunteer_Interest`)

**Examples**: Environment, Education, Healthcare, Animal Welfare, Arts & Culture, Community Development, etc.

---

### 19. Skill

**Purpose**: Lookup table for volunteer skills (coding, teaching, medical, etc.).

**Attributes**:

| Field        | Type         | Constraints      | Description                                          |
| ------------ | ------------ | ---------------- | ---------------------------------------------------- |
| `id`         | UUID         | PRIMARY KEY      | Unique identifier                                    |
| `name`       | VARCHAR(100) | UNIQUE, NOT NULL | Skill name                                           |
| `slug`       | VARCHAR(100) | UNIQUE, NOT NULL | URL-friendly name                                    |
| `category`   | VARCHAR(50)  | NULL             | Skill category (technical, medical, education, etc.) |
| `created_at` | TIMESTAMP    | NOT NULL         | Record creation timestamp                            |

**Relationships**:

- N:M with `Volunteer_Profile` (via `Volunteer_Skill`)
- N:M with `Opportunity` (via `Opportunity_Skill` - required skills)

**Examples**: Web Development, Graphic Design, Teaching, Medical/Nursing, Event Planning, Marketing, etc.

---

### 20. Volunteer_Skill

**Purpose**: Junction table linking volunteers to their skills.

**Attributes**:

| Field                  | Type      | Constraints                      | Description               |
| ---------------------- | --------- | -------------------------------- | ------------------------- |
| `volunteer_profile_id` | UUID      | NOT NULL, FK → Volunteer_Profile | Volunteer reference       |
| `skill_id`             | UUID      | NOT NULL, FK → Skill             | Skill reference           |
| `created_at`           | TIMESTAMP | NOT NULL                         | Record creation timestamp |

**Primary Key**: Composite (`volunteer_profile_id`, `skill_id`)

---

### 21. Volunteer_Interest

**Purpose**: Junction table linking volunteers to their cause interests.

**Attributes**:

| Field                  | Type      | Constraints                      | Description               |
| ---------------------- | --------- | -------------------------------- | ------------------------- |
| `volunteer_profile_id` | UUID      | NOT NULL, FK → Volunteer_Profile | Volunteer reference       |
| `cause_category_id`    | UUID      | NOT NULL, FK → Cause_Category    | Cause reference           |
| `created_at`           | TIMESTAMP | NOT NULL                         | Record creation timestamp |

**Primary Key**: Composite (`volunteer_profile_id`, `cause_category_id`)

---

### 22. Organization_Cause

**Purpose**: Junction table linking organizations to their cause areas.

**Attributes**:

| Field               | Type      | Constraints                   | Description               |
| ------------------- | --------- | ----------------------------- | ------------------------- |
| `organization_id`   | UUID      | NOT NULL, FK → Organization   | Organization reference    |
| `cause_category_id` | UUID      | NOT NULL, FK → Cause_Category | Cause reference           |
| `created_at`        | TIMESTAMP | NOT NULL                      | Record creation timestamp |

**Primary Key**: Composite (`organization_id`, `cause_category_id`)

---

### 23. Opportunity_Cause

**Purpose**: Junction table linking opportunities to their cause areas.

**Attributes**:

| Field               | Type      | Constraints                   | Description               |
| ------------------- | --------- | ----------------------------- | ------------------------- |
| `opportunity_id`    | UUID      | NOT NULL, FK → Opportunity    | Opportunity reference     |
| `cause_category_id` | UUID      | NOT NULL, FK → Cause_Category | Cause reference           |
| `created_at`        | TIMESTAMP | NOT NULL                      | Record creation timestamp |

**Primary Key**: Composite (`opportunity_id`, `cause_category_id`)

---

### 24. Opportunity_Skill

**Purpose**: Junction table linking opportunities to required skills.

**Attributes**:

| Field            | Type      | Constraints                | Description                            |
| ---------------- | --------- | -------------------------- | -------------------------------------- |
| `opportunity_id` | UUID      | NOT NULL, FK → Opportunity | Opportunity reference                  |
| `skill_id`       | UUID      | NOT NULL, FK → Skill       | Skill reference                        |
| `is_required`    | BOOLEAN   | NOT NULL, DEFAULT true     | Whether skill is required or preferred |
| `created_at`     | TIMESTAMP | NOT NULL                   | Record creation timestamp              |

**Primary Key**: Composite (`opportunity_id`, `skill_id`)

---

### 25. Opportunity_Document

**Purpose**: Junction table linking opportunities to required documents.

**Attributes**:

| Field            | Type      | Constraints                | Description                                   |
| ---------------- | --------- | -------------------------- | --------------------------------------------- |
| `opportunity_id` | UUID      | NOT NULL, FK → Opportunity | Opportunity reference                         |
| `document_id`    | UUID      | NOT NULL, FK → Document    | Document reference                            |
| `is_required`    | BOOLEAN   | NOT NULL, DEFAULT true     | Whether document is required for registration |
| `created_at`     | TIMESTAMP | NOT NULL                   | Record creation timestamp                     |

**Primary Key**: Composite (`opportunity_id`, `document_id`)

---

### 26. Volunteer_Favorite_Organization

**Purpose**: Junction table for volunteers' favorite organizations.

**Attributes**:

| Field                  | Type      | Constraints                      | Description                     |
| ---------------------- | --------- | -------------------------------- | ------------------------------- |
| `volunteer_profile_id` | UUID      | NOT NULL, FK → Volunteer_Profile | Volunteer reference             |
| `organization_id`      | UUID      | NOT NULL, FK → Organization      | Organization reference          |
| `favorited_at`         | TIMESTAMP | NOT NULL                         | When organization was favorited |
| `created_at`           | TIMESTAMP | NOT NULL                         | Record creation timestamp       |

**Primary Key**: Composite (`volunteer_profile_id`, `organization_id`)

---

## Database Migrations Strategy

**Migration Management**: golang-migrate or GORM AutoMigrate with version control

**Migration Files**:

- `001_create_users_table.up.sql`
- `002_create_organizations_table.up.sql`
- `003_create_volunteer_profiles_table.up.sql`
- ... (one migration per entity or logical grouping)

**Best Practices**:

- One migration per schema change
- Always include rollback (`.down.sql` files)
- Test migrations on dev/staging before production
- Never modify committed migrations (create new migration instead)
- Include seed data for lookup tables (Cause_Category, Skill, Achievement)

---

## Performance Considerations

### Indexes Summary

**High-Priority Indexes** (for frequent queries):

- `User.email` (unique, login)
- `Organization.slug` (unique, public URLs)
- `Organization.city, state` (composite, location search)
- `Opportunity.start_date, start_time` (composite, date range queries)
- `Opportunity.latitude, longitude` (geospatial queries)
- `Registration.opportunity_id, volunteer_profile_id` (composite unique)
- `Notification.recipient_id, read_at` (composite, unread notifications)
- All foreign keys (join performance)

### Query Optimization

- Use GORM Preload for eager loading (prevent N+1)
- Paginate large result sets (20-50 items per page)
- Use database-level full-text search for opportunity search (PostgreSQL `tsvector`)
- Cache frequently accessed data in Redis (organization profiles, opportunity listings)
- Use database views for complex analytics queries
- Implement read replicas for heavy read workloads (future)

---

## Data Model Completeness Checklist

- [x] All 12 primary entities from spec defined
- [x] All relationships documented
- [x] Foreign key constraints specified
- [x] Indexes on all foreign keys
- [x] Indexes on frequently queried fields
- [x] Soft delete support (`deleted_at`)
- [x] Audit timestamps (`created_at`, `updated_at`)
- [x] Junction tables for many-to-many relationships
- [x] Enum types for status fields
- [x] Business rules documented
- [x] Performance considerations addressed

**Status**: Data model complete and ready for contract generation ✅

---

**Next Phase**: Generate API contracts (OpenAPI specifications) based on functional requirements and data model.
