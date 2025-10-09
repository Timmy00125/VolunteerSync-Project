package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/pdf"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for service operations
var (
	// ErrVolunteerNotFound is returned when a volunteer profile cannot be found
	ErrVolunteerNotFound = errors.New("volunteer profile not found")
	// ErrOrganizationNotFound is returned when an organization cannot be found
	ErrOrganizationNotFound = errors.New("organization not found")
	// ErrUnauthorized is returned when a user doesn't have permission for an operation
	ErrUnauthorized = errors.New("unauthorized to access analytics")
	// ErrInvalidDateRange is returned when date range parameters are invalid
	ErrInvalidDateRange = errors.New("invalid date range")
)

// AnalyticsService encapsulates analytics and reporting business logic
// Provides methods for volunteer, organization, and platform analytics
type AnalyticsService interface {
	// GetVolunteerAnalytics retrieves analytics data for a volunteer
	// Includes hours over time, events by cause category (FR-078)
	GetVolunteerAnalytics(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange) (*VolunteerAnalytics, error)

	// GetOrganizationAnalytics retrieves analytics data for an organization
	// Includes volunteers recruited, hours contributed, retention rate (FR-077)
	GetOrganizationAnalytics(ctx context.Context, organizationID uuid.UUID, dateRange DateRange) (*OrganizationAnalytics, error)

	// GetPlatformAnalytics retrieves platform-wide analytics (admin only)
	// Includes total volunteers, orgs, hours, growth trends (FR-080)
	GetPlatformAnalytics(ctx context.Context, dateRange DateRange) (*PlatformAnalytics, error)

	// GenerateReport generates a PDF report for a volunteer or organization
	// Returns the PDF file content as bytes (FR-079)
	GenerateReport(ctx context.Context, reportType string, entityID uuid.UUID, dateRange DateRange) ([]byte, error)
}

// DateRange represents a date range for analytics queries
type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

// Validate checks if the date range is valid
func (dr DateRange) Validate() error {
	if dr.StartDate.IsZero() || dr.EndDate.IsZero() {
		return fmt.Errorf("%w: start and end dates are required", ErrInvalidDateRange)
	}
	if dr.EndDate.Before(dr.StartDate) {
		return fmt.Errorf("%w: end date must be after start date", ErrInvalidDateRange)
	}
	return nil
}

// VolunteerAnalytics represents analytics data for a volunteer
type VolunteerAnalytics struct {
	VolunteerProfileID   uuid.UUID          `json:"volunteer_profile_id"`
	TotalHours           float64            `json:"total_hours"`
	TotalEvents          int                `json:"total_events"`
	TotalOrganizations   int                `json:"total_organizations"`
	AverageHoursPerEvent float64            `json:"average_hours_per_event"`
	HoursOverTime        []DataPoint        `json:"hours_over_time"`
	EventsByCause        []CategoryCount    `json:"events_by_cause"`
	HoursByCause         []CategoryCount    `json:"hours_by_cause"`
	OrganizationStats    []OrganizationStat `json:"organization_stats"`
	MonthlyTrend         []MonthlyDataPoint `json:"monthly_trend"`
}

// OrganizationAnalytics represents analytics data for an organization
type OrganizationAnalytics struct {
	OrganizationID     uuid.UUID          `json:"organization_id"`
	TotalVolunteers    int                `json:"total_volunteers"`
	ActiveVolunteers   int                `json:"active_volunteers"` // Volunteered in last 90 days
	TotalHours         float64            `json:"total_hours"`
	TotalOpportunities int                `json:"total_opportunities"`
	CompletedEvents    int                `json:"completed_events"`
	RetentionRate      float64            `json:"retention_rate"`    // % of volunteers who returned
	AverageFillRate    float64            `json:"average_fill_rate"` // % of spots filled
	VolunteersOverTime []DataPoint        `json:"volunteers_over_time"`
	HoursOverTime      []DataPoint        `json:"hours_over_time"`
	EventsByCause      []CategoryCount    `json:"events_by_cause"`
	TopVolunteers      []VolunteerSummary `json:"top_volunteers"` // By hours contributed
	MonthlyTrend       []MonthlyDataPoint `json:"monthly_trend"`
}

// PlatformAnalytics represents platform-wide analytics data
type PlatformAnalytics struct {
	TotalVolunteers        int                `json:"total_volunteers"`
	ActiveVolunteers       int                `json:"active_volunteers"` // Active in last 90 days
	TotalOrganizations     int                `json:"total_organizations"`
	ActiveOrganizations    int                `json:"active_organizations"` // Posted event in last 90 days
	TotalHours             float64            `json:"total_hours"`
	TotalOpportunities     int                `json:"total_opportunities"`
	CompletedOpportunities int                `json:"completed_opportunities"`
	UserGrowth             []DataPoint        `json:"user_growth"`
	OrganizationGrowth     []DataPoint        `json:"organization_growth"`
	HoursOverTime          []DataPoint        `json:"hours_over_time"`
	EventsByCause          []CategoryCount    `json:"events_by_cause"`
	GeographicDistribution []GeographicStat   `json:"geographic_distribution"`
	MonthlyTrend           []MonthlyDataPoint `json:"monthly_trend"`
}

// DataPoint represents a single data point for time-series charts
type DataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// MonthlyDataPoint represents aggregated data for a month
type MonthlyDataPoint struct {
	Month         string  `json:"month"` // YYYY-MM format
	Hours         float64 `json:"hours"`
	Events        int     `json:"events"`
	Volunteers    int     `json:"volunteers"`
	Organizations int     `json:"organizations,omitempty"` // Platform analytics only
}

// CategoryCount represents a count by category (cause category)
type CategoryCount struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Count        int       `json:"count"`
	Hours        float64   `json:"hours,omitempty"` // For hours by cause
}

