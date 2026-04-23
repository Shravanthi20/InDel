import axios from 'axios'

const currentHost = typeof window !== 'undefined' && window.location?.hostname
  ? window.location.hostname
  : '127.0.0.1'
const defaultGatewayBaseUrl = `http://${currentHost}:8004`

const INSURER_API_URL = import.meta.env.VITE_INSURER_API_URL || defaultGatewayBaseUrl
// coreClient is used for platform-gateway routes (/api/v1/platform/*)
// In production set VITE_PLATFORM_API_URL to the platform-gateway Render URL
const GATEWAY_API_URL =
  import.meta.env.VITE_PLATFORM_API_URL ||
  import.meta.env.VITE_GATEWAY_API_URL ||
  import.meta.env.VITE_CORE_API_URL ||
  defaultGatewayBaseUrl
const WORKER_API_URL =
  import.meta.env.VITE_WORKER_API_URL ||
  defaultGatewayBaseUrl
const ML_API_URL =
  import.meta.env.VITE_ML_API_URL ||
  defaultGatewayBaseUrl
const ENABLE_NETWORK_LOGS = import.meta.env.DEV && import.meta.env.VITE_ENABLE_API_DEBUG === 'true'

const insurerClient = axios.create({
  baseURL: INSURER_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

const coreClient = axios.create({
  baseURL: GATEWAY_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

const workerClient = axios.create({
  baseURL: WORKER_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

const forecastClient = axios.create({
  baseURL: ML_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add JWT token to requests
const attachAuthToken = (config: any) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
}

insurerClient.interceptors.request.use(attachAuthToken)
coreClient.interceptors.request.use(attachAuthToken)
workerClient.interceptors.request.use(attachAuthToken)
forecastClient.interceptors.request.use(attachAuthToken)

// Handle token expiration
const handleUnauthorized = (
  response: any
) => response

const rejectUnauthorized = (error: any) => {
  if (error.response?.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/'
  }
  return Promise.reject(error)
}

// Network Debugger Interceptor
const logNetwork = (clientName: string) => {
  const logResponse = (response: any) => {
    if (ENABLE_NETWORK_LOGS) {
      console.log(`%c[API-SUCCESS] ${clientName} %c${response.config.method?.toUpperCase()} %c${response.config.url}`, 'color: #10b981; font-weight: bold;', 'color: #f97316; font-weight: bold;', 'color: #94a3b8;', response.data)
    }
    return response
  }
  const logError = (error: any) => {
    if (ENABLE_NETWORK_LOGS) {
      console.error(`%c[API-ERROR] ${clientName} %c${error.config?.method?.toUpperCase()} %c${error.config?.url}`, 'color: #f43f5e; font-weight: bold;', 'color: #f97316; font-weight: bold;', 'color: #94a3b8;', error.response?.data || error.message)
    }
    return Promise.reject(error)
  }
  return { logResponse, logError }
}

const insurerLogs = logNetwork('Insurer-Gateway')
const coreLogs = logNetwork('Gateway-Service')
const workerLogs = logNetwork('Worker-Gateway')
const forecastLogs = logNetwork('Forecast-Service')

insurerClient.interceptors.response.use(
  (response) => {
    handleUnauthorized(response)
    return insurerLogs.logResponse(response)
  },
  (error) => {
    rejectUnauthorized(error)
    return insurerLogs.logError(error)
  }
)

coreClient.interceptors.response.use(
  (response) => {
    handleUnauthorized(response)
    return coreLogs.logResponse(response)
  },
  (error) => {
    rejectUnauthorized(error)
    return coreLogs.logError(error)
  }
)

workerClient.interceptors.response.use(
  (response) => {
    handleUnauthorized(response)
    return workerLogs.logResponse(response)
  },
  (error) => {
    rejectUnauthorized(error)
    return workerLogs.logError(error)
  }
)

forecastClient.interceptors.response.use(
  (response) => {
    handleUnauthorized(response)
    return forecastLogs.logResponse(response)
  },
  (error) => {
    rejectUnauthorized(error)
    return forecastLogs.logError(error)
  }
)

export { coreClient, workerClient, forecastClient }

export default insurerClient
