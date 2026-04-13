import axios from 'axios'

const currentHost = typeof window !== 'undefined' && window.location?.hostname
  ? window.location.hostname
  : '192.168.1.6'
const defaultGatewayBaseUrl = `http://${currentHost}:8004`

const client = axios.create({
  baseURL: import.meta.env.VITE_PLATFORM_API_URL || defaultGatewayBaseUrl
})

export const coreClient = client

const WEBHOOK_KEY = import.meta.env.VITE_PLATFORM_WEBHOOK_KEY as string | undefined
let bootstrapTokenPromise: Promise<string | null> | null = null
let bootstrapAttempted = false

async function ensureBootstrapToken(forceRefresh = false): Promise<string | null> {
  const existingToken = localStorage.getItem('token')
  if (existingToken && !forceRefresh) {
    return existingToken
  }

  if (!bootstrapTokenPromise) {
    const baseURL = (import.meta.env.VITE_PLATFORM_API_URL || defaultGatewayBaseUrl).replace(/\/$/, '')
    const demoPhone = (import.meta.env.VITE_DEMO_WORKER_PHONE as string | undefined) || '+919999999999'
    const demoOtp = (import.meta.env.VITE_DEMO_WORKER_OTP as string | undefined) || '123456'

    bootstrapTokenPromise = axios
      .post(
        `${baseURL}/api/v1/auth/otp/verify`,
        { phone: demoPhone, otp: demoOtp },
        {
          headers: { 'Content-Type': 'application/json' },
          timeout: 5000,
        }
      )
      .then((res) => {
        const token = res.data?.token as string | undefined
        if (token) {
          localStorage.setItem('token', token)
          return token
        }
        return null
      })
      .catch(() => null)
      .finally(() => {
        bootstrapTokenPromise = null
      })
  }

  return bootstrapTokenPromise
}

client.interceptors.request.use(async (config) => {
  const requestUrl = String(config.url || '')
  const isAuthBootstrapCall = requestUrl.includes('/api/v1/auth/otp/verify')

  let token = localStorage.getItem('token')
  if (!isAuthBootstrapCall && !bootstrapAttempted) {
    bootstrapAttempted = true
    token = await ensureBootstrapToken(true)
  } else if (!token && !isAuthBootstrapCall) {
    token = await ensureBootstrapToken()
  }
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  if (WEBHOOK_KEY && typeof WEBHOOK_KEY === 'string' && config.url?.includes('/api/v1/platform/webhooks/')) {
    config.headers['X-Platform-Webhook-Key'] = WEBHOOK_KEY
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const status = error?.response?.status
    const originalConfig = error?.config as any
    const requestUrl = String(originalConfig?.url || '')
    const isAuthBootstrapCall = requestUrl.includes('/api/v1/auth/otp/verify')

    if (status === 401 && originalConfig && !originalConfig._retry && !isAuthBootstrapCall) {
      originalConfig._retry = true
      localStorage.removeItem('token')

      const token = await ensureBootstrapToken()
      if (token) {
        originalConfig.headers = originalConfig.headers || {}
        originalConfig.headers.Authorization = `Bearer ${token}`
        return client(originalConfig)
      }
    }

    return Promise.reject(error)
  }
)

export default client