// OrganizationStat represents statistics for an organization
type OrganizationStat struct {
	OrganizationID   uuid.UUID `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	EventsAttended   int       `json:"events_attended"`
	HoursContributed float64   `json:"hours_contributed"`
}

// VolunteerSummary represents a summary of a volunteer
type VolunteerSummary struct {
	VolunteerProfileID uuid.UUID `json:"volunteer_profile_id"`
	UserID             uuid.UUID `json:"user_id"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	TotalHours         float64   `json:"total_hours"`
	TotalEvents        int       `json:"total_events"`
}

// GeographicStat represents geographic distribution statistics
type GeographicStat struct {
	State         string  `json:"state"`
	City          string  `json:"city,omitempty"`
	Volunteers    int     `json:"volunteers"`
	Organizations int     `json:"organizations"`
	Hours         float64 `json:"hours"`
}

// analyticsService is the implementation of AnalyticsService
type analyticsService struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewAnalyticsService creates a new instance of AnalyticsService
func NewAnalyticsService(
	db *gorm.DB,
	logger logger.Logger,
) AnalyticsService {
	return &analyticsService{
		db:     db,
		logger: logger,
	}
}

// GetVolunteerAnalytics retrieves analytics data for a volunteer
func (s *analyticsService) GetVolunteerAnalytics(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange) (*VolunteerAnalytics, error) {
	// Validate inputs
	if volunteerProfileID == uuid.Nil {
		return nil, fmt.Errorf("volunteer profile ID is required")
	}
	if err := dateRange.Validate(); err != nil {
		return nil, err
	}

	// Check if volunteer profile exists
	var exists bool
	if err := s.db.WithContext(ctx).
		Model(&struct{ ID uuid.UUID }{}).
		Table("volunteer_profiles").
		Select("COUNT(*) > 0").
		Where("id = ? AND deleted_at IS NULL", volunteerProfileID).
		Find(&exists).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to check volunteer profile existence", err)
		return nil, fmt.Errorf("failed to check volunteer profile: %w", err)
	}
	if !exists {
		return nil, ErrVolunteerNotFound
	}

	analytics := &VolunteerAnalytics{
		VolunteerProfileID: volunteerProfileID,
	}

	// Get total hours and events
	if err := s.getVolunteerTotals(ctx, volunteerProfileID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get hours over time (daily data points)
	if err := s.getVolunteerHoursOverTime(ctx, volunteerProfileID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get events by cause category
	if err := s.getVolunteerEventsByCause(ctx, volunteerProfileID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get hours by cause category
	if err := s.getVolunteerHoursByCause(ctx, volunteerProfileID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get organization stats
	if err := s.getVolunteerOrganizationStats(ctx, volunteerProfileID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get monthly trend
	if err := s.getVolunteerMonthlyTrend(ctx, volunteerProfileID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Calculate average hours per event
	if analytics.TotalEvents > 0 {
		analytics.AverageHoursPerEvent = analytics.TotalHours / float64(analytics.TotalEvents)
	}

	return analytics, nil
}

// GetOrganizationAnalytics retrieves analytics data for an organization
func (s *analyticsService) GetOrganizationAnalytics(ctx context.Context, organizationID uuid.UUID, dateRange DateRange) (*OrganizationAnalytics, error) {
	// Validate inputs
	if organizationID == uuid.Nil {
		return nil, fmt.Errorf("organization ID is required")
	}
	if err := dateRange.Validate(); err != nil {
		return nil, err
	}

	// Check if organization exists
	var exists bool
	if err := s.db.WithContext(ctx).
		Model(&struct{ ID uuid.UUID }{}).
		Table("organizations").
		Select("COUNT(*) > 0").
		Where("id = ? AND deleted_at IS NULL", organizationID).
		Find(&exists).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to check organization existence", err)
		return nil, fmt.Errorf("failed to check organization: %w", err)
	}
	if !exists {
		return nil, ErrOrganizationNotFound
	}

	analytics := &OrganizationAnalytics{
		OrganizationID: organizationID,
	}

	// Get totals
	if err := s.getOrganizationTotals(ctx, organizationID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get volunteers over time
	if err := s.getOrganizationVolunteersOverTime(ctx, organizationID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get hours over time
	if err := s.getOrganizationHoursOverTime(ctx, organizationID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get events by cause
	if err := s.getOrganizationEventsByCause(ctx, organizationID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get top volunteers
	if err := s.getOrganizationTopVolunteers(ctx, organizationID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get monthly trend
	if err := s.getOrganizationMonthlyTrend(ctx, organizationID, dateRange, analytics); err != nil {
		return nil, err
	}

	// Calculate retention rate (volunteers who attended 2+ events)
	if err := s.calculateOrganizationRetentionRate(ctx, organizationID, dateRange, analytics); err != nil {
		return nil, err
	}

	return analytics, nil
}

// GetPlatformAnalytics retrieves platform-wide analytics
func (s *analyticsService) GetPlatformAnalytics(ctx context.Context, dateRange DateRange) (*PlatformAnalytics, error) {
	// Validate date range
	if err := dateRange.Validate(); err != nil {
		return nil, err
	}

	analytics := &PlatformAnalytics{}

	// Get totals
	if err := s.getPlatformTotals(ctx, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get user growth over time
	if err := s.getPlatformUserGrowth(ctx, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get organization growth over time
	if err := s.getPlatformOrganizationGrowth(ctx, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get hours over time
	if err := s.getPlatformHoursOverTime(ctx, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get events by cause
	if err := s.getPlatformEventsByCause(ctx, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get geographic distribution
	if err := s.getPlatformGeographicDistribution(ctx, dateRange, analytics); err != nil {
		return nil, err
	}

	// Get monthly trend
	if err := s.getPlatformMonthlyTrend(ctx, dateRange, analytics); err != nil {
		return nil, err
	}

	return analytics, nil
}

// GenerateReport generates a PDF report
func (s *analyticsService) GenerateReport(ctx context.Context, reportType string, entityID uuid.UUID, dateRange DateRange) ([]byte, error) {
	// Validate inputs
	if reportType == "" {
		return nil, fmt.Errorf("report type is required")
	}
	if entityID == uuid.Nil {
		return nil, fmt.Errorf("entity ID is required")
	}
	if err := dateRange.Validate(); err != nil {
		return nil, err
	}

	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"report_type": reportType,
		"entity_id":   entityID.String(),
		"date_range":  fmt.Sprintf("%s to %s", dateRange.StartDate.Format("2006-01-02"), dateRange.EndDate.Format("2006-01-02")),
	}).Info("Generating PDF report")

	// Generate report based on type
	switch reportType {
	case "volunteer":
		return s.generateVolunteerReport(ctx, entityID, dateRange)
	case "organization":
		return s.generateOrganizationReport(ctx, entityID, dateRange)
	case "platform":
		return s.generatePlatformReport(ctx, dateRange)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", reportType)
	}
}

// generateVolunteerReport generates a PDF report for a volunteer
func (s *analyticsService) generateVolunteerReport(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange) ([]byte, error) {
	// Get volunteer analytics data
	analytics, err := s.GetVolunteerAnalytics(ctx, volunteerProfileID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get volunteer analytics: %w", err)
	}

	// Create PDF generator
	generator := pdf.NewGenerator()
	generator.SetTitle("Volunteer Analytics Report")
	generator.SetAuthor("VolunteerSync Platform")

	// Add header
	generator.AddHeader(
		"Volunteer Analytics Report",
		fmt.Sprintf("Period: %s to %s", dateRange.StartDate.Format("Jan 2, 2006"), dateRange.EndDate.Format("Jan 2, 2006")),
	)

	generator.AddFooter(time.Now())

	// Summary Metrics
	generator.AddSectionTitle("Summary")

	metrics := []pdf.Metric{
		{Label: "Total Hours", Value: fmt.Sprintf("%.1f", analytics.TotalHours)},
		{Label: "Total Events", Value: fmt.Sprintf("%d", analytics.TotalEvents)},
		{Label: "Organizations", Value: fmt.Sprintf("%d", analytics.TotalOrganizations)},
		{Label: "Avg Hours/Event", Value: fmt.Sprintf("%.1f", analytics.AverageHoursPerEvent)},
	}

	generator.AddMetricsRow(metrics)

	// Events by Cause
	if len(analytics.EventsByCause) > 0 {
		generator.AddDivider()
		chartData := make([]pdf.ChartData, 0, len(analytics.EventsByCause))
		maxValue := 0.0

		for _, item := range analytics.EventsByCause {
			value := float64(item.Count)
			chartData = append(chartData, pdf.ChartData{
				Label: item.CategoryName,
				Value: value,
			})
			if value > maxValue {
				maxValue = value
			}
		}

		generator.AddSimpleBarChart("Events by Cause Category", chartData, maxValue)
	}

	// Hours by Cause
	if len(analytics.HoursByCause) > 0 {
		generator.AddDivider()
		chartData := make([]pdf.ChartData, 0, len(analytics.HoursByCause))
		maxValue := 0.0

		for _, item := range analytics.HoursByCause {
			chartData = append(chartData, pdf.ChartData{
				Label: item.CategoryName,
				Value: item.Hours,
			})
			if item.Hours > maxValue {
				maxValue = item.Hours
			}
		}

		generator.AddSimpleBarChart("Hours by Cause Category", chartData, maxValue)
	}

	// Organization Stats Table
	if len(analytics.OrganizationStats) > 0 {
		generator.AddDivider()
		generator.AddSectionTitle("Organization Breakdown")

		headers := []string{"Organization", "Events", "Hours"}
		colWidths := []float64{90, 40, 40}

		rows := make([][]string, 0, len(analytics.OrganizationStats))
		for _, org := range analytics.OrganizationStats {
			rows = append(rows, []string{
				org.OrganizationName,
				fmt.Sprintf("%d", org.EventsAttended),
				fmt.Sprintf("%.1f", org.HoursContributed),
			})
		}

		generator.AddTable(headers, rows, colWidths)
	}

	// Generate and return PDF
	pdfBytes, err := generator.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"volunteer_profile_id": volunteerProfileID.String(),
		"pdf_size":             len(pdfBytes),
	}).Info("Volunteer report generated successfully")

	return pdfBytes, nil
}

// generateOrganizationReport generates a PDF report for an organization
func (s *analyticsService) generateOrganizationReport(ctx context.Context, organizationID uuid.UUID, dateRange DateRange) ([]byte, error) {
	// Get organization analytics data
	analytics, err := s.GetOrganizationAnalytics(ctx, organizationID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization analytics: %w", err)
	}

	// Create PDF generator
	generator := pdf.NewGenerator()
	generator.SetTitle("Organization Analytics Report")
	generator.SetAuthor("VolunteerSync Platform")

	// Add header
	generator.AddHeader(
		"Organization Analytics Report",
		fmt.Sprintf("Period: %s to %s", dateRange.StartDate.Format("Jan 2, 2006"), dateRange.EndDate.Format("Jan 2, 2006")),
	)

	generator.AddFooter(time.Now())

	// Summary Metrics
	generator.AddSectionTitle("Summary")

	metrics := []pdf.Metric{
		{Label: "Total Volunteers", Value: fmt.Sprintf("%d", analytics.TotalVolunteers)},
		{Label: "Active Volunteers", Value: fmt.Sprintf("%d", analytics.ActiveVolunteers)},
		{Label: "Total Hours", Value: fmt.Sprintf("%.1f", analytics.TotalHours)},
	}

	generator.AddMetricsRow(metrics)

	// Second row of metrics
	metrics2 := []pdf.Metric{
		{Label: "Opportunities", Value: fmt.Sprintf("%d", analytics.TotalOpportunities)},
		{Label: "Completed Events", Value: fmt.Sprintf("%d", analytics.CompletedEvents)},
		{Label: "Retention Rate", Value: fmt.Sprintf("%.1f%%", analytics.RetentionRate)},
	}

	generator.AddMetricsRow(metrics2)

	// Events by Cause
	if len(analytics.EventsByCause) > 0 {
		generator.AddDivider()
		chartData := make([]pdf.ChartData, 0, len(analytics.EventsByCause))
		maxValue := 0.0

		for _, item := range analytics.EventsByCause {
			value := float64(item.Count)
			chartData = append(chartData, pdf.ChartData{
				Label: item.CategoryName,
				Value: value,
			})
			if value > maxValue {
				maxValue = value
			}
		}

		generator.AddSimpleBarChart("Events by Cause Category", chartData, maxValue)
	}

	// Top Volunteers Table
	if len(analytics.TopVolunteers) > 0 {
		generator.AddDivider()
		generator.AddSectionTitle("Top Volunteers")

		headers := []string{"Volunteer", "Hours", "Events"}
		colWidths := []float64{90, 40, 40}

		rows := make([][]string, 0, len(analytics.TopVolunteers))
		for _, vol := range analytics.TopVolunteers {
			rows = append(rows, []string{
				vol.FirstName + " " + vol.LastName,
				fmt.Sprintf("%.1f", vol.TotalHours),
				fmt.Sprintf("%d", vol.TotalEvents),
			})
		}

		generator.AddTable(headers, rows, colWidths)
	}

	// Generate and return PDF
	pdfBytes, err := generator.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"organization_id": organizationID.String(),
		"pdf_size":        len(pdfBytes),
	}).Info("Organization report generated successfully")

	return pdfBytes, nil
}

// generatePlatformReport generates a PDF report for platform-wide analytics
func (s *analyticsService) generatePlatformReport(ctx context.Context, dateRange DateRange) ([]byte, error) {
	// Get platform analytics data
	analytics, err := s.GetPlatformAnalytics(ctx, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform analytics: %w", err)
	}

	// Create PDF generator
	generator := pdf.NewGenerator()
	generator.SetTitle("Platform Analytics Report")
	generator.SetAuthor("VolunteerSync Platform")

	// Add header
	generator.AddHeader(
		"Platform Analytics Report",
		fmt.Sprintf("Period: %s to %s", dateRange.StartDate.Format("Jan 2, 2006"), dateRange.EndDate.Format("Jan 2, 2006")),
	)

	generator.AddFooter(time.Now())

	// Summary Metrics
	generator.AddSectionTitle("Platform Overview")

	metrics := []pdf.Metric{
		{Label: "Total Volunteers", Value: fmt.Sprintf("%d", analytics.TotalVolunteers)},
		{Label: "Active Volunteers", Value: fmt.Sprintf("%d", analytics.ActiveVolunteers)},
		{Label: "Total Orgs", Value: fmt.Sprintf("%d", analytics.TotalOrganizations)},
	}

	generator.AddMetricsRow(metrics)

	// Second row of metrics
	avgHoursPerVol := 0.0
	if analytics.TotalVolunteers > 0 {
		avgHoursPerVol = analytics.TotalHours / float64(analytics.TotalVolunteers)
	}

	metrics2 := []pdf.Metric{
		{Label: "Total Hours", Value: fmt.Sprintf("%.0f", analytics.TotalHours)},
		{Label: "Opportunities", Value: fmt.Sprintf("%d", analytics.TotalOpportunities)},
		{Label: "Avg Hours/Vol", Value: fmt.Sprintf("%.1f", avgHoursPerVol)},
	}

	generator.AddMetricsRow(metrics2)

	// Events by Cause
	if len(analytics.EventsByCause) > 0 {
		generator.AddDivider()
		chartData := make([]pdf.ChartData, 0, len(analytics.EventsByCause))
		maxValue := 0.0

		for _, item := range analytics.EventsByCause {
			value := float64(item.Count)
			chartData = append(chartData, pdf.ChartData{
				Label: item.CategoryName,
				Value: value,
			})
			if value > maxValue {
				maxValue = value
			}
		}

		generator.AddSimpleBarChart("Events by Cause Category", chartData, maxValue)
	}

	// Geographic Distribution Table
	if len(analytics.GeographicDistribution) > 0 {
		generator.AddDivider()
		generator.AddSectionTitle("Geographic Distribution")

		headers := []string{"Location", "Volunteers", "Hours"}
		colWidths := []float64{90, 40, 40}

		rows := make([][]string, 0, len(analytics.GeographicDistribution))
		for _, geo := range analytics.GeographicDistribution {
			location := geo.State
			if geo.City != "" {
				location = geo.City + ", " + geo.State
			}
			rows = append(rows, []string{
				location,
				fmt.Sprintf("%d", geo.Volunteers),
				fmt.Sprintf("%.1f", geo.Hours),
			})
		}

		generator.AddTable(headers, rows, colWidths)
	}

	// Generate and return PDF
	pdfBytes, err := generator.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"pdf_size": len(pdfBytes),
	}).Info("Platform report generated successfully")

	return pdfBytes, nil
}

// Helper methods for volunteer analytics

func (s *analyticsService) getVolunteerTotals(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange, analytics *VolunteerAnalytics) error {
	// Get total hours from hours_logs
	type Result struct {
		TotalHours         float64
		TotalEvents        int
		TotalOrganizations int
	}
	var result Result

	// Query hours_logs joined with registrations and opportunities
	query := `
		SELECT 
			COALESCE(SUM(hl.hours), 0) as total_hours,
			COUNT(DISTINCT hl.registration_id) as total_events,
			COUNT(DISTINCT o.organization_id) as total_organizations
		FROM hours_logs hl
		INNER JOIN registrations r ON hl.registration_id = r.id
		INNER JOIN opportunities o ON r.opportunity_id = o.id
		WHERE r.volunteer_profile_id = ?
			AND hl.status = 'verified'
			AND hl.logged_at BETWEEN ? AND ?
			AND hl.deleted_at IS NULL
			AND r.deleted_at IS NULL
			AND o.deleted_at IS NULL
	`

	if err := s.db.WithContext(ctx).Raw(query, volunteerProfileID, dateRange.StartDate, dateRange.EndDate).Scan(&result).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get volunteer totals", err)
		return fmt.Errorf("failed to get volunteer totals: %w", err)
	}

	analytics.TotalHours = result.TotalHours
	analytics.TotalEvents = result.TotalEvents
	analytics.TotalOrganizations = result.TotalOrganizations

	return nil
}

func (s *analyticsService) getVolunteerHoursOverTime(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange, analytics *VolunteerAnalytics) error {
	// Get daily hours aggregated
	query := `
		SELECT 
			DATE(hl.logged_at) as date,
			SUM(hl.hours) as value
		FROM hours_logs hl
		INNER JOIN registrations r ON hl.registration_id = r.id
		WHERE r.volunteer_profile_id = ?
			AND hl.status = 'verified'
			AND hl.logged_at BETWEEN ? AND ?
			AND hl.deleted_at IS NULL
			AND r.deleted_at IS NULL
		GROUP BY DATE(hl.logged_at)
		ORDER BY date
	`

	var dataPoints []DataPoint
	if err := s.db.WithContext(ctx).Raw(query, volunteerProfileID, dateRange.StartDate, dateRange.EndDate).Scan(&dataPoints).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get volunteer hours over time", err)
		return fmt.Errorf("failed to get volunteer hours over time: %w", err)
	}

	analytics.HoursOverTime = dataPoints
	return nil
}

func (s *analyticsService) getVolunteerEventsByCause(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange, analytics *VolunteerAnalytics) error {
	query := `
		SELECT 
			cc.id as category_id,
			cc.name as category_name,
			COUNT(DISTINCT r.id) as count
		FROM registrations r
		INNER JOIN opportunities o ON r.opportunity_id = o.id
		INNER JOIN cause_categories cc ON o.cause_category_id = cc.id
		WHERE r.volunteer_profile_id = ?
			AND r.status = 'completed'
			AND o.start_date BETWEEN ? AND ?
			AND r.deleted_at IS NULL
			AND o.deleted_at IS NULL
		GROUP BY cc.id, cc.name
		ORDER BY count DESC
	`

	var counts []CategoryCount
	if err := s.db.WithContext(ctx).Raw(query, volunteerProfileID, dateRange.StartDate, dateRange.EndDate).Scan(&counts).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get volunteer events by cause", err)
		return fmt.Errorf("failed to get volunteer events by cause: %w", err)
	}

	analytics.EventsByCause = counts
	return nil
}

func (s *analyticsService) getVolunteerHoursByCause(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange, analytics *VolunteerAnalytics) error {
	query := `
		SELECT 
			cc.id as category_id,
			cc.name as category_name,
			COUNT(DISTINCT r.id) as count,
			COALESCE(SUM(hl.hours), 0) as hours
		FROM registrations r
		INNER JOIN opportunities o ON r.opportunity_id = o.id
		INNER JOIN cause_categories cc ON o.cause_category_id = cc.id
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE r.volunteer_profile_id = ?
			AND o.start_date BETWEEN ? AND ?
			AND r.deleted_at IS NULL
			AND o.deleted_at IS NULL
		GROUP BY cc.id, cc.name
		ORDER BY hours DESC
	`

	var counts []CategoryCount
	if err := s.db.WithContext(ctx).Raw(query, volunteerProfileID, dateRange.StartDate, dateRange.EndDate).Scan(&counts).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get volunteer hours by cause", err)
		return fmt.Errorf("failed to get volunteer hours by cause: %w", err)
	}

	analytics.HoursByCause = counts
	return nil
}

func (s *analyticsService) getVolunteerOrganizationStats(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange, analytics *VolunteerAnalytics) error {
	query := `
		SELECT 
			o.id as organization_id,
			o.name as organization_name,
			COUNT(DISTINCT r.id) as events_attended,
			COALESCE(SUM(hl.hours), 0) as hours_contributed
		FROM registrations r
		INNER JOIN opportunities o ON r.opportunity_id = o.id
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE r.volunteer_profile_id = ?
			AND r.status IN ('confirmed', 'checked_in', 'completed')
			AND o.start_date BETWEEN ? AND ?
			AND r.deleted_at IS NULL
			AND o.deleted_at IS NULL
		GROUP BY o.id, o.name
		ORDER BY hours_contributed DESC
	`

	var stats []OrganizationStat
	if err := s.db.WithContext(ctx).Raw(query, volunteerProfileID, dateRange.StartDate, dateRange.EndDate).Scan(&stats).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get volunteer organization stats", err)
		return fmt.Errorf("failed to get volunteer organization stats: %w", err)
	}

	analytics.OrganizationStats = stats
	return nil
}

func (s *analyticsService) getVolunteerMonthlyTrend(ctx context.Context, volunteerProfileID uuid.UUID, dateRange DateRange, analytics *VolunteerAnalytics) error {
	query := `
		SELECT 
			TO_CHAR(hl.logged_at, 'YYYY-MM') as month,
			COALESCE(SUM(hl.hours), 0) as hours,
			COUNT(DISTINCT hl.registration_id) as events,
			0 as volunteers
		FROM hours_logs hl
		INNER JOIN registrations r ON hl.registration_id = r.id
		WHERE r.volunteer_profile_id = ?
			AND hl.status = 'verified'
			AND hl.logged_at BETWEEN ? AND ?
			AND hl.deleted_at IS NULL
			AND r.deleted_at IS NULL
		GROUP BY TO_CHAR(hl.logged_at, 'YYYY-MM')
		ORDER BY month
	`

	var trend []MonthlyDataPoint
	if err := s.db.WithContext(ctx).Raw(query, volunteerProfileID, dateRange.StartDate, dateRange.EndDate).Scan(&trend).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get volunteer monthly trend", err)
		return fmt.Errorf("failed to get volunteer monthly trend: %w", err)
	}

	analytics.MonthlyTrend = trend
	return nil
}

