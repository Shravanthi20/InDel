import axios from 'axios'

const client = axios.create({
  baseURL: import.meta.env.VITE_PLATFORM_API_URL || 'http://192.168.1.6:8004'
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export default client
