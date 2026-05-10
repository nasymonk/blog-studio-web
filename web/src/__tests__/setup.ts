import { vi, beforeEach } from 'vitest'

// Provide a full localStorage mock (happy-dom 17.x has a restricted localStorage)
function makeLocalStorage() {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, val: string) => { store[key] = String(val) },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
    get length() { return Object.keys(store).length },
    key: (i: number) => Object.keys(store)[i] ?? null,
  }
}

const localStorageMock = makeLocalStorage()
vi.stubGlobal('localStorage', localStorageMock)

beforeEach(() => {
  localStorageMock.clear()
  vi.clearAllMocks()
})