// Helper methods for organization analytics

func (s *analyticsService) getOrganizationTotals(ctx context.Context, organizationID uuid.UUID, dateRange DateRange, analytics *OrganizationAnalytics) error {
	type Result struct {
		TotalVolunteers    int
		ActiveVolunteers   int
		TotalHours         float64
		TotalOpportunities int
		CompletedEvents    int
	}
	var result Result

	// Calculate active threshold (90 days ago)
	activeThreshold := time.Now().AddDate(0, 0, -90)

	query := `
		SELECT 
			COUNT(DISTINCT r.volunteer_profile_id) as total_volunteers,
			COUNT(DISTINCT CASE WHEN r.created_at >= ? THEN r.volunteer_profile_id END) as active_volunteers,
			COALESCE(SUM(hl.hours), 0) as total_hours,
			COUNT(DISTINCT o.id) as total_opportunities,
			COUNT(DISTINCT CASE WHEN o.status = 'completed' THEN o.id END) as completed_events
		FROM opportunities o
		LEFT JOIN registrations r ON o.id = r.opportunity_id AND r.deleted_at IS NULL
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE o.organization_id = ?
			AND o.start_date BETWEEN ? AND ?
			AND o.deleted_at IS NULL
	`

	if err := s.db.WithContext(ctx).Raw(query, activeThreshold, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&result).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get organization totals", err)
		return fmt.Errorf("failed to get organization totals: %w", err)
	}

	analytics.TotalVolunteers = result.TotalVolunteers
	analytics.ActiveVolunteers = result.ActiveVolunteers
	analytics.TotalHours = result.TotalHours
	analytics.TotalOpportunities = result.TotalOpportunities
	analytics.CompletedEvents = result.CompletedEvents

	// Calculate average fill rate
	if result.TotalOpportunities > 0 {
		var fillRateResult struct {
			AvgFillRate float64
		}
		fillRateQuery := `
			SELECT 
				AVG(CASE WHEN o.max_volunteers > 0 
					THEN (o.registered_volunteers::float / o.max_volunteers::float) * 100 
					ELSE 0 
				END) as avg_fill_rate
			FROM opportunities o
			WHERE o.organization_id = ?
				AND o.start_date BETWEEN ? AND ?
				AND o.deleted_at IS NULL
		`
		if err := s.db.WithContext(ctx).Raw(fillRateQuery, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&fillRateResult).Error; err != nil {
			s.logger.WithContext(ctx).ErrorWithErr("Failed to calculate average fill rate", err)
			// Don't fail the entire request, just log and continue
		} else {
			analytics.AverageFillRate = fillRateResult.AvgFillRate
		}
	}

	return nil
}

