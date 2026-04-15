import React, { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Marker, Tooltip, GeoJSON } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';

// Helper to fetch zones from backend
async function fetchZones() {
  const res = await fetch(import.meta.env.VITE_PLATFORM_API_URL + '/api/v1/platform/zones');
  const data = await res.json();
  return data.zones || [];
}

// Helper to fetch India/city borders GeoJSON
async function fetchIndiaGeoJson() {
  // You should host/download a proper India/city GeoJSON and serve it statically
  const res = await fetch('/india.geojson');
  return await res.json();
}

export default function ZoneMap() {
  const [zones, setZones] = useState([]);
  const [indiaGeoJson, setIndiaGeoJson] = useState(null);

  useEffect(() => {
    fetchZones().then(setZones);
    fetchIndiaGeoJson().then(setIndiaGeoJson);
  }, []);

  return (
    <div style={{ height: '600px', width: '100%' }}>
      <MapContainer center={[22.5937, 78.9629]} zoom={5} style={{ height: '100%', width: '100%' }} minZoom={4} maxBounds={[[6, 68], [38, 98]]}>
        <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
        {indiaGeoJson && <GeoJSON data={indiaGeoJson} style={{ color: '#222', weight: 1, fillOpacity: 0 }} />}
        {zones.map((zone: any) => (
          <Marker key={zone.zone_id || zone.city} position={[zone.lat, zone.long]}>
            <Tooltip direction="top" offset={[0, -10]} opacity={1} permanent={false}>
              <div>
                <strong>{zone.city || zone.zone_name}</strong><br/>
                State: {zone.state}<br/>
                Lat: {zone.lat}<br/>
                Long: {zone.long}
              </div>
            </Tooltip>
          </Marker>
        ))}
      </MapContainer>
    </div>
  );
}
