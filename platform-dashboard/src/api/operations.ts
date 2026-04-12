import { coreClient } from './client'

export type SyntheticScenario = 'normal_week' | 'mild_disruption' | 'severe_disruption' | 'fraud_burst'

export type SyntheticPayload = {
  seed: number
  scenario: SyntheticScenario
  output_dir?: string
  force_reset?: boolean
  reason?: string
}

export const runWeeklyCycle = () => coreClient.post('/api/v1/internal/policy/weekly-cycle/run')

export const generateClaimsForDisruption = (disruptionId: string) =>
  coreClient.post(`/api/v1/internal/claims/generate-for-disruption/${disruptionId}`)

export const autoProcessDisruption = (disruptionId: string) =>
  coreClient.post(`/api/v1/internal/claims/auto-process/${disruptionId}`)

export const queuePayoutForClaim = (claimId: string) =>
  coreClient.post(`/api/v1/internal/payouts/queue/${claimId}`)

export const processQueuedPayouts = () => coreClient.post('/api/v1/internal/payouts/process')

export const getPayoutReconciliation = (params: { from: string; to: string }) =>
  coreClient.get('/api/v1/internal/payouts/reconciliation', { params })

export const generateSyntheticData = (payload: SyntheticPayload) =>
  coreClient.post('/api/v1/internal/data/synthetic/generate', {
    force_reset: true,
    reason: payload.reason || 'platform synthetic refresh',
    ...payload,
  })
