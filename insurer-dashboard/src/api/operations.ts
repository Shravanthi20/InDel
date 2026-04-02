import client from './client'

export type SyntheticScenario = 'normal_week' | 'mild_disruption' | 'severe_disruption' | 'fraud_burst'

export type SyntheticPayload = {
  seed: number
  scenario: SyntheticScenario
  output_dir?: string
}

export const runWeeklyCycle = () => client.post('/api/v1/internal/policy/weekly-cycle/run')

export const generateClaimsForDisruption = (disruptionId: string) =>
  client.post(`/api/v1/internal/claims/generate-for-disruption/${disruptionId}`)

export const queuePayoutForClaim = (claimId: string) =>
  client.post(`/api/v1/internal/payouts/queue/${claimId}`)

export const processQueuedPayouts = () => client.post('/api/v1/internal/payouts/process')

export const getPayoutReconciliation = (params: { from: string; to: string }) =>
  client.get('/api/v1/internal/payouts/reconciliation', { params })

export const generateSyntheticData = (payload: SyntheticPayload) =>
  client.post('/api/v1/internal/data/synthetic/generate', payload)
