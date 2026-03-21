import client from './client'

export const getWorkers = () => client.get('/api/workers')
export const getZones = () => client.get('/api/zones')
export const getAnalytics = () => client.get('/api/analytics')
export const postOrderWebhook = (data: any) => client.post('/api/webhooks/orders', data)
export const postEarningsWebhook = (data: any) => client.post('/api/webhooks/earnings', data)
