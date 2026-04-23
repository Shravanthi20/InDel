import { coreClient } from './client';

// Get zone paths for a given type (a, b, c)
export const getZonePaths = (type: 'a' | 'b' | 'c') =>
  coreClient.get(`/api/v1/platform/zone-paths?type=${type}`);
