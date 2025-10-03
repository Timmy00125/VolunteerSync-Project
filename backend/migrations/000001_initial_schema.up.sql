-- VolunteerSync Initial Schema Migration (UP)
-- Created: 2025-10-03
-- Task: T046
-- Description: Complete database schema for VolunteerSync platform
-- PostgreSQL 16 compatible

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- LOOKUP TABLES (No dependencies)
-- ============================================================================

-- Cause Categories lookup table
CREATE TABLE cause_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    icon_name VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cause_categories_slug ON cause_categories(slug);

-- Skills lookup table
CREATE TABLE skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_skills_slug ON skills(slug);
CREATE INDEX idx_skills_category ON skills(category);

-- ============================================================================
-- CORE USER ENTITIES
-- ============================================================================

-- Users table with authentication and security questions
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    account_status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (account_status IN ('active', 'inactive', 'suspended')),
    last_login_at TIMESTAMP,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    security_question_1 VARCHAR(255) NOT NULL,
    security_answer_1_hash VARCHAR(255) NOT NULL,
    security_question_2 VARCHAR(255) NOT NULL,
    security_answer_2_hash VARCHAR(255) NOT NULL,
    security_question_3 VARCHAR(255) NOT NULL,
    security_answer_3_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_account_status ON users(account_status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Volunteer profiles (extended data for volunteer users)
CREATE TABLE volunteer_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_photo_url VARCHAR(500),
    biography TEXT,
    location VARCHAR(255),
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    availability_monday BOOLEAN NOT NULL DEFAULT false,
    availability_tuesday BOOLEAN NOT NULL DEFAULT false,
    availability_wednesday BOOLEAN NOT NULL DEFAULT false,
    availability_thursday BOOLEAN NOT NULL DEFAULT false,
    availability_friday BOOLEAN NOT NULL DEFAULT false,
    availability_saturday BOOLEAN NOT NULL DEFAULT false,
    availability_sunday BOOLEAN NOT NULL DEFAULT false,
    preferred_time VARCHAR(20) CHECK (preferred_time IN ('morning', 'afternoon', 'evening', 'flexible')),
    total_hours DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_events INT NOT NULL DEFAULT 0,
    emergency_contact_name VARCHAR(200),
    emergency_contact_phone VARCHAR(20),
    privacy_show_hours BOOLEAN NOT NULL DEFAULT true,
    privacy_show_events BOOLEAN NOT NULL DEFAULT true,
    privacy_show_organizations BOOLEAN NOT NULL DEFAULT true,
    notification_in_app BOOLEAN NOT NULL DEFAULT true,
    notification_browser_push BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_volunteer_profiles_user_id ON volunteer_profiles(user_id);
CREATE INDEX idx_volunteer_profiles_location ON volunteer_profiles(location);
CREATE INDEX idx_volunteer_profiles_lat_lng ON volunteer_profiles(latitude, longitude);

-- ============================================================================
-- ORGANIZATION ENTITIES
-- ============================================================================

-- Organizations table
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    mission_statement TEXT,
    description TEXT,
    website VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address_line_1 VARCHAR(255),
    address_line_2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'United States',
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    logo_url VARCHAR(500),
    banner_url VARCHAR(500),
    verification_status VARCHAR(20) NOT NULL DEFAULT 'verified' CHECK (verification_status IN ('verified', 'unverified')),
    verified_at TIMESTAMP,
    total_volunteers INT NOT NULL DEFAULT 0,
    total_hours DECIMAL(10,2) NOT NULL DEFAULT 0,
    avg_rating DECIMAL(3,2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_name ON organizations(name);
CREATE INDEX idx_organizations_city_state ON organizations(city, state);
CREATE INDEX idx_organizations_verification_status ON organizations(verification_status);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at);
CREATE INDEX idx_organizations_lat_lng ON organizations(latitude, longitude);

-- Organization members (team members with roles)
CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'coordinator')),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    invited_at TIMESTAMP,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_organization_members_org_id ON organization_members(organization_id);
CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX idx_organization_members_deleted_at ON organization_members(deleted_at);

-- ============================================================================
-- ACHIEVEMENT / GAMIFICATION
-- ============================================================================

-- Achievements (badges)
CREATE TABLE achievements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    icon_url VARCHAR(500),
    badge_type VARCHAR(20) NOT NULL CHECK (badge_type IN ('system', 'organization_custom')),
    criteria_type VARCHAR(30) CHECK (criteria_type IN ('hours_milestone', 'events_milestone', 'consistency', 'custom')),
    criteria_value INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_achievements_org_id ON achievements(organization_id);
CREATE INDEX idx_achievements_badge_type ON achievements(badge_type);

-- Volunteer achievements (junction table with earned timestamp)
CREATE TABLE volunteer_achievements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    awarded_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(volunteer_profile_id, achievement_id)
);

