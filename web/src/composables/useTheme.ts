import { watch } from 'vue'
import { useStorage } from '@vueuse/core'

// Detect system preference for initial default when no stored value exists
const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
const defaultTheme: 'light' | 'dark' = systemPrefersDark ? 'dark' : 'light'

const theme = useStorage<'light' | 'dark'>('blog-studio-theme', defaultTheme)

watch(theme, (val) => {
  document.documentElement.classList.toggle('dark', val === 'dark')
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', val === 'dark' ? '#17181a' : '#fdfcf8')
}, { immediate: true })

// Listen for system preference changes — only apply if user hasn't set a manual preference
const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
mediaQuery.addEventListener('change', (e) => {
  if (!localStorage.getItem('blog-studio-theme')) {
    theme.value = e.matches ? 'dark' : 'light'
  }
})

export function useTheme() {
  return {
    theme,
    toggle: () => { theme.value = theme.value === 'light' ? 'dark' : 'light' }
  }
}
