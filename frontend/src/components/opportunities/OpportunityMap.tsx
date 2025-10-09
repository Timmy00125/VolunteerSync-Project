'use client';

import { useMemo } from 'react';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import { Icon } from 'leaflet';
import 'leaflet/dist/leaflet.css';
import type { Opportunity } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { format, parseISO } from 'date-fns';
import { MapPin, Calendar, Users } from 'lucide-react';

interface OpportunityMapProps {
  opportunities: Opportunity[];
  center?: [number, number];
  zoom?: number;
}

// Fix Leaflet default marker icon issue with Next.js
const customIcon = new Icon({
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});

/**
 * OpportunityMap Component
 *
 * Displays opportunities on an interactive map using Leaflet.
 * Each opportunity is marked with a pin that shows a popup with details on click.
 */
export default function OpportunityMap({
  opportunities,
  center = [37.7749, -122.4194], // Default to San Francisco
  zoom = 10,
}: OpportunityMapProps) {
  // Filter opportunities that have valid coordinates
  const opportunitiesWithCoordinates = opportunities.filter(
    (opp) =>
      opp.latitude !== null &&
      opp.latitude !== undefined &&
      opp.longitude !== null &&
      opp.longitude !== undefined
  );

  // Calculate center based on opportunities if available
  const mapCenter = useMemo(() => {
    if (opportunitiesWithCoordinates.length > 0) {
      const avgLat =
        opportunitiesWithCoordinates.reduce((sum, opp) => sum + (opp.latitude || 0), 0) /
        opportunitiesWithCoordinates.length;
      const avgLng =
        opportunitiesWithCoordinates.reduce((sum, opp) => sum + (opp.longitude || 0), 0) /
        opportunitiesWithCoordinates.length;
      return [avgLat, avgLng] as [number, number];
    }
    return center;
  }, [opportunitiesWithCoordinates, center]);

  if (opportunitiesWithCoordinates.length === 0) {
    return (
      <div className="flex h-[500px] items-center justify-center rounded-md border bg-muted">
        <div className="text-center">
          <MapPin className="mx-auto h-12 w-12 text-muted-foreground" />
          <p className="mt-2 text-sm text-muted-foreground">
            No opportunities with location data to display on map
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-[500px] w-full overflow-hidden rounded-md border">
      <MapContainer center={mapCenter} zoom={zoom} scrollWheelZoom={true} className="h-full w-full">
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        {opportunitiesWithCoordinates.map((opportunity) => (
          <Marker
            key={opportunity.id}
            position={[opportunity.latitude!, opportunity.longitude!]}
            icon={customIcon}
          >
            <Popup className="w-64">
              <div className="space-y-2">
                <div>
                  <h3 className="font-semibold">{opportunity.title}</h3>
                  <Badge
                    variant={opportunity.status === 'published' ? 'default' : 'secondary'}
                    className="mt-1"
                  >
                    {opportunity.status}
                  </Badge>
                </div>

                <p className="line-clamp-2 text-xs text-muted-foreground">
                  {opportunity.description}
                </p>

                <div className="space-y-1 text-xs">
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Calendar className="h-3 w-3" />
                    <span>{format(parseISO(opportunity.start_time), 'MMM d, yyyy')}</span>
                  </div>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <MapPin className="h-3 w-3" />
                    <span>
                      {opportunity.city}, {opportunity.state}
                    </span>
                  </div>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Users className="h-3 w-3" />
                    <span>
                      {opportunity.registered_count}/{opportunity.capacity} registered
                    </span>
                  </div>
                </div>

                <Button
                  size="sm"
                  className="w-full"
                  onClick={() => (window.location.href = `/opportunities/${opportunity.id}`)}
                >
                  View Details
                </Button>
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>
    </div>
  );
}
