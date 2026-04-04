import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: process.env.VITE_PLATFORM_API_URL || 'http://192.168.1.6:8004',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist'
  }
})
