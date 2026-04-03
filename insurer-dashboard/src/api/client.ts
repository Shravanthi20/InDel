import axios from 'axios'

const INSURER_API_URL = import.meta.env.VITE_INSURER_API_URL || 'http://localhost:8002'
const CORE_API_URL = import.meta.env.VITE_CORE_API_URL || 'http://localhost:8000'

const insurerClient = axios.create({
  baseURL: INSURER_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

const coreClient = axios.create({
  baseURL: CORE_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add JWT token to requests
const attachAuthToken = (config: any) => {
  const token = localStorage.getItem('token')
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
}

insurerClient.interceptors.request.use(attachAuthToken)
coreClient.interceptors.request.use(attachAuthToken)

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

insurerClient.interceptors.response.use(
  handleUnauthorized,
  rejectUnauthorized
)

coreClient.interceptors.response.use(
  handleUnauthorized,
  rejectUnauthorized
)

export { coreClient }

export default insurerClient
