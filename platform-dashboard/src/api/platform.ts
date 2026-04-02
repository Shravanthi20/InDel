import client from './client'

export const getWorkers = () => client.get('/api/v1/platform/workers')
export const getZones = () => client.get('/api/v1/platform/zones')
export const getZoneHealth = () => client.get('/api/v1/platform/zones/health')
export const getDisruptions = () => client.get('/api/v1/platform/disruptions')
export const postTriggerDemo = (data: {zone_id: number, force_order_drop: boolean, external_signal: string}) => 
  client.post('/api/v1/demo/trigger-disruption', data)
export const postExternalSignal = (data: {zone_id: number, source: string, status: string}) =>
  client.post('/api/v1/platform/webhooks/external-signal', data)
