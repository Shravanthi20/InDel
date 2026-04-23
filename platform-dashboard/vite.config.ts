import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

function resolveProxyTarget(primary: string | undefined, secondary: string | undefined, fallback: string) {
  return primary?.trim() || secondary?.trim() || fallback
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react()],
    server: {
      port: 5176,
      strictPort: true,
      proxy: {
        '/api/v1/forecast': {
          target: resolveProxyTarget(env.VITE_FORECAST_API_URL, undefined, 'http://127.0.0.1:9003'),
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api\/v1\/forecast/, '/forecast'),
        },
        '/api': {
          target: resolveProxyTarget(env.VITE_PLATFORM_API_URL, env.PLATFORM_API_URL, 'http://127.0.0.1:8004'),
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: 'dist',
    },
  }
})
