import client from './client';

// Get zone paths for a given type (a, b, c)
export const getZonePaths = (type: 'a' | 'b' | 'c') =>
  client.get(`/api/v1/platform/zone-paths?type=${type}`);
