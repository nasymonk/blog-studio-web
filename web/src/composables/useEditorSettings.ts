import { ref, watch } from 'vue'

export type CodeTheme = 'oneDark' | 'githubLight' | 'catppuccin' | 'solarized'

export interface EditorSettings {
  fontFamily: string  // 'system' | 'mono' | 'serif'
  fontSize: number    // 12-24
  lineHeight: number  // 1.2-2.0
  lineNumbers: boolean
  codeTheme: CodeTheme
}

const defaults: EditorSettings = {
  fontFamily: 'system',
  fontSize: 14,
  lineHeight: 1.6,
  lineNumbers: true,
  codeTheme: 'oneDark',
}

export function useEditorSettings() {
  const stored = localStorage.getItem('editor:settings')
  const settings = ref<EditorSettings>(
    stored ? { ...defaults, ...JSON.parse(stored) } : defaults
  )

  watch(settings, (val) => {
    localStorage.setItem('editor:settings', JSON.stringify(val))
  }, { deep: true })

  function applyPreset(preset: 'default' | 'typewriter' | 'compact') {
    switch (preset) {
      case 'typewriter':
        settings.value = { ...settings.value, fontFamily: 'serif', fontSize: 18, lineHeight: 2.0 }
        break
      case 'compact':
        settings.value = { ...settings.value, fontFamily: 'system', fontSize: 13, lineHeight: 1.4 }
        break
      default:
        settings.value = { ...settings.value, ...defaults }
    }
  }

  return { settings, applyPreset }
}
