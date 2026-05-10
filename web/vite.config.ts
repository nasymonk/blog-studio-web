import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  base: '/studio/',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    sourcemap: true,
    chunkSizeWarningLimit: 600,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-codemirror': [
            '@codemirror/view',
            '@codemirror/state',
            '@codemirror/commands',
            '@codemirror/language',
            '@codemirror/lang-markdown',
            '@codemirror/theme-one-dark',
            'vue-codemirror6',
          ],
          'vendor-codemirror-languages': [
            '@codemirror/language-data',
          ],
          'vendor-vue': ['vue', 'vue-router', '@vueuse/core'],
          'vendor-icons': ['lucide-vue-next'],
          'vendor-marked': ['marked', 'dompurify'],
          'vendor-ui': ['reka-ui'],
        }
      }
    }
  },
  server: {
    proxy: {
      '/studio/api': 'http://localhost:8080',
      '/studio/preview': 'http://localhost:8080'
    }
  },
  test: {
    environment: 'happy-dom',
    globals: true,
    setupFiles: ['./src/__tests__/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/store/**', 'src/composables/**', 'src/services/**'],
      exclude: ['src/__tests__/**'],
    }
  }
})