func (s *analyticsService) getOrganizationVolunteersOverTime(ctx context.Context, organizationID uuid.UUID, dateRange DateRange, analytics *OrganizationAnalytics) error {
	query := `
		SELECT 
			DATE(r.created_at) as date,
			COUNT(DISTINCT r.volunteer_profile_id) as value
		FROM registrations r
		INNER JOIN opportunities o ON r.opportunity_id = o.id
		WHERE o.organization_id = ?
			AND r.created_at BETWEEN ? AND ?
			AND r.deleted_at IS NULL
			AND o.deleted_at IS NULL
		GROUP BY DATE(r.created_at)
		ORDER BY date
	`

	var dataPoints []DataPoint
	if err := s.db.WithContext(ctx).Raw(query, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&dataPoints).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get organization volunteers over time", err)
		return fmt.Errorf("failed to get organization volunteers over time: %w", err)
	}

	analytics.VolunteersOverTime = dataPoints
	return nil
}

func (s *analyticsService) getOrganizationHoursOverTime(ctx context.Context, organizationID uuid.UUID, dateRange DateRange, analytics *OrganizationAnalytics) error {
	query := `
		SELECT 
			DATE(hl.logged_at) as date,
			SUM(hl.hours) as value
		FROM hours_logs hl
		INNER JOIN registrations r ON hl.registration_id = r.id
		INNER JOIN opportunities o ON r.opportunity_id = o.id
		WHERE o.organization_id = ?
			AND hl.status = 'verified'
			AND hl.logged_at BETWEEN ? AND ?
			AND hl.deleted_at IS NULL
			AND r.deleted_at IS NULL
			AND o.deleted_at IS NULL
		GROUP BY DATE(hl.logged_at)
		ORDER BY date
	`

	var dataPoints []DataPoint
	if err := s.db.WithContext(ctx).Raw(query, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&dataPoints).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get organization hours over time", err)
		return fmt.Errorf("failed to get organization hours over time: %w", err)
	}

	analytics.HoursOverTime = dataPoints
	return nil
}

