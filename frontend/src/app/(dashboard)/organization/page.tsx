'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Users, Clock, Calendar, TrendingUp, UserCheck, Award } from 'lucide-react';
import { useOrganizationDashboard } from '@/lib/api';
import { MetricCard } from '@/components/dashboard/MetricCard';
import { format, parseISO } from 'date-fns';
import { useEffect, useState } from 'react';
import apiClient from '@/lib/api/client';

/**
 * Organization Dashboard Page (T115)
 *
 * Main dashboard view for organization administrators and coordinators showing:
 * - Key metrics (volunteers recruited, hours contributed, events hosted)
 * - Upcoming events list
 * - Recent registrations
 * - Quick analytics summaries
 */
export default function OrganizationDashboardPage() {
  const [organizationId, setOrganizationId] = useState<string>('');
  const [isLoadingOrg, setIsLoadingOrg] = useState(true);

  // Fetch user's organizations on mount
  useEffect(() => {
    const fetchUserOrganizations = async () => {
      try {
        const orgs = await apiClient.getUserOrganizations();
        if (orgs && orgs.length > 0) {
          // Use the first organization (in a real app, you might let the user select)
          setOrganizationId(orgs[0].id);
        }
      } catch (error) {
        console.error('Failed to fetch user organizations:', error);
      } finally {
        setIsLoadingOrg(false);
      }
    };

    fetchUserOrganizations();
  }, []);

  const { data: dashboard, isLoading, error } = useOrganizationDashboard(organizationId);

  // Loading state
  if (isLoading || isLoadingOrg || !organizationId) {
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
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <div className="h-6 w-32 animate-pulse rounded-md bg-muted" />
            </CardHeader>
            <CardContent>
              <div className="h-48 animate-pulse rounded-md bg-muted" />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <div className="h-6 w-32 animate-pulse rounded-md bg-muted" />
            </CardHeader>
            <CardContent>
              <div className="h-48 animate-pulse rounded-md bg-muted" />
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Organization Dashboard</h1>
          <p className="text-muted-foreground">Welcome to your organization management portal</p>
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

  // No data state
  if (!dashboard) {
    return (
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Organization Dashboard</h1>
          <p className="text-muted-foreground">Welcome to your organization management portal</p>
        </div>
        <Card>
          <CardContent className="py-8 text-center">
            <p className="text-sm text-muted-foreground">No dashboard data available.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const { organization, metrics, upcoming_events, recent_registrations } = dashboard;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">{organization.name} Dashboard</h1>
        <p className="text-muted-foreground">
          Overview of your organization&apos;s volunteer activities and impact
        </p>
      </div>

      {/* Metrics Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Total Volunteers"
          value={metrics.total_volunteers}
          icon={Users}
          subtitle={`${metrics.active_volunteers} active this month`}
        />
        <MetricCard
          title="Hours Contributed"
          value={metrics.total_hours.toFixed(1)}
          icon={Clock}
          subtitle={`${metrics.hours_this_month.toFixed(1)} hours this month`}
        />
        <MetricCard
          title="Events Hosted"
          value={metrics.total_events}
          icon={Calendar}
          subtitle={`${metrics.events_this_month} this month`}
        />
        <MetricCard
          title="Retention Rate"
          value={`${metrics.volunteer_retention_rate.toFixed(0)}%`}
          icon={TrendingUp}
          subtitle="Volunteer retention"
        />
      </div>

      {/* Content Grid */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Upcoming Events */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Calendar className="h-5 w-5" />
              Upcoming Events
            </CardTitle>
          </CardHeader>
          <CardContent>
            {upcoming_events.length === 0 ? (
              <p className="text-sm text-muted-foreground">No upcoming events scheduled.</p>
            ) : (
              <div className="space-y-4">
                {upcoming_events.slice(0, 5).map((event) => (
                  <div
                    key={event.id}
                    className="flex items-start justify-between border-b pb-3 last:border-0 last:pb-0"
                  >
                    <div className="space-y-1">
                      <p className="text-sm font-medium leading-none">{event.title}</p>
                      <p className="text-xs text-muted-foreground">
                        {format(parseISO(event.date), 'MMM d, yyyy')} at {event.start_time}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm font-medium">
                        {event.registered_count}/{event.capacity}
                      </p>
                      <p className="text-xs text-muted-foreground">registered</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Recent Registrations */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <UserCheck className="h-5 w-5" />
              Recent Registrations
            </CardTitle>
          </CardHeader>
          <CardContent>
            {recent_registrations.length === 0 ? (
              <p className="text-sm text-muted-foreground">No recent registrations.</p>
            ) : (
              <div className="space-y-4">
                {recent_registrations.slice(0, 5).map((registration) => (
                  <div
                    key={registration.id}
                    className="flex items-start justify-between border-b pb-3 last:border-0 last:pb-0"
                  >
                    <div className="space-y-1">
                      <p className="text-sm font-medium leading-none">
                        {registration.volunteer_name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {registration.opportunity_title}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-muted-foreground">
                        {format(parseISO(registration.registered_at), 'MMM d')}
                      </p>
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
                          registration.status === 'registered'
                            ? 'bg-green-50 text-green-700'
                            : registration.status === 'waitlisted'
                              ? 'bg-yellow-50 text-yellow-700'
                              : 'bg-gray-50 text-gray-700'
                        }`}
                      >
                        {registration.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Quick Stats */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Award className="h-5 w-5" />
            Quick Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <p className="text-sm font-medium text-muted-foreground">Upcoming Events</p>
              <p className="text-2xl font-bold">{metrics.upcoming_events}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm font-medium text-muted-foreground">Hours This Month</p>
              <p className="text-2xl font-bold">{metrics.hours_this_month.toFixed(1)}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm font-medium text-muted-foreground">Active Volunteers</p>
              <p className="text-2xl font-bold">{metrics.active_volunteers}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
