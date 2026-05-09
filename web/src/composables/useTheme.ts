import { watch } from 'vue'
import { useStorage } from '@vueuse/core'

const theme = useStorage<'light' | 'dark'>('blog-studio-theme', 'light')

watch(theme, (val) => {
  document.documentElement.classList.toggle('dark', val === 'dark')
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', val === 'dark' ? '#17181a' : '#fdfcf8')
}, { immediate: true })

export function useTheme() {
  return {
    theme,
    toggle: () => { theme.value = theme.value === 'light' ? 'dark' : 'light' }
  }
}
