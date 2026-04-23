import client, { coreClient, workerClient, forecastClient } from './client'

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

type PlanUsersPayload<T> = {
  users: T[]
}

type PlanActionPayload<T> = {
  success: boolean
  user: T
}

export type LossRatioRow = {
  zone_id: number
  zone_name: string
  city: string
  premiums: number
  claims: number
  loss_ratio: number
}

export type ForecastRow = {
  city: string
  zone: string
  date: string
  probability: number
}

export type ClaimListRow = {
  claim_id: number
  disruption_id: number
  worker_id: number
  zone_name: string
  claim_amount: number
  status: string
  fraud_verdict: string
  created_at: string
}

export type FraudQueueRow = {
  claim_id: number
  worker_id: number
  status: string
  fraud_verdict: string
  fraud_score: number
  violations?: string[]
  created_at: string
}

export type LedgerRow = {
  timestamp: string
  worker_id: number
  zone: string
  event_type: 'premium' | 'payout'
  amount: number
  status: string
  reference_id: string
}

const unwrapSuccess = <T>(response: { data: SuccessEnvelope<T> }) => response.data.data
const unwrapPaginated = <T>(response: { data: PaginatedEnvelope<T> }) => response.data

export const getOverview = async <T = any>(): Promise<T> =>
  unwrapSuccess<T>(await client.get<SuccessEnvelope<T>>('/api/v1/insurer/overview'))

export const getLossRatio = async (params?: { zone_id?: string }): Promise<LossRatioRow[]> =>
  unwrapSuccess<LossRatioRow[]>(await client.get<SuccessEnvelope<LossRatioRow[]>>('/api/v1/insurer/loss-ratio', { params }))

export const getClaims = async (params?: {
  page?: number
  limit?: number
  status?: string
  fraud_verdict?: string
}): Promise<PaginatedEnvelope<ClaimListRow[]>> =>
  unwrapPaginated<ClaimListRow[]>(await client.get<PaginatedEnvelope<ClaimListRow[]>>('/api/v1/insurer/claims', { params }))

export const getClaimDetail = async <T = any>(claimId: string): Promise<T> =>
  unwrapSuccess<T>(await client.get<SuccessEnvelope<T>>(`/api/v1/insurer/claims/${claimId}`))

export const reviewClaim = async (claimId: number, status: string, fraud_verdict: string, notes: string = ''): Promise<void> => {
  await client.post(`/api/v1/insurer/claims/${claimId}/review`, { status, fraud_verdict, notes })
}

export const getFraudQueue = async (params?: {
  page?: number
  limit?: number
}): Promise<PaginatedEnvelope<FraudQueueRow[]>> =>
  unwrapPaginated<FraudQueueRow[]>(await client.get<PaginatedEnvelope<FraudQueueRow[]>>('/api/v1/insurer/claims/fraud-queue', { params }))

export const getForecast = async (): Promise<ForecastRow[]> => {
  const response = await client.get<{ forecast: ForecastRow[] }>('/api/v1/insurer/forecast')
  return response.data.forecast
}

export const getPoolHealth = async <T = any>(): Promise<T> => {
  const response = await client.get<T>('/api/v1/insurer/pool/health')
  return response.data
}

export const getMoneyExchange = async <T = any>(params?: { level?: string; zone?: string }): Promise<T> =>
  unwrapSuccess<T>(await client.get<SuccessEnvelope<T>>('/api/v1/insurer/money-exchange', { params }))

export const getLedger = async (params?: {
  page?: number
  limit?: number
}): Promise<PaginatedEnvelope<LedgerRow[]>> =>
  unwrapPaginated<LedgerRow[]>(await client.get<PaginatedEnvelope<LedgerRow[]>>('/api/v1/insurer/ledger', { params }))

export const getPlanUsers = async <T = any>(): Promise<T[]> => {
  const payload = await unwrapSuccess<PlanUsersPayload<T>>(await client.get<SuccessEnvelope<PlanUsersPayload<T>>>('/api/v1/insurer/users'))
  return payload.users
}

// Kept for compatibility with the existing plan status dashboard import.
export const getPlanStats = getPlanUsers

export const startUserPlan = async <T = any>(userId: string | number): Promise<T> => {
  const payload = await unwrapSuccess<PlanActionPayload<T>>(
    await client.post<SuccessEnvelope<PlanActionPayload<T>>>(`/api/v1/insurer/users/${userId}/plan/start`)
  )
  return payload.user
}

export const endUserPlan = async <T = any>(userId: string | number): Promise<T> => {
  const payload = await unwrapSuccess<PlanActionPayload<T>>(
    await client.post<SuccessEnvelope<PlanActionPayload<T>>>(`/api/v1/insurer/users/${userId}/plan/end`)
  )
  return payload.user
}

// Kept for compatibility with legacy Register flow.
export const getZones = () => coreClient.get('/api/v1/platform/zones')
export const getZoneHealth = () => coreClient.get('/api/v1/platform/zones/health')
export const getDisruptions = () => coreClient.get('/api/v1/platform/disruptions')
export const getMLForecast = (zone_id: number) => forecastClient.post('/api/v1/ml/forecast', { zone_id })

export const getAvailableBatches = async <T = any>(): Promise<T> => {
  const response = await workerClient.get<T>('/api/v1/worker/batches')
  return response.data
}

export const getAssignedBatches = async <T = any>(): Promise<T> => {
  const response = await workerClient.get<T>('/api/v1/worker/batches/assigned')
  return response.data
}
