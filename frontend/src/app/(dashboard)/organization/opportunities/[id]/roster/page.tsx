'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useOpportunity, useOpportunityRegistrations, useCheckIn } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  ArrowLeft,
  Users,
  Search,
  CheckCircle,
  Clock,
  MessageSquare,
  FileText,
  Download,
} from 'lucide-react';
import { format, parseISO } from 'date-fns';
import type { Registration } from '@/lib/api/types';

/**
 * Event Roster Page (T119)
 *
 * Manage volunteers registered for a specific opportunity:
 * - View list of registered volunteers with status
 * - Check in volunteers on event day
 * - Send broadcast messages to all volunteers
 * - View volunteer details
 * - Export roster to CSV
 *
 * Statuses:
 * - registered: Confirmed registration
 * - checked_in: Volunteer has arrived
 * - completed: Event completed
 * - cancelled: Volunteer cancelled
 * - waitlisted: On waitlist (at capacity)
 */
export default function EventRosterPage() {
  const params = useParams();
  const router = useRouter();
  const opportunityId = params.id as string;

  const [searchQuery, setSearchQuery] = useState('');

  const { data: opportunity, isLoading: loadingOpportunity } = useOpportunity(opportunityId);
  const { data: registrations, isLoading: loadingRegistrations } =
    useOpportunityRegistrations(opportunityId);
  const checkInMutation = useCheckIn();

  const isLoading = loadingOpportunity || loadingRegistrations;

  // Filter registrations by search query
  const filteredRegistrations = registrations?.filter((reg: Registration) => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    // Note: In real implementation, we'd have volunteer name from the registration object
    return reg.id.toLowerCase().includes(query) || reg.status.toLowerCase().includes(query);
  });

  // Group registrations by status
  const registeredVolunteers = filteredRegistrations?.filter(
    (r: Registration) => r.status === 'registered'
  );
  const checkedInVolunteers = filteredRegistrations?.filter(
    (r: Registration) => r.status === 'checked_in'
  );
  const waitlistedVolunteers = filteredRegistrations?.filter(
    (r: Registration) => r.status === 'waitlisted'
  );
  const cancelledVolunteers = filteredRegistrations?.filter(
    (r: Registration) => r.status === 'cancelled'
  );

  const handleCheckIn = async (registrationId: string) => {
    try {
      await checkInMutation.mutateAsync(registrationId);
    } catch (error) {
      console.error('Failed to check in volunteer:', error);
    }
  };

  const handleSendMessage = () => {
    // TODO: Navigate to message compose with all registered volunteers
    router.push(`/organization/opportunities/${opportunityId}/message`);
  };

  const handleExportRoster = () => {
    // TODO: Implement CSV export
    console.log('Exporting roster...');
  };

  if (isLoading) {
    return (
      <div className="space-y-8">
        <div className="h-8 w-64 animate-pulse rounded bg-muted" />
        <Card>
          <CardContent className="py-12 text-center">
            <div className="text-muted-foreground">Loading roster...</div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!opportunity) {
    return (
      <div className="space-y-8">
        <Card className="border-destructive">
          <CardContent className="py-12 text-center">
            <p className="text-destructive">Opportunity not found</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <Button variant="ghost" size="sm" onClick={() => router.back()} className="mb-4">
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Opportunities
        </Button>
        <h1 className="text-3xl font-bold tracking-tight">{opportunity.title} - Roster</h1>
        <p className="text-muted-foreground">
          {format(parseISO(opportunity.start_time), 'MMMM d, yyyy')} • {opportunity.city},{' '}
          {opportunity.state}
        </p>
      </div>

      {/* Stats and Actions */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Registered</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {opportunity.registered_count}/{opportunity.capacity}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Checked In</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{checkedInVolunteers?.length || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Waitlisted</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{opportunity.waitlist_count}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Cancelled</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{cancelledVolunteers?.length || 0}</div>
          </CardContent>
        </Card>
      </div>

      {/* Actions Bar */}
      <div className="flex items-center justify-between gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search volunteers..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleExportRoster}>
            <Download className="mr-2 h-4 w-4" />
            Export CSV
          </Button>
          <Button onClick={handleSendMessage}>
            <MessageSquare className="mr-2 h-4 w-4" />
            Message All
          </Button>
          <Button asChild>
            <Link href={`/organization/opportunities/${opportunityId}/hours`}>
              <FileText className="mr-2 h-4 w-4" />
              Log Hours
            </Link>
          </Button>
        </div>
      </div>

      {/* Volunteer Lists */}
      <div className="space-y-6">
        {/* Registered Volunteers */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Registered Volunteers ({registeredVolunteers?.length || 0})
            </CardTitle>
            <CardDescription>Volunteers confirmed for this event</CardDescription>
          </CardHeader>
          <CardContent>
            {!registeredVolunteers || registeredVolunteers.length === 0 ? (
              <p className="text-center text-sm text-muted-foreground">
                No registered volunteers yet
              </p>
            ) : (
              <div className="space-y-2">
                {registeredVolunteers.map((registration) => (
                  <VolunteerRow
                    key={registration.id}
                    registration={registration}
                    onCheckIn={() => handleCheckIn(registration.id)}
                  />
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Checked In Volunteers */}
        {checkedInVolunteers && checkedInVolunteers.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CheckCircle className="h-5 w-5 text-green-600" />
                Checked In ({checkedInVolunteers.length})
              </CardTitle>
              <CardDescription>Volunteers who have arrived</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {checkedInVolunteers.map((registration) => (
                  <VolunteerRow key={registration.id} registration={registration} isCheckedIn />
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Waitlisted Volunteers */}
        {waitlistedVolunteers && waitlistedVolunteers.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Clock className="h-5 w-5 text-yellow-600" />
                Waitlisted ({waitlistedVolunteers.length})
              </CardTitle>
              <CardDescription>Volunteers on the waitlist</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {waitlistedVolunteers.map((registration) => (
                  <VolunteerRow key={registration.id} registration={registration} />
                ))}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}

/**
 * Volunteer Row Component
 */
interface VolunteerRowProps {
  registration: Registration;
  isCheckedIn?: boolean;
  onCheckIn?: () => void;
}

function VolunteerRow({ registration, isCheckedIn, onCheckIn }: VolunteerRowProps) {
  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-4">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">
          <Users className="h-5 w-5 text-muted-foreground" />
        </div>
        <div>
          <p className="font-medium">Volunteer #{registration.volunteer_id}</p>
          <p className="text-sm text-muted-foreground">
            Registered: {format(parseISO(registration.created_at), 'MMM d, yyyy')}
          </p>
          {registration.checked_in_at && (
            <p className="text-sm text-green-600">
              Checked in: {format(parseISO(registration.checked_in_at), 'MMM d, h:mm a')}
            </p>
          )}
        </div>
      </div>
      <div className="flex items-center gap-2">
        <span
          className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
            registration.status === 'registered'
              ? 'bg-blue-50 text-blue-700'
              : registration.status === 'checked_in'
                ? 'bg-green-50 text-green-700'
                : registration.status === 'waitlisted'
                  ? 'bg-yellow-50 text-yellow-700'
                  : 'bg-gray-50 text-gray-700'
          }`}
        >
          {registration.status}
        </span>
        {!isCheckedIn && onCheckIn && registration.status === 'registered' && (
          <Button size="sm" onClick={onCheckIn}>
            <CheckCircle className="mr-2 h-4 w-4" />
            Check In
          </Button>
        )}
      </div>
    </div>
  );
}
