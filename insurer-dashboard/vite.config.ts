import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react()],
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: env.VITE_GATEWAY_API_URL || env.VITE_INSURER_API_URL || 'http://127.0.0.1:8004',
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
  }
})
