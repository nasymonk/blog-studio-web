import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  base: '/studio/',
  server: {
    proxy: {
      '/studio/api': 'http://localhost:8080',
      '/studio/preview': 'http://localhost:8080'
    }
  }
})
