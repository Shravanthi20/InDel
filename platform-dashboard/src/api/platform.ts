import client from './client'

export const getWorkers = () => client.get('/api/v1/platform/workers')
export const getZones = () => client.get('/api/v1/platform/zones')
export const getZonePaths = (type: 'a' | 'b' | 'c') => client.get(`/api/v1/platform/zone-paths?type=${type}`)
export const getZoneHealth = () => client.get('/api/v1/platform/zones/health')
export const getDisruptions = () => client.get('/api/v1/platform/disruptions')
export const getOrders = () => client.get('/api/v1/worker/orders')
export const getAvailableBatches = () => client.get('/api/v1/worker/batches')
export const getAssignedBatches = () => client.get('/api/v1/worker/batches/assigned')
export const postSimulateOrders = (data: { count: number }) =>
  client.post('/api/v1/demo/simulate-orders', data)
export const postIngestDemoOrder = (data: {
  order_id: string
  customer_name: string
  customer_id: string
  customer_contact_number: string
  address: string
  payment_method: string
  order_value: number
  payment_amount: number
  package_size: string
  package_weight_kg: number
  zone_id: number
  from_city: string
  to_city: string
  from_state?: string
  to_state?: string
  pickup_area: string
  drop_area: string
  distance_km: number
  tip_inr: number
  delivery_fee_inr: number
  status: string
  source: string
}) => client.post('/api/v1/demo/orders/ingest', data)
export const postTriggerDemo = (data: { zone_id: number; force_order_drop: boolean; external_signal: string }) =>
  client.post('/api/v1/platform/demo/trigger-disruption', data)
export const generateClaimsForDisruption = (disruptionId: number) =>
  client.post(`/api/v1/internal/claims/generate-for-disruption/${disruptionId}`)
export const postExternalSignal = (data: {zone_id: number, source: string, status: string}) =>
  client.post('/api/v1/platform/webhooks/external-signal', data)
