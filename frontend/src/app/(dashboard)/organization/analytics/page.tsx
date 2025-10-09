/**
 * Organization Analytics Page
 *
 * Task: T122 - Organization Analytics Page
 *
 * Comprehensive analytics dashboard for organization coordinators to view:
 * - Volunteers by cause breakdown
 * - Hours contributed over time
 * - Volunteer retention rate
 * - Event completion rate
 * - Average volunteers per event
 * - Top volunteers
 *
 * Features:
 * - Time period selection (Last 30 days, 3 months, 6 months, 1 year, All time)
 * - Visual charts for key metrics
 * - Export functionality
 * - Loading and error states
 */

'use client';

import { useState } from 'react';
import { useOrganizationAnalytics } from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { HoursChart } from '@/components/dashboard/HoursChart';
import {
  TrendingUp,
  Users,
  Target,
  Award,
  Download,
  Calendar,
  BarChart3,
  PieChart,
} from 'lucide-react';

type TimePeriod = '30d' | '3m' | '6m' | '1y' | 'all';

interface MetricCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  trend?: {
    value: number;
    label: string;
  };
}

function MetricCard({ title, value, subtitle, icon, trend }: MetricCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {subtitle && <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>}
        {trend && (
          <div
            className={`flex items-center gap-1 mt-2 text-xs ${
              trend.value >= 0 ? 'text-green-600' : 'text-red-600'
            }`}
          >
            <TrendingUp className={`h-3 w-3 ${trend.value < 0 ? 'rotate-180' : ''}`} />
            <span>
              {trend.value >= 0 ? '+' : ''}
              {trend.value}% {trend.label}
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function VolunteersByCauseChart({ data }: { data: Record<string, number> }) {
  const causes = Object.entries(data).sort((a, b) => b[1] - a[1]);
  const total = causes.reduce((sum, [_, count]) => sum + count, 0);
  const maxCount = Math.max(...causes.map(([_, count]) => count), 1);

  if (causes.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <PieChart className="h-5 w-5" />
            Volunteers by Cause
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No data available</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <PieChart className="h-5 w-5" />
          Volunteers by Cause
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {causes.map(([cause, count]) => {
            const percentage = total > 0 ? (count / total) * 100 : 0;
            const barWidth = (count / maxCount) * 100;

            return (
              <div key={cause} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium capitalize">{cause.replace(/_/g, ' ')}</span>
                  <span className="text-muted-foreground">
                    {count} ({percentage.toFixed(0)}%)
                  </span>
                </div>
                <div className="h-2 bg-secondary rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary rounded-full transition-all"
                    style={{ width: `${barWidth}%` }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}

function TopVolunteersTable({
  volunteers,
}: {
  volunteers: Array<{
    id: string;
    name: string;
    hours: number;
    events: number;
  }>;
}) {
  if (volunteers.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Award className="h-5 w-5" />
            Top Volunteers
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No volunteers yet</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Award className="h-5 w-5" />
          Top Volunteers
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {volunteers.map((volunteer, index) => (
            <div
              key={volunteer.id}
              className="flex items-center justify-between p-3 bg-secondary/50 rounded-lg"
            >
              <div className="flex items-center gap-3">
                <div
                  className={`flex items-center justify-center w-8 h-8 rounded-full font-bold text-sm ${
                    index === 0
                      ? 'bg-yellow-500 text-white'
                      : index === 1
                        ? 'bg-gray-400 text-white'
                        : index === 2
                          ? 'bg-orange-600 text-white'
                          : 'bg-primary/10 text-primary'
                  }`}
                >
                  {index + 1}
                </div>
                <div>
                  <div className="font-medium">{volunteer.name}</div>
                  <div className="text-sm text-muted-foreground">
                    {volunteer.events} event{volunteer.events !== 1 ? 's' : ''}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="font-bold text-primary">{volunteer.hours.toFixed(1)}h</div>
                <div className="text-xs text-muted-foreground">total hours</div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

export default function OrganizationAnalyticsPage() {
  const [timePeriod, setTimePeriod] = useState<TimePeriod>('3m');
  const { data: analytics, isLoading, error } = useOrganizationAnalytics(timePeriod);

  const handleExport = () => {
    if (!analytics) return;

    // Prepare CSV data
    const csvData = [
      ['Organization Analytics Report'],
      ['Time Period', timePeriod],
      ['Generated', new Date().toISOString()],
      [''],
      ['Key Metrics'],
      ['Volunteer Retention Rate', `${analytics.volunteer_retention_rate.toFixed(1)}%`],
      ['Event Completion Rate', `${analytics.event_completion_rate.toFixed(1)}%`],
      ['Average Volunteers per Event', analytics.average_volunteers_per_event.toFixed(1)],
      [''],
      ['Hours Over Time'],
      ['Period', 'Hours'],
      ...analytics.hours_over_time.map((item) => [item.month, item.hours.toString()]),
      [''],
      ['Volunteers by Cause'],
      ['Cause', 'Count'],
      ...analytics.volunteers_by_cause.map((item) => [item.cause, item.count.toString()]),
      [''],
      ['Top Volunteers'],
      ['Rank', 'Name', 'Total Hours', 'Events Attended'],
      ...analytics.top_volunteers.map((v, i) => [
        (i + 1).toString(),
        v.name,
        v.hours.toFixed(1),
        v.events.toString(),
      ]),
    ];

    const csv = csvData.map((row) => row.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `organization-analytics-${timePeriod}-${Date.now()}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  if (isLoading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex items-center justify-center h-64">
          <div className="text-center space-y-2">
            <BarChart3 className="h-12 w-12 animate-pulse mx-auto text-muted-foreground" />
            <p className="text-muted-foreground">Loading analytics...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-8">
        <Card className="border-destructive">
          <CardHeader>
            <CardTitle className="text-destructive">Error Loading Analytics</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              {error instanceof Error
                ? error.message
                : 'Failed to load analytics data. Please try again.'}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!analytics) {
    return (
      <div className="container mx-auto py-8">
        <Card>
          <CardHeader>
            <CardTitle>No Analytics Data</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              No analytics data available yet. Data will appear once you have volunteers and events.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Prepare hours over time data for the chart
  const hoursChartData = analytics.hours_over_time.map((item) => ({
    label: item.month,
    hours: item.hours,
  }));

  // Prepare volunteers by cause data for the chart
  const volunteersByCauseData = analytics.volunteers_by_cause.reduce(
    (acc, item) => {
      acc[item.cause] = item.count;
      return acc;
    },
    {} as Record<string, number>
  );

  return (
    <div className="container mx-auto py-8 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Organization Analytics</h1>
          <p className="text-muted-foreground mt-1">Insights and metrics for your organization</p>
        </div>
        <div className="flex items-center gap-3">
          <Select value={timePeriod} onValueChange={(value) => setTimePeriod(value as TimePeriod)}>
            <SelectTrigger className="w-[180px]">
              <Calendar className="h-4 w-4 mr-2" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="30d">Last 30 days</SelectItem>
              <SelectItem value="3m">Last 3 months</SelectItem>
              <SelectItem value="6m">Last 6 months</SelectItem>
              <SelectItem value="1y">Last year</SelectItem>
              <SelectItem value="all">All time</SelectItem>
            </SelectContent>
          </Select>
          <Button onClick={handleExport} variant="outline">
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Volunteer Retention"
          value={`${analytics.volunteer_retention_rate.toFixed(1)}%`}
          subtitle="of volunteers returned"
          icon={<Users className="h-4 w-4 text-muted-foreground" />}
        />
        <MetricCard
          title="Event Completion"
          value={`${analytics.event_completion_rate.toFixed(1)}%`}
          subtitle="of events completed"
          icon={<Target className="h-4 w-4 text-muted-foreground" />}
        />
        <MetricCard
          title="Avg Volunteers/Event"
          value={analytics.average_volunteers_per_event.toFixed(1)}
          subtitle="average attendance"
          icon={<Users className="h-4 w-4 text-muted-foreground" />}
        />
        <MetricCard
          title="Total Hours"
          value={analytics.hours_over_time.reduce((sum, item) => sum + item.hours, 0).toFixed(0)}
          subtitle="volunteer hours contributed"
          icon={<BarChart3 className="h-4 w-4 text-muted-foreground" />}
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Hours Over Time Chart */}
        <HoursChart title="Hours Contributed Over Time" data={hoursChartData} />

        {/* Volunteers by Cause Chart */}
        <VolunteersByCauseChart data={volunteersByCauseData} />
      </div>

      {/* Top Volunteers Table */}
      <TopVolunteersTable volunteers={analytics.top_volunteers} />

      {/* Additional Insights */}
      <Card>
        <CardHeader>
          <CardTitle>Insights</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-4 bg-secondary/50 rounded-lg">
              <div className="flex items-start gap-3">
                <TrendingUp className="h-5 w-5 text-green-600 mt-0.5" />
                <div>
                  <h4 className="font-medium">Retention Rate</h4>
                  <p className="text-sm text-muted-foreground mt-1">
                    {analytics.volunteer_retention_rate >= 70 ? (
                      <>
                        Great! Your retention rate of{' '}
                        {analytics.volunteer_retention_rate.toFixed(0)}% indicates volunteers are
                        enjoying their experience.
                      </>
                    ) : analytics.volunteer_retention_rate >= 50 ? (
                      <>
                        Your retention rate of {analytics.volunteer_retention_rate.toFixed(0)}% is
                        moderate. Consider gathering feedback to improve volunteer experience.
                      </>
                    ) : (
                      <>
                        Your retention rate of {analytics.volunteer_retention_rate.toFixed(0)}%
                        could be improved. Focus on volunteer engagement and communication.
                      </>
                    )}
                  </p>
                </div>
              </div>
            </div>

            <div className="p-4 bg-secondary/50 rounded-lg">
              <div className="flex items-start gap-3">
                <Target className="h-5 w-5 text-blue-600 mt-0.5" />
                <div>
                  <h4 className="font-medium">Event Completion</h4>
                  <p className="text-sm text-muted-foreground mt-1">
                    {analytics.event_completion_rate >= 90 ? (
                      <>
                        Excellent! {analytics.event_completion_rate.toFixed(0)}% of your events are
                        being completed successfully.
                      </>
                    ) : analytics.event_completion_rate >= 75 ? (
                      <>
                        {analytics.event_completion_rate.toFixed(0)}% completion rate is good.
                        Review cancelled events to identify patterns.
                      </>
                    ) : (
                      <>
                        {analytics.event_completion_rate.toFixed(0)}% completion rate suggests room
                        for improvement in event planning and volunteer commitment.
                      </>
                    )}
                  </p>
                </div>
              </div>
            </div>

            <div className="p-4 bg-secondary/50 rounded-lg">
              <div className="flex items-start gap-3">
                <Users className="h-5 w-5 text-purple-600 mt-0.5" />
                <div>
                  <h4 className="font-medium">Volunteer Engagement</h4>
                  <p className="text-sm text-muted-foreground mt-1">
                    Average of {analytics.average_volunteers_per_event.toFixed(1)} volunteers per
                    event.
                    {analytics.average_volunteers_per_event < 5 &&
                      ' Consider promoting events to attract more volunteers.'}
                    {analytics.average_volunteers_per_event >= 5 &&
                      analytics.average_volunteers_per_event < 15 &&
                      ' Good volunteer turnout!'}
                    {analytics.average_volunteers_per_event >= 15 &&
                      ' Excellent volunteer engagement!'}
                  </p>
                </div>
              </div>
            </div>

            <div className="p-4 bg-secondary/50 rounded-lg">
              <div className="flex items-start gap-3">
                <PieChart className="h-5 w-5 text-orange-600 mt-0.5" />
                <div>
                  <h4 className="font-medium">Cause Diversity</h4>
                  <p className="text-sm text-muted-foreground mt-1">
                    You have volunteers interested in{' '}
                    {Object.keys(analytics.volunteers_by_cause).length} different cause
                    {Object.keys(analytics.volunteers_by_cause).length !== 1 ? 's' : ''}.
                    {Object.keys(analytics.volunteers_by_cause).length >= 5 &&
                      ' Great diversity in volunteer interests!'}
                    {Object.keys(analytics.volunteers_by_cause).length < 5 &&
                      Object.keys(analytics.volunteers_by_cause).length > 0 &&
                      ' Consider diversifying your opportunities to attract more volunteers.'}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
