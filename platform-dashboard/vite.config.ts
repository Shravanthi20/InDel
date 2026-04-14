import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5176,
    strictPort: true,
    proxy: {
      '/api/v1/forecast': {
        target: 'http://localhost:9003',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/v1\/forecast/, '/forecast'),
      },
      '/api': {
        target: process.env.VITE_PLATFORM_API_URL || 'http://192.168.1.6:8004',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
})