CREATE INDEX idx_volunteer_achievements_volunteer_id ON volunteer_achievements(volunteer_profile_id);
CREATE INDEX idx_volunteer_achievements_achievement_id ON volunteer_achievements(achievement_id);

-- ============================================================================
-- DOCUMENT MANAGEMENT
-- ============================================================================

-- Documents (waivers, policies, etc.)
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    document_type VARCHAR(30) NOT NULL CHECK (document_type IN ('waiver', 'policy', 'certification', 'guidelines')),
    version INT NOT NULL DEFAULT 1,
    is_required BOOLEAN NOT NULL DEFAULT false,
    uploaded_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_documents_org_id ON documents(organization_id);
CREATE INDEX idx_documents_document_type ON documents(document_type);
CREATE INDEX idx_documents_uploaded_by ON documents(uploaded_by_user_id);

-- Document acknowledgments
CREATE TABLE document_acknowledgments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    acknowledged_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, volunteer_profile_id)
);

CREATE INDEX idx_document_acknowledgments_document_id ON document_acknowledgments(document_id);
CREATE INDEX idx_document_acknowledgments_volunteer_id ON document_acknowledgments(volunteer_profile_id);

-- ============================================================================
-- TEAM MANAGEMENT
-- ============================================================================

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    team_leader_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    total_hours DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_events INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_teams_team_leader_id ON teams(team_leader_id);
CREATE INDEX idx_teams_name ON teams(name);

-- Team members
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(team_id, volunteer_profile_id)
);

CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_volunteer_id ON team_members(volunteer_profile_id);

-- ============================================================================
-- OPPORTUNITIES
-- ============================================================================

-- Opportunities (volunteer events/shifts)
CREATE TABLE opportunities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'cancelled', 'completed')),
    start_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_date DATE NOT NULL,
    end_time TIME NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    address_line_1 VARCHAR(255) NOT NULL,
    address_line_2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(50) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL DEFAULT 'United States',
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    capacity INT NOT NULL,
    current_registrations INT NOT NULL DEFAULT 0,
    min_age INT,
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    recurrence_pattern VARCHAR(50) CHECK (recurrence_pattern IN ('daily', 'weekly', 'monthly', 'custom')),
    recurrence_end_date DATE,
    parent_opportunity_id UUID REFERENCES opportunities(id) ON DELETE CASCADE,
    published_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    completed_at TIMESTAMP,
    auto_complete_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_opportunities_org_id ON opportunities(organization_id);
CREATE INDEX idx_opportunities_created_by ON opportunities(created_by_user_id);
CREATE INDEX idx_opportunities_status ON opportunities(status);
CREATE INDEX idx_opportunities_start_date_time ON opportunities(start_date, start_time);
CREATE INDEX idx_opportunities_city_state ON opportunities(city, state);
CREATE INDEX idx_opportunities_lat_lng ON opportunities(latitude, longitude);
CREATE INDEX idx_opportunities_created_at ON opportunities(created_at);
CREATE INDEX idx_opportunities_parent_id ON opportunities(parent_opportunity_id);

-- ============================================================================
-- REGISTRATIONS
-- ============================================================================

-- Registrations (volunteer sign-ups for opportunities)
CREATE TABLE registrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled', 'waitlisted', 'completed')),
    registered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checked_in_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    hours_worked DECIMAL(5,2),
    hours_status VARCHAR(20) CHECK (hours_status IN ('pending', 'verified', 'disputed')),
    hours_logged_at TIMESTAMP,
    hours_verified_at TIMESTAMP,
    volunteer_rating INT CHECK (volunteer_rating >= 1 AND volunteer_rating <= 5),
    volunteer_review TEXT,
    review_submitted_at TIMESTAMP,
    coordinator_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(opportunity_id, volunteer_profile_id)
);

CREATE INDEX idx_registrations_opportunity_id ON registrations(opportunity_id);
CREATE INDEX idx_registrations_volunteer_id ON registrations(volunteer_profile_id);
CREATE INDEX idx_registrations_status ON registrations(status);
CREATE INDEX idx_registrations_registered_at ON registrations(registered_at);
CREATE INDEX idx_registrations_hours_status ON registrations(hours_status);

-- ============================================================================
-- HOURS TRACKING (AUDIT LOG)
-- ============================================================================

-- Hours log (immutable audit trail)
CREATE TABLE hours_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_id UUID NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
    hours DECIMAL(5,2) NOT NULL CHECK (hours > 0),
    logged_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'verified', 'disputed')),
    coordinator_notes TEXT,
    volunteer_notes TEXT,
    dispute_reason TEXT,
    disputed_at TIMESTAMP,
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    logged_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP,
    auto_verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_hours_logs_registration_id ON hours_logs(registration_id);
