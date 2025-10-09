'use client';

import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import { Icon } from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { MapPin } from 'lucide-react';

interface OpportunityDetailMapProps {
  latitude: number;
  longitude: number;
  title: string;
  address: string;
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
 * OpportunityDetailMap Component
 *
 * Displays a single opportunity location on an interactive map.
 * Used on the opportunity detail page.
 */
export default function OpportunityDetailMap({
  latitude,
  longitude,
  title,
  address,
}: OpportunityDetailMapProps) {
  return (
    <div className="h-[400px] w-full overflow-hidden rounded-md border">
      <MapContainer
        center={[latitude, longitude]}
        zoom={14}
        scrollWheelZoom={true}
        className="h-full w-full"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        <Marker position={[latitude, longitude]} icon={customIcon}>
          <Popup>
            <div className="space-y-1">
              <h3 className="font-semibold">{title}</h3>
              <div className="flex items-start gap-1 text-sm text-muted-foreground">
                <MapPin className="h-3 w-3 mt-0.5" />
                <span>{address}</span>
              </div>
            </div>
          </Popup>
        </Marker>
      </MapContainer>
    </div>
  );
}
