import { ref } from 'vue'

const STORAGE_KEY = 'command_history'
const MAX_HISTORY = 50

export function useCommandHistory() {
  const history = ref<string[]>([])
  const currentIndex = ref(-1)

  // Load history from localStorage
  const loadHistory = () => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        history.value = JSON.parse(stored)
      }
    } catch (error) {
      console.error('Failed to load command history:', error)
    }
  }

  // Save history to localStorage
  const saveHistory = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(history.value))
    } catch (error) {
      console.error('Failed to save command history:', error)
    }
  }

  // Add command to history
  const addCommand = (command: string) => {
    const trimmed = command.trim()
    if (!trimmed) return

    // Remove duplicate if exists
    const existingIndex = history.value.indexOf(trimmed)
    if (existingIndex !== -1) {
      history.value.splice(existingIndex, 1)
    }

    // Add to end
    history.value.push(trimmed)

    // Limit size
    if (history.value.length > MAX_HISTORY) {
      history.value.shift()
    }

    saveHistory()
    resetIndex()
  }

  // Get previous command
  const getPrevious = (): string | null => {
    if (history.value.length === 0) return null

    if (currentIndex.value === -1) {
      currentIndex.value = history.value.length - 1
    } else if (currentIndex.value > 0) {
      currentIndex.value--
    }

    return history.value[currentIndex.value] || null
  }

  // Get next command
  const getNext = (): string | null => {
    if (history.value.length === 0 || currentIndex.value === -1) return null

    if (currentIndex.value < history.value.length - 1) {
      currentIndex.value++
      return history.value[currentIndex.value]
    } else {
      currentIndex.value = -1
      return ''
    }
  }

  // Reset index
  const resetIndex = () => {
    currentIndex.value = -1
  }

  // Initialize
  loadHistory()

  return {
    history,
    addCommand,
    getPrevious,
    getNext,
    resetIndex
  }
}
