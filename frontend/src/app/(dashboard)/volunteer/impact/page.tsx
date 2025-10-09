'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import {
  Download,
  Award,
  Clock,
  Calendar,
  Building2,
  TrendingUp,
  Trophy,
  Target,
  Heart,
} from 'lucide-react';
import { useVolunteerDashboard } from '@/lib/api';
import { HoursChart } from '@/components/dashboard/HoursChart';
import type { Achievement } from '@/lib/api/types';

/**
 * Impact Page
 *
 * Displays volunteer's impact metrics and achievements:
 * - Total impact metrics (hours, events, organizations)
 * - Achievement badges display
 * - Hours breakdown by cause
 * - Download impact report button
 *
 * Task: T114
 */
export default function ImpactPage() {
  const { data: dashboard, isLoading, error } = useVolunteerDashboard();

  const handleDownloadReport = async () => {
    try {
      // TODO: Implement actual API call to generate and download PDF report
      // This would call something like: GET /api/v1/volunteers/me/impact-report
      alert('Report download feature coming soon!');
    } catch (error) {
      alert('Failed to download report. Please try again.');
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <div className="h-8 w-48 animate-pulse rounded bg-muted" />
          <div className="grid gap-4 md:grid-cols-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-32 animate-pulse rounded-lg bg-muted" />
            ))}
          </div>
          <div className="h-96 animate-pulse rounded-lg bg-muted" />
        </div>
      </div>
    );
  }

  if (error || !dashboard) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <h2 className="text-2xl font-semibold text-destructive">Failed to Load Impact Data</h2>
            <p className="mt-2 text-muted-foreground">
              Please try refreshing the page or contact support if the issue persists.
            </p>
            <Button onClick={() => window.location.reload()} className="mt-6">
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Calculate hours by cause (mock data for now)
  // TODO: Fetch actual hours breakdown from API
  const hoursByCause = [
    { cause: 'Education', hours: 24, color: 'bg-blue-500' },
    { cause: 'Environment', hours: 18, color: 'bg-green-500' },
    { cause: 'Health', hours: 15, color: 'bg-red-500' },
    { cause: 'Community', hours: 12, color: 'bg-purple-500' },
    { cause: 'Animal Welfare', hours: 8, color: 'bg-orange-500' },
  ];

  const totalCauseHours = hoursByCause.reduce((sum, item) => sum + item.hours, 0);

  const renderAchievementBadge = (achievement: Achievement) => (
    <Card key={achievement.id} className="hover:shadow-md transition-shadow">
      <CardContent className="flex flex-col items-center justify-center p-6 text-center">
        <div className="h-16 w-16 rounded-full bg-primary/10 flex items-center justify-center mb-3">
          {achievement.icon_url ? (
            <img src={achievement.icon_url} alt={achievement.name} className="h-10 w-10" />
          ) : (
            <Trophy className="h-10 w-10 text-primary" />
          )}
        </div>
        <h3 className="font-semibold">{achievement.name}</h3>
        <p className="text-xs text-muted-foreground mt-1">{achievement.description}</p>
        {achievement.earned_at && (
          <Badge variant="secondary" className="mt-3">
            Earned {new Date(achievement.earned_at).toLocaleDateString()}
          </Badge>
        )}
      </CardContent>
    </Card>
  );

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">My Impact</h1>
          <p className="text-muted-foreground mt-2">
            Track your volunteer journey and celebrate your achievements
          </p>
        </div>
        <Button onClick={handleDownloadReport}>
          <Download className="mr-2 h-4 w-4" />
          Download Report
        </Button>
      </div>

      {/* Impact Metrics */}
      <div className="grid gap-4 md:grid-cols-3 mb-8">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Hours</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{dashboard.total_hours}</div>
            <p className="text-xs text-muted-foreground mt-1">
              <TrendingUp className="inline h-3 w-3 mr-1" />
              {dashboard.hours_this_month} hours this month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Events Attended</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{dashboard.total_events}</div>
            <p className="text-xs text-muted-foreground mt-1">
              <TrendingUp className="inline h-3 w-3 mr-1" />
              {dashboard.events_this_month} events this month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Organizations</CardTitle>
            <Building2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{dashboard.total_organizations}</div>
            <p className="text-xs text-muted-foreground mt-1">Different organizations served</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2 mb-8">
        {/* Hours by Cause */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Heart className="h-5 w-5" />
              Hours by Cause
            </CardTitle>
          </CardHeader>
          <CardContent>
            {hoursByCause.length > 0 ? (
              <div className="space-y-4">
                {hoursByCause.map((item) => (
                  <div key={item.cause}>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm font-medium">{item.cause}</span>
                      <span className="text-sm text-muted-foreground">{item.hours} hours</span>
                    </div>
                    <div className="w-full bg-muted rounded-full h-2">
                      <div
                        className={`${item.color} h-2 rounded-full transition-all`}
                        style={{ width: `${(item.hours / totalCauseHours) * 100}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <Target className="h-12 w-12 text-muted-foreground mb-3" />
                <p className="text-sm text-muted-foreground">
                  No hours logged yet. Start volunteering to track your impact!
                </p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Monthly Hours Trend */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Hours Over Time
            </CardTitle>
          </CardHeader>
          <CardContent>
            {/* TODO: Use actual monthly hours data from dashboard.hours_by_month */}
            <HoursChart
              data={[
                { label: 'Jan', hours: 5 },
                { label: 'Feb', hours: 8 },
                { label: 'Mar', hours: 12 },
                { label: 'Apr', hours: 10 },
                { label: 'May', hours: 15 },
                { label: 'Jun', hours: 18 },
              ]}
              title=""
            />
          </CardContent>
        </Card>
      </div>

      {/* Achievements Section */}
      <Card className="mb-8">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Award className="h-5 w-5" />
              Achievements
            </CardTitle>
            <Badge variant="secondary">{dashboard.achievements.length} earned</Badge>
          </div>
        </CardHeader>
        <CardContent>
          {dashboard.achievements.length > 0 ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {dashboard.achievements.map(renderAchievementBadge)}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Trophy className="h-16 w-16 text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold">No Achievements Yet</h3>
              <p className="text-sm text-muted-foreground mt-2 max-w-md">
                Keep volunteering to unlock achievements! Complete events, log hours, and make an
                impact to earn badges.
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Impact Summary */}
      <Card>
        <CardHeader>
          <CardTitle>Your Impact Story</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="prose prose-sm max-w-none">
            <p className="text-muted-foreground">
              You've dedicated <strong>{dashboard.total_hours} hours</strong> of your time to make a
              difference in your community. Through{' '}
              <strong>{dashboard.total_events} volunteer events</strong> with{' '}
              <strong>{dashboard.total_organizations} organizations</strong>, you've contributed to
              causes you care about and helped build a better world.
            </p>
            {dashboard.achievements.length > 0 && (
              <p className="text-muted-foreground mt-4">
                Your dedication has earned you{' '}
                <strong>{dashboard.achievements.length} achievement badges</strong>. Keep up the
                amazing work!
              </p>
            )}
          </div>
          <Separator className="my-6" />
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              Share your impact story with friends and inspire others to volunteer!
            </p>
            <Button variant="outline">Share</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
