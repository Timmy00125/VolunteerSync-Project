'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Heart, Clock, Award, Calendar, CheckCircle2, TrendingUp } from 'lucide-react';
import { useVolunteerDashboard } from '@/lib/api';
import { MetricCard } from '@/components/dashboard/MetricCard';
import { HoursChart } from '@/components/dashboard/HoursChart';
import { format, parseISO, differenceInDays } from 'date-fns';

/**
 * Volunteer Dashboard Page
 *
 * Main dashboard view for volunteers showing:
 * - Quick stats (hours volunteered, events attended, organizations supported)
 * - Hours over time chart
 * - Upcoming events
 * - Recent events
 * - Achievement highlights
 */
export default function VolunteerDashboardPage() {
  const { data: dashboard, isLoading, error } = useVolunteerDashboard();

  // Loading state
  if (isLoading) {
    return (
      <div className="space-y-8">
        <div className="space-y-2">
          <div className="h-9 w-64 animate-pulse rounded-md bg-muted" />
          <div className="h-5 w-96 animate-pulse rounded-md bg-muted" />
        </div>

        {/* Skeleton for metric cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardHeader className="space-y-0 pb-2">
                <div className="h-4 w-24 animate-pulse rounded-md bg-muted" />
              </CardHeader>
              <CardContent>
                <div className="h-8 w-16 animate-pulse rounded-md bg-muted" />
                <div className="mt-1 h-3 w-32 animate-pulse rounded-md bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Skeleton for content sections */}
        <Card>
          <CardHeader>
            <div className="h-6 w-32 animate-pulse rounded-md bg-muted" />
          </CardHeader>
          <CardContent>
            <div className="h-48 animate-pulse rounded-md bg-muted" />
          </CardContent>
        </Card>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        </div>
        <Card className="border-destructive">
          <CardHeader>
            <CardTitle className="text-destructive">Error Loading Dashboard</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              {error instanceof Error
                ? error.message
                : 'Failed to load dashboard data. Please try again later.'}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  // No data state (shouldn't happen in normal flow)
  if (!dashboard) {
    return (
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        </div>
        <Card>
          <CardContent className="py-8 text-center">
            <p className="text-sm text-muted-foreground">No dashboard data available.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Get user's first name from profile if available
  const userName = dashboard.profile?.user_id ? 'there' : 'there'; // Placeholder until we have user name

  // Prepare chart data (last 6 months)
  // Note: Backend doesn't currently return monthly breakdown, so we'll show a placeholder
  // This can be enhanced when the analytics endpoint is implemented
  const chartData = [
    { label: 'Jan', hours: 0 },
    { label: 'Feb', hours: 0 },
    { label: 'Mar', hours: 0 },
    { label: 'Apr', hours: 0 },
    { label: 'May', hours: 0 },
    { label: 'Jun', hours: dashboard.hours_this_month },
  ];

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          Welcome back{userName !== 'there' ? `, ${userName}` : ''}!
        </h1>
        <p className="text-muted-foreground">
          Here's an overview of your volunteer impact and upcoming events.
        </p>
      </div>

      {/* Quick Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Total Hours"
          value={dashboard.total_hours.toFixed(1)}
          icon={Clock}
          subtitle={`+${dashboard.hours_this_month.toFixed(1)} hours this month`}
        />

        <MetricCard
          title="Events Attended"
          value={dashboard.total_events}
          icon={Calendar}
          subtitle={`+${dashboard.events_this_month} events this month`}
        />

        <MetricCard
          title="Organizations"
          value={dashboard.total_organizations}
          icon={Heart}
          subtitle="Across the community"
        />

        <MetricCard
          title="Achievements"
          value={dashboard.achievements.length}
          icon={Award}
          subtitle={
            dashboard.achievements.length > 0
              ? 'Badges earned'
              : 'Start volunteering to earn badges'
          }
        />
      </div>

      {/* Hours Over Time Chart */}
      <HoursChart title="Hours This Year" data={chartData} />

      {/* Upcoming Events Section */}
      {dashboard.upcoming_events.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Upcoming Events</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {dashboard.upcoming_events.slice(0, 5).map((event) => {
                const eventDate = parseISO(event.date);
                const daysUntil = differenceInDays(eventDate, new Date());

                return (
                  <div
                    key={event.id}
                    className="flex items-center justify-between rounded-lg border p-4"
                  >
                    <div className="space-y-1">
                      <p className="font-medium">{event.opportunity_title}</p>
                      <p className="text-sm text-muted-foreground">
                        {event.organization_name} • {format(eventDate, 'EEEE, MMM d, yyyy')} •{' '}
                        {format(parseISO(event.start_time), 'h:mm a')} -{' '}
                        {format(parseISO(event.end_time), 'h:mm a')}
                      </p>
                      <p className="text-sm text-muted-foreground">{event.location}</p>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {daysUntil === 0
                        ? 'Today'
                        : daysUntil === 1
                          ? 'Tomorrow'
                          : `In ${daysUntil} days`}
                    </div>
                  </div>
                );
              })}
              {dashboard.upcoming_events.length > 5 && (
                <p className="text-center text-sm text-muted-foreground">
                  + {dashboard.upcoming_events.length - 5} more events in{' '}
                  <a href="/volunteer/events" className="text-primary hover:underline">
                    My Events
                  </a>
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Recent Events Section */}
      {dashboard.recent_events.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {dashboard.recent_events.slice(0, 5).map((event) => (
                <div key={event.id} className="flex items-start gap-4">
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
                    <CheckCircle2 className="h-4 w-4 text-primary" />
                  </div>
                  <div className="flex-1 space-y-1">
                    <p className="text-sm font-medium">{event.opportunity_title}</p>
                    <p className="text-sm text-muted-foreground">
                      {event.organization_name} • {format(parseISO(event.date), 'MMM d, yyyy')}
                    </p>
                    <div className="flex items-center gap-2">
                      <p className="text-sm text-muted-foreground">
                        {event.hours_logged} hours • Status: {event.status || 'completed'}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Achievements Section */}
      {dashboard.achievements.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Achievements</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {dashboard.achievements.slice(0, 3).map((achievement) => (
                <div key={achievement.id} className="flex items-start gap-4">
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
                    <Award className="h-4 w-4 text-primary" />
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium">{achievement.name}</p>
                    <p className="text-sm text-muted-foreground">{achievement.description}</p>
                    {achievement.earned_at && (
                      <p className="text-xs text-muted-foreground">
                        Earned {format(parseISO(achievement.earned_at), 'MMM d, yyyy')}
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Empty State */}
      {dashboard.upcoming_events.length === 0 &&
        dashboard.recent_events.length === 0 &&
        dashboard.achievements.length === 0 && (
          <Card>
            <CardContent className="py-8 text-center">
              <TrendingUp className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">Start Your Volunteer Journey</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                You haven't signed up for any events yet. Browse opportunities to get started!
              </p>
              <a
                href="/volunteer/opportunities"
                className="mt-4 inline-block rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
              >
                Find Opportunities
              </a>
            </CardContent>
          </Card>
        )}
    </div>
  );
}