func (s *analyticsService) getOrganizationEventsByCause(ctx context.Context, organizationID uuid.UUID, dateRange DateRange, analytics *OrganizationAnalytics) error {
	query := `
		SELECT 
			cc.id as category_id,
			cc.name as category_name,
			COUNT(o.id) as count
		FROM opportunities o
		INNER JOIN cause_categories cc ON o.cause_category_id = cc.id
		WHERE o.organization_id = ?
			AND o.start_date BETWEEN ? AND ?
			AND o.deleted_at IS NULL
		GROUP BY cc.id, cc.name
		ORDER BY count DESC
	`

	var counts []CategoryCount
	if err := s.db.WithContext(ctx).Raw(query, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&counts).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get organization events by cause", err)
		return fmt.Errorf("failed to get organization events by cause: %w", err)
	}

	analytics.EventsByCause = counts
	return nil
}

func (s *analyticsService) getOrganizationTopVolunteers(ctx context.Context, organizationID uuid.UUID, dateRange DateRange, analytics *OrganizationAnalytics) error {
	query := `
		SELECT 
			vp.id as volunteer_profile_id,
			vp.user_id,
			u.first_name,
			u.last_name,
			COALESCE(SUM(hl.hours), 0) as total_hours,
			COUNT(DISTINCT r.id) as total_events
		FROM registrations r
		INNER JOIN opportunities o ON r.opportunity_id = o.id
		INNER JOIN volunteer_profiles vp ON r.volunteer_profile_id = vp.id
		INNER JOIN users u ON vp.user_id = u.id
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE o.organization_id = ?
			AND o.start_date BETWEEN ? AND ?
			AND r.deleted_at IS NULL
			AND o.deleted_at IS NULL
			AND vp.deleted_at IS NULL
			AND u.deleted_at IS NULL
		GROUP BY vp.id, vp.user_id, u.first_name, u.last_name
		ORDER BY total_hours DESC
		LIMIT 10
	`

	var volunteers []VolunteerSummary
	if err := s.db.WithContext(ctx).Raw(query, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&volunteers).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get organization top volunteers", err)
		return fmt.Errorf("failed to get organization top volunteers: %w", err)
	}

	analytics.TopVolunteers = volunteers
	return nil
}

