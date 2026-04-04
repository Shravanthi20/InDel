import axios from 'axios'
import client from './client'

const platformClient = axios.create({
  baseURL: import.meta.env.VITE_PLATFORM_API_URL || import.meta.env.VITE_INSURER_API_URL || 'http://192.168.1.6:8004'
})

export const getOverview = () => client.get('/api/v1/insurer/overview')
export const getLossRatio = (params?: any) => client.get('/api/v1/insurer/loss-ratio', { params })
export const getClaims = (params?: any) => client.get('/api/v1/insurer/claims', { params })
export const getClaimDetail = (claimId: string) => client.get(`/api/v1/insurer/claims/${claimId}`)
export const getFraudQueue = () => client.get('/api/v1/insurer/claims/fraud-queue')
export const getFraudSignals = (claimId: string) => client.get(`/api/v1/insurer/claims/${claimId}`)
export const getForecast = () => client.get('/api/v1/insurer/forecast')
export const getWorkers = () => platformClient.get('/api/v1/platform/workers')
export const getMaintenanceChecks = () => client.get('/api/v1/insurer/pool/health')
export const respondToCheck = (checkId: string, response: any) => 
  client.post(`/api/v1/insurer/claims/${checkId}/review`, response)
export const reviewClaim = (claimId: string, response: any) =>
  client.post(`/api/v1/insurer/claims/${claimId}/review`, response)
export const getZones = () => platformClient.get('/api/v1/platform/zones')
export const getZonePaths = (type: 'a' | 'b' | 'c') => platformClient.get(`/api/v1/platform/zone-paths?type=${type}`)
