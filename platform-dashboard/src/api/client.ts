import axios from 'axios'

function resolveHost() {
  if (typeof window !== 'undefined' && window.location?.hostname) {
    return window.location.hostname
  }
  return '127.0.0.1'
}

function resolveBaseUrl(envValue: unknown, port: number) {
  const configured = typeof envValue === 'string' ? envValue.trim() : ''
  if (configured) {
    return configured
  }
  return `http://${resolveHost()}:${port}`
}

const defaultGatewayBaseUrl = resolveBaseUrl(import.meta.env.VITE_PLATFORM_API_URL, 8004)
const coreBackendBaseUrl = resolveBaseUrl(import.meta.env.VITE_CORE_API_URL, 8000)
const forecastGatewayBaseUrl = resolveBaseUrl(import.meta.env.VITE_FORECAST_API_URL, 8004)
const workerGatewayBaseUrl = resolveBaseUrl(import.meta.env.VITE_WORKER_API_URL, 8001)

const client = axios.create({
  baseURL: defaultGatewayBaseUrl
})

export const coreClient = axios.create({ baseURL: coreBackendBaseUrl })
export const forecastClient = axios.create({ baseURL: forecastGatewayBaseUrl })
export const workerClient = axios.create({ baseURL: workerGatewayBaseUrl })

const WEBHOOK_KEY = import.meta.env.VITE_PLATFORM_WEBHOOK_KEY as string | undefined

function attachAuthHeader(config: any) {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
}

client.interceptors.request.use(async (config) => {
  attachAuthHeader(config)
  if (WEBHOOK_KEY && typeof WEBHOOK_KEY === 'string' && config.url?.includes('/api/v1/platform/webhooks/')) {
    config.headers['X-Platform-Webhook-Key'] = WEBHOOK_KEY
  }
  return config
})

coreClient.interceptors.request.use((config) => attachAuthHeader(config))
forecastClient.interceptors.request.use((config) => attachAuthHeader(config))
workerClient.interceptors.request.use((config) => attachAuthHeader(config))

client.interceptors.response.use(
  (response) => response,
  (error) => Promise.reject(error)
)

export default client
