'use client';

import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import {
  MapPin,
  Calendar,
  Clock,
  Users,
  Building2,
  Phone,
  Mail,
  Globe,
  Star,
  ArrowLeft,
  Share2,
} from 'lucide-react';
import { useOpportunity } from '@/lib/api';
import { format, parseISO } from 'date-fns';
import dynamic from 'next/dynamic';
import Link from 'next/link';

// Dynamically import map component (only loads on client side)
const OpportunityDetailMap = dynamic(
  () => import('@/components/opportunities/OpportunityDetailMap'),
  {
    ssr: false,
    loading: () => (
      <div className="flex h-[400px] items-center justify-center rounded-md border bg-muted">
        <p className="text-sm text-muted-foreground">Loading map...</p>
      </div>
    ),
  }
);

/**
 * Opportunity Detail Page
 *
 * Displays comprehensive information about a single opportunity:
 * - Opportunity information (title, description, date, location, capacity)
 * - Organization information
 * - Map with location
 * - "Register" button
 * - Reviews from past volunteers (placeholder for now)
 *
 * Task: T112
 */
export default function OpportunityDetailPage() {
  const params = useParams();
  const router = useRouter();
  const opportunityId = params.id as string;

  const { data: opportunity, isLoading, error } = useOpportunity(opportunityId);

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <div className="h-8 w-32 animate-pulse rounded bg-muted" />
          <div className="h-64 animate-pulse rounded-lg bg-muted" />
          <div className="grid gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2 space-y-6">
              <div className="h-96 animate-pulse rounded-lg bg-muted" />
            </div>
            <div className="space-y-6">
              <div className="h-64 animate-pulse rounded-lg bg-muted" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !opportunity) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <h2 className="text-2xl font-semibold text-destructive">Opportunity Not Found</h2>
            <p className="mt-2 text-muted-foreground">
              The opportunity you're looking for doesn't exist or has been removed.
            </p>
            <Button onClick={() => router.push('/opportunities')} className="mt-6">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Opportunities
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  const isFull = opportunity.registered_count >= opportunity.capacity;
  const spotsLeft = opportunity.capacity - opportunity.registered_count;

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <Button variant="ghost" onClick={() => router.back()}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back
        </Button>
        <Button variant="outline" size="icon">
          <Share2 className="h-4 w-4" />
        </Button>
      </div>

      {/* Hero Section */}
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="space-y-2 flex-1">
              <div className="flex items-center gap-2">
                <Badge
                  variant={
                    opportunity.status === 'published'
                      ? 'default'
                      : opportunity.status === 'cancelled'
                        ? 'destructive'
                        : 'secondary'
                  }
                >
                  {opportunity.status}
                </Badge>
                {opportunity.is_recurring && <Badge variant="outline">Recurring</Badge>}
                {isFull && <Badge variant="destructive">Full</Badge>}
              </div>
              <CardTitle className="text-3xl">{opportunity.title}</CardTitle>
              <div className="flex flex-wrap gap-4 text-muted-foreground">
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  <span>{format(parseISO(opportunity.start_time), 'PPP')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4" />
                  <span>
                    {format(parseISO(opportunity.start_time), 'p')} -{' '}
                    {format(parseISO(opportunity.end_time), 'p')}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <MapPin className="h-4 w-4" />
                  <span>
                    {opportunity.city}, {opportunity.state}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              <span>
                {opportunity.registered_count} / {opportunity.capacity} registered
              </span>
            </div>
            {!isFull && (
              <Badge variant="secondary" className="ml-auto">
                {spotsLeft} {spotsLeft === 1 ? 'spot' : 'spots'} left
              </Badge>
            )}
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Description */}
          <Card>
            <CardHeader>
              <CardTitle>About This Opportunity</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="whitespace-pre-wrap text-muted-foreground">{opportunity.description}</p>
            </CardContent>
          </Card>

          {/* Map */}
          {opportunity.latitude && opportunity.longitude && (
            <Card>
              <CardHeader>
                <CardTitle>Location</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-start gap-2">
                    <MapPin className="h-5 w-5 text-muted-foreground mt-0.5" />
                    <div>
                      <p className="font-medium">{opportunity.address}</p>
                      <p className="text-sm text-muted-foreground">
                        {opportunity.city}, {opportunity.state} {opportunity.zip_code}
                      </p>
                    </div>
                  </div>
                  <OpportunityDetailMap
                    latitude={opportunity.latitude}
                    longitude={opportunity.longitude}
                    title={opportunity.title}
                    address={opportunity.address}
                  />
                </div>
              </CardContent>
            </Card>
          )}

          {/* Requirements */}
          {(opportunity.required_skills.length > 0 || opportunity.min_age) && (
            <Card>
              <CardHeader>
                <CardTitle>Requirements</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {opportunity.min_age && (
                    <div>
                      <h4 className="font-medium mb-2">Age Requirement</h4>
                      <p className="text-sm text-muted-foreground">
                        Minimum age: {opportunity.min_age} years old
                      </p>
                    </div>
                  )}
                  {opportunity.required_skills.length > 0 && (
                    <div>
                      <h4 className="font-medium mb-2">Required Skills</h4>
                      <div className="flex flex-wrap gap-2">
                        {opportunity.required_skills.map((skill) => (
                          <Badge key={skill} variant="secondary">
                            {skill}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Reviews Section - Placeholder */}
          <Card>
            <CardHeader>
              <CardTitle>Reviews from Past Volunteers</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <Star className="h-12 w-12 text-muted-foreground mb-3" />
                <p className="text-sm text-muted-foreground">
                  No reviews yet. Be the first to volunteer and share your experience!
                </p>
              </div>
              {/* TODO: Implement reviews feature in future phase */}
            </CardContent>
          </Card>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Registration Card */}
          <Card>
            <CardHeader>
              <CardTitle>Register for Event</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Status</span>
                  <Badge variant={isFull ? 'destructive' : 'default'}>
                    {isFull ? 'Full' : 'Available'}
                  </Badge>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Capacity</span>
                  <span className="font-medium">
                    {opportunity.registered_count} / {opportunity.capacity}
                  </span>
                </div>
                {opportunity.waitlist_count > 0 && (
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Waitlist</span>
                    <span className="font-medium">{opportunity.waitlist_count}</span>
                  </div>
                )}
              </div>
              <Separator />
              <Button className="w-full" disabled={opportunity.status !== 'published'} size="lg">
                {isFull ? 'Join Waitlist' : 'Register Now'}
              </Button>
              <p className="text-xs text-center text-muted-foreground">
                {opportunity.status === 'published'
                  ? 'By registering, you commit to attending this event'
                  : 'Registration is currently unavailable'}
              </p>
            </CardContent>
          </Card>

          {/* Organization Info */}
          <Card>
            <CardHeader>
              <CardTitle>Hosted By</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-start gap-3">
                <Building2 className="h-5 w-5 text-muted-foreground mt-0.5" />
                <div className="flex-1">
                  <h4 className="font-medium">Organization Name</h4>
                  <p className="text-sm text-muted-foreground">View organization profile →</p>
                </div>
              </div>
              {/* TODO: Fetch and display actual organization details */}
              <Separator />
              <div className="space-y-3 text-sm">
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Mail className="h-4 w-4" />
                  <span>contact@example.org</span>
                </div>
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Phone className="h-4 w-4" />
                  <span>(555) 123-4567</span>
                </div>
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Globe className="h-4 w-4" />
                  <Link href="#" className="hover:underline">
                    www.example.org
                  </Link>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Causes */}
          {opportunity.causes.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Causes</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {opportunity.causes.map((cause) => (
                    <Badge key={cause} variant="outline">
                      {cause}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
