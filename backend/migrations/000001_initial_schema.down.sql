-- VolunteerSync Initial Schema Migration (DOWN)
-- Created: 2025-10-03
-- Task: T047
-- Description: Rollback complete database schema for VolunteerSync platform
-- Drops all tables in reverse dependency order

-- ============================================================================
-- DROP JUNCTION TABLES (No foreign key dependencies)
-- ============================================================================

DROP TABLE IF EXISTS volunteer_favorite_organizations CASCADE;
DROP TABLE IF EXISTS opportunity_documents CASCADE;
DROP TABLE IF EXISTS opportunity_skills CASCADE;
DROP TABLE IF EXISTS opportunity_causes CASCADE;
DROP TABLE IF EXISTS organization_causes CASCADE;
DROP TABLE IF EXISTS volunteer_interests CASCADE;
DROP TABLE IF EXISTS volunteer_skills CASCADE;

-- ============================================================================
-- DROP SAVED SEARCHES (depends on volunteer_profiles)
-- ============================================================================

DROP TABLE IF EXISTS saved_searches CASCADE;

-- ============================================================================
-- DROP COMMUNICATIONS (depends on users, opportunities, messages)
-- ============================================================================

DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS message_recipients CASCADE;
DROP TABLE IF EXISTS messages CASCADE;

-- ============================================================================
-- DROP HOURS TRACKING (depends on registrations, users)
-- ============================================================================

DROP TABLE IF EXISTS hours_logs CASCADE;

-- ============================================================================
-- DROP REGISTRATIONS (depends on opportunities, volunteer_profiles)
-- ============================================================================

DROP TABLE IF EXISTS registrations CASCADE;

-- ============================================================================
-- DROP OPPORTUNITIES (depends on organizations, users)
-- ============================================================================

DROP TABLE IF EXISTS opportunities CASCADE;

-- ============================================================================
-- DROP TEAM MANAGEMENT (depends on teams, volunteer_profiles)
-- ============================================================================

DROP TABLE IF EXISTS team_members CASCADE;
DROP TABLE IF EXISTS teams CASCADE;

-- ============================================================================
-- DROP DOCUMENT MANAGEMENT (depends on documents, volunteer_profiles)
-- ============================================================================

DROP TABLE IF EXISTS document_acknowledgments CASCADE;
DROP TABLE IF EXISTS documents CASCADE;

-- ============================================================================
-- DROP ACHIEVEMENT / GAMIFICATION (depends on achievements, volunteer_profiles)
-- ============================================================================

DROP TABLE IF EXISTS volunteer_achievements CASCADE;
DROP TABLE IF EXISTS achievements CASCADE;

-- ============================================================================
-- DROP ORGANIZATION ENTITIES (depends on organizations, users)
-- ============================================================================

DROP TABLE IF EXISTS organization_members CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

-- ============================================================================
-- DROP CORE USER ENTITIES (depends on users)
-- ============================================================================

DROP TABLE IF EXISTS volunteer_profiles CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- ============================================================================
-- DROP LOOKUP TABLES (no dependencies)
-- ============================================================================

DROP TABLE IF EXISTS skills CASCADE;
DROP TABLE IF EXISTS cause_categories CASCADE;

-- ============================================================================
-- DROP FUNCTIONS
-- ============================================================================

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- ============================================================================
-- DROP EXTENSIONS
-- ============================================================================

DROP EXTENSION IF EXISTS "uuid-ossp";

-- Migration rollback completed successfully