CREATE INDEX idx_hours_logs_logged_by ON hours_logs(logged_by_user_id);
CREATE INDEX idx_hours_logs_status ON hours_logs(status);
CREATE INDEX idx_hours_logs_logged_at ON hours_logs(logged_at);

-- ============================================================================
-- COMMUNICATIONS
-- ============================================================================

-- Messages (direct and broadcast)
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id UUID REFERENCES opportunities(id) ON DELETE CASCADE,
    message_type VARCHAR(20) NOT NULL CHECK (message_type IN ('direct', 'broadcast')),
    subject VARCHAR(255),
    content TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_opportunity_id ON messages(opportunity_id);
CREATE INDEX idx_messages_sent_at ON messages(sent_at);
CREATE INDEX idx_messages_message_type ON messages(message_type);

-- Message recipients
CREATE TABLE message_recipients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_message_recipients_message_id ON message_recipients(message_id);
CREATE INDEX idx_message_recipients_recipient_id ON message_recipients(recipient_id);
CREATE INDEX idx_message_recipients_read_at ON message_recipients(read_at);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    action_url VARCHAR(500),
    priority VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'critical')),
    related_entity_type VARCHAR(50),
    related_entity_id UUID,
    read_at TIMESTAMP,
    delivered_at TIMESTAMP,
    delivery_method VARCHAR(20) NOT NULL DEFAULT 'in_app' CHECK (delivery_method IN ('in_app', 'browser_push')),
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX idx_notifications_recipient_read ON notifications(recipient_id, read_at);
CREATE INDEX idx_notifications_sent_at ON notifications(sent_at);
CREATE INDEX idx_notifications_notification_type ON notifications(notification_type);

-- ============================================================================
-- SAVED SEARCHES
-- ============================================================================

-- Saved searches (volunteer search alerts)
CREATE TABLE saved_searches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    search_filters JSONB NOT NULL,
    alert_frequency VARCHAR(20) NOT NULL CHECK (alert_frequency IN ('immediate', 'daily', 'weekly')),
    last_notification_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_saved_searches_volunteer_id ON saved_searches(volunteer_profile_id);
CREATE INDEX idx_saved_searches_is_active ON saved_searches(is_active);

-- ============================================================================
-- JUNCTION TABLES (Many-to-Many Relationships)
-- ============================================================================

-- Volunteer skills
CREATE TABLE volunteer_skills (
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (volunteer_profile_id, skill_id)
);

CREATE INDEX idx_volunteer_skills_volunteer_id ON volunteer_skills(volunteer_profile_id);
CREATE INDEX idx_volunteer_skills_skill_id ON volunteer_skills(skill_id);

-- Volunteer interests (causes)
CREATE TABLE volunteer_interests (
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    cause_category_id UUID NOT NULL REFERENCES cause_categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (volunteer_profile_id, cause_category_id)
);

CREATE INDEX idx_volunteer_interests_volunteer_id ON volunteer_interests(volunteer_profile_id);
CREATE INDEX idx_volunteer_interests_cause_id ON volunteer_interests(cause_category_id);

-- Organization causes
CREATE TABLE organization_causes (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cause_category_id UUID NOT NULL REFERENCES cause_categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (organization_id, cause_category_id)
);

CREATE INDEX idx_organization_causes_org_id ON organization_causes(organization_id);
CREATE INDEX idx_organization_causes_cause_id ON organization_causes(cause_category_id);

-- Opportunity causes
CREATE TABLE opportunity_causes (
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    cause_category_id UUID NOT NULL REFERENCES cause_categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (opportunity_id, cause_category_id)
);

CREATE INDEX idx_opportunity_causes_opportunity_id ON opportunity_causes(opportunity_id);
CREATE INDEX idx_opportunity_causes_cause_id ON opportunity_causes(cause_category_id);

-- Opportunity skills (required skills)
CREATE TABLE opportunity_skills (
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    is_required BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (opportunity_id, skill_id)
);

CREATE INDEX idx_opportunity_skills_opportunity_id ON opportunity_skills(opportunity_id);
CREATE INDEX idx_opportunity_skills_skill_id ON opportunity_skills(skill_id);

-- Opportunity documents (required documents)
CREATE TABLE opportunity_documents (
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    is_required BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (opportunity_id, document_id)
);

CREATE INDEX idx_opportunity_documents_opportunity_id ON opportunity_documents(opportunity_id);
CREATE INDEX idx_opportunity_documents_document_id ON opportunity_documents(document_id);

