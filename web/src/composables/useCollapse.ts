import { ref } from 'vue'

export function useCollapse(initialCollapsed = true) {
  const isCollapsed = ref(initialCollapsed)

  const toggle = () => {
    isCollapsed.value = !isCollapsed.value
  }

  const expand = () => {
    isCollapsed.value = false
  }

  const collapse = () => {
    isCollapsed.value = true
  }

  return {
    isCollapsed,
    toggle,
    expand,
    collapse
  }
}
