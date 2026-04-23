import client, { forecastClient, coreClient, workerClient } from './client'

export type ZoneLevelOption = {
  level: 'A' | 'B' | 'C'
  label: string
  description?: string
}

export const getWorkers = () => client.get('/api/v1/platform/workers')
export const getZones = (level?: 'A' | 'B' | 'C' | 'ALL') =>
  client.get('/api/v1/platform/zones', { params: level && level !== 'ALL' ? { level } : undefined })
export const getZoneLevels = () => client.get<{ levels: ZoneLevelOption[] }>('/api/v1/platform/zone-levels')
export const getZonePaths = (type: 'a' | 'b' | 'c') => client.get(`/api/v1/platform/zone-paths?type=${type}`)
export const getZoneHealth = () => client.get('/api/v1/platform/zones/health')
export const getDisruptions = () => client.get('/api/v1/platform/disruptions')
export const getOrders = () => workerClient.get('/api/v1/worker/orders')
export const getAvailableBatches = () => workerClient.get('/api/v1/worker/batches')
export const getAssignedBatches = () => workerClient.get('/api/v1/worker/batches/assigned')
export const getSimulationBatches = (status?: 'assigned' | 'picked_up' | 'delivered') => workerClient.get('/api/v1/demo/batches', { params: status ? { status } : undefined })
export const putAcceptBatch = (batchId: string, data: { orderIds: string[]; pickupCode: string }) =>
  workerClient.put(`/api/v1/worker/batches/${encodeURIComponent(batchId)}/accept`, data)
export const putDeliverBatch = (batchId: string, data: { deliveryCode: string }) =>
  workerClient.put(`/api/v1/worker/batches/${encodeURIComponent(batchId)}/deliver`, data)
export const postSimulateOrders = (data: { count: number }) =>
  workerClient.post('/api/v1/demo/simulate-orders', data)
export const postAddBatches = (data: {
  count: number
  zone_id?: number
  zone_level?: 'A' | 'B' | 'C'
  from_city?: string
  to_city?: string
  from_state?: string
  to_state?: string
  distance_km?: number
}) => client.post('/api/v1/platform/demo/add-batches', data)
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
}) => workerClient.post('/api/v1/demo/orders/ingest', data)
export const postTriggerDemo = (data: {
  zone_id: number
  force_order_drop: boolean
  external_signal: string
  generate_claims?: boolean
  aqi?: number
  rain?: number
  traffic?: number
  temperature?: number
  max_payout_inr?: number
  max_payout_per_day?: number
  coverage_ratio?: number
}) =>
  client.post('/api/v1/platform/demo/trigger-disruption', data)
export const generateClaimsForDisruption = (disruptionId: number) =>
  coreClient.post(`/api/v1/internal/claims/generate-for-disruption/${disruptionId}`)
export const postExternalSignal = (data: {zone_id: number, source: string, status: string}) =>
  client.post('/api/v1/platform/webhooks/external-signal', data)

export const getForecast = (zone_id: number) =>
  forecastClient.post('/api/v1/ml/forecast', { zone_id })
