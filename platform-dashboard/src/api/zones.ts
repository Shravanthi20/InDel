import client from './client';

// Get all zones (optionally by level)
export const getZones = (level?: 'a' | 'b' | 'c' | 'A' | 'B' | 'C') =>
  client.get('/api/v1/platform/zones', { params: level ? { level: level.toUpperCase() } : undefined });

// Get zone health
export const getZoneHealth = () => client.get('/api/v1/platform/zones/health');

// Get zone paths for a given type (a, b, c)
export const getZonePaths = (type: 'a' | 'b' | 'c') =>
  client.get(`/api/v1/platform/zone-paths?type=${type}`);
