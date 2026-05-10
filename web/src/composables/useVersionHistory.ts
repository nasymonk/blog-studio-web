import { ref } from 'vue'

export interface Version {
  id: string
  timestamp: Date
  body: string
  wordCount: number
}

export function useVersionHistory(maxVersions = 10) {
  const versions = ref<Version[]>([])
  const isOpen = ref(false)

  function saveVersion(body: string, wordCount: number) {
    const last = versions.value[0]
    if (last && last.body === body) return // Don't save if unchanged

    versions.value.unshift({
      id: Date.now().toString(36),
      timestamp: new Date(),
      body,
      wordCount,
    })

    if (versions.value.length > maxVersions) {
      versions.value = versions.value.slice(0, maxVersions)
    }

    persist()
  }

  function persist() {
    try {
      const key = `editor:history:${window.location.pathname}`
      localStorage.setItem(key, JSON.stringify(versions.value.map(v => ({
        ...v,
        timestamp: v.timestamp.toISOString(),
      }))))
    } catch { /* storage full, ignore */ }
  }

  function loadHistory(slug: string) {
    try {
      const key = `editor:history:/posts/${encodeURIComponent(slug)}`
      const stored = localStorage.getItem(key)
      if (stored) {
        versions.value = JSON.parse(stored).map((v: any) => ({
          ...v,
          timestamp: new Date(v.timestamp),
        }))
      }
    } catch { /* ignore */ }
  }

  function toggle() { isOpen.value = !isOpen.value }

  return { versions, isOpen, saveVersion, loadHistory, toggle }
}
