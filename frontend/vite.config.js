import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const backendTarget =
  process.env.VITE_BACKEND_HOST || 'http://gateway:8080'

const usePolling =
  String(process.env.CHOKIDAR_USEPOLLING || process.env.VITE_USE_POLLING || '')
    .toLowerCase() === 'true'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    allowedHosts: ['qrcode-itip.freedynamicdns.net'],
    watch: {
      usePolling,
    },
    proxy: {
      '/api': {
        target: backendTarget,
        changeOrigin: true,
        secure: false,
      },
      '/redirect': {
        target: backendTarget,
        changeOrigin: true,
        secure: false,
      }
    },
  },
})