'use client';

import { useMemo, useCallback } from 'react';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import { Icon, type LatLngExpression } from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { MapPin } from 'lucide-react';
import { cn } from '@/lib/utils';

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

export interface MapMarker {
  id: string;
  latitude: number;
  longitude: number;
  title: string;
  description?: string;
  [key: string]: any; // Allow additional properties
}

interface MapProps {
  /**
   * Markers to display on the map
   */
  markers: MapMarker[];

  /**
   * Optional center coordinates [lat, lng]
   * If not provided, will auto-center based on markers
   */
  center?: [number, number];

  /**
   * Initial zoom level (default: 10)
   */
  zoom?: number;

  /**
   * Height of the map container (default: '500px')
   */
  height?: string;

  /**
   * Additional CSS classes for the map container
   */
  className?: string;

  /**
   * Enable scroll wheel zoom (default: true)
   */
  scrollWheelZoom?: boolean;

  /**
   * Callback when a marker is clicked
   */
  onMarkerClick?: (marker: MapMarker) => void;

  /**
   * Custom render function for popup content
   */
  renderPopup?: (marker: MapMarker) => React.ReactNode;

  /**
   * Show message when no markers are available (default: true)
   */
  showEmptyState?: boolean;
}

/**
 * Map Component
 *
 * A reusable map component using Leaflet and react-leaflet.
 * Features:
 * - Display markers with custom popup content
 * - Auto-center based on marker positions
 * - Click handlers for markers
 * - Customizable zoom, height, and styling
 * - Empty state when no markers are available
 *
 * @example
 * ```tsx
 * <Map
 *   markers={[
 *     { id: '1', latitude: 37.7749, longitude: -122.4194, title: 'Event 1' }
 *   ]}
 *   onMarkerClick={(marker) => console.log(marker)}
 *   height="600px"
 * />
 * ```
 */
export default function Map({
  markers,
  center,
  zoom = 10,
  height = '500px',
  className,
  scrollWheelZoom = true,
  onMarkerClick,
  renderPopup,
  showEmptyState = true,
}: MapProps) {
  // Filter markers that have valid coordinates
  const validMarkers = useMemo(
    () =>
      markers.filter(
        (marker) =>
          marker.latitude !== null &&
          marker.latitude !== undefined &&
          marker.longitude !== null &&
          marker.longitude !== undefined &&
          !isNaN(marker.latitude) &&
          !isNaN(marker.longitude)
      ),
    [markers]
  );

  // Calculate center based on markers if not provided
  const mapCenter: LatLngExpression = useMemo(() => {
    if (center) {
      return center;
    }

    if (validMarkers.length > 0) {
      const avgLat =
        validMarkers.reduce((sum, marker) => sum + marker.latitude, 0) / validMarkers.length;
      const avgLng =
        validMarkers.reduce((sum, marker) => sum + marker.longitude, 0) / validMarkers.length;
      return [avgLat, avgLng];
    }

    // Default to San Francisco
    return [37.7749, -122.4194];
  }, [validMarkers, center]);

  // Calculate zoom level based on marker spread if not provided
  const mapZoom = useMemo(() => {
    if (zoom) {
      return zoom;
    }

    if (validMarkers.length === 0) {
      return 10;
    }

    if (validMarkers.length === 1) {
      return 13;
    }

    // Calculate bounds
    const lats = validMarkers.map((m) => m.latitude);
    const lngs = validMarkers.map((m) => m.longitude);
    const latDiff = Math.max(...lats) - Math.min(...lats);
    const lngDiff = Math.max(...lngs) - Math.min(...lngs);
    const maxDiff = Math.max(latDiff, lngDiff);

    // Approximate zoom level based on coordinate spread
    if (maxDiff > 10) return 4;
    if (maxDiff > 5) return 6;
    if (maxDiff > 2) return 8;
    if (maxDiff > 1) return 9;
    if (maxDiff > 0.5) return 10;
    if (maxDiff > 0.1) return 12;
    return 13;
  }, [validMarkers, zoom]);

  // Handle marker click
  const handleMarkerClick = useCallback(
    (marker: MapMarker) => {
      if (onMarkerClick) {
        onMarkerClick(marker);
      }
    },
    [onMarkerClick]
  );

  // Default popup renderer
  const defaultRenderPopup = useCallback((marker: MapMarker) => {
    return (
      <div className="min-w-[200px] max-w-[300px]">
        <h3 className="font-semibold text-sm mb-1">{marker.title}</h3>
        {marker.description && (
          <p className="text-xs text-muted-foreground line-clamp-3">{marker.description}</p>
        )}
      </div>
    );
  }, []);

  const popupRenderer = renderPopup || defaultRenderPopup;

  // Empty state
  if (showEmptyState && validMarkers.length === 0) {
    return (
      <div
        className={cn('flex items-center justify-center rounded-md border bg-muted', className)}
        style={{ height }}
      >
        <div className="text-center px-4">
          <MapPin className="mx-auto h-12 w-12 text-muted-foreground" />
          <p className="mt-2 text-sm text-muted-foreground">No locations to display on map</p>
        </div>
      </div>
    );
  }

  return (
    <div className={cn('w-full overflow-hidden rounded-md border', className)} style={{ height }}>
      <MapContainer
        center={mapCenter}
        zoom={mapZoom}
        scrollWheelZoom={scrollWheelZoom}
        className="h-full w-full"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        {validMarkers.map((marker) => (
          <Marker
            key={marker.id}
            position={[marker.latitude, marker.longitude]}
            icon={customIcon}
            eventHandlers={{
              click: () => handleMarkerClick(marker),
            }}
          >
            <Popup className="leaflet-popup-custom">{popupRenderer(marker)}</Popup>
          </Marker>
        ))}
      </MapContainer>
    </div>
  );
}
