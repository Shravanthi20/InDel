import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

function resolveProxyTarget(envValue: string | undefined, fallbackPort: number) {
  const configured = envValue?.trim()
  if (configured) {
    return configured
  }
  return `http://127.0.0.1:${fallbackPort}`
}

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5176,
    strictPort: true,
    proxy: {
      '/api/v1/forecast': {
        target: resolveProxyTarget(process.env.VITE_FORECAST_API_URL, 9003),
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/v1\/forecast/, '/forecast'),
      },
      '/api': {
        target: resolveProxyTarget(process.env.VITE_PLATFORM_API_URL, 8004),
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
})
