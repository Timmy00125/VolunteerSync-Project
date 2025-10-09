'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { useOpportunity, useOpportunityRegistrations } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ArrowLeft, Clock, Save, CheckCircle, AlertCircle } from 'lucide-react';
import { format, parseISO } from 'date-fns';
import type { Registration } from '@/lib/api/types';
import apiClient from '@/lib/api/client';

/**
 * Hours Logging Page (T120)
 *
 * Coordinator logs volunteer hours after an event:
 * - View table of checked-in volunteers
 * - Input hours for each volunteer
 * - Add coordinator notes
 * - Submit hours (creates pending records)
 *
 * Requirements:
 * - Hours logged with status 'pending' (FR-024)
 * - Volunteer receives notification (FR-025)
 * - Volunteer can verify or dispute (FR-026, FR-027)
 * - Auto-verify after 7 days (FR-028)
 */
export default function HoursLoggingPage() {
  const params = useParams();
  const router = useRouter();
  const opportunityId = params.id as string;

  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: opportunity, isLoading: loadingOpportunity } = useOpportunity(opportunityId);
  const { data: registrations, isLoading: loadingRegistrations } =
    useOpportunityRegistrations(opportunityId);

  const isLoading = loadingOpportunity || loadingRegistrations;

  // Filter to only checked-in volunteers
  const checkedInVolunteers = registrations?.filter((r: Registration) => r.status === 'checked_in');

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm();

  const onSubmit = async (data: any) => {
    setSubmitting(true);
    setError(null);

    try {
      // Create hours log entries for each volunteer
      const hoursEntries = checkedInVolunteers?.map((registration: Registration) => ({
        registration_id: registration.id,
        volunteer_id: registration.volunteer_id,
        opportunity_id: opportunityId,
        hours: parseFloat(data[`hours_${registration.id}`] || '0'),
        coordinator_notes: data[`notes_${registration.id}`] || '',
      }));

      // Submit to API
      await apiClient.post('/hours/batch-log', { hours_entries: hoursEntries });

      setSuccess(true);
      setTimeout(() => {
        router.push(`/organization/opportunities/${opportunityId}/roster`);
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to log hours. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-8">
        <div className="h-8 w-64 animate-pulse rounded bg-muted" />
        <Card>
          <CardContent className="py-12 text-center">
            <div className="text-muted-foreground">Loading volunteers...</div>
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

  if (!checkedInVolunteers || checkedInVolunteers.length === 0) {
    return (
      <div className="space-y-8">
        <div>
          <Button variant="ghost" size="sm" onClick={() => router.back()} className="mb-4">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Roster
          </Button>
          <h1 className="text-3xl font-bold tracking-tight">Log Volunteer Hours</h1>
        </div>
        <Card>
          <CardContent className="py-12 text-center">
            <Clock className="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
            <h3 className="mb-2 text-lg font-semibold">No Checked-In Volunteers</h3>
            <p className="text-sm text-muted-foreground">
              You can only log hours for volunteers who have been checked in
            </p>
            <Button className="mt-4" onClick={() => router.back()}>
              Return to Roster
            </Button>
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
          Back to Roster
        </Button>
        <h1 className="text-3xl font-bold tracking-tight">Log Volunteer Hours</h1>
        <p className="text-muted-foreground">
          {opportunity.title} • {format(parseISO(opportunity.start_time), 'MMMM d, yyyy')}
        </p>
      </div>

      {/* Success Message */}
      {success && (
        <Card className="border-green-500">
          <CardContent className="flex items-center gap-2 py-4">
            <CheckCircle className="h-5 w-5 text-green-600" />
            <div>
              <p className="font-medium text-green-600">Hours logged successfully!</p>
              <p className="text-sm text-muted-foreground">
                Volunteers have been notified and can now verify their hours.
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Error Message */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="flex items-center gap-2 py-4">
            <AlertCircle className="h-5 w-5 text-destructive" />
            <p className="text-sm text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Instructions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5" />
            Instructions
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>• Enter the number of hours each volunteer worked during this event</p>
          <p>• Hours will be marked as &quot;pending&quot; and volunteers will be notified</p>
          <p>• Volunteers have 7 days to verify or dispute the logged hours</p>
          <p>• Hours will be automatically verified after 7 days if no action is taken</p>
        </CardContent>
      </Card>

      {/* Hours Form */}
      <form onSubmit={handleSubmit(onSubmit)}>
        <Card>
          <CardHeader>
            <CardTitle>Volunteer Hours ({checkedInVolunteers.length} volunteers)</CardTitle>
            <CardDescription>Enter hours for each checked-in volunteer</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              {checkedInVolunteers.map((registration: Registration) => (
                <div key={registration.id} className="rounded-lg border p-4 space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium">Volunteer #{registration.volunteer_id}</p>
                      <p className="text-sm text-muted-foreground">
                        Checked in:{' '}
                        {registration.checked_in_at
                          ? format(parseISO(registration.checked_in_at), 'MMM d, h:mm a')
                          : 'N/A'}
                      </p>
                    </div>
                    <span className="rounded-full bg-green-50 px-2.5 py-0.5 text-xs font-medium text-green-700">
                      Checked In
                    </span>
                  </div>

                  <div className="grid gap-4 md:grid-cols-2">
                    {/* Hours Input */}
                    <div className="space-y-2">
                      <Label htmlFor={`hours_${registration.id}`}>
                        Hours Worked <span className="text-destructive">*</span>
                      </Label>
                      <Input
                        id={`hours_${registration.id}`}
                        type="number"
                        step="0.5"
                        min="0"
                        max="24"
                        placeholder="4.5"
                        {...register(`hours_${registration.id}`, {
                          required: 'Hours required',
                          min: { value: 0, message: 'Hours must be positive' },
                          max: { value: 24, message: 'Hours cannot exceed 24' },
                        })}
                      />
                      {errors[`hours_${registration.id}`] && (
                        <p className="text-sm text-destructive">
                          {errors[`hours_${registration.id}`]?.message as string}
                        </p>
                      )}
                    </div>

                    {/* Notes Input */}
                    <div className="space-y-2">
                      <Label htmlFor={`notes_${registration.id}`}>
                        Coordinator Notes (Optional)
                      </Label>
                      <Input
                        id={`notes_${registration.id}`}
                        placeholder="e.g., Arrived late, left early"
                        {...register(`notes_${registration.id}`)}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Submit Button */}
            <div className="mt-6 flex items-center gap-4">
              <Button type="submit" size="lg" disabled={submitting || success}>
                {submitting ? (
                  <>
                    <Clock className="mr-2 h-4 w-4 animate-spin" />
                    Logging Hours...
                  </>
                ) : (
                  <>
                    <Save className="mr-2 h-4 w-4" />
                    Submit Hours
                  </>
                )}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => router.back()}
                disabled={submitting}
              >
                Cancel
              </Button>
            </div>
          </CardContent>
        </Card>
      </form>

      {/* Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">What happens next?</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>1. Hours are logged with &quot;pending&quot; status</p>
          <p>2. Each volunteer receives a notification to verify their hours</p>
          <p>3. Volunteers can verify (approve) or dispute the logged hours</p>
          <p>4. Disputed hours require coordinator resolution</p>
          <p>5. Hours are automatically verified after 7 days if no action is taken</p>
          <p>6. Verified hours are added to the volunteer&apos;s total impact metrics</p>
        </CardContent>
      </Card>
    </div>
  );
}
