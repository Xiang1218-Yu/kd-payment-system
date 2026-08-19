import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Dev server proxies /api to the Go backend on :3600 so the frontend can be
// developed with hot reload while the API runs separately. In production the
// built assets are embedded into the Go binary, so no proxy is needed.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:3600',
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
