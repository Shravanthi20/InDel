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

type PlanUsersPayload<T> = {
  users: T[]
}

type PlanActionPayload<T> = {
  success: boolean
  user: T
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

export const getMoneyExchange = async <T = any>(params?: { level?: string; zone?: string }): Promise<T> =>
  unwrapSuccess<T>(await client.get<SuccessEnvelope<T>>('/api/v1/insurer/money-exchange', { params }))

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
export const getZones = () => client.get('/api/v1/platform/zones')

export const getAvailableBatches = async <T = any>(): Promise<T> => {
  const response = await coreClient.get<T>('/api/v1/worker/batches')
  return response.data
}

export const getAssignedBatches = async <T = any>(): Promise<T> => {
  const response = await coreClient.get<T>('/api/v1/worker/batches/assigned')
  return response.data
}