func (s *analyticsService) getOrganizationMonthlyTrend(ctx context.Context, organizationID uuid.UUID, dateRange DateRange, analytics *OrganizationAnalytics) error {
	query := `
		SELECT 
			TO_CHAR(o.start_date, 'YYYY-MM') as month,
			COALESCE(SUM(hl.hours), 0) as hours,
			COUNT(DISTINCT o.id) as events,
			COUNT(DISTINCT r.volunteer_profile_id) as volunteers
		FROM opportunities o
		LEFT JOIN registrations r ON o.id = r.opportunity_id AND r.deleted_at IS NULL
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE o.organization_id = ?
			AND o.start_date BETWEEN ? AND ?
			AND o.deleted_at IS NULL
		GROUP BY TO_CHAR(o.start_date, 'YYYY-MM')
		ORDER BY month
	`

	var trend []MonthlyDataPoint
	if err := s.db.WithContext(ctx).Raw(query, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&trend).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get organization monthly trend", err)
		return fmt.Errorf("failed to get organization monthly trend: %w", err)
	}

	analytics.MonthlyTrend = trend
	return nil
}

func (s *analyticsService) calculateOrganizationRetentionRate(ctx context.Context, organizationID uuid.UUID, dateRange DateRange, analytics *OrganizationAnalytics) error {
	// Retention rate = (volunteers with 2+ events) / (total volunteers) * 100
	query := `
		SELECT 
			COUNT(DISTINCT CASE WHEN event_count >= 2 THEN volunteer_profile_id END)::float / 
			NULLIF(COUNT(DISTINCT volunteer_profile_id), 0) * 100 as retention_rate
		FROM (
			SELECT 
				r.volunteer_profile_id,
				COUNT(DISTINCT r.id) as event_count
			FROM registrations r
			INNER JOIN opportunities o ON r.opportunity_id = o.id
			WHERE o.organization_id = ?
				AND r.status IN ('confirmed', 'checked_in', 'completed')
				AND o.start_date BETWEEN ? AND ?
				AND r.deleted_at IS NULL
				AND o.deleted_at IS NULL
			GROUP BY r.volunteer_profile_id
		) AS volunteer_events
	`

	var result struct {
		RetentionRate float64
	}

	if err := s.db.WithContext(ctx).Raw(query, organizationID, dateRange.StartDate, dateRange.EndDate).Scan(&result).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to calculate retention rate", err)
		// Don't fail the entire request
		return nil
	}

	analytics.RetentionRate = result.RetentionRate
	return nil
}

