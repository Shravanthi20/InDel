import axios from 'axios'

const currentHost = typeof window !== 'undefined' && window.location?.hostname
  ? window.location.hostname
  : '192.168.1.6'
const defaultGatewayBaseUrl = `http://${currentHost}:8004`

const client = axios.create({
  baseURL: import.meta.env.VITE_PLATFORM_API_URL || defaultGatewayBaseUrl
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export default client