-- Volunteer favorite organizations
CREATE TABLE volunteer_favorite_organizations (
    volunteer_profile_id UUID NOT NULL REFERENCES volunteer_profiles(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    favorited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (volunteer_profile_id, organization_id)
);

CREATE INDEX idx_volunteer_favorites_volunteer_id ON volunteer_favorite_organizations(volunteer_profile_id);
CREATE INDEX idx_volunteer_favorites_org_id ON volunteer_favorite_organizations(organization_id);

-- ============================================================================
-- TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to all tables with updated_at column
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_volunteer_profiles_updated_at BEFORE UPDATE ON volunteer_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_members_updated_at BEFORE UPDATE ON organization_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_achievements_updated_at BEFORE UPDATE ON achievements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_opportunities_updated_at BEFORE UPDATE ON opportunities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_registrations_updated_at BEFORE UPDATE ON registrations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hours_logs_updated_at BEFORE UPDATE ON hours_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_saved_searches_updated_at BEFORE UPDATE ON saved_searches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- INITIAL SEED DATA (Lookup Tables)
-- ============================================================================

-- Seed cause categories
INSERT INTO cause_categories (name, slug, description, icon_name) VALUES
    ('Environment', 'environment', 'Environmental conservation and sustainability', 'leaf'),
    ('Education', 'education', 'Teaching, tutoring, and educational programs', 'book'),
    ('Healthcare', 'healthcare', 'Medical services and health awareness', 'heart'),
    ('Animal Welfare', 'animal-welfare', 'Animal rescue and care', 'paw'),
    ('Arts & Culture', 'arts-culture', 'Arts, music, and cultural programs', 'palette'),
    ('Community Development', 'community-development', 'Community building and social programs', 'users'),
    ('Hunger & Homelessness', 'hunger-homelessness', 'Food banks and homeless services', 'home'),
    ('Youth & Children', 'youth-children', 'Programs for youth and children', 'child'),
    ('Seniors', 'seniors', 'Services for elderly and seniors', 'user-plus'),
    ('Disaster Relief', 'disaster-relief', 'Emergency response and disaster recovery', 'alert-triangle')
ON CONFLICT (slug) DO NOTHING;

-- Seed common skills
INSERT INTO skills (name, slug, category) VALUES
    ('Web Development', 'web-development', 'technical'),
    ('Graphic Design', 'graphic-design', 'creative'),
    ('Teaching', 'teaching', 'education'),
    ('Medical/Nursing', 'medical-nursing', 'medical'),
    ('Event Planning', 'event-planning', 'organizational'),
    ('Marketing', 'marketing', 'business'),
    ('Photography', 'photography', 'creative'),
    ('Writing', 'writing', 'communication'),
    ('Public Speaking', 'public-speaking', 'communication'),
    ('Fundraising', 'fundraising', 'business'),
    ('Social Media', 'social-media', 'communication'),
    ('Translation', 'translation', 'communication'),
    ('Construction', 'construction', 'manual'),
    ('Cooking', 'cooking', 'service'),
    ('Counseling', 'counseling', 'support')
ON CONFLICT (slug) DO NOTHING;

-- Seed system achievements
INSERT INTO achievements (name, description, badge_type, criteria_type, criteria_value, icon_url) VALUES
    ('First Event', 'Completed your first volunteer event!', 'system', 'events_milestone', 1, '/badges/first-event.svg'),
    ('10 Hours', 'Contributed 10 volunteer hours', 'system', 'hours_milestone', 10, '/badges/10-hours.svg'),
    ('25 Hours', 'Contributed 25 volunteer hours', 'system', 'hours_milestone', 25, '/badges/25-hours.svg'),
    ('50 Hours', 'Contributed 50 volunteer hours', 'system', 'hours_milestone', 50, '/badges/50-hours.svg'),
    ('100 Hours', 'Contributed 100 volunteer hours!', 'system', 'hours_milestone', 100, '/badges/100-hours.svg'),
    ('10 Events', 'Completed 10 volunteer events', 'system', 'events_milestone', 10, '/badges/10-events.svg'),
    ('3 Months Consistent', 'Volunteered consistently for 3 months', 'system', 'consistency', 90, '/badges/consistent.svg')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE users IS 'Core user accounts with authentication and security questions';
COMMENT ON TABLE volunteer_profiles IS 'Extended profile data for volunteer users';
COMMENT ON TABLE organizations IS 'Nonprofit organizations managing volunteer programs';
COMMENT ON TABLE organization_members IS 'Team members with role-based access to organizations';
COMMENT ON TABLE opportunities IS 'Volunteer events and shifts';
COMMENT ON TABLE registrations IS 'Volunteer sign-ups for opportunities with hours tracking';
COMMENT ON TABLE hours_logs IS 'Immutable audit log for volunteer hours verification';
COMMENT ON TABLE achievements IS 'Recognition badges for volunteer milestones';
COMMENT ON TABLE notifications IS 'System-generated notifications for users';
COMMENT ON TABLE messages IS 'Direct and broadcast messages between users';
COMMENT ON TABLE documents IS 'Organization documents (waivers, policies) with versioning';
COMMENT ON TABLE teams IS 'Volunteer groups for coordinated participation';

-- Migration completed successfully
