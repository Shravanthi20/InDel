import client, { coreClient } from './client'

type SuccessEnvelope<T> = {
  data: T
}

type PaginatedEnvelope<T> = {
  data: T
  pagination: {
    page: number
    limit: number
    total: number
    has_next: boolean
  }
}

const unwrapSuccess = <T>(response: { data: SuccessEnvelope<T> }) => response.data.data
const unwrapPaginated = <T>(response: { data: PaginatedEnvelope<T> }) => response.data

export const getOverview = async <T = any>(): Promise<T> =>
  unwrapSuccess<T>(await client.get<SuccessEnvelope<T>>('/api/v1/insurer/overview'))

export const getLossRatio = async <T = any>(params?: any): Promise<T> =>
  unwrapSuccess<T>(await client.get<SuccessEnvelope<T>>('/api/v1/insurer/loss-ratio', { params }))

export const getClaims = async <T = any>(params?: any): Promise<PaginatedEnvelope<T>> =>
  unwrapPaginated<T>(await client.get<PaginatedEnvelope<T>>('/api/v1/insurer/claims', { params }))

export const getClaimDetail = async <T = any>(claimId: string): Promise<T> =>
  unwrapSuccess<T>(await client.get<SuccessEnvelope<T>>(`/api/v1/insurer/claims/${claimId}`))

export const getFraudQueue = async <T = any>(params?: any): Promise<PaginatedEnvelope<T>> =>
  unwrapPaginated<T>(await client.get<PaginatedEnvelope<T>>('/api/v1/insurer/claims/fraud-queue', { params }))

export const getForecast = async <T = any[]>(): Promise<T> => {
  const response = await client.get<{ forecast: T }>('/api/v1/insurer/forecast')
  return response.data.forecast
}

export const getPoolHealth = async <T = any>(): Promise<T> => {
  const response = await client.get<T>('/api/v1/insurer/pool/health')
  return response.data
}

// Kept for compatibility with legacy Register flow.
export const getZones = () => client.get('/api/v1/platform/zones')

export const getMaintenanceChecks = async <T = any>(params?: any): Promise<PaginatedEnvelope<T>> =>
  unwrapPaginated<T>(await client.get<PaginatedEnvelope<T>>('/api/v1/insurer/maintenance-checks', { params }))

export const respondToCheck = async <T = any>(checkId: string, response: { findings: string }): Promise<T> =>
  unwrapSuccess<T>(await client.post<SuccessEnvelope<T>>(`/api/v1/insurer/maintenance-checks/${checkId}/respond`, response))

export const getAvailableBatches = async <T = any>(): Promise<T> => {
  const response = await coreClient.get<T>('/api/v1/worker/batches')
  return response.data
}

export const getAssignedBatches = async <T = any>(): Promise<T> => {
  const response = await coreClient.get<T>('/api/v1/worker/batches/assigned')
  return response.data
}
