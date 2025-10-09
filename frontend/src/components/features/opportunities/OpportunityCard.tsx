'use client';

import { format, parseISO } from 'date-fns';
import { MapPin, Calendar, Users, Clock } from 'lucide-react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import type { Opportunity } from '@/lib/api/types';
import { cn } from '@/lib/utils';

interface OpportunityCardProps {
  opportunity: Opportunity;
  showOrganization?: boolean;
  className?: string;
  onRegisterClick?: (opportunityId: string) => void;
}

/**
 * OpportunityCard Component
 *
 * Displays a summary card for an opportunity with:
 * - Opportunity title and description
 * - Organization info (if showOrganization is true)
 * - Date and time information
 * - Location details
 * - Capacity indicator with visual progress
 * - Cause badges
 * - Call-to-action button
 *
 * @param opportunity - The opportunity data to display
 * @param showOrganization - Whether to show organization information
 * @param className - Additional CSS classes
 * @param onRegisterClick - Callback when register button is clicked
 */
export default function OpportunityCard({
  opportunity,
  showOrganization = true,
  className,
  onRegisterClick,
}: OpportunityCardProps) {
  const registeredCount = opportunity.registered_count || 0;
  const capacity = opportunity.capacity || 0;
  const isFull = registeredCount >= capacity;
  const capacityPercentage = capacity > 0 ? (registeredCount / capacity) * 100 : 0;

  // Format date and time
  const formattedDate = opportunity.start_time
    ? format(parseISO(opportunity.start_time), 'EEEE, MMMM d, yyyy')
    : '';
  const formattedTime = opportunity.start_time
    ? `${format(parseISO(opportunity.start_time), 'h:mm a')}${
        opportunity.end_time ? ` - ${format(parseISO(opportunity.end_time), 'h:mm a')}` : ''
      }`
    : '';

  // Determine status color and text
  const getStatusBadge = () => {
    if (opportunity.status === 'cancelled') {
      return (
        <Badge variant="destructive" className="shrink-0">
          Cancelled
        </Badge>
      );
    }
    if (opportunity.status === 'completed') {
      return (
        <Badge variant="secondary" className="shrink-0">
          Completed
        </Badge>
      );
    }
    if (isFull) {
      return (
        <Badge variant="secondary" className="shrink-0">
          Full
        </Badge>
      );
    }
    if (opportunity.status === 'published') {
      return (
        <Badge variant="default" className="shrink-0">
          Open
        </Badge>
      );
    }
    return null;
  };

  const handleRegisterClick = (e: React.MouseEvent) => {
    e.preventDefault();
    if (onRegisterClick) {
      onRegisterClick(opportunity.id);
    }
  };

  return (
    <Card className={cn('hover:shadow-lg transition-shadow', className)}>
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <CardTitle className="text-xl line-clamp-2">
              <Link href={`/opportunities/${opportunity.id}`} className="hover:underline">
                {opportunity.title}
              </Link>
            </CardTitle>
            {showOrganization && opportunity.organization_id && (
              <CardDescription className="mt-1">
                {/* Organization name would come from a separate API call or be included in the opportunity object */}
                Organization ID: {opportunity.organization_id}
              </CardDescription>
            )}
          </div>
          {getStatusBadge()}
        </div>

        {/* Cause Badges */}
        {opportunity.causes && opportunity.causes.length > 0 && (
          <div className="flex flex-wrap gap-2 mt-3">
            {opportunity.causes.map((cause, index) => (
              <Badge key={index} variant="outline" className="text-xs">
                {cause}
              </Badge>
            ))}
          </div>
        )}
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Description */}
        <p className="text-sm text-muted-foreground line-clamp-3">{opportunity.description}</p>

        {/* Date and Time */}
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm">
            <Calendar className="h-4 w-4 text-muted-foreground shrink-0" />
            <span>{formattedDate}</span>
          </div>
          <div className="flex items-center gap-2 text-sm">
            <Clock className="h-4 w-4 text-muted-foreground shrink-0" />
            <span>{formattedTime}</span>
          </div>
        </div>

        {/* Location */}
        <div className="flex items-start gap-2 text-sm">
          <MapPin className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
          <div className="flex-1 min-w-0">
            <div className="line-clamp-2">
              {opportunity.location || `${opportunity.city}, ${opportunity.state}`}
            </div>
          </div>
        </div>

        {/* Capacity Indicator */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">
                {registeredCount} / {capacity} volunteers
              </span>
            </div>
            <span className="text-muted-foreground">{Math.round(capacityPercentage)}% full</span>
          </div>

          {/* Visual progress bar */}
          <div className="w-full h-2 bg-secondary rounded-full overflow-hidden">
            <div
              className={cn(
                'h-full transition-all duration-300',
                isFull ? 'bg-destructive' : 'bg-primary'
              )}
              style={{ width: `${Math.min(capacityPercentage, 100)}%` }}
            />
          </div>
        </div>

        {/* Required Skills */}
        {opportunity.required_skills && opportunity.required_skills.length > 0 && (
          <div className="pt-2 border-t">
            <p className="text-xs text-muted-foreground mb-2">Required Skills:</p>
            <div className="flex flex-wrap gap-1">
              {opportunity.required_skills.map((skill, index) => (
                <Badge key={index} variant="secondary" className="text-xs">
                  {skill}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Action Button */}
        {opportunity.status === 'published' && !isFull && (
          <div className="pt-2">
            <Button onClick={handleRegisterClick} className="w-full" disabled={isFull}>
              Register for Event
            </Button>
          </div>
        )}

        {/* Recurring indicator */}
        {opportunity.is_recurring && (
          <div className="pt-2 border-t">
            <p className="text-xs text-muted-foreground flex items-center gap-1">
              <span className="inline-block w-2 h-2 bg-primary rounded-full"></span>
              Recurring Event{' '}
              {opportunity.recurrence_pattern && `(${opportunity.recurrence_pattern})`}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
