import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: process.env.VITE_INSURER_API_URL || 'http://192.168.1.6:8004',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('react-router-dom')) {
              return 'router'
            }
            if (id.includes('recharts')) {
              return 'charts'
            }
            if (id.includes('@tremor/react')) {
              return 'tremor'
            }
            if (id.includes('axios')) {
              return 'axios'
            }
            if (id.includes('react') || id.includes('scheduler')) {
              return 'react'
            }
          }
        }
      }
    }
  }
})