// Helper methods for platform analytics

func (s *analyticsService) getPlatformTotals(ctx context.Context, dateRange DateRange, analytics *PlatformAnalytics) error {
	type Result struct {
		TotalVolunteers        int
		ActiveVolunteers       int
		TotalOrganizations     int
		ActiveOrganizations    int
		TotalHours             float64
		TotalOpportunities     int
		CompletedOpportunities int
	}
	var result Result

	activeThreshold := time.Now().AddDate(0, 0, -90)

	// Get volunteer counts
	volunteerQuery := `
		SELECT 
			COUNT(DISTINCT vp.id) as total_volunteers,
			COUNT(DISTINCT CASE WHEN r.created_at >= ? THEN vp.id END) as active_volunteers
		FROM volunteer_profiles vp
		LEFT JOIN registrations r ON vp.id = r.volunteer_profile_id AND r.deleted_at IS NULL
		WHERE vp.created_at <= ?
			AND vp.deleted_at IS NULL
	`
	if err := s.db.WithContext(ctx).Raw(volunteerQuery, activeThreshold, dateRange.EndDate).Scan(&result).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform volunteer counts", err)
		return fmt.Errorf("failed to get platform volunteer counts: %w", err)
	}

	// Get organization counts
	orgQuery := `
		SELECT 
			COUNT(DISTINCT o.id) as total_organizations,
			COUNT(DISTINCT CASE WHEN opp.created_at >= ? THEN o.id END) as active_organizations
		FROM organizations o
		LEFT JOIN opportunities opp ON o.id = opp.organization_id AND opp.deleted_at IS NULL
		WHERE o.created_at <= ?
			AND o.deleted_at IS NULL
	`
	var orgResult struct {
		TotalOrganizations  int
		ActiveOrganizations int
	}
	if err := s.db.WithContext(ctx).Raw(orgQuery, activeThreshold, dateRange.EndDate).Scan(&orgResult).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform organization counts", err)
		return fmt.Errorf("failed to get platform organization counts: %w", err)
	}
	result.TotalOrganizations = orgResult.TotalOrganizations
	result.ActiveOrganizations = orgResult.ActiveOrganizations

	// Get hours and opportunities
	hoursQuery := `
		SELECT 
			COALESCE(SUM(hl.hours), 0) as total_hours,
			COUNT(DISTINCT o.id) as total_opportunities,
			COUNT(DISTINCT CASE WHEN o.status = 'completed' THEN o.id END) as completed_opportunities
		FROM opportunities o
		LEFT JOIN registrations r ON o.id = r.opportunity_id AND r.deleted_at IS NULL
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE o.start_date BETWEEN ? AND ?
			AND o.deleted_at IS NULL
	`
	var hoursResult struct {
		TotalHours             float64
		TotalOpportunities     int
		CompletedOpportunities int
	}
	if err := s.db.WithContext(ctx).Raw(hoursQuery, dateRange.StartDate, dateRange.EndDate).Scan(&hoursResult).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform hours and opportunities", err)
		return fmt.Errorf("failed to get platform hours and opportunities: %w", err)
	}
	result.TotalHours = hoursResult.TotalHours
	result.TotalOpportunities = hoursResult.TotalOpportunities
	result.CompletedOpportunities = hoursResult.CompletedOpportunities

	analytics.TotalVolunteers = result.TotalVolunteers
	analytics.ActiveVolunteers = result.ActiveVolunteers
	analytics.TotalOrganizations = result.TotalOrganizations
	analytics.ActiveOrganizations = result.ActiveOrganizations
	analytics.TotalHours = result.TotalHours
	analytics.TotalOpportunities = result.TotalOpportunities
	analytics.CompletedOpportunities = result.CompletedOpportunities

	return nil
}

