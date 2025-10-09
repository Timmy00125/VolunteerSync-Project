'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useSearchOpportunities } from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Calendar,
  MapPin,
  Users,
  Plus,
  Edit,
  Trash2,
  CheckCircle,
  XCircle,
  Eye,
} from 'lucide-react';
import { format, parseISO } from 'date-fns';
import type { Opportunity } from '@/lib/api/types';

/**
 * Opportunity Management Page (T118)
 *
 * Organization's opportunity management interface with tabs:
 * - Published: Active opportunities that volunteers can see
 * - Draft: Unpublished opportunities still being edited
 * - Completed: Past opportunities that have ended
 * - Cancelled: Opportunities that were cancelled
 *
 * Features:
 * - View, edit, cancel, complete actions
 * - View registrations for each opportunity
 * - Create new opportunities
 */
export default function OpportunityManagementPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<'published' | 'draft' | 'completed' | 'cancelled'>(
    'published'
  );

  // Fetch opportunities - in real implementation, filter by organization and status
  const {
    data: opportunitiesData,
    isLoading,
    error,
  } = useSearchOpportunities({
    // TODO: Add organization filter and status filter based on active tab
    page: 1,
    limit: 50,
  });

  // Filter opportunities by status (client-side for now)
  const opportunities = opportunitiesData?.data || [];
  const filteredOpportunities = opportunities.filter((opp) => opp.status === activeTab);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Manage Opportunities</h1>
          <p className="text-muted-foreground">View and manage your volunteer opportunities</p>
        </div>
        <Button asChild>
          <Link href="/organization/opportunities/new">
            <Plus className="mr-2 h-4 w-4" />
            Create Opportunity
          </Link>
        </Button>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={(value: any) => setActiveTab(value)}>
        <TabsList>
          <TabsTrigger value="published">
            <CheckCircle className="mr-2 h-4 w-4" />
            Published
          </TabsTrigger>
          <TabsTrigger value="draft">
            <Edit className="mr-2 h-4 w-4" />
            Draft
          </TabsTrigger>
          <TabsTrigger value="completed">
            <CheckCircle className="mr-2 h-4 w-4" />
            Completed
          </TabsTrigger>
          <TabsTrigger value="cancelled">
            <XCircle className="mr-2 h-4 w-4" />
            Cancelled
          </TabsTrigger>
        </TabsList>

        {/* Published Tab */}
        <TabsContent value="published" className="mt-6">
          {isLoading ? (
            <OpportunityListSkeleton />
          ) : error ? (
            <Card className="border-destructive">
              <CardContent className="py-8 text-center">
                <p className="text-sm text-destructive">
                  Failed to load opportunities. Please try again.
                </p>
              </CardContent>
            </Card>
          ) : filteredOpportunities.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <Calendar className="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
                <h3 className="mb-2 text-lg font-semibold">No Published Opportunities</h3>
                <p className="mb-4 text-sm text-muted-foreground">
                  Create your first volunteer opportunity to start recruiting volunteers
                </p>
                <Button asChild>
                  <Link href="/organization/opportunities/new">
                    <Plus className="mr-2 h-4 w-4" />
                    Create Opportunity
                  </Link>
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {filteredOpportunities.map((opportunity) => (
                <OpportunityCard
                  key={opportunity.id}
                  opportunity={opportunity}
                  onEdit={() => router.push(`/organization/opportunities/${opportunity.id}/edit`)}
                  onViewRoster={() =>
                    router.push(`/organization/opportunities/${opportunity.id}/roster`)
                  }
                  onCancel={() => {
                    /* TODO: Implement cancel logic */
                  }}
                />
              ))}
            </div>
          )}
        </TabsContent>

        {/* Draft Tab */}
        <TabsContent value="draft" className="mt-6">
          {isLoading ? (
            <OpportunityListSkeleton />
          ) : filteredOpportunities.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <Edit className="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
                <h3 className="mb-2 text-lg font-semibold">No Draft Opportunities</h3>
                <p className="text-sm text-muted-foreground">
                  Draft opportunities will appear here before publishing
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {filteredOpportunities.map((opportunity) => (
                <OpportunityCard
                  key={opportunity.id}
                  opportunity={opportunity}
                  isDraft
                  onEdit={() => router.push(`/organization/opportunities/${opportunity.id}/edit`)}
                  onPublish={() => {
                    /* TODO: Implement publish logic */
                  }}
                  onDelete={() => {
                    /* TODO: Implement delete logic */
                  }}
                />
              ))}
            </div>
          )}
        </TabsContent>

        {/* Completed Tab */}
        <TabsContent value="completed" className="mt-6">
          {isLoading ? (
            <OpportunityListSkeleton />
          ) : filteredOpportunities.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <CheckCircle className="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
                <h3 className="mb-2 text-lg font-semibold">No Completed Opportunities</h3>
                <p className="text-sm text-muted-foreground">
                  Completed opportunities will appear here
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {filteredOpportunities.map((opportunity) => (
                <OpportunityCard
                  key={opportunity.id}
                  opportunity={opportunity}
                  isCompleted
                  onViewRoster={() =>
                    router.push(`/organization/opportunities/${opportunity.id}/roster`)
                  }
                />
              ))}
            </div>
          )}
        </TabsContent>

        {/* Cancelled Tab */}
        <TabsContent value="cancelled" className="mt-6">
          {isLoading ? (
            <OpportunityListSkeleton />
          ) : filteredOpportunities.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <XCircle className="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
                <h3 className="mb-2 text-lg font-semibold">No Cancelled Opportunities</h3>
                <p className="text-sm text-muted-foreground">
                  Cancelled opportunities will appear here
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {filteredOpportunities.map((opportunity) => (
                <OpportunityCard key={opportunity.id} opportunity={opportunity} isCancelled />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

/**
 * Opportunity Card Component
 */
interface OpportunityCardProps {
  opportunity: Opportunity;
  isDraft?: boolean;
  isCompleted?: boolean;
  isCancelled?: boolean;
  onEdit?: () => void;
  onViewRoster?: () => void;
  onPublish?: () => void;
  onCancel?: () => void;
  onDelete?: () => void;
}

function OpportunityCard({
  opportunity,
  isDraft,
  isCompleted,
  isCancelled,
  onEdit,
  onViewRoster,
  onPublish,
  onCancel,
  onDelete,
}: OpportunityCardProps) {
  const capacityPercentage = (opportunity.registered_count / opportunity.capacity) * 100;

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="flex-1 space-y-3">
            {/* Title and Status */}
            <div className="flex items-start gap-3">
              <div className="flex-1">
                <h3 className="text-lg font-semibold">{opportunity.title}</h3>
                <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
                  {opportunity.description}
                </p>
              </div>
              <span
                className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                  opportunity.status === 'published'
                    ? 'bg-green-50 text-green-700'
                    : opportunity.status === 'draft'
                      ? 'bg-yellow-50 text-yellow-700'
                      : opportunity.status === 'completed'
                        ? 'bg-blue-50 text-blue-700'
                        : 'bg-red-50 text-red-700'
                }`}
              >
                {opportunity.status}
              </span>
            </div>

            {/* Details */}
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1">
                <Calendar className="h-4 w-4" />
                {format(parseISO(opportunity.start_time), 'MMM d, yyyy')}
              </div>
              <div className="flex items-center gap-1">
                <MapPin className="h-4 w-4" />
                {opportunity.city}, {opportunity.state}
              </div>
              {!isDraft && (
                <div className="flex items-center gap-1">
                  <Users className="h-4 w-4" />
                  {opportunity.registered_count}/{opportunity.capacity} registered
                  {opportunity.waitlist_count > 0 && (
                    <span className="text-yellow-600">
                      ({opportunity.waitlist_count} waitlisted)
                    </span>
                  )}
                </div>
              )}
            </div>

            {/* Capacity Bar (for published opportunities) */}
            {!isDraft && !isCancelled && (
              <div className="space-y-1">
                <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                  <div
                    className={`h-full transition-all ${
                      capacityPercentage >= 100
                        ? 'bg-red-500'
                        : capacityPercentage >= 75
                          ? 'bg-yellow-500'
                          : 'bg-green-500'
                    }`}
                    style={{ width: `${Math.min(capacityPercentage, 100)}%` }}
                  />
                </div>
                {capacityPercentage >= 100 && <p className="text-xs text-red-600">At capacity</p>}
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="ml-4 flex flex-col gap-2">
            {isDraft ? (
              <>
                <Button size="sm" onClick={onPublish}>
                  Publish
                </Button>
                <Button size="sm" variant="outline" onClick={onEdit}>
                  <Edit className="mr-2 h-4 w-4" />
                  Edit
                </Button>
                <Button size="sm" variant="destructive" onClick={onDelete}>
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </Button>
              </>
            ) : isCompleted || isCancelled ? (
              <Button size="sm" variant="outline" onClick={onViewRoster}>
                <Eye className="mr-2 h-4 w-4" />
                View Details
              </Button>
            ) : (
              <>
                <Button size="sm" onClick={onViewRoster}>
                  <Users className="mr-2 h-4 w-4" />
                  View Roster
                </Button>
                <Button size="sm" variant="outline" onClick={onEdit}>
                  <Edit className="mr-2 h-4 w-4" />
                  Edit
                </Button>
                <Button size="sm" variant="destructive" onClick={onCancel}>
                  <XCircle className="mr-2 h-4 w-4" />
                  Cancel
                </Button>
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Loading Skeleton
 */
function OpportunityListSkeleton() {
  return (
    <div className="space-y-4">
      {[...Array(3)].map((_, i) => (
        <Card key={i}>
          <CardContent className="p-6">
            <div className="space-y-3">
              <div className="h-6 w-3/4 animate-pulse rounded bg-muted" />
              <div className="h-4 w-full animate-pulse rounded bg-muted" />
              <div className="flex gap-4">
                <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                <div className="h-4 w-24 animate-pulse rounded bg-muted" />
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
