'use client';

import { useState, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import {
  Search,
  MapPin,
  Calendar,
  Users,
  Map as MapIcon,
  List,
  ChevronLeft,
  ChevronRight,
  Filter,
  X,
} from 'lucide-react';
import { useSearchOpportunities } from '@/lib/api';
import type { OpportunitySearchParams, Opportunity } from '@/lib/api/types';
import { format, parseISO } from 'date-fns';
import dynamic from 'next/dynamic';

// Dynamically import map component (only loads on client side)
const OpportunityMap = dynamic(() => import('@/components/opportunities/OpportunityMap'), {
  ssr: false,
  loading: () => (
    <div className="flex h-[500px] items-center justify-center rounded-md border bg-muted">
      <p className="text-sm text-muted-foreground">Loading map...</p>
    </div>
  ),
});

/**
 * Opportunity Search Page
 *
 * Allows volunteers to search for opportunities with:
 * - Text search
 * - Location-based filtering (radius search)
 * - Cause/category filtering
 * - Date range filtering
 * - Skills filtering
 * - Map view and list view toggle
 * - Pagination
 *
 * Performance target: <2s search results
 */
export default function OpportunitySearchPage() {
  // View state
  const [viewMode, setViewMode] = useState<'list' | 'map'>('list');
  const [showFilters, setShowFilters] = useState(true);

  // Search filters state
  const [filters, setFilters] = useState<OpportunitySearchParams>({
    search: '',
    page: 1,
    limit: 20,
  });

  // Temporary filter state for form inputs
  const [tempFilters, setTempFilters] = useState({
    search: '',
    location: '',
    radius: '25',
    cause: '',
    startDate: '',
    endDate: '',
    minAge: '',
  });

  // Fetch opportunities with current filters
  const { data, isLoading, error } = useSearchOpportunities(filters);

  // Handle search submit
  const handleSearch = () => {
    const newFilters: OpportunitySearchParams = {
      page: 1,
      limit: 20,
    };

    if (tempFilters.search) newFilters.search = tempFilters.search;
    if (tempFilters.cause) newFilters.cause = tempFilters.cause;
    if (tempFilters.startDate) newFilters.start_date = tempFilters.startDate;
    if (tempFilters.endDate) newFilters.end_date = tempFilters.endDate;
    if (tempFilters.minAge) newFilters.min_age = parseInt(tempFilters.minAge);
    if (tempFilters.radius) newFilters.radius = parseFloat(tempFilters.radius);

    // TODO: Add geocoding for location to lat/lng conversion
    // For now, we'll use a placeholder location (San Francisco)
    if (tempFilters.location) {
      newFilters.latitude = 37.7749;
      newFilters.longitude = -122.4194;
    }

    setFilters(newFilters);
  };

  // Handle clear filters
  const handleClearFilters = () => {
    setTempFilters({
      search: '',
      location: '',
      radius: '25',
      cause: '',
      startDate: '',
      endDate: '',
      minAge: '',
    });
    setFilters({
      page: 1,
      limit: 20,
    });
  };

  // Handle pagination
  const handlePageChange = (newPage: number) => {
    setFilters((prev) => ({ ...prev, page: newPage }));
    // Scroll to top when changing pages
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Check if any filters are active
  const hasActiveFilters = useMemo(() => {
    return (
      tempFilters.search ||
      tempFilters.location ||
      tempFilters.cause ||
      tempFilters.startDate ||
      tempFilters.endDate ||
      tempFilters.minAge
    );
  }, [tempFilters]);

  // Render opportunity card
  const renderOpportunityCard = (opportunity: Opportunity) => (
    <Card
      key={opportunity.id}
      className="cursor-pointer transition-shadow hover:shadow-md"
      onClick={() => (window.location.href = `/opportunities/${opportunity.id}`)}
    >
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-xl">{opportunity.title}</CardTitle>
            <p className="mt-1 text-sm text-muted-foreground">
              Organization: {opportunity.organization_id}
            </p>
          </div>
          <Badge
            variant={opportunity.status === 'published' ? 'default' : 'secondary'}
            className="ml-2"
          >
            {opportunity.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="line-clamp-2 text-sm">{opportunity.description}</p>

        <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
          <div className="flex items-center gap-1">
            <Calendar className="h-4 w-4" />
            <span>{format(parseISO(opportunity.start_time), 'MMM d, yyyy')}</span>
          </div>
          <div className="flex items-center gap-1">
            <MapPin className="h-4 w-4" />
            <span>
              {opportunity.city}, {opportunity.state}
            </span>
          </div>
          <div className="flex items-center gap-1">
            <Users className="h-4 w-4" />
            <span>
              {opportunity.registered_count}/{opportunity.capacity} registered
            </span>
          </div>
        </div>

        {opportunity.causes && opportunity.causes.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {opportunity.causes.map((cause, index) => (
              <Badge key={index} variant="outline">
                {cause}
              </Badge>
            ))}
          </div>
        )}

        {opportunity.required_skills && opportunity.required_skills.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {opportunity.required_skills.map((skill, index) => (
              <Badge key={index} variant="secondary" className="text-xs">
                {skill}
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Find Opportunities</h1>
        <p className="mt-2 text-muted-foreground">
          Discover volunteer opportunities that match your interests and availability
        </p>
      </div>

      {/* Search and View Toggle */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2">
          <Button
            variant={showFilters ? 'default' : 'outline'}
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter className="mr-2 h-4 w-4" />
            Filters
          </Button>
          {hasActiveFilters && (
            <Button variant="ghost" size="sm" onClick={handleClearFilters}>
              <X className="mr-2 h-4 w-4" />
              Clear Filters
            </Button>
          )}
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant={viewMode === 'list' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setViewMode('list')}
          >
            <List className="mr-2 h-4 w-4" />
            List
          </Button>
          <Button
            variant={viewMode === 'map' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setViewMode('map')}
          >
            <MapIcon className="mr-2 h-4 w-4" />
            Map
          </Button>
        </div>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Search Filters</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Text Search */}
            <div className="space-y-2">
              <Label htmlFor="search">Search</Label>
              <div className="relative">
                <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="search"
                  placeholder="Search by title or description..."
                  value={tempFilters.search}
                  onChange={(e) => setTempFilters({ ...tempFilters, search: e.target.value })}
                  className="pl-9"
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                />
              </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {/* Location */}
              <div className="space-y-2">
                <Label htmlFor="location">Location</Label>
                <Input
                  id="location"
                  placeholder="City, State or ZIP"
                  value={tempFilters.location}
                  onChange={(e) => setTempFilters({ ...tempFilters, location: e.target.value })}
                />
              </div>

              {/* Radius */}
              <div className="space-y-2">
                <Label htmlFor="radius">Radius (miles)</Label>
                <Select
                  value={tempFilters.radius}
                  onValueChange={(value) => setTempFilters({ ...tempFilters, radius: value })}
                >
                  <SelectTrigger id="radius">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="5">5 miles</SelectItem>
                    <SelectItem value="10">10 miles</SelectItem>
                    <SelectItem value="25">25 miles</SelectItem>
                    <SelectItem value="50">50 miles</SelectItem>
                    <SelectItem value="100">100 miles</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Cause */}
              <div className="space-y-2">
                <Label htmlFor="cause">Cause</Label>
                <Select
                  value={tempFilters.cause}
                  onValueChange={(value) => setTempFilters({ ...tempFilters, cause: value })}
                >
                  <SelectTrigger id="cause">
                    <SelectValue placeholder="All causes" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">All causes</SelectItem>
                    <SelectItem value="education">Education</SelectItem>
                    <SelectItem value="environment">Environment</SelectItem>
                    <SelectItem value="health">Health</SelectItem>
                    <SelectItem value="animals">Animals</SelectItem>
                    <SelectItem value="community">Community</SelectItem>
                    <SelectItem value="arts">Arts & Culture</SelectItem>
                    <SelectItem value="seniors">Seniors</SelectItem>
                    <SelectItem value="youth">Youth</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Start Date */}
              <div className="space-y-2">
                <Label htmlFor="startDate">Start Date</Label>
                <Input
                  id="startDate"
                  type="date"
                  value={tempFilters.startDate}
                  onChange={(e) => setTempFilters({ ...tempFilters, startDate: e.target.value })}
                />
              </div>

              {/* End Date */}
              <div className="space-y-2">
                <Label htmlFor="endDate">End Date</Label>
                <Input
                  id="endDate"
                  type="date"
                  value={tempFilters.endDate}
                  onChange={(e) => setTempFilters({ ...tempFilters, endDate: e.target.value })}
                />
              </div>

              {/* Minimum Age */}
              <div className="space-y-2">
                <Label htmlFor="minAge">Minimum Age</Label>
                <Input
                  id="minAge"
                  type="number"
                  placeholder="Any age"
                  value={tempFilters.minAge}
                  onChange={(e) => setTempFilters({ ...tempFilters, minAge: e.target.value })}
                  min="0"
                  max="100"
                />
              </div>
            </div>

            <div className="flex gap-2">
              <Button onClick={handleSearch} className="flex-1">
                <Search className="mr-2 h-4 w-4" />
                Search
              </Button>
              {hasActiveFilters && (
                <Button variant="outline" onClick={handleClearFilters}>
                  Clear
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Results Section */}
      <div>
        {/* Loading State */}
        {isLoading && (
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <div className="h-6 w-3/4 animate-pulse rounded bg-muted" />
                  <div className="mt-2 h-4 w-1/2 animate-pulse rounded bg-muted" />
                </CardHeader>
                <CardContent>
                  <div className="h-20 animate-pulse rounded bg-muted" />
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Error State */}
        {error && (
          <Card className="border-destructive">
            <CardHeader>
              <CardTitle className="text-destructive">Error Loading Opportunities</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error
                  ? error.message
                  : 'Failed to load opportunities. Please try again later.'}
              </p>
            </CardContent>
          </Card>
        )}

        {/* Results */}
        {!isLoading && !error && data && (
          <>
            {/* Results Header */}
            <div className="mb-4 flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                Showing {data.data.length} of {data.pagination.total} opportunities
              </p>
              <p className="text-sm text-muted-foreground">
                Page {data.pagination.page} of {data.pagination.total_pages}
              </p>
            </div>

            {/* Map View */}
            {viewMode === 'map' && (
              <div className="mb-6">
                <OpportunityMap opportunities={data.data} />
              </div>
            )}

            {/* List View */}
            {viewMode === 'list' && (
              <>
                {data.data.length === 0 ? (
                  <Card>
                    <CardContent className="py-12 text-center">
                      <p className="text-muted-foreground">
                        No opportunities found matching your criteria.
                      </p>
                      <Button variant="link" onClick={handleClearFilters} className="mt-2">
                        Clear filters to see all opportunities
                      </Button>
                    </CardContent>
                  </Card>
                ) : (
                  <div className="space-y-4">{data.data.map(renderOpportunityCard)}</div>
                )}
              </>
            )}

            {/* Pagination */}
            {data.pagination.total_pages > 1 && (
              <div className="mt-6 flex items-center justify-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(data.pagination.page - 1)}
                  disabled={data.pagination.page === 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                  Previous
                </Button>

                <div className="flex items-center gap-1">
                  {[...Array(Math.min(5, data.pagination.total_pages))].map((_, i) => {
                    const pageNum = i + 1;
                    return (
                      <Button
                        key={pageNum}
                        variant={pageNum === data.pagination.page ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => handlePageChange(pageNum)}
                      >
                        {pageNum}
                      </Button>
                    );
                  })}
                </div>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(data.pagination.page + 1)}
                  disabled={data.pagination.page === data.pagination.total_pages}
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
