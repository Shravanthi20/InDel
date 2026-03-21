import client from './client'

export const getOverview = () => client.get('/api/overview')
export const getLossRatio = () => client.get('/api/loss-ratio')
export const getClaims = (params?: any) => client.get('/api/claims', { params })
export const getClaimDetail = (claimId: string) => client.get(`/api/claims/${claimId}`)
export const getFraudQueue = () => client.get('/api/fraud-queue')
export const getFraudSignals = (claimId: string) => client.get(`/api/fraud/${claimId}`)
export const getForecast = () => client.get('/api/forecast')
export const getWorkers = () => client.get('/api/workers')
export const getMaintenanceChecks = () => client.get('/api/maintenance-checks')
export const respondToCheck = (checkId: string, response: any) => 
  client.post(`/api/maintenance-checks/${checkId}/respond`, response)
