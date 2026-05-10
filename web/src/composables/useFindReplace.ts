import { ref, watch } from 'vue'
import type { EditorView } from '@codemirror/view'

export function useFindReplace(getView: () => EditorView | null) {
  const open = ref(false)
  const searchText = ref('')
  const replaceText = ref('')
  const caseSensitive = ref(false)
  const wholeWord = ref(false)
  const useRegex = ref(false)
  const matchCount = ref(0)
  const currentMatch = ref(0)

  // Lazily loaded search module
  let searchModule: typeof import('@codemirror/search') | null = null

  async function loadSearch() {
    if (!searchModule) searchModule = await import('@codemirror/search')
    return searchModule
  }

  function openFind() {
    open.value = true
    // Pre-fill with current selection (single line only)
    const v = getView()
    if (v) {
      const sel = v.state.sliceDoc(v.state.selection.main.from, v.state.selection.main.to)
      if (sel && !sel.includes('\n')) searchText.value = sel
    }
  }

  function close() {
    open.value = false
    clearHighlights(getView())
  }

  function clearHighlights(v: EditorView | null) {
    if (!v || !searchModule) return
    const { SearchQuery, setSearchQuery } = searchModule
    v.dispatch({
      effects: setSearchQuery.of(new SearchQuery({ search: '' }))
    })
  }

  async function syncQuery() {
    const v = getView()
    if (!v) return
    const mod = await loadSearch()
    const { SearchQuery, setSearchQuery } = mod

    if (!searchText.value) {
      v.dispatch({
        effects: setSearchQuery.of(new SearchQuery({ search: '' }))
      })
      matchCount.value = 0
      currentMatch.value = 0
      return
    }

    const query = new SearchQuery({
      search: searchText.value,
      caseSensitive: caseSensitive.value,
      wholeWord: wholeWord.value,
      regexp: useRegex.value,
    })
    v.dispatch({ effects: setSearchQuery.of(query) })
    countMatches()
  }

  // Re-sync query when options change
  watch([searchText, caseSensitive, wholeWord, useRegex], () => {
    syncQuery()
  })

  function countMatches() {
    const v = getView()
    if (!v || !searchModule || !searchText.value) {
      matchCount.value = 0
      currentMatch.value = 0
      return
    }

    const { SearchQuery } = searchModule
    const query = new SearchQuery({
      search: searchText.value,
      caseSensitive: caseSensitive.value,
      wholeWord: wholeWord.value,
      regexp: useRegex.value,
    })

    let count = 0
    const cursor = query.getCursor(v.state)
    while (!cursor.next().done) {
      count++
    }
    matchCount.value = count

    // Determine current match index
    const pos = v.state.selection.main.head
    const cursor2 = query.getCursor(v.state)
    let idx = 0
    let result = cursor2.next()
    while (!result.done) {
      idx++
      if (result.value.from >= pos) {
        currentMatch.value = idx
        return
      }
      result = cursor2.next()
    }
    currentMatch.value = count > 0 ? 1 : 0
  }

  async function findNext() {
    const v = getView()
    if (!v) return
    // Ensure query is synced first
    await syncQuery()
    const mod = await loadSearch()
    mod.findNext(v)
    countMatches()
  }

  async function findPrev() {
    const v = getView()
    if (!v) return
    await syncQuery()
    const mod = await loadSearch()
    mod.findPrevious(v)
    countMatches()
  }

  async function replaceMatch() {
    const v = getView()
    if (!v) return
    const mod = await loadSearch()
    mod.replaceNext(v)
    countMatches()
  }

  async function replaceAllMatches() {
    const v = getView()
    if (!v) return
    const mod = await loadSearch()
    mod.replaceAll(v)
    countMatches()
  }

  return {
    open, searchText, replaceText, caseSensitive, wholeWord, useRegex,
    matchCount, currentMatch,
    openFind, close, findNext, findPrev, replaceMatch, replaceAllMatches,
  }
}
