'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Calendar,
  Clock,
  MapPin,
  Building2,
  CheckCircle,
  XCircle,
  Star,
  Download,
  AlertCircle,
} from 'lucide-react';
import { useMyRegistrations, useCheckIn, useCancelRegistration } from '@/lib/api';
import { format, parseISO, isPast, isToday, isFuture } from 'date-fns';
import Link from 'next/link';
import type { Registration } from '@/lib/api/types';

/**
 * My Events Page
 *
 * Displays volunteer's event registrations with tabs:
 * - Upcoming: Future events
 * - Past: Completed events
 * - Cancelled: Cancelled registrations
 *
 * Features:
 * - Check-in button (on event day)
 * - Hours verification status
 * - Write review (after completion)
 * - Cancel registration with reason
 *
 * Task: T113
 */
export default function MyEventsPage() {
  const [activeTab, setActiveTab] = useState<'upcoming' | 'past' | 'cancelled'>('upcoming');
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [selectedRegistration, setSelectedRegistration] = useState<Registration | null>(null);
  const [cancellationReason, setCancellationReason] = useState('');

  // Fetch registrations based on active tab
  const {
    data: upcomingData,
    isLoading: upcomingLoading,
    error: upcomingError,
  } = useMyRegistrations({
    status: activeTab === 'upcoming' ? 'registered' : undefined,
  });

  const {
    data: pastData,
    isLoading: pastLoading,
    error: pastError,
  } = useMyRegistrations({
    status: activeTab === 'past' ? 'completed' : undefined,
  });

  const {
    data: cancelledData,
    isLoading: cancelledLoading,
    error: cancelledError,
  } = useMyRegistrations({
    status: activeTab === 'cancelled' ? 'cancelled' : undefined,
  });

  const checkInMutation = useCheckIn();
  const cancelMutation = useCancelRegistration();

  const handleCheckIn = async (registrationId: string) => {
    try {
      await checkInMutation.mutateAsync(registrationId);
      // Show success toast notification
      alert('Successfully checked in!');
    } catch (error) {
      alert('Failed to check in. Please try again.');
    }
  };

  const handleCancelClick = (registration: Registration) => {
    setSelectedRegistration(registration);
    setCancelDialogOpen(true);
  };

  const handleCancelConfirm = async () => {
    if (!selectedRegistration) return;

    try {
      await cancelMutation.mutateAsync({
        registrationId: selectedRegistration.id,
        reason: cancellationReason,
      });
      setCancelDialogOpen(false);
      setSelectedRegistration(null);
      setCancellationReason('');
      alert('Registration cancelled successfully.');
    } catch (error) {
      alert('Failed to cancel registration. Please try again.');
    }
  };

  const renderEventCard = (registration: Registration, showActions: boolean = true) => {
    // TODO: Fetch actual opportunity details
    // For now, using mock data structure
    const eventDate = parseISO(registration.created_at); // Placeholder
    const canCheckIn = isToday(eventDate) && registration.status === 'registered';
    const isCompleted = registration.status === 'completed';
    const isCancelled = registration.status === 'cancelled';

    return (
      <Card key={registration.id} className="hover:shadow-md transition-shadow">
        <CardContent className="p-6">
          <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
            {/* Event Info */}
            <div className="flex-1 space-y-3">
              <div>
                <Link
                  href={`/opportunities/${registration.opportunity_id}`}
                  className="text-lg font-semibold hover:underline"
                >
                  Opportunity Title
                  {/* TODO: Replace with actual opportunity.title */}
                </Link>
                <div className="flex items-center gap-2 mt-1 text-sm text-muted-foreground">
                  <Building2 className="h-4 w-4" />
                  <span>Organization Name</span>
                  {/* TODO: Replace with actual organization.name */}
                </div>
              </div>

              <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  <span>{format(eventDate, 'PPP')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4" />
                  <span>9:00 AM - 5:00 PM</span>
                  {/* TODO: Replace with actual times */}
                </div>
                <div className="flex items-center gap-2">
                  <MapPin className="h-4 w-4" />
                  <span>Location Name</span>
                  {/* TODO: Replace with actual location */}
                </div>
              </div>

              {/* Status Badges */}
              <div className="flex flex-wrap gap-2">
                <Badge
                  variant={
                    registration.status === 'completed'
                      ? 'default'
                      : registration.status === 'cancelled'
                        ? 'destructive'
                        : registration.status === 'checked_in'
                          ? 'secondary'
                          : 'outline'
                  }
                >
                  {registration.status.replace('_', ' ').toUpperCase()}
                </Badge>

                {registration.hours_status && (
                  <Badge
                    variant={
                      registration.hours_status === 'verified'
                        ? 'default'
                        : registration.hours_status === 'disputed'
                          ? 'destructive'
                          : 'secondary'
                    }
                  >
                    Hours: {registration.hours_status}
                  </Badge>
                )}

                {registration.hours_logged && (
                  <Badge variant="outline">{registration.hours_logged} hours logged</Badge>
                )}
              </div>
            </div>

            {/* Actions */}
            {showActions && (
              <div className="flex flex-col gap-2 md:min-w-[160px]">
                {canCheckIn && (
                  <Button
                    onClick={() => handleCheckIn(registration.id)}
                    disabled={checkInMutation.isPending}
                    className="w-full"
                  >
                    <CheckCircle className="mr-2 h-4 w-4" />
                    Check In
                  </Button>
                )}

                {isCompleted && (
                  <>
                    <Button variant="outline" className="w-full">
                      <Star className="mr-2 h-4 w-4" />
                      Write Review
                    </Button>
                    <Button variant="outline" className="w-full">
                      <Download className="mr-2 h-4 w-4" />
                      Download .ics
                    </Button>
                  </>
                )}

                {registration.status === 'registered' && (
                  <Button
                    variant="outline"
                    onClick={() => handleCancelClick(registration)}
                    disabled={cancelMutation.isPending}
                    className="w-full text-destructive hover:text-destructive"
                  >
                    <XCircle className="mr-2 h-4 w-4" />
                    Cancel
                  </Button>
                )}

                {registration.hours_status === 'pending' && (
                  <Button variant="outline" className="w-full">
                    <AlertCircle className="mr-2 h-4 w-4" />
                    Hours Details
                  </Button>
                )}
              </div>
            )}
          </div>

          {/* Cancellation Info */}
          {isCancelled && registration.cancelled_at && (
            <div className="mt-4 p-3 rounded-md bg-destructive/10 text-destructive text-sm">
              <p className="font-medium">
                Cancelled on {format(parseISO(registration.cancelled_at), 'PPP')}
              </p>
              {registration.cancellation_reason && (
                <p className="mt-1 text-muted-foreground">
                  Reason: {registration.cancellation_reason}
                </p>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    );
  };

  const renderTabContent = (
    data: typeof upcomingData,
    isLoading: boolean,
    error: any,
    emptyMessage: string
  ) => {
    if (isLoading) {
      return (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="space-y-3">
                  <div className="h-6 w-3/4 animate-pulse rounded bg-muted" />
                  <div className="h-4 w-1/2 animate-pulse rounded bg-muted" />
                  <div className="h-4 w-2/3 animate-pulse rounded bg-muted" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      );
    }

    if (error) {
      return (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <AlertCircle className="h-12 w-12 text-destructive mb-3" />
            <h3 className="text-lg font-semibold text-destructive">Failed to Load Events</h3>
            <p className="text-sm text-muted-foreground mt-2">Please try refreshing the page.</p>
          </CardContent>
        </Card>
      );
    }

    if (!data || data.data.length === 0) {
      return (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <Calendar className="h-12 w-12 text-muted-foreground mb-3" />
            <h3 className="text-lg font-semibold">{emptyMessage}</h3>
            <p className="text-sm text-muted-foreground mt-2">
              Explore opportunities to find events to join.
            </p>
            <Button asChild className="mt-4">
              <Link href="/volunteer/opportunities">Browse Opportunities</Link>
            </Button>
          </CardContent>
        </Card>
      );
    }

    return (
      <div className="space-y-4">
        {data.data.map((registration) => renderEventCard(registration, activeTab !== 'cancelled'))}
      </div>
    );
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">My Events</h1>
        <p className="text-muted-foreground mt-2">
          View and manage your volunteer event registrations
        </p>
      </div>

      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)}>
        <TabsList className="grid w-full max-w-md grid-cols-3">
          <TabsTrigger value="upcoming">
            Upcoming
            {upcomingData && upcomingData.data.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {upcomingData.data.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="past">
            Past
            {pastData && pastData.data.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {pastData.data.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="cancelled">
            Cancelled
            {cancelledData && cancelledData.data.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {cancelledData.data.length}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="upcoming" className="mt-6">
          {renderTabContent(upcomingData, upcomingLoading, upcomingError, 'No upcoming events')}
        </TabsContent>

        <TabsContent value="past" className="mt-6">
          {renderTabContent(pastData, pastLoading, pastError, 'No past events')}
        </TabsContent>

        <TabsContent value="cancelled" className="mt-6">
          {renderTabContent(cancelledData, cancelledLoading, cancelledError, 'No cancelled events')}
        </TabsContent>
      </Tabs>

      {/* Cancel Registration Dialog */}
      <Dialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Cancel Registration</DialogTitle>
            <DialogDescription>
              Are you sure you want to cancel your registration for this event? This action cannot
              be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="reason">Reason for Cancellation (Optional)</Label>
              <Textarea
                id="reason"
                placeholder="Let the organization know why you need to cancel..."
                value={cancellationReason}
                onChange={(e) => setCancellationReason(e.target.value)}
                rows={4}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setCancelDialogOpen(false);
                setSelectedRegistration(null);
                setCancellationReason('');
              }}
            >
              Keep Registration
            </Button>
            <Button
              variant="destructive"
              onClick={handleCancelConfirm}
              disabled={cancelMutation.isPending}
            >
              {cancelMutation.isPending ? 'Cancelling...' : 'Cancel Registration'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