func (s *analyticsService) getPlatformUserGrowth(ctx context.Context, dateRange DateRange, analytics *PlatformAnalytics) error {
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as value
		FROM users
		WHERE created_at BETWEEN ? AND ?
			AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	var dataPoints []DataPoint
	if err := s.db.WithContext(ctx).Raw(query, dateRange.StartDate, dateRange.EndDate).Scan(&dataPoints).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform user growth", err)
		return fmt.Errorf("failed to get platform user growth: %w", err)
	}

	analytics.UserGrowth = dataPoints
	return nil
}

func (s *analyticsService) getPlatformOrganizationGrowth(ctx context.Context, dateRange DateRange, analytics *PlatformAnalytics) error {
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as value
		FROM organizations
		WHERE created_at BETWEEN ? AND ?
			AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	var dataPoints []DataPoint
	if err := s.db.WithContext(ctx).Raw(query, dateRange.StartDate, dateRange.EndDate).Scan(&dataPoints).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform organization growth", err)
		return fmt.Errorf("failed to get platform organization growth: %w", err)
	}

	analytics.OrganizationGrowth = dataPoints
	return nil
}

func (s *analyticsService) getPlatformHoursOverTime(ctx context.Context, dateRange DateRange, analytics *PlatformAnalytics) error {
	query := `
		SELECT 
			DATE(hl.logged_at) as date,
			SUM(hl.hours) as value
		FROM hours_logs hl
		WHERE hl.status = 'verified'
			AND hl.logged_at BETWEEN ? AND ?
			AND hl.deleted_at IS NULL
		GROUP BY DATE(hl.logged_at)
		ORDER BY date
	`

	var dataPoints []DataPoint
	if err := s.db.WithContext(ctx).Raw(query, dateRange.StartDate, dateRange.EndDate).Scan(&dataPoints).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform hours over time", err)
		return fmt.Errorf("failed to get platform hours over time: %w", err)
	}

	analytics.HoursOverTime = dataPoints
	return nil
}

func (s *analyticsService) getPlatformEventsByCause(ctx context.Context, dateRange DateRange, analytics *PlatformAnalytics) error {
	query := `
		SELECT 
			cc.id as category_id,
			cc.name as category_name,
			COUNT(o.id) as count
		FROM opportunities o
		INNER JOIN cause_categories cc ON o.cause_category_id = cc.id
		WHERE o.start_date BETWEEN ? AND ?
			AND o.deleted_at IS NULL
		GROUP BY cc.id, cc.name
		ORDER BY count DESC
	`

	var counts []CategoryCount
	if err := s.db.WithContext(ctx).Raw(query, dateRange.StartDate, dateRange.EndDate).Scan(&counts).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform events by cause", err)
		return fmt.Errorf("failed to get platform events by cause: %w", err)
	}

	analytics.EventsByCause = counts
	return nil
}

func (s *analyticsService) getPlatformGeographicDistribution(ctx context.Context, dateRange DateRange, analytics *PlatformAnalytics) error {
	query := `
		SELECT 
			COALESCE(o.state, 'Unknown') as state,
			COUNT(DISTINCT vp.id) as volunteers,
			COUNT(DISTINCT o.id) as organizations,
			COALESCE(SUM(hl.hours), 0) as hours
		FROM organizations o
		LEFT JOIN opportunities opp ON o.id = opp.organization_id AND opp.deleted_at IS NULL
		LEFT JOIN registrations r ON opp.id = r.opportunity_id AND r.deleted_at IS NULL
		LEFT JOIN volunteer_profiles vp ON r.volunteer_profile_id = vp.id AND vp.deleted_at IS NULL
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE o.created_at <= ?
			AND o.deleted_at IS NULL
		GROUP BY o.state
		ORDER BY hours DESC
		LIMIT 20
	`

	var stats []GeographicStat
	if err := s.db.WithContext(ctx).Raw(query, dateRange.EndDate).Scan(&stats).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform geographic distribution", err)
		return fmt.Errorf("failed to get platform geographic distribution: %w", err)
	}

	analytics.GeographicDistribution = stats
	return nil
}

func (s *analyticsService) getPlatformMonthlyTrend(ctx context.Context, dateRange DateRange, analytics *PlatformAnalytics) error {
	query := `
		SELECT 
			TO_CHAR(o.start_date, 'YYYY-MM') as month,
			COALESCE(SUM(hl.hours), 0) as hours,
			COUNT(DISTINCT o.id) as events,
			COUNT(DISTINCT r.volunteer_profile_id) as volunteers,
			COUNT(DISTINCT o.organization_id) as organizations
		FROM opportunities o
		LEFT JOIN registrations r ON o.id = r.opportunity_id AND r.deleted_at IS NULL
		LEFT JOIN hours_logs hl ON r.id = hl.registration_id AND hl.status = 'verified' AND hl.deleted_at IS NULL
		WHERE o.start_date BETWEEN ? AND ?
			AND o.deleted_at IS NULL
		GROUP BY TO_CHAR(o.start_date, 'YYYY-MM')
		ORDER BY month
	`

	var trend []MonthlyDataPoint
	if err := s.db.WithContext(ctx).Raw(query, dateRange.StartDate, dateRange.EndDate).Scan(&trend).Error; err != nil {
		s.logger.WithContext(ctx).ErrorWithErr("Failed to get platform monthly trend", err)
		return fmt.Errorf("failed to get platform monthly trend: %w", err)
	}

	analytics.MonthlyTrend = trend
	return nil
}
